package errSrv

import (
	"github.com/ip2location/ip2location-go/v9"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

var totalLogs int64

var firewallRules = make([]*FirewallRule, 0)
var geoCache = make(map[string]string)

var geoDb *ip2location.DB

func syncFirewall() {
	db.Model(FirewallRule{}).
		Order("saved desc").
		Find(&firewallRules)
	for i := range firewallRules {
		firewallRules[i].TmpID = rand.Int63n(1e10)
	}
}

func getCity(ip string) string {
	if geoDb != nil {
		if c, ok := geoCache[ip]; ok {
			return c
		}
		rec, err := geoDb.Get_all(ip)
		if err == nil {
			ct := rec.City + ", " + rec.Country_short
			geoCache[ip] = ct
			return ct
		}
	}
	return ""
}

func downGeoDb() {
	if sys.GeoToken == "" {
		return
	}
	dbCode := "DB3LITEBINIPV6"
	u := "https://www.ip2location.com/download" +
		"/?token=" + sys.GeoToken + "&file=" + dbCode
	var e error
	fn := "db" + strconv.FormatInt(now(), 32) + ".bin"
	wait(func() {
		resp, err := http.Get(u)
		e = err
		if err != nil {
			return
		}
		defer resp.Body.Close()
		out, err := os.Create(fn)
		if err != nil {
			return
		}
		defer out.Close()
		_, err = io.Copy(out, resp.Body)
		return
	})

	if e == nil {
		if geoDb != nil {
			geoDb.Close()
			geoDb = nil
		}
		err := os.Remove("geo.bin")
		if err == nil {
			err = os.Rename(fn, "geo.bin")
		}
		if err != nil {
			geoDb, _ = ip2location.OpenDB(fn)
		} else {
			geoDb, _ = ip2location.OpenDB("geo.bin")
		}
	}
}

func initFirewall(app *iris.Application) {
	addJob(downGeoDb)
	syncFirewall()
	syncTotal("access_logs", &totalLogs)
	geoDb, _ = ip2location.OpenDB("geo.bin")
	log := app.Party("/log")
	ft := app.Party("/ft")
	auth(log.Get, "/{page}", pageQuery(AccessLog{}, &totalLogs, "ip%", "path%", "%ua%", "%refer%"))
	auth(log.Get, "/refresh", func(c *context.Context) {
		syncFirewall()
	})
	auth(ft.Get, "", ftGet)
	auth(ft.Post, "", ftPost)
	auth(ft.Patch, "", ftPath)
	auth(ft.Delete, "/{id}", ftDel)
}

func ftDel(c *context.Context) {
	id, err := c.Params().GetInt64("id")
	if id == 0 {
		c.StatusCode(200)
	} else {
		for i, f := range firewallRules {
			if f.TmpID == id {
				err = db.Delete(&FirewallRule{}, id).Error
				if err == nil {
					firewallRules = append(firewallRules[:i], firewallRules[i+1:]...)
				}
				break
			}
		}
	}
	handleErr(c, err)
}

func ftPost(c *context.Context) {
	f := &FirewallRule{}
	err := c.ReadJSON(f)
	f.ID = 0
	f.TmpID = rand.Int63n(1e10)
	f.Saved = now()
	if err == nil {
		firewallRules = append([]*FirewallRule{f}, firewallRules...)
		err = db.Create(f).Error
	}
	if err == nil {
		_, err = c.Writef("%d", f.ID)
	}
	handleErr(c, err)
}
func ftPath(c *context.Context) {
	f := &FirewallRule{}
	err := c.ReadJSON(f)
	if f.Saved == 0 {
		f.Saved = now()
	}
	if err == nil {
		for i, ft := range firewallRules {
			if ft.TmpID == f.TmpID {
				if f.ID == 0 {
					err = db.Create(f).Error
				} else {
					err = db.Save(f).Error
				}
				if err == nil {
					firewallRules[i] = f
				}
				break
			}
		}
	}
	if err == nil {
		_, err = c.Writef("%d", f.ID)
	}
	handleErr(c, err)
}

func ftGet(c *context.Context) {
	_, _ = c.JSON(firewallRules)
}
