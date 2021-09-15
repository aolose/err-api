package errSrv

import (
	"errors"
	"fmt"
	"github.com/cosmos72/gomacro/fast"
	xr "github.com/cosmos72/gomacro/xreflect"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	cliLife  = 60 * 60 * 3 // 3h
	qaLife   = 60 * 2      // 2 min
	tryTimes = 1
	ticks    = 1
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
		if strings.ContainsAny(a, `"+-*/()`) {
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
	tryTimes     int64
	delay        int64
	tick         int
	nextTickTime int64
	ip           string
	expire       int64
	qs           map[string]*QATicket
}

func now() int64 {
	return time.Now().Unix()
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
		vv := int64(mi)
		if mi != ma {
			c := int64(1)
			if mi > ma {
				c = -1
			}
			vv = c * (vv + rand.Int63n(int64(ma-mi)*c))
		}
		v := strconv.FormatInt(vv, 10)
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
		tryTimes: tryTimes,
		tick:     ticks,
		delay:    qaLife,
		ip:       ip,
		expire:   now() + cliLife,
		qs:       make(map[string]*QATicket),
	}
	for i := 0; i < l; i++ {
		cli := qaClients[i]
		if cli.ip == ip {
			return cli
		}
	}
	qaClients = append(qaClients, c)
	return c
}
func randQa() (*QATicket, error) {
	l := len(qaCache)
	return qaCache[rand.Intn(l)].build()
}

func (cli *QAClient) getWaitTime() int64 {
	if cli.tick == 0 {
		n := now()
		if cli.nextTickTime > n {
			return n - cli.nextTickTime
		}
	}
	return 0
}

func (cli *QAClient) getQA(k string) (string, *QATicket, int64) {
	if cli.tick == 0 {
		n := now()
		if cli.nextTickTime > cli.expire {
			bm.add(BlackList{
				IP:   cli.ip,
				Type: BkLogin,
			})
			cli.expire = 0
			cleanQA()
			return "", nil, -1
		}
		if cli.nextTickTime > n {
			return "", nil, cli.nextTickTime - n
		}
		cli.delay = cli.delay * 2
		cli.tick = ticks
	} else {
		if cli.tick == ticks {
			cli.nextTickTime = now() + cli.delay
		}
		cli.tick = cli.tick - 1
	}
	delete(cli.qs, k)
	k = randKey()
	q, e := randQa()
	if e != nil {
		return e.Error(), nil, -1
	}
	cli.qs[k] = q
	cli.expire = now() + cliLife
	return k, q, 0
}

func (cli *QAClient) checkA(k string, a string) (string, *QATicket, int64) {
	if a != "" {
		if _a, ok := cli.qs[k]; ok {
			if _a.A == a {
				delete(cli.qs, k)
				return "", nil, 0
			}
		}
	}
	return cli.getQA(k)
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
