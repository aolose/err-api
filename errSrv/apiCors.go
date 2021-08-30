package errSrv

import (
	"fmt"
	"github.com/kataras/iris/v12"
)

func allowCors(app *iris.Application) {
	app.UseRouter(func(ctx iris.Context) {
		r := ctx.Request()
		fmt.Printf("%v \t %v\n", r.Method, r.URL)
		ctx.Header("Access-Control-Allow-Origin", r.Header.Get("origin"))
		ctx.Header("Access-Control-Allow-Credentials", "true")
		if ctx.Method() == iris.MethodOptions {
			ctx.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS")
			ctx.Header("Access-Control-Max-Age", "86400")
			ctx.StatusCode(204)
		} else {
			ctx.Next()
		}
	})
}
