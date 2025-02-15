package app

import (
	"github.com/Akvicor/glog"
	"net/http"
	"sms/config"
)

func SessionVerify(r *http.Request) bool {
	ses, err := Global.session.Get(r, config.Global.Session.Name)
	if err != nil {
		return false
	}
	username, ok := ses.Values["username"].(string)
	if !ok {
		return false
	}
	return username == config.Global.Security.Username
}

func sessionUpdate(w http.ResponseWriter, r *http.Request, user string) {
	ses, _ := Global.session.Get(r, config.Global.Session.Name)
	if user == config.Global.Security.Username {
		ses.Values["username"] = config.Global.Security.Username
		ses.Values["password"] = config.Global.Security.Password
	}
	ses.Options.MaxAge = config.Global.Session.MaxAge
	err := ses.Save(r, w)
	if err != nil {
		glog.Warning("failed to update session")
		return
	}
}

func sessionDelete(w http.ResponseWriter, r *http.Request) {
	ses, _ := Global.session.Get(r, config.Global.Session.Name)
	ses.Values["username"] = ""
	ses.Values["password"] = ""
	ses.Options.MaxAge = 1
	delete(ses.Values, "username")
	delete(ses.Values, "password")
	err := ses.Save(r, w)
	if err != nil {
		glog.Warning("failed to delete session")
		return
	}
}
