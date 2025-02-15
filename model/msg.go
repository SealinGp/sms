package model

import (
	"encoding/json"
	"github.com/Akvicor/util"
)

const (
	_ int = iota
	MsgTagSmsReceived
	MsgTagSmsSend
	MsgTagSmsACK
)

type MSG struct {
	Tag  int    `json:"tag"`
	Md5  string `json:"md5"`
	Data string `json:"data"`
	SMS  *SMS   `json:"-"`
}

func NewMSG(tag int, sms []*SMS) []*MSG {
	msg := make([]*MSG, 0, len(sms))
	for _, v := range sms {
		m := &MSG{
			Tag:  tag,
			Md5:  "",
			Data: v.String(),
			SMS:  v,
		}
		m.GenerateMd5()
		msg = append(msg, m)
	}
	return msg
}

func UnmarshalMSG(data []byte) *MSG {
	msg := &MSG{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil
	}
	return msg
}

func (m *MSG) GenerateMd5() {
	m.Md5 = util.NewMD5().FromString(m.Data).Upper()
}

func (m *MSG) Bytes() []byte {
	data, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return data
}

func (m *MSG) String() string {
	data, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(data)
}
