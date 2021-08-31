package errSrv

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"time"
)

var jobs []func()

func doJobs() {
	for {
		l := len(jobs)
		if l > 0 {
			j := jobs[0]
			jobs = jobs[1:]
			j()
		} else {
			time.Sleep(time.Millisecond * 300)
		}
	}
}

func addJob(fn func()) {
	if jobs == nil {
		jobs = make([]func(), 0)
	}
	jobs = append(jobs, fn)
}

func Run(addr string) {
	go doJobs()
	go func() {
		for {
			time.Sleep(time.Second * 5)
			n := time.Now()
			if nextSyncSys.Before(n) {
				syncSys()
				nextSyncSys = nextSyncSys.Add(time.Hour * 999)
			}
		}
	}()
	app := iris.New()
	allowCors(app)
	app.OnAnyErrorCode(func(ctx iris.Context) {
		r := ctx.Request()
		fmt.Printf("Error %v \t %v %v\n", r.Method, r.URL, ctx.GetErr())
		ctx.Next()
	})

	initArtApi(app)
	initTagsApi(app)
	initResApi(app)
	initHisApi(app)
	_ = app.Listen(addr)
}

func handleErr(ctx iris.Context, err error) {
	if err == nil {
		ctx.StatusCode(200)
	} else {
		_ = fmt.Errorf("%v", err)
		ctx.StatusCode(500)
		_, _ = ctx.JSON(err)
	}
}
func auth(next func(ctx iris.Context)) func(ctx iris.Context) {
	return func(ctx iris.Context) {
		// -- todo check
		next(ctx)
	}
}
