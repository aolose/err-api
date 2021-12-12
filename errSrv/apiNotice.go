package errSrv

import (
	"github.com/kataras/iris/v12"
)

var totalNotice int64

func initNoticeApi(app *iris.Application) {
	n := app.Party("n")
	auth(n.Get, "/{page}", pageQuery(Notice{}, &totalNotice, "read", "type", "%msg%"))
}

func pushNotice(n *Notice) {
	n.Date = now()
	db.Create(n)
}
