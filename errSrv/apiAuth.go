package errSrv

import (
	"github.com/kataras/iris/v12"
)

type NewTick struct {
	Msg      string `json:"m"`
	Wait     int64  `json:"w"`
	Key      string `json:"k"`
	Question string `json:"q"`
}

func (cli *QAClient) passed(ctx iris.Context, key, aws, msg string) bool {
	k, q, t := cli.checkA(key, aws)
	if k != "" || t != 0 {
		ctx.StatusCode(403)
		tk := &NewTick{
			Msg:  msg,
			Wait: t,
			Key:  k,
		}
		if q != nil {
			tk.Question = q.Q
		}
		ctx.JSON(tk)
		return false
	}
	return true
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
						usr, pwd, key, aws, er := upk(s)
						var cli *QAClient
						if er == nil {
							if l := len(qaCache); sys.LoginProtect && l > 0 {
								cli = getQaCli(getIP(ctx))
								if cli.tryTimes < 1 {
									if !cli.passed(ctx, key, aws, "incorrect") {
										return
									}
								}
							}
							if sys.Admin == usr && sys.Pwd == pwd {
								pass = true
								ctx.StatusCode(200)
								_, _ = ctx.WriteString(newTk())
							} else if cli != nil {
								cli.tryTimes = cli.tryTimes - 1
								if cli.tryTimes < 1 {
									cli.passed(ctx, key, "", "auth fail")
									return
								}
							}
						} else {
							err = er
						}
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
			ck := ctx.GetCookie("session_id", iris.CookieAllowSubdomains())
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
