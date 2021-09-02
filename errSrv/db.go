package errSrv

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"time"
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

var nextClean time.Time

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

func nextCleanDelay(v time.Duration) {
	n := time.Now()
	t := n.Add(v)
	if nextClean.Before(n) || (nextClean.After(t) && n.Before(t)) {
		nextClean = t
	}
}

func countPos() {
	var countPost int64
	var countPubPost int64
	db.Model(&Art{}).Where("version > ?", 0).Count(&countPubPost)
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
		sys.Pwd = "admin"
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
