package errSrv

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

const dbName = "errDB.db"

var db *gorm.DB

func DB() *gorm.DB {
	return db
}

func Connect() {
	var err error
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             0,           // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,        // 禁用彩色打印
		},
	)
	db, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{Logger: newLogger})
	if err != nil {
		log.Fatal("Open: ", err)
	}
	fmt.Printf("Connected to %q", dbName)
	dbInit()
}
