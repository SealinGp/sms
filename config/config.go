package config

import (
	"github.com/Akvicor/glog"
	"github.com/go-ini/ini"
)

var cfg *ini.File
var Global *Model

func Load(path string) {
	var err error
	cfg, err = ini.Load(path)
	if err != nil {
		glog.Fatal("unable to read config [%s][%s]", path, err.Error())
	}
	Global = new(Model)
	err = cfg.MapTo(Global)
	if err != nil {
		glog.Fatal("unable to parse config [%s]", err.Error())
	}
}
