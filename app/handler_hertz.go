package app

import (
	"context"
	"github.com/Akvicor/glog"
	"github.com/Akvicor/util"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"sms/config"
	"sms/db"
	"sms/model"
	"sms/serial"
	"sms/static"
	"strconv"
)

func staticFavicon(ctx context.Context, c *app.RequestContext) {
	c.Response.Header.Set("Content-Type", "image/x-icon")
	c.Write(static.Favicon)
}

func index(ctx context.Context, c *app.RequestContext) {
	glog.Debug("[%-4s][%-32s] %s", c.Method(), "/index", c.Path())
	if !sessionVerify(ctx, c) {
		loginGet(ctx, c)
		return
	}

	if string(c.Method()) == "GET" {
		c.Response.Header.Set("Content-Type", "text/html; charset=utf-8")
		_ = static.Index.Execute(c.Response.BodyWriter(), map[string]interface{}{"title": "SMS Pusher"})
	}
}

func loginGet(ctx context.Context, c *app.RequestContext) {
	glog.Debug("[%-4s][%-32s] %s", c.Method(), "/login", c.Path())

	c.Response.Header.Set("Content-Type", "text/html; charset=utf-8")
	_ = static.Login.Execute(c.Response.BodyWriter(), map[string]interface{}{"title": "Login"})
}

func loginPost(ctx context.Context, c *app.RequestContext) {
	glog.Debug("[%-4s][%-32s] %s", c.Method(), "/login", c.Path())

	username := string(c.PostForm("username"))
	password := string(c.PostForm("password"))

	if username == config.Global.Security.Username && password == config.Global.Security.Password {
		glog.Info("Login successful [%s]", username)
		sessionUpdate(ctx, c, config.Global.Security.Username)
		c.Redirect(consts.StatusFound, []byte(string(c.URI().RequestURI())))
	} else {
		glog.Info("Login failed: [%s][%s]", username, password)
		c.Redirect(consts.StatusFound, []byte(string(c.URI().RequestURI())))
	}
}

func randomKey(ctx context.Context, c *app.RequestContext) {
	glog.Debug("[%-4s][%-32s] %s", c.Method(), "/random_key", c.Path())

	if !sessionVerify(ctx, c) && !keyVerify(ctx, c) {
		writeHTTPRespAPINotAuthorized(c)
		return
	}

	rangeStr := string(c.Query("range"))
	lengthStr := string(c.Query("length"))

	var rangeVal uint64 = 100000000
	var lengthVal int = 8

	if rangeStr != "" {
		if val, err := strconv.ParseUint(rangeStr, 10, 64); err == nil {
			rangeVal = val
		}
	}

	if lengthStr != "" {
		if val, err := strconv.Atoi(lengthStr); err == nil {
			lengthVal = val
		}
	}

	key := util.RandomString(int(rangeVal), strconv.Itoa(lengthVal))
	writeHTTPRespAPIOk(c, map[string]interface{}{"key": key})
}

func sendSMSCN(ctx context.Context, c *app.RequestContext) {
	glog.Debug("[%-4s][%-32s] %s", c.Method(), "/send_sms_cn", c.Path())

	if !sessionVerify(ctx, c) && !keyVerify(ctx, c) {
		writeHTTPRespAPINotAuthorized(c)
		return
	}

	sender := string(c.PostForm("sender"))
	phone := string(c.PostForm("phone"))
	message := string(c.PostForm("message"))

	if len(phone) < 1 {
		writeHTTPRespAPIInvalidInput(c, "invalid phone number")
		return
	}
	if len(message) < 1 {
		writeHTTPRespAPIInvalidInput(c, "invalid message")
		return
	}
	if len(sender) < 1 {
		writeHTTPRespAPIInvalidInput(c, "invalid sender")
		return
	}

	serial.Send("cn", sender, model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(phone, message)))

	writeHTTPRespAPIOk(c, nil)
}

func historyCN(ctx context.Context, c *app.RequestContext) {
	glog.Debug("[%-4s][%-32s] %s", c.Method(), "/history_cn", c.Path())
	if !sessionVerify(ctx, c) {
		loginGet(ctx, c)
		return
	}

	if string(c.Method()) == "GET" {
		his := db.GetAllHistories("CN", true)
		glog.Debug("%#v", his)
		c.Response.Header.Set("Content-Type", "text/html; charset=utf-8")
		_ = static.History.Execute(c.Response.BodyWriter(), map[string]interface{}{"title": "History CN", "histories": his})
	}
}

func sendSMSUS(ctx context.Context, c *app.RequestContext) {
	glog.Debug("[%-4s][%-32s] %s", c.Method(), "/send_sms_us", c.Path())

	if !sessionVerify(ctx, c) && !keyVerify(ctx, c) {
		writeHTTPRespAPINotAuthorized(c)
		return
	}

	sender := string(c.PostForm("sender"))
	phone := string(c.PostForm("phone"))
	message := string(c.PostForm("message"))

	if len(phone) < 1 {
		writeHTTPRespAPIInvalidInput(c, "invalid phone number")
		return
	}
	if len(message) < 1 {
		writeHTTPRespAPIInvalidInput(c, "invalid message")
		return
	}
	if len(sender) < 1 {
		writeHTTPRespAPIInvalidInput(c, "invalid sender")
		return
	}

	serial.Send("us", sender, model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(phone, message)))

	writeHTTPRespAPIOk(c, nil)
}

func historyUS(ctx context.Context, c *app.RequestContext) {
	glog.Debug("[%-4s][%-32s] %s", c.Method(), "/history_us", c.Path())
	if !sessionVerify(ctx, c) {
		loginGet(ctx, c)
		return
	}

	if string(c.Method()) == "GET" {
		his := db.GetAllHistories("US", true)
		glog.Debug("%#v", his)
		c.Response.Header.Set("Content-Type", "text/html; charset=utf-8")
		_ = static.History.Execute(c.Response.BodyWriter(), map[string]interface{}{"title": "History US", "histories": his})
	}
}

func help(ctx context.Context, c *app.RequestContext) {
	glog.Debug("[%-4s][%-32s] %s", c.Method(), "/help", c.Path())
	if !sessionVerify(ctx, c) {
		loginGet(ctx, c)
		return
	}

	if string(c.Method()) == "GET" {
		c.Response.Header.Set("Content-Type", "text/html; charset=utf-8")
		_ = static.Index.Execute(c.Response.BodyWriter(), map[string]interface{}{"title": "Help"})
	}
}

// Helper functions for HTTP responses
func writeHTTPRespAPIOk(c *app.RequestContext, data interface{}) {
	c.JSON(consts.StatusOK, map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
}

func writeHTTPRespAPIInvalidInput(c *app.RequestContext, msg string) {
	c.JSON(consts.StatusBadRequest, map[string]interface{}{
		"code": 1,
		"msg":  msg,
		"data": nil,
	})
}

func writeHTTPRespAPINotAuthorized(c *app.RequestContext) {
	c.JSON(consts.StatusUnauthorized, map[string]interface{}{
		"code": 2,
		"msg":  "not authorized",
		"data": nil,
	})
}