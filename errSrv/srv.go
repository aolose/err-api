package errSrv

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"time"
)

func Run(addr string) {
	go doJobs()
	go func() {
		for {
			time.Sleep(time.Second * 5)
			cleanQA()
			cleanToken()
		}
	}()
	app := iris.New()
	allowCors(app)
	app.OnAnyErrorCode(func(ctx iris.Context) {
		r := ctx.Request()
		fmt.Printf("Error %v \t %v %v\n", r.Method, r.URL, ctx.GetErr())
		ctx.Next()
	})
	app.Get("/k", func(ctx iris.Context) {
		k := randK(8)
		authKeys[k] = now() + 60
		ctx.StatusCode(200)
		_, _ = ctx.WriteString(k)
	})
	app.Post("/auth", auth(nil))
	app.Get("/ot", auth(func(ctx iris.Context) {
		sys.Token = ""
		ctx.StatusCode(200)
	}))
	initSettingApi(app)
	initArtApi(app)
	initTagsApi(app)
	initResApi(app)
	initHisApi(app)
	initQa(app)
	initBlackList(app)
	initCmApi(app)
	_ = app.Listen(addr)
}
