package errSrv

import (
	"github.com/kataras/iris/v12"
)

var totalQA int64

func initQa(app *iris.Application) {
	syncTotal("qa", &totalQA)
	qa := app.Party("/qa")
	qa.Get("/", pageQuery("qa", &totalQA, "%q%", "%tp%"))
	qa.Post("/", qaSave)
	qa.Delete("/", qaDel)
	qa.Patch("/", qaSave)
}

func qaSave(ctx iris.Context) {
	syncTotal("qa", &totalQA)
}
func qaDel(ctx iris.Context) {
	syncTotal("qa", &totalQA)
}
