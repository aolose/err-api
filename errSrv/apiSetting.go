package errSrv

import "github.com/kataras/iris/v12"

func initSettingApi(app *iris.Application) {
	sit := app.Party("/sys")
	sit.Post("/acc", func(ctx iris.Context) {
		b, err := ctx.GetBody()
		if err == nil {
			usr, pwd, _, _, er := upk(string(b))
			if er != nil {
				err = er
			} else {
				err = db.Model(sys).Updates(System{
					Admin: usr,
					Pwd:   pwd,
				}).Error
			}
		}
		handleErr(ctx, err)
	})
}
