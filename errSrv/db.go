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
	refreshQaCache()
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
	if sys.Admin == "" {
		sys.Admin = "admin"
		sys.Pwd = md5Enc("admin")
	}
	db.Save(sys)
}

func Connect() {
	var err error
	db, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal("Open: ", err)
	}
	fmt.Printf("Connected to %q", dbName)
	dbInit()
	initSys()
}

type ListResult struct {
	Total int         `json:"total"`
	List  interface{} `json:"ls"`
	Cur   int         `json:"cur"`
}

func pageQuery(table interface{}, total *int64, field ...string) func(ctx iris.Context) {
	return auth(func(ctx iris.Context) {
		page := ctx.Params().GetIntDefault("page", 1)
		count := ctx.URLParamIntDefault("count", 20)
		q := make([]string, 0)
		qs := make([]interface{}, 0)
		noQ := true
		for _, f := range field {
			re, _ := regexp.Compile("/([_%]*)([a-b_]+?)([_%]*)/")
			s := re.FindStringSubmatch(f)
			if len(s) == 3 {
				ff := s[1]
				v := ctx.URLParam(ff)
				if v != "" {
					noQ = false
					qq := " = ?"
					if len(s[0]) > 0 || len(s[2]) > 0 {
						qq = " like ?"
						v = s[0] + v + s[1]
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
		tx = tx.Order("saved desc")
		re := &ListResult{
			Cur:   page,
			Total: (int(c) + count - 1) / count,
		}
		var err error
		switch table.(type) {
		case Qa:
			res := make([]Qa, 0)
			err = tx.Find(&res).Error
			re.List = res
		case BlackList:
			res := make([]BlackList, 0)
			err = tx.Find(&res).Error
			re.List = res
		}
		if err != nil {
			handleErr(ctx, err)
		} else {
			ctx.StatusCode(200)
			_, _ = ctx.JSON(re)
		}
	})
}

func syncTotal(table string, c *int64) {
	addJob(func() {
		db.Table(table).Count(c)
	})
}
