package errSrv

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"strings"
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

var nextToken = int64(0)
var day = int64(60 * 60 * 24)

func cleanToken() {
	n := now()
	if nextToken < n {
		nextToken = n + day*2
		sys.Token = ""
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
	app.Post("/auth", auth(nil))
	app.Get("/ot", auth(func(ctx iris.Context) {
		sys.Token = ""
	}))
	initArtApi(app)
	initTagsApi(app)
	initResApi(app)
	initHisApi(app)
	initQa(app)
	initBlackList(app)
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

func enc(str string) string {
	return string(b64.StdEncoding.EncodeToString([]byte(str)))
}

func dec(str string) (string, error) {
	des, err := b64.StdEncoding.DecodeString(str)
	return string(des), err
}

func newTk() string {
	tk := enc(uuid.New().String())
	sys.Token = tk
	return tk
}

func upk(s string) (string, string, error) {
	s = s[1:]
	st, er := dec(s)
	if er == nil {
		v := strings.Split(st, "\u0001")
		if len(v) == 2 && len(v[0]) > 2 && len(v[1]) > 3 {
			v0, e0 := dec(v[0])
			v1, e1 := dec(v[1])
			if e0 == nil && e1 == nil {
				return v0, v1, nil
			}
		} else {
			er = errors.New("wrong text")
		}
	}
	return "", "", er
}

func authFail(ctx iris.Context) {
	ctx.StatusCode(403)
	d := "auth fail"
	if sys.Token == "" {
		d = "session expired"
	}
	ctx.WriteString(d)
}
