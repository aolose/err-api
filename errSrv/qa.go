package errSrv

import (
	"errors"
	"fmt"
	"github.com/cosmos72/gomacro/fast"
	xr "github.com/cosmos72/gomacro/xreflect"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type QATicket struct {
	Q      string `gorm:"index" json:"q"`
	A      string
	expire int64
}

func eval(s string) ([]xr.Value, error) {
	wa := sync.WaitGroup{}
	ipt := ""
	for _, v := range []string{
		"regexp",
		"strconv",
		"strings",
		"time",
		"fmt",
		"errors",
	} {
		if strings.Contains(s, v+".") {
			ipt = ipt + `import "` + v + `"` + "\n"
		}
	}
	var r []xr.Value
	var err error
	wa.Add(1)
	go func() {
		defer func() {
			rec := recover()
			if rec != nil {
				r = nil
				err = errors.New(fmt.Sprintf("%v", rec))
			}
			wa.Done()
		}()
		vm := fast.New()
		r, _ = vm.Eval(ipt + s)
	}()
	wa.Wait()
	return r, err
}

func getAnswer(a string) (string, error) {
	if strings.HasPrefix(a, "func(") {
		res, er := eval(`func run(){` + a + "\nrun()")
		if er != nil {
			return "", er
		}
		return fmt.Sprintf("%v", res[0].ReflectValue()), nil
	} else {
		if strings.ContainsAny(a, "+-*/()") {
			res, er := eval(a)
			if er != nil {
				return "", er
			}
			return fmt.Sprintf("%v", res[0].ReflectValue()), nil
		} else {
			return a, nil
		}
	}
}

type QAClient struct {
	delay        int64
	tick         int
	nextTickTime int64
	ip           string
	expire       int64
	qs           map[string]*QATicket
}

const cliLife = 60 * 60 * 3 // 3h
const qaLife = 60 * 2       // 2 min

func now() int64 {
	return time.Now().Unix()
}

func RunGomacro(toeval string) reflect.Value {
	interp := fast.New()
	vals, _ := interp.Eval(toeval)
	return vals[0].ReflectValue()
}

var qaCache []Qa

var qaClients []*QAClient

func (qa *Qa) build() (*QATicket, error) {
	q := qa.Q
	a := qa.A
	s := strings.Split(qa.Params, ",")
	l := len(s) / 3
	for i := 0; i < l; i++ {
		k := s[i*3]
		mi, _ := strconv.Atoi(s[i*3+1])
		ma, _ := strconv.Atoi(s[i*3+2])
		v := strconv.FormatInt(int64(mi+rand.Intn(ma-mi)), 10)
		rs := `{` + k + `}`
		q = strings.ReplaceAll(q, rs, v)
		a = strings.ReplaceAll(a, rs, v)
	}
	as, er := getAnswer(a)
	if er == nil {
		return &QATicket{
			Q:      q,
			A:      as,
			expire: now() + qaLife,
		}, nil
	}
	return nil, errors.New(er.Error() + "\n" + a)
}

func randKey() string {
	return strconv.FormatInt(time.Now().UnixMicro()+rand.Int63n(1e8)*1e8, 36)
}

func getQaCli(ip string) *QAClient {
	l := len(qaClients)
	c := &QAClient{
		tick:   10,
		delay:  qaLife,
		ip:     ip,
		expire: now() + cliLife,
		qs:     make(map[string]*QATicket),
	}
	for i := 0; i < l; i++ {
		cli := qaClients[i]
		if cli.ip == ip {
			c = cli
			break
		}
	}
	return c
}
func randQa() (*QATicket, error) {
	l := len(qaCache)
	return qaCache[rand.Intn(l)].build()
}

func (c *QAClient) getWaitTime() int64 {
	if c.tick == 0 {
		n := now()
		if c.nextTickTime > n {
			return n - c.nextTickTime
		}
	}
	return 0
}

func (c *QAClient) getQA(k string) (string, *QATicket, int64) {
	if c.tick == 10 {
		c.nextTickTime = now() + c.delay
	} else if c.tick == 0 {
		n := now()
		if c.nextTickTime > c.expire {
			bm.add(BlackList{
				IP:   c.ip,
				Type: BkLogin,
			})
			c.expire = 0
			cleanQA()
			return "", nil, -1
		}
		if c.nextTickTime > n {
			return "", nil, n - c.nextTickTime
		}
		c.delay = c.delay * 2
		c.tick = 10
	}
	delete(c.qs, k)
	c.tick = c.tick - 1
	k = randKey()
	q, e := randQa()
	if e != nil {
		return e.Error(), nil, -1
	}
	c.qs[k] = q
	c.expire = now() + cliLife
	return k, q, 0
}

func (c *QAClient) checkA(k string, a string) (string, *QATicket, int64) {
	if a != "" {
		if _a, ok := c.qs[k]; ok {
			if _a.A == a {
				delete(c.qs, k)
				return "", nil, 0
			}
		}
	}
	return c.getQA(k)
}

var nextQaClean = int64(0)

func cleanQA() {
	n := now()
	if nextQaClean < n {
		nextQaClean = n + qaLife
		cc := make([]*QAClient, 0)
		for _, cli := range qaClients {
			if cli.expire > n {
				for k, v := range cli.qs {
					if v.expire < n {
						delete(cli.qs, k)
					}
				}
				cc = append(cc, cli)
			}
		}
		qaClients = cc
	}
}

func refreshQaCache() {
	db.Find(&qaCache)
	qaClients = make([]*QAClient, 0)
}
