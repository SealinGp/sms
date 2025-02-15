package app

import (
	"github.com/Akvicor/glog"
	"github.com/Akvicor/util"
	"github.com/gorilla/sessions"
	"net/http"
	"sync"
)

type app struct {
	mutex   sync.RWMutex
	session *sessions.CookieStore
	handler map[string]func(w http.ResponseWriter, r *http.Request)
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	head, tail := util.SplitPath(r.URL.Path)
	glog.Debug("[%-4s][%-32s] [%s][%s]", r.Method, "/", head, tail)

	var handler func(w http.ResponseWriter, r *http.Request)
	var ok bool

	handler, ok = a.handler[head]
	if ok {
		handler(w, r)
		return
	}

	glog.Debug("Unhandled [%-4s][%-32s] [%s][%s]", r.Method, "/", head, tail)
}
