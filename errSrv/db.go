package errSrv

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"regexp"
	"strings"
)

const dbName = "errDB.db"

var db *gorm.DB
var sys *System

func initSys() {
	sys = &System{}
	db.FirstOrCreate(sys)
	countPos()
	countRes()
	syncSys()
	var tas []TagArt
	db.Find(&tas)
loop:
	for _, t := range tas {
		ok, v := hasTag(t.AID, t.Name)
		if ok {
			continue loop
		}
		v = append(v, t.AID)
		tagsCache[t.Name] = v
	}
}

func slugCount(str string, id uint) int64 {
	var c int64
	cd := "id !=? and slug like ?"
	db.Model(&Art{}).Where(cd, id, str).Count(&c)
	if c > 0 {
		var b int64
		db.Model(&Art{}).Where(cd, id, str+"_").Count(&b)
		c += b
		if c > 9 {
			db.Model(&Art{}).Where(cd, id, str+"__").Count(&b)
			c += b
		}
	}
	return c
}

func nextTokenCleanDelay() {
	nextToken = now() + day*2
}

func countPos() {
	var countPost int64
	var countPubPost int64
	db.Model(&Art{}).Where("updated > ?", 0).Count(&countPubPost)
	db.Model(&Art{}).Count(&countPost)
	sys.TotalPosts = int(countPost)
	sys.TotalPubPosts = int(countPubPost)
}

func countRes() {
	var countRes int64
	db.Model(&Res{}).Count(&countRes)
	sys.TotalRes = int(countRes)
}

func syncSys() {
	if sys.CnLen == 0 {
		sys.CnLen = 64
	}
	if sys.CmLen == 0 {
		sys.CmLen = 512
	}
	if errCfg.User != "" || sys.Admin == "" {
		if errCfg.User == "" {
			errCfg.User = "admin"
			errCfg.Pass = "admin"
		}
		//sys.DisCm =1
		sys.Admin = errCfg.User
		sys.Pwd = md5Enc(errCfg.Pass, "")
		errCfg.User = ""
		errCfg.Pass = ""
		errCfg.Update()
	}
	db.Updates(sys)
}

func Connect() {
	var err error
	db, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal("Open: ", err)
	}
	fmt.Printf("Connected to %q", dbName)
	dbInit()
	initialCfg()
	initSys()
}

type ListResult struct {
	Total int         `json:"total"`
	List  interface{} `json:"ls"`
	Cur   int         `json:"cur"`
}

func pageQuery(table interface{}, total *int64, field ...string) func(ctx iris.Context) {
	return func(ctx iris.Context) {
		page := ctx.Params().GetIntDefault("page", 1)
		count := ctx.URLParamIntDefault("c", 20)
		q := make([]string, 0)
		qs := make([]interface{}, 0)
		noQ := true
		for _, f := range field {
			re := regexp.MustCompile("^([_%]*)([a-z_]+?)([_%]*)$")
			s := re.FindStringSubmatch(f)
			if len(s) == 4 {
				s = s[1:]
				ff := s[1]
				v := ctx.URLParam(ff)
				if v != "" {
					noQ = false
					qq := " = ?"
					if len(s[0]) > 0 || len(s[2]) > 0 {
						qq = " like ?"
						v = s[0] + v + s[2]
					}
					q = append(q, ff+qq)
					qs = append(qs, v)
				}
			}
		}
		var c int64
		tx := db.Model(table)
		tx1 := db.Model(table).Offset((page - 1) * count).Limit(count)
		if noQ && total != nil {
			c = *total
		} else {
			if len(q) > 0 {
				qq := strings.Join(q, " and ")
				tx = tx.Where(qq, qs...)
				tx1 = tx1.Where(qq, qs...)
			}
			tx.Count(&c)
		}
		tx1 = tx1.Order("saved desc")
		re := &ListResult{
			Cur:   page,
			Total: (int(c) + count - 1) / count,
		}
		var err error
		switch table.(type) {
		case Notice:
			res := make([]Notice, 0)
			err = tx1.Find(&res).Error
			re.List = res
		case AccessLog:
			res := make([]AccessLog, 0)
			err = tx1.Find(&res).Error
			for n, r := range res {
				if r.Saved == 0 {
					res[n].Saved = r.Date
				}
			}
			re.List = res
		case Comment:
			res := make([]Comment, 0)
			art := make([]Art, 0)
			err = tx1.Find(&res).Error
			if err == nil {
				ids := make([]uint, len(res))
				for n, i := range res {
					ids[n] = i.ArtID
				}
				err = db.Find(&art, ids).Error
			}
			if err == nil {
				for n, r := range res {
					res[n].From = r.IP
					for _, a := range art {
						if r.ArtID == a.ID {
							res[n].Inf = ArtInf{
								Title: a.Title,
								Slug:  a.Slug,
								Date:  a.Created,
							}
							if a.OverrideCreate > 0 {
								res[n].Inf.Date = a.OverrideCreate
							}
						}
					}
				}
			}
			re.List = res
		case BlackList:
			res := make([]BlackList, 0)
			err = tx1.Find(&res).Error
			re.List = res
		}
		if err != nil {
			handleErr(ctx, err)
		} else {
			ctx.StatusCode(200)
			_, _ = ctx.JSON(re)
		}
	}
}

func syncTotal(table string, c *int64) {
	addJob(func() {
		db.Table(table).Count(c)
	})
}
