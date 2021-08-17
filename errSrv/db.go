package errSrv

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
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
	db.Model(&Post{}).Where("status = ?", 1).Count(&countPubPost)
	db.Model(&Post{}).Count(&countPost)
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
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             10 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	db, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{Logger: newLogger})
	if err != nil {
		log.Fatal("Open: ", err)
	}
	fmt.Printf("Connected to %q", dbName)
	dbInit()
	initSys()
}
