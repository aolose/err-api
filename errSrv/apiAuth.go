package errSrv

import (
	"github.com/kataras/iris/v12"
)

func auth(next func(ctx iris.Context)) func(ctx iris.Context) {
	return func(ctx iris.Context) {
		pass := false
		if next == nil {
			b, err := ctx.GetBody()
			if err == nil {
				s := string(b)
				if len(s) > 10 {
					if s[0] == '_' {
						usr, pwd, er := upk(s)
						if er == nil {
							if sys.Admin == usr && sys.Pwd == pwd {
								pass = true
								ctx.StatusCode(200)
								_, _ = ctx.WriteString(newTk())
							}
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
