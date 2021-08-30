package errSrv

import "github.com/kataras/iris/v12"

func initHisApi(app *iris.Application) {
	his := app.Party("/his")
	his.Get("/{id}/{ver}", func(ctx iris.Context) {})
	his.Delete("/{id}/{ver}", func(ctx iris.Context) {})
}
