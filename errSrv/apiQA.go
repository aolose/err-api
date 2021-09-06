package errSrv

import (
	"errors"
	"github.com/kataras/iris/v12"
)

var totalQA int64

func initQa(app *iris.Application) {
	syncTotal("qas", &totalQA)
	qa := app.Party("/qa")
	qa.Get("/{page}", pageQuery(Qa{}, &totalQA, "%q%", "%a%"))
	qa.Post("/", auth(qaSave))
	qa.Post("/test", auth(testQa))
	qa.Delete("/{id}", auth(qaDel))
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
	p := &Qa{}
	err := ctx.ReadJSON(p)
	if err == nil {
		tk, er := p.build()
		if er == nil {
			p.Saved = now()
			if p.ID == 0 {
				er = db.Create(p).Error
			} else {
				er = db.Save(p).Error
			}
			if er == nil {
				r := p.RQA
				r.Q = tk.Q
				r.A = tk.A
				ctx.JSON(r)
			}
		}
		err = er
	}
	if err != nil {
		handleErr(ctx, err)
	}
	syncTotal("qas", &totalQA)
}
func qaDel(ctx iris.Context) {
	id := ctx.Params().GetUintDefault("id", 0)
	err := errors.New("error id")
	if id > 0 {
		err = db.Delete(&Qa{}, id).Error
	}
	if err == nil {
		syncTotal("qas", &totalQA)
	}
	handleErr(ctx, err)
}
