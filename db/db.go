package db

import (
	"github.com/Akvicor/glog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sms/config"
	"sync"
)

var (
	db        *gorm.DB
	dbLock    = sync.RWMutex{}
	connected = false
)

func Connect() *gorm.DB {
	if connected {
		return db
	}
	dbLock.Lock()
	defer dbLock.Unlock()
	if connected {
		return db
	}
	var err error
	db, err = gorm.Open(sqlite.Open(config.Global.Database.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		glog.Warning("failed to connect database [%s]", err.Error())
		return nil
	}
	connected = true
	return db
}
