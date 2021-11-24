package errSrv

import (
	"github.com/kataras/iris/v12"
	"github.com/oschwald/geoip2-golang"
	"net"
)

var bm = &BKManager{}
var totalBL int64
var totalLogs int64

var geoCache = make(map[string]string)

var geoDb *geoip2.Reader

func getCity(ip string) string {
	if geoDb != nil {
		if c, ok := geoCache[ip]; ok {
			return c
		}
		p := net.ParseIP(ip)
		rec, err := geoDb.City(p)
		if err == nil {
			return rec.Country.Names["en"] + "\t" + rec.City.Names["en"]
		}
	}
	return ""
}

func initBlackList(app *iris.Application) {
	syncTotal("access_logs", &totalLogs)
	syncTotal("black_lists", &totalBL)
	geoDb, _ = geoip2.Open("geo.mmdb")
	blackCache = &BlCAche{}
	blackCache.load()
	bk := app.Party("/bk")
	log := app.Party("/log")
	auth(log.Get, "/{page}", pageQuery(AccessLog{}, &totalLogs, "ip%", "%from%", "path%", "%ua%"))
	auth(bk.Get, "/{page}", pageQuery(BlackList{}, &totalBL, "ip", "type"))
	bk.Post("/", bkSave)
	bk.Delete("/{id}", bkDel)
}

func bkSave(ctx iris.Context) {
	syncTotal("black_lists", &totalBL)
}
func bkDel(ctx iris.Context) {
	id := ctx.Params().GetIntDefault("id", 0)
	bm.rm(id)
	syncTotal("black_lists", &totalBL)
	ctx.StatusCode(200)
	blackCache = &BlCAche{}
}
