package errSrv

import "github.com/kataras/iris/v12"
import "github.com/kataras/iris/v12/context"

func initHisApi(app *iris.Application) {
	his := app.Party("/his")
	his.Get("/{id}/{ver}", func(ctx *context.Context) {})
	his.Delete("/{id}/{ver}", func(ctx *context.Context) {})
}
