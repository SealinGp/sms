package model

import (
	"encoding/json"
	"strings"
	"time"
)

type SMS struct {
	Phone   string `json:"phone"`
	Message string `json:"msg"`
	Time    string `json:"time"`
}

func NewSMSLong(phone, msg string) []*SMS {
	if phone[0] != '+' {
		phone = "+86" + phone
	}
	var smsLen = 0
	smsArray := make([]string, 0, 2)
	buf := strings.Builder{}
	buf.Grow(150)
	lenStep := 1
	for _, v := range msg {
		if uint32(v) >= 128 {
			lenStep = 2
			break
		}
	}
	for _, v := range msg {
		if smsLen == 0 && (v == '\n' || v == ' ') {
			continue
		}
		smsLen += lenStep
		if smsLen < 140 {
			buf.WriteRune(v)
			continue
		} else if smsLen == 140 {
			buf.WriteRune(v)
			smsArray = append(smsArray, buf.String())
			buf.Reset()
			smsLen = 0
			continue
		} else {
			smsArray = append(smsArray, buf.String())
			buf.Reset()
			smsLen = 0
			buf.WriteRune(v)
			continue
		}
	}
	if buf.Len() > 0 {
		smsArray = append(smsArray, buf.String())
	}

	sms := make([]*SMS, 0, len(smsArray))
	t := time.Now().Format("2006-01-02 15:04:05")
	for _, v := range smsArray {
		sms = append(sms, &SMS{
			Phone:   phone,
			Message: v,
			Time:    t,
		})
	}
	return sms
}

func UnmarshalSMS(data []byte) *SMS {
	sms := &SMS{}
	err := json.Unmarshal(data, sms)
	if err != nil {
		return nil
	}
	return sms
}

func (s *SMS) Bytes() []byte {
	data, err := json.Marshal(s)
	if err != nil {
		return nil
	}
	return data
}

func (s *SMS) String() string {
	data, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(data)
}
