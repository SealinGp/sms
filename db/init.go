package db

import (
	"github.com/Akvicor/glog"
	"github.com/Akvicor/util"
	"sms/config"
)

func CreateDatabase() {
	if util.FileStat(config.Global.Database.Path).IsExist() {
		glog.Fatal("database file exist!")
	}
	d := Connect()
	if d == nil {
		glog.Fatal("con not connect to database!")
	}
	err := db.AutoMigrate(&HistoryModel{})
	if err != nil {
		glog.Fatal(err.Error())
	}
	glog.Info("database create finished")
}
