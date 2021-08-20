package errSrv

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

const dbName = "errDB.db"

var db *gorm.DB
var sys *System

func DB() *gorm.DB {
	return db
}

func initSys() {
	sys = &System{}
	db.FirstOrCreate(sys)
	var countPost int64
	var countPubPost int64
	db.Model(&Art{}).Where("version > ?", 0).Count(&countPubPost)
	db.Model(&Art{}).Count(&countPost)
	if sys.Admin == "" {
		sys.Admin = "admin"
		sys.Pwd = "admin"
		sys.Token = uuid.New().String()
	}
	sys.TotalPosts = int(countPost)
	sys.TotalPubPosts = int(countPubPost)
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
