package errSrv

import (
	"fmt"
	"github.com/kataras/iris/v12"
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
			ctx.WriteString("Access Forbidden!")
		} else {
			r := ctx.Request()
			fmt.Printf("%v \t %v\n", r.Method, r.URL)
			origin := r.Header.Get("origin")
			if r.Method == "GET" &&
				strings.HasPrefix(r.URL.Path, "/r/") {

				if origin == "" {
					origin = r.Referer()
					origin = origin[0 : len(origin)-1]
				}

				if !strings.HasSuffix(r.URL.Path, ".png") {
					id := r.URL.Path[3:]
					re := &Res{ID: id}
					err := db.First(re).Error
					if err == nil {
						m1 := regexp.MustCompile(`^(.*?)\.\w+$`)
						//Content-Disposition:
						ctx.Header(
							"Content-Disposition",
							" attachment; filename=\""+
								m1.ReplaceAllString(re.Name, "$1")+"."+re.Ext+"\"",
						)
					}
				}
			}
			ua := r.Header.Get("User-agent")
			if origin == "https://www.err.name" || ua == "node-fetch" || origin == "null" || origin == "http://localhost:3000" {
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
