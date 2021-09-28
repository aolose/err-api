package errSrv

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"net/url"
	"regexp"
	"strings"
)

func getIP(ctx iris.Context) string {
	ip := ctx.RemoteAddr()
	return strings.Split(ip, ":")[0]
}

func allowCors(app *iris.Application) {
	app.UseRouter(func(ctx iris.Context) {
		if blackCache.has(getIP(ctx)) {
			ctx.StatusCode(403)
			ctx.WriteString("forbidden ip")
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
			fmt.Printf("%v \t %v\n", r.Method, r.URL)
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
						if origin == DOMAIN {
							ctx.Next()
							return
						}
					}
				} else {
					ctx.Next()
					return
				}
			}

			ua := r.Header.Get("User-agent")
			if origin == DOMAIN || ua == "node-fetch" || origin == "null" {
				ctx.Header("Access-Control-Allow-Origin", origin)
				ctx.Header("Access-Control-Allow-Credentials", "true")
				ctx.Header("Access-Control-Allow-Headers", "token")
				if ctx.Method() == iris.MethodOptions {
					ctx.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS")
					ctx.Header("Access-Control-Max-Age", "86400")
					ctx.StatusCode(204)
				} else {
					ctx.Next()
				}
			} else {
				ctx.StatusCode(403)
			}
		}
	})
}
