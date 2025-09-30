package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Akvicor/glog"
	"github.com/Akvicor/protocol"
	"github.com/Akvicor/util"
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
	serial.EnableSerial()
	initApp()

	addr := fmt.Sprintf("%s:%d", config.Global.Server.HTTPAddr, config.Global.Server.HTTPPort)
	if config.Global.Server.EnableHTTPS {
		glog.Info("Starting Hertz server: https://%s", addr)
	} else {
		glog.Info("Starting Hertz server: http://%s", addr)
	}

	err = app.StartServer()
	if err != nil {
		glog.Fatal("failed to start Hertz server [%s]", err.Error())
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

		glog.Info("stopping Hertz server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = app.StopServer(ctx)

		glog.Info("close serial")
		serial.KillSerial()

		glog.Info("close log file")
		if config.Global.Log.LogToFile {
			glog.CloseFile()
		}
		glog.Info("log file closed")

		os.Exit(0)
	}()
}
