package errSrv

import (
	"math/rand"
	"strconv"
	"time"
)

type QATicket struct {
	QA
	expire int64
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

var qaCache []QA

var qaClients []*QAClient

func (qa *QA) build() QA {

	return QA{}
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
func randQa() *QATicket {
	l := len(qaCache)
	q := qaCache[rand.Intn(l)].build()
	return &QATicket{q, now() + qaLife}
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
		if c.nextTickTime > n {
			return "", nil, n - c.nextTickTime
		}
		c.delay = c.delay * 2
		c.tick = 10
	}
	delete(c.qs, k)
	c.tick = c.tick - 1
	k = randKey()
	q := randQa()
	c.qs[k] = q
	c.expire = now() + cliLife
	return k, q, 0
}

func (c *QAClient) checkA(k string, a string) (string, *QATicket, int64) {
	if a != "" {
		if _a, ok := c.qs[k]; ok && _a.A == a {
			delete(c.qs, k)
			c.tick = 10
			c.delay = qaLife
			return "", nil, 0
		}
	}
	return c.getQA(k)
}

var nextQaClean = int64(0)

func cleanQuestion() {
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
