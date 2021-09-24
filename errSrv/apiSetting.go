package errSrv

import "github.com/kataras/iris/v12"

func initSettingApi(app *iris.Application) {
	sit := app.Party("/sys")
	sit.Get("", auth(sysInfo))
	sit.Post("/acc", auth(setAcc))
}

func setAcc(ctx iris.Context) {
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
func sysInfo(ctx iris.Context) {
	ctx.JSON(sys)
}
