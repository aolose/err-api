package errSrv

import "github.com/kataras/iris/v12"
import "github.com/kataras/iris/v12/context"

func initSettingApi(app *iris.Application) {
	sit := app.Party("/sys")
	auth(sit.Get, "", sysInfo)
	auth(sit.Post, "/acc", setAcc)
}

func setAcc(ctx *context.Context) {
	b, err := ctx.GetBody()
	if err == nil {
		usr, pwd, er := upk(string(b))
		if er != nil {
			err = er
		} else {
			err = db.Model(sys).Updates(System{
				Admin: usr,
				Pwd:   md5Enc(pwd, ""),
			}).Error
		}
	}
	handleErr(ctx, err)
}
func sysInfo(ctx *context.Context) {
	ctx.JSON(sys)
}
