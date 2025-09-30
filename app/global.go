package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	smsConfig "sms/config"
)

var Global *server.Hertz
var SessionStore *sessions.CookieStore

func Generate() {
	// Initialize session store
	SessionStore = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
	SessionStore.Options.Domain = smsConfig.Global.Session.Domain
	SessionStore.Options.Path = smsConfig.Global.Session.Path
	SessionStore.Options.MaxAge = smsConfig.Global.Session.MaxAge

	// Create Hertz server with configuration
	opts := []config.Option{
		server.WithHostPorts(fmt.Sprintf("%s:%d", smsConfig.Global.Server.HTTPAddr, smsConfig.Global.Server.HTTPPort)),
	}

	if smsConfig.Global.Server.EnableHTTPS {
		cert, err := tls.LoadX509KeyPair(smsConfig.Global.Server.SSLCert, smsConfig.Global.Server.SSLKey)
		if err == nil {
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			opts = append(opts, server.WithTLS(tlsConfig))
		}
	}

	Global = server.Default(opts...)

	// Register routes
	registerRoutes()
}

func registerRoutes() {
	// Static routes
	Global.GET("/favicon.ico", staticFavicon)

	// Main routes
	Global.GET("/", index)
	Global.GET("/login", loginGet)
	Global.POST("/login", loginPost)
	Global.GET("/random_key", randomKey)
	Global.POST("/send_sms", sendSMSCN)
	Global.GET("/history", historyCN)
	Global.POST("/send_sms_cn", sendSMSCN)
	Global.GET("/history_cn", historyCN)
	Global.POST("/send_sms_us", sendSMSUS)
	Global.GET("/history_us", historyUS)
	Global.GET("/help", help)
}

func StartServer() error {
	return Global.Run()
}

func StopServer(ctx context.Context) error {
	return Global.Shutdown(ctx)
}
