package errSrv

import (
	"github.com/kataras/iris/v12"
	"strconv"
	"time"
)

const (
	tkLife   = 30
	cliLife  = 3600 * 24 // 24h
	tryTimes = 3
	ticks    = 3
)

var cliMap = make(map[string]*Client)

func getCli(ip string) *Client {
	cli, ok := cliMap[ip]
	if !ok {
		cli = &Client{
			ip:           ip,
			key:          randK(8),
			tryTimes:     tryTimes,
			ticks:        ticks,
			delay:        0,
			nextTickTime: 0,
		}
		cliMap[ip] = cli
	}
	cli.expire = now() + cliLife
	return cli
}

type Client struct {
	ip           string
	key          string
	tryTimes     int64
	ticks        int
	delay        int64
	nextTickTime int64
	expire       int64
}

func str(ctx iris.Context, s string) {
	_, _ = ctx.WriteString(s)
}

func login(ctx iris.Context, s string) {
	ip := getIP(ctx)
	c := getCli(ip)
	n := now()
	if blackCache.has(ip) {
		delete(cliMap, ip)
		ctx.StatusCode(403)
		str(ctx, "forbidden ip")
		return
	}
	if c.nextTickTime > 0 && c.nextTickTime < n && c.tryTimes == 0 {
		c.nextTickTime = 0
		c.tryTimes = tryTimes
		c.ticks = c.ticks - 1
	}

	if c.nextTickTime > n {
		ctx.StatusCode(403)
		str(ctx, "w:"+strconv.FormatInt(c.nextTickTime-n, 10))
	} else {
		usr, pwd, err := upk(s)
		if err == nil {
			if sys.Admin == usr {
				if md5Enc(sys.Pwd, c.key) == pwd {
					delete(cliMap, c.ip)
					ctx.StatusCode(200)
					_, _ = ctx.WriteString(newTk())
					return
				}
			}
		}
		if c.tryTimes == 0 {
			if c.ticks == 0 {
				ctx.StatusCode(403)
				bm.add(&BlackList{
					IP:   ip,
					Type: BkLogin,
				})
				delete(cliMap, ip)
				str(ctx, "forbidden ip")
			} else {
				c.delay = c.delay + tkLife
				c.nextTickTime = n + c.delay
				ctx.StatusCode(403)
				ctx.WriteString("w:" + strconv.FormatInt(c.delay, 10))
			}
		} else {
			c.tryTimes = c.tryTimes - 1
			ctx.StatusCode(403)
			_, _ = ctx.WriteString("wrong name or password")
		}
	}
}

func now() int64 {
	return time.Now().Unix()
}

type NewTick struct {
	Msg      string `json:"m"`
	Wait     int64  `json:"w"`
	Key      string `json:"k"`
	Question string `json:"q"`
}

func cleanCli() {
	for k, t := range cliMap {
		if t.expire < now() {
			delete(cliMap, k)
		}
	}
}

func auth(next func(ctx iris.Context)) func(ctx iris.Context) {
	return func(ctx iris.Context) {
		pass := false
		if next == nil {
			b, err := ctx.GetBody()
			if err == nil {
				s := string(b)
				if len(s) > 10 {
					if s[0] == '_' {
						login(ctx, s)
						return
					} else {
						if len(sys.Token) > 20 && sys.Token == s {
							pass = true
							ctx.StatusCode(200)
							ctx.SetCookie(&iris.Cookie{
								Name:     "session_id",
								Value:    sys.Token,
								HttpOnly: true,
								MaxAge:   60 * 60 * 24 * 7,
								SameSite: iris.SameSiteLaxMode,
								Path:     "/",
							}, iris.CookieAllowSubdomains())
						}
					}
				}
			} else {
				handleErr(ctx, err)
			}
		} else {
			ck := ctx.GetCookie("session_id")
			if ck == "" {
				ck = ctx.GetHeader("token")
			}
			if len(ck) > 10 && ck == sys.Token {
				pass = true
			}
		}
		if !pass {
			authFail(ctx)
		} else {
			nextTokenCleanDelay()
			if next != nil {
				next(ctx)
			}
		}
	}
}
