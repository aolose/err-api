package errSrv

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

type Cfg struct {
	Bind   string
	Domain string
	User   string
	Pass   string
}

func (c *Cfg) Update() {
	errCfg.User = ""
	errCfg.Pass = ""
	d, _ := yaml.Marshal(&errCfg)
	_ = ioutil.WriteFile("cfg.yaml", d, os.ModePerm)
}

var errCfg Cfg

//const DOMAIN="http://localhost:3000"
func Run() {
	go doJobs()
	go func() {
		for {
			time.Sleep(time.Second * 5)
			cleanToken()
			cleanCli()
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
		k := getCli(getIP(ctx)).key
		ctx.StatusCode(200)
		_, _ = ctx.WriteString(k)
	})
	app.Post("/auth", auth(nil))
	app.Get("/ot", auth(func(ctx iris.Context) {
		sys.Token = ""
		ctx.StatusCode(200)
		setSession(ctx, "")
	}))
	initSettingApi(app)
	initArtApi(app)
	initTagsApi(app)
	initResApi(app)
	initHisApi(app)
	initBlackList(app)
	initCmApi(app)
	_ = app.Run(iris.Addr(errCfg.Bind), iris.WithConfiguration(iris.Configuration{
		RemoteAddrHeaders: []string{"X-Real-Ip"},
	}))
}
