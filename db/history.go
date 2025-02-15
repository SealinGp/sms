package db

import (
	"github.com/Akvicor/glog"
	"html"
	"sms/model"
	"sync"
	"time"
)

var historyLock = sync.RWMutex{}

type HistoryModel struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Country    string `gorm:"column:country"`
	Sender     string `gorm:"column:sender"`
	RecordTime int64  `gorm:"column:record_time"`
	Phone      string `gorm:"column:phone"`
	Message    string `gorm:"column:message"`
	Time       int64  `gorm:"column:time"`
	SentTime   int64  `gorm:"column:sent_time"`
}

func (HistoryModel) TableName() string {
	return "history"
}

func (h *HistoryModel) Format() HistoryFormatModel {
	his := HistoryFormatModel{
		ID:         h.ID,
		Sender:     html.EscapeString(h.Sender),
		Country:    html.EscapeString(h.Country),
		RecordTime: "",
		Phone:      html.EscapeString(h.Phone),
		Message:    html.EscapeString(h.Message),
		Time:       "",
		SentTime:   "",
	}
	if h.RecordTime != 0 {
		his.RecordTime = html.EscapeString(time.Unix(h.RecordTime, 0).Format("2006-01-02 15:04:05"))
	}
	if h.Time != 0 {
		his.Time = html.EscapeString(time.Unix(h.Time, 0).Format("2006-01-02 15:04:05"))
	}
	if h.SentTime != 0 {
		his.SentTime = html.EscapeString(time.Unix(h.SentTime, 0).Format("2006-01-02 15:04:05"))
	}
	return his
}

type HistoryFormatModel struct {
	ID         int64
	Country    string
	Sender     string
	RecordTime string
	Phone      string
	Message    string
	Time       string
	SentTime   string
}

func GetAllHistories(country string, desc bool) []HistoryModel {
	d := Connect()
	if d == nil {
		return nil
	}
	d = d.Model(&HistoryModel{}).Where("country = ?", country)
	historyLock.RLock()
	defer historyLock.RUnlock()

	histories := make([]HistoryModel, 0)
	if desc {
		d = d.Order("id DESC").Find(&histories)
	} else {
		d = d.Find(&histories)
	}
	if d.Error != nil {
		glog.Warning("get all histories failed [%v] [%v]", d.Error, d.RowsAffected)
		return nil
	}
	return histories
}

func InsertHistory(country string, sender string, sms *model.SMS) int64 {
	if sms == nil {
		return 0
	}
	d := Connect()
	if d == nil {
		return -1
	}
	d = d.Model(&HistoryModel{})
	historyLock.RLock()
	defer historyLock.RUnlock()

	tu := int64(0)
	t, err := time.ParseInLocation("2006-01-02 15:04:05", sms.Time, time.Local)
	if err == nil {
		tu = t.Unix()
	}
	now := time.Now().Unix()
	his := &HistoryModel{
		Country:    country,
		Sender:     sender,
		RecordTime: now,
		Phone:      sms.Phone,
		Message:    sms.Message,
		Time:       tu,
		SentTime:   0,
	}
	res := d.Create(his)
	if res.Error != nil || res.RowsAffected != 1 {
		glog.Warning("insert history failed [%v] [%v]", res.Error, res.RowsAffected)
		return -1
	}
	return his.ID
}

func UpdateHistorySent(id int64) bool {
	d := Connect()
	if d == nil {
		return false
	}
	d = d.Model(&HistoryModel{})
	historyLock.RLock()
	defer historyLock.RUnlock()

	res := d.Where("id = ?", id).Update("sent_time", time.Now().Unix())

	if res.Error != nil || res.RowsAffected != 1 {
		glog.Warning("update [%d] history sent failed [%v] [%v]", id, res.Error, res.RowsAffected)
		return false
	}
	return true
}
