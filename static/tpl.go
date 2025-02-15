package static

import (
	"embed"
	_ "embed"
	"github.com/Akvicor/glog"
	"html/template"
)

//go:embed img/favicon.ico
var Favicon []byte

//go:embed gohtml/*
var html embed.FS

var Login *template.Template
var Index *template.Template
var SendSMS *template.Template
var History *template.Template

func init() {
	t := template.Must(template.ParseFS(html, "gohtml/*"))

	Login = t.Lookup("login.gohtml")
	if Login == nil {
		glog.Fatal("missing gohtml template [login.gohtml]")
	}
	Index = t.Lookup("index.gohtml")
	if Index == nil {
		glog.Fatal("missing gohtml template [index.gohtml]")
	}
	SendSMS = t.Lookup("send_sms.gohtml")
	if SendSMS == nil {
		glog.Fatal("missing gohtml template [send_sms.gohtml]")
	}
	History = t.Lookup("history.gohtml")
	if History == nil {
		glog.Fatal("missing gohtml template [history.gohtml]")
	}
}
