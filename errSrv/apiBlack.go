package errSrv

import (
	"github.com/ip2location/ip2location-go/v9"
	"github.com/kataras/iris/v12"
)

var bm = &BKManager{}
var totalBL int64
var totalLogs int64

var geoCache = make(map[string]string)

var geoDb *ip2location.DB

func getCity(ip string) string {
	if geoDb != nil {
		if c, ok := geoCache[ip]; ok {
			return c
		}
		rec, err := geoDb.Get_all(ip)
		if err == nil {
			return rec.Country_long + "\t" + rec.City
		}
	}
	return ""
}

func initBlackList(app *iris.Application) {
	syncTotal("access_logs", &totalLogs)
	syncTotal("black_lists", &totalBL)
	geoDb, _ = ip2location.OpenDB("geo.bin")
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
