package errSrv

import (
	"github.com/kataras/iris/v12"
)

var totalQA int64

func initQa(app *iris.Application) {
	syncTotal("qas", &totalQA)
	qa := app.Party("/qa")
	qa.Get("/{page}", pageQuery("qas", &totalQA, "%q%", "%a%"))
	qa.Post("/test", auth(testQa))
	qa.Post("/", auth(qaSave))
	qa.Delete("/", auth(qaDel))
	qa.Patch("/", auth(qaSave))
}

func testQa(ctx iris.Context) {
	p := &Qa{}
	err := ctx.ReadJSON(p)
	if err == nil {
		tk, er := p.build()
		if er == nil {
			ctx.JSON(map[string]string{
				"q": tk.Q,
				"a": tk.A,
			})
		} else {
			handleErr(ctx, er)
		}
	} else {
		handleErr(ctx, err)
	}

}

func qaSave(ctx iris.Context) {
	syncTotal("qas", &totalQA)
}
func qaDel(ctx iris.Context) {
	syncTotal("qas", &totalQA)
}
