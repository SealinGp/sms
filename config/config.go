package config

import (
	"github.com/Akvicor/glog"
	"github.com/go-ini/ini"
	"strings"
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

	loadSerialDevices()
}

func loadSerialDevices() {
	Global.SerialDevices = []SerialDevice{}

	for _, section := range cfg.Sections() {
		if strings.HasPrefix(section.Name(), "serial-device-") {
			device := SerialDevice{}
			err := section.MapTo(&device)
			if err != nil {
				glog.Error("unable to parse serial device section [%s]: %s", section.Name(), err.Error())
				continue
			}
			Global.SerialDevices = append(Global.SerialDevices, device)
		}
	}

	glog.Info("Loaded %d serial devices", len(Global.SerialDevices))
}
