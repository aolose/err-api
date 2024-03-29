package errSrv

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"log"
	"net"
	"net/url"
	"regexp"
	"strings"
)

func gIP(str string) string {
	addr := strings.TrimSpace(str)
	if addr != "" {
		if k, _, err := net.SplitHostPort(addr); err == nil {
			return k
		}
	}
	return addr
}

func getIP(ctx iris.Context) string {
	return gIP(ctx.RemoteAddr())
}

func logAccess(c *context.Context) {
	c.Next()
	p := c.GetCurrentRoute()
	if p == nil {
		log.Printf("Access no router matched: %s \n", c.Path())
		return
	}
	// skip auth path
	for _, a := range authPaths {
		if a == p.Tmpl().Src {
			return
		}
	}
	ip := getIP(c)
	if c.Method() != "OPTIONS" && ip != "127.0.0.1" {
		if firewall(c, SkipLog) {
			return
		}
		ag := c.GetHeader("User-Agent")
		if ag == "node-fetch" {
			ag = c.GetHeader("node-user-agent")
		}
		db.Create(&AccessLog{
			Ip:    ip,
			Refer: c.GetReferrer().String(),
			Saved: now(),
			Path:  c.Path(),
			UA:    c.GetHeader("User-Agent"),
		})
		totalLogs++
	}
}

func allowCors(app *iris.Application) {
	app.UseRouter(func(ctx *context.Context) {
		if firewall(ctx, BlockAccess) {
			ctx.StatusCode(403)
			ctx.WriteString("access forbidden")
		} else {
			r := ctx.Request()
			origin := r.Header.Get("origin")
			if origin == "" {
				rf := r.Referer()
				if len(rf) > 10 {
					u, _ := url.Parse(rf)
					origin = u.Scheme + "://" + u.Host
				}
			}
			log.Printf("%v \t %v\n", r.Method, r.URL)
			if r.Method == "GET" &&
				strings.HasPrefix(r.URL.Path, "/r/") {
				if !strings.HasSuffix(r.URL.Path, ".webp") {
					id := r.URL.Path[3:]
					re := &Res{ID: id}
					err := db.First(re).Error
					if err == nil {
						if strings.Contains(re.Type, "image") {
							ctx.Header("Content-Type", "image/webp")
							ctx.Header("Accept-Ranges", "bytes")
						} else {
							m1 := regexp.MustCompile(`^(.*?)\.\w+$`)
							ctx.Header(
								"Content-Disposition",
								" attachment; filename=\""+
									m1.ReplaceAllString(re.Name, "$1")+"."+re.Ext+"\"",
							)
						}
						if origin == errCfg.Domain {
							ctx.Next()
							return
						}
					}
				} else {
					ctx.Next()
					return
				}
			}

			if origin == errCfg.Domain ||
				errCfg.Host == gIP(r.RemoteAddr) ||
				(sys.Token != "" && sys.Token == ctx.GetHeader("token")) {
				ctx.Header("Access-Control-Allow-Origin", ctx.GetHeader("origin"))
				ctx.Header("Access-Control-Allow-Credentials", "true")
				ctx.Header("Access-Control-Allow-Headers", "token, cache-control")
				ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				ctx.Header("Access-Control-Max-Age", "86400")
				if ctx.Method() == iris.MethodOptions {
					ctx.StatusCode(204)
				} else {
					ctx.Next()
				}
			} else {
				log.Printf("403: %s %s", errCfg.Host, gIP(r.RemoteAddr))
				ctx.StatusCode(403)
			}
		}
	})
}
