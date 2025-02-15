package serial

import (
	"github.com/Akvicor/glog"
	"github.com/Akvicor/protocol"
	"github.com/patrickmn/go-cache"
	"github.com/tarm/serial"
	"sms/config"
	"sms/db"
	"sms/model"
	"strings"
	"time"
)

var usSentCache *cache.Cache
var protUS *protocol.Protocol

const tagUS = "Air780E-US"

func EnableSerialUS() {
	// 防止重复发送
	usSentCache = cache.New(3*time.Minute, 5*time.Minute)
	// 打开串口
	conn, err := serial.OpenPort(&serial.Config{Name: config.Global.SerialUS.Name, Baud: config.Global.SerialUS.Baud})
	if err != nil {
		glog.Fatal("failed to open serial %v", err)
	}
	protUS = protocol.New(tagUS, conn, conn, config.Global.SerialUS.SendQueueSize,
		readCallbackUS, heartbeatFailedUS, nil, nil, func() {
			time.Sleep(3 * time.Second)
		}, nil, func() {
			_ = conn.Close()
		})
	protUS.SetHeartbeatInterval(uint8(config.Global.SerialUS.HeartbeatSendInterval))
	protUS.SetHeartbeatTimeout(uint8(config.Global.SerialUS.HeartbeatReceiveTimeout))
	protUS.Connect(true)
}

func heartbeatFailedUS(p *protocol.Protocol) bool {
	glog.Trace("[%s] heartbeat failed", p.GetTag())
	return true
}

func readCallbackUS(data []byte) {
	msg := model.UnmarshalMSG(data)
	if msg == nil {
		glog.Warning("[%s] unmarshal msg failed, rev: %s\n", tagUS, string(data))
		return
	}
	if msg.Tag == model.MsgTagSmsReceived {
		sms := model.UnmarshalSMS([]byte(msg.Data))
		if sms == nil {
			glog.Warning("[%s] unmarshal sms failed, data: %s\n", tagUS, msg.Data)
			return
		}
		db.InsertHistory("US", tagUS, sms)
		glog.Info("[%s] [Received] sms Phone:[%s] Time:[%s] Message:[%s]", tagUS, sms.Phone, sms.Time, sms.Message)
		if sms.Message == "hello" {
			if strings.Contains(sms.Phone, selfPhoneCN) {
				SendUS("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, "Hello Akvicor! here is sms")))
			} else {
				SendUS("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, "Hello! here is sms")))
			}
		} else if sms.Message == "你好" {
			if strings.Contains(sms.Phone, selfPhoneCN) {
				SendUS("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, "你好Akvicor！这里是sms")))
			} else {
				SendUS("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, "你好！这里是sms")))
			}
		}
	} else if msg.Tag == model.MsgTagSmsACK {
		ack := model.UnmarshalACK([]byte(msg.Data))
		if ack == nil {
			glog.Warning("[%s] unmarshal ack failed, data: %s\n", tagUS, msg.Data)
			return
		}
		sendUSMap.Trick(ack.Key)
	}
}

func SendUS(sender string, msg []*model.MSG) {
	for _, v := range msg {
		go sendUS(sender, v)
	}
}

var sendUSMap = NewSyncMap()

func sendUS(sender string, msg *model.MSG) {
	_, ok := usSentCache.Get(msg.SMS.Phone + msg.SMS.Message)
	if ok {
		msg.SMS.Time = "D:" + msg.SMS.Time
		msg.GenerateMd5()
	} else {
		usSentCache.Set(msg.SMS.Phone+msg.SMS.Message, struct{}{}, 5*time.Minute)
	}
	id := db.InsertHistory("US", sender, msg.SMS)
	if !ok {
		c := sendUSMap.Put(msg.Md5)
		send := func() {
			err := protUS.Write(msg.Bytes())
			for err != nil {
				time.Sleep(3 * time.Second)
				err = protUS.Write(msg.Bytes())
			}
		}
		send()
		func() {
			retry := 0
			for {
				select {
				case <-time.After(30 * time.Second):
					send()
				case <-c:
					return
				}
				if retry >= 10 {
					return
				}
				retry += 1
			}
		}()
		sendUSMap.Delete(msg.Md5)
	}
	db.UpdateHistorySent(id)
	glog.Trace("[%s] [Send] send Sender:[%s] Message:[%s]", tagUS, sender, msg.String())
}

func KillUS() {
	if protUS != nil {
		protUS.Kill()
	}
}
