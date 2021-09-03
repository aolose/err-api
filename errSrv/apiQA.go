package errSrv

import (
	"github.com/kataras/iris/v12"
)

var totalQA int64

func initQa(app *iris.Application) {
	syncTotal("qas", &totalQA)
	qa := app.Party("/qa")
	qa.Get("/", pageQuery("qas", &totalQA, "%q%", "%a%"))
	qa.Post("/test", auth(testQa))
	qa.Post("/", auth(qaSave))
	qa.Delete("/", auth(qaDel))
	qa.Patch("/", auth(qaSave))
}

type Tqa struct {
	Qa
	test string
}

func testQa(ctx iris.Context) {
	p := &Tqa{}
	var pass bool
	err := ctx.ReadJSON(p)
	if err == nil {
		pass, err = p.Qa.build().check(p.test)
	}
	ctx.StatusCode(200)
	if !pass {
		ctx.WriteString(err.Error())
	}
}

func qaSave(ctx iris.Context) {
	syncTotal("qas", &totalQA)
}
func qaDel(ctx iris.Context) {
	syncTotal("qas", &totalQA)
}
