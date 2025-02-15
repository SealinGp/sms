package main

import (
	"flag"
	"fmt"
	"github.com/Akvicor/glog"
	"github.com/Akvicor/protocol"
	"github.com/Akvicor/util"
	"net/http"
	"os"
	"os/signal"
	"sms/app"
	"sms/config"
	"sms/db"
	"sms/serial"
	"syscall"
	"time"
)

func main() {
	var err error

	isInit := flag.Bool("i", false, "init database")
	c := flag.String("c", "config.ini", "path to config file")
	flag.Parse()

	if util.FileStat(*c).NotFile() {
		glog.Fatal("missing config [%s]!", *c)
	}

	config.Load(*c)
	setGlog()

	if *isInit {
		initDatabase()
	}
	if util.FileStat(config.Global.Database.Path).NotFile() {
		glog.Fatal("missing database [%s]!", config.Global.Database.Path)
	}

	EnableShutDownListener()
	serial.EnableSerialCN()
	serial.EnableSerialUS()
	initApp()

	addr := fmt.Sprintf("%s:%d", config.Global.Server.HTTPAddr, config.Global.Server.HTTPPort)
	if config.Global.Server.EnableHTTPS {
		glog.Info("ListenAndServe: https://%s", addr)
		err = http.ListenAndServeTLS(addr, config.Global.Server.SSLCert, config.Global.Server.SSLKey, app.Global)
	} else {
		glog.Info("ListenAndServe: http://%s", addr)
		err = http.ListenAndServe(addr, app.Global)
	}
	if err != nil {
		glog.Fatal("failed to listen and serve [%s]", err.Error())
	}

}

func initApp() {
	app.Generate()
}

func initDatabase() {
	db.CreateDatabase()

	os.Exit(0)
}

func setGlog() {
	if config.Global.Log.LogToFile {
		err := glog.SetLogFile(config.Global.Log.FilePath)
		if err != nil {
			glog.Fatal("failed to set log file [%s]", err.Error())
		}
	}
	if config.Global.Prod {
		glog.SetFlag(glog.FlagStd)
		protocol.SetLogProd(true)
	} else {
		glog.SetFlag(glog.FlagStd | glog.FlagShortFile | glog.FlagFunc | glog.FlagSuffix)
		protocol.SetLogProd(false)
	}
}

func EnableShutDownListener() {
	go func() {
		down := make(chan os.Signal, 1)
		signal.Notify(down, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-down
		go func() {
			ticker := time.NewTicker(3 * time.Second)
			<-ticker.C
			glog.Fatal("Ticker Finished")
		}()

		glog.Info("close serial")
		serial.KillCN()
		serial.KillUS()

		glog.Info("close log file")
		if config.Global.Log.LogToFile {
			glog.CloseFile()
		}
		glog.Info("log file closed")

		os.Exit(0)
	}()
}
