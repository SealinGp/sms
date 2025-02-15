package app

import (
	"github.com/Akvicor/glog"
	"github.com/Akvicor/util"
	"net/http"
	"sms/config"
	"sms/db"
	"sms/model"
	"sms/serial"
	"sms/static"
	"strconv"
)

func staticFavicon(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write(static.Favicon)
}

func index(w http.ResponseWriter, r *http.Request) {
	head, tail := util.SplitPathRepeat(r.URL.Path, 1)
	glog.Debug("[%-4s][%-32s] [%s][%s]", r.Method, "/index", head, tail)
	if !SessionVerify(r) {
		Login(w, r)
		return
	}

	if r.Method == http.MethodGet {
		_ = static.Index.Execute(w, map[string]interface{}{"title": "SMS Pusher"})
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	head, tail := util.SplitPathRepeat(r.URL.Path, 1)
	glog.Debug("[%-4s][%-32s] [%s][%s]", r.Method, "/login", head, tail)

	if r.Method == http.MethodGet {
		_ = static.Login.Execute(w, map[string]interface{}{"title": "Login"})
		return
	} else if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == config.Global.Security.Username && password == config.Global.Security.Password {
			glog.Info("Login successful [%s]", username)
			sessionUpdate(w, r, config.Global.Security.Username)
			util.RespRedirect(w, r, r.URL.String())
		} else {
			glog.Info("Login failed: [%s][%s]", username, password)
			util.RespRedirect(w, r, r.URL.String())
		}
	}
}

func randomKey(w http.ResponseWriter, r *http.Request) {
	head, tail := util.SplitPathRepeat(r.URL.Path, 1)
	glog.Debug("[%-4s][%-32s] [%s][%s]", r.Method, "/random_key", head, tail)
	if !SessionVerify(r) {
		Login(w, r)
		return
	}
	rRange := r.FormValue("range")
	if rRange == "" {
		rRange = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	}
	rLength := r.FormValue("length")
	cLength := 8
	if rLength != "" {
		clen, err := strconv.Atoi(rLength)
		if err == nil {
			cLength = clen
		}
	}
	_, _ = w.Write([]byte(util.RandomString(cLength, rRange)))
	return

}

func sendSMSCN(w http.ResponseWriter, r *http.Request) {
	head, tail := util.SplitPathRepeat(r.URL.Path, 1)
	glog.Debug("[%-4s][%-32s] [%s][%s]", r.Method, "/send_sms_cn", head, tail)

	key := ""
	phone := ""
	message := ""
	sender := ""
	if r.Method == "GET" {
		values := r.URL.Query()
		key = values.Get("key")
		phone = values.Get("phone")
		message = values.Get("message")
		sender = values.Get("sender")
	} else if r.Method == "POST" {
		key = r.PostFormValue("key")
		phone = r.PostFormValue("phone")
		message = r.PostFormValue("message")
		sender = r.PostFormValue("sender")
	}

	if r.Method == "GET" && key == "" {
		if !SessionVerify(r) {
			Login(w, r)
			return
		}
		_ = static.SendSMS.Execute(w, map[string]any{"title": "Send SMS", "url": "/send_sms_cn"})
		return
	}

	if (key != config.Global.Security.AccessKey) && (!SessionVerify(r)) {
		util.WriteHTTPRespAPIInvalidInput(w, "invalid access key")
		return
	}
	if len(phone) < 1 {
		util.WriteHTTPRespAPIInvalidInput(w, "invalid phone number")
		return
	}
	if len(message) < 1 {
		util.WriteHTTPRespAPIInvalidInput(w, "invalid message")
		return
	}
	if len(sender) < 1 {
		util.WriteHTTPRespAPIInvalidInput(w, "invalid sender")
		return
	}

	serial.SendCN(sender, model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(phone, message)))

	util.WriteHTTPRespAPIOk(w, nil)
}

func historyCN(w http.ResponseWriter, r *http.Request) {
	head, tail := util.SplitPathRepeat(r.URL.Path, 1)
	glog.Debug("[%-4s][%-32s] [%s][%s]", r.Method, "/history_cn", head, tail)
	if !SessionVerify(r) {
		Login(w, r)
		return
	}

	if r.Method == http.MethodGet {
		his := db.GetAllHistories("CN", true)
		glog.Debug("%#v", his)
		histories := make([]db.HistoryFormatModel, len(his))
		for k, v := range his {
			histories[k] = v.Format()
		}
		_ = static.History.Execute(w, map[string]any{"title": "History", "histories": histories})
		return
	}
}

func sendSMSUS(w http.ResponseWriter, r *http.Request) {
	head, tail := util.SplitPathRepeat(r.URL.Path, 1)
	glog.Debug("[%-4s][%-32s] [%s][%s]", r.Method, "/send_sms_us", head, tail)

	key := ""
	phone := ""
	message := ""
	sender := ""
	if r.Method == "GET" {
		values := r.URL.Query()
		key = values.Get("key")
		phone = values.Get("phone")
		message = values.Get("message")
		sender = values.Get("sender")
	} else if r.Method == "POST" {
		key = r.PostFormValue("key")
		phone = r.PostFormValue("phone")
		message = r.PostFormValue("message")
		sender = r.PostFormValue("sender")
	}

	if r.Method == "GET" && key == "" {
		if !SessionVerify(r) {
			Login(w, r)
			return
		}
		_ = static.SendSMS.Execute(w, map[string]any{"title": "Send SMS", "url": "/send_sms_us"})
		return
	}

	if (key != config.Global.Security.AccessKey) && (!SessionVerify(r)) {
		util.WriteHTTPRespAPIInvalidInput(w, "invalid access key")
		return
	}
	if len(phone) < 1 {
		util.WriteHTTPRespAPIInvalidInput(w, "invalid phone number")
		return
	}
	if len(message) < 1 {
		util.WriteHTTPRespAPIInvalidInput(w, "invalid message")
		return
	}
	if len(sender) < 1 {
		util.WriteHTTPRespAPIInvalidInput(w, "invalid sender")
		return
	}

	serial.SendUS(sender, model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(phone, message)))

	util.WriteHTTPRespAPIOk(w, nil)
}

func historyUS(w http.ResponseWriter, r *http.Request) {
	head, tail := util.SplitPathRepeat(r.URL.Path, 1)
	glog.Debug("[%-4s][%-32s] [%s][%s]", r.Method, "/history_us", head, tail)
	if !SessionVerify(r) {
		Login(w, r)
		return
	}

	if r.Method == http.MethodGet {
		his := db.GetAllHistories("US", true)
		glog.Debug("%#v", his)
		histories := make([]db.HistoryFormatModel, len(his))
		for k, v := range his {
			histories[k] = v.Format()
		}
		_ = static.History.Execute(w, map[string]any{"title": "History", "histories": histories})
		return
	}
}

func help(w http.ResponseWriter, r *http.Request) {
	head, tail := util.SplitPathRepeat(r.URL.Path, 1)
	glog.Debug("[%-4s][%-32s] [%s][%s]", r.Method, "/help", head, tail)
	if !SessionVerify(r) {
		Login(w, r)
		return
	}

	if r.Method == http.MethodGet {
		w.Write([]byte(`
GET:
  /random_key?range=(不提供则使用默认值)&length=(默认为8)
POST:
  /send_sms?key=(访问密钥,如果已通过网页登录则不需要)&sender=(发送者)&phone=(手机号)&message=(短信内容)
`))
		return
	}
}
