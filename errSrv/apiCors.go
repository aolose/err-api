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
			if r.Method == "GET" &&
				strings.HasPrefix(r.URL.Path, "/r/") &&
				!strings.HasSuffix(r.URL.Path, ".png") {
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
			ctx.Header("Access-Control-Allow-Origin", r.Header.Get("origin"))
			ctx.Header("Access-Control-Allow-Credentials", "true")
			if ctx.Method() == iris.MethodOptions {
				ctx.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS")
				ctx.Header("Access-Control-Max-Age", "86400")
				ctx.StatusCode(204)
			} else {
				ctx.Next()
			}
		}
	})
}
