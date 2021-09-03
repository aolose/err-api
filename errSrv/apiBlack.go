package errSrv

import (
	"github.com/kataras/iris/v12"
)

var bm = &BKManager{}
var totalBL int64

func initBlackList(app *iris.Application) {
	syncTotal("black_lists", &totalBL)
	bk := app.Party("/bk")
	bk.Get("/", pageQuery("black_lists", &totalBL, "ip", "tp"))
	bk.Post("/", bkSave)
	bk.Delete("/{id}", bkDel)
}

func bkSave(ctx iris.Context) {
	syncTotal("black_lists", &totalBL)
}
func bkDel(ctx iris.Context) {
	id := ctx.Params().GetIntDefault("id", 0)
	bm.rm(id)
	syncTotal("black_lists", &totalBL)
	ctx.StatusCode(200)
}
