package app

import (
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
	"sms/config"
	"sync"
)

var Global *app

func Generate() {
	Global = new(app)
	Global.session = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
	Global.session.Options.Domain = config.Global.Session.Domain
	Global.session.Options.Path = config.Global.Session.Path
	Global.session.Options.MaxAge = config.Global.Session.MaxAge
	Global.handler = make(map[string]func(w http.ResponseWriter, r *http.Request))
	Global.mutex = sync.RWMutex{}

	Global.handler["/favicon.ico"] = staticFavicon

	Global.handler["/"] = index
	Global.handler["/login"] = Login
	Global.handler["/random_key"] = randomKey
	Global.handler["/send_sms"] = sendSMSCN
	Global.handler["/history"] = historyCN
	Global.handler["/send_sms_cn"] = sendSMSCN
	Global.handler["/history_cn"] = historyCN
	Global.handler["/send_sms_us"] = sendSMSUS
	Global.handler["/history_us"] = historyUS
	Global.handler["/help"] = help
}
