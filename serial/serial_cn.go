package serial

import (
	"github.com/Akvicor/glog"
	"github.com/Akvicor/protocol"
	"github.com/Akvicor/util"
	"github.com/patrickmn/go-cache"
	"github.com/tarm/serial"
	"sms/config"
	"sms/db"
	"sms/model"
	"strings"
	"time"
)

var cnSentCache *cache.Cache
var protCN *protocol.Protocol

const tagCN = "Air780E-CN"
const selfPhoneCN = "12345678900"

func EnableSerialCN() {
	// 防止重复发送
	cnSentCache = cache.New(3*time.Minute, 5*time.Minute)
	// 打开串口
	conn, err := serial.OpenPort(&serial.Config{Name: config.Global.SerialCN.Name, Baud: config.Global.SerialCN.Baud})
	if err != nil {
		glog.Fatal("failed to open serial %v", err)
	}
	protCN = protocol.New(tagCN, conn, conn, config.Global.SerialCN.SendQueueSize,
		readCallbackCN, heartbeatFailedCN, nil, nil, func() {
			time.Sleep(3 * time.Second)
		}, nil, func() {
			_ = conn.Close()
		})
	protCN.SetHeartbeatInterval(uint8(config.Global.SerialCN.HeartbeatSendInterval))
	protCN.SetHeartbeatTimeout(uint8(config.Global.SerialCN.HeartbeatReceiveTimeout))
	protCN.Connect(true)
}

func heartbeatFailedCN(p *protocol.Protocol) bool {
	glog.Trace("[%s] heartbeat failed", p.GetTag())
	return true
}

func readCallbackCN(data []byte) {
	msg := model.UnmarshalMSG(data)
	if msg == nil {
		glog.Warning("[%s] unmarshal msg failed, rev: %s\n", tagCN, string(data))
		return
	}
	if msg.Tag == model.MsgTagSmsReceived {
		sms := model.UnmarshalSMS([]byte(msg.Data))
		if sms == nil {
			glog.Warning("[%s] unmarshal sms failed, data: %s\n", tagCN, msg.Data)
			return
		}
		db.InsertHistory("CN", tagCN, sms)
		glog.Info("[%s] [Received] sms Phone:[%s] Time:[%s] Message:[%s]", tagCN, sms.Phone, sms.Time, sms.Message)
		if sms.Message == "hello" {
			if strings.Contains(sms.Phone, selfPhoneCN) {
				SendCN("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, "Hello Akvicor! here is sms")))
			} else {
				SendCN("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, "Hello! here is sms")))
			}
		} else if sms.Message == "你好" {
			if strings.Contains(sms.Phone, selfPhoneCN) {
				SendCN("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, "你好Akvicor！这里是sms")))
			} else {
				SendCN("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, "你好！这里是sms")))
			}
		} else if sms.Message == "ha.help" && strings.Contains(sms.Phone, selfPhoneCN) {
			SendCN("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, "[HA][HELP]\nha.op.reboot - Reboot OP")))
		} else if sms.Message == "ha.op.reboot" && strings.Contains(sms.Phone, selfPhoneCN) {
			_, _ = util.HttpPost("http://127.0.0.1/api/services/script/reboot_router", nil, util.HTTPContentTypeJson, map[string]string{"Authorization": "Bearer xxxxxxx"})
			SendCN("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, "Reboot OP")))
		}
	} else if msg.Tag == model.MsgTagSmsACK {
		ack := model.UnmarshalACK([]byte(msg.Data))
		if ack == nil {
			glog.Warning("[%s] unmarshal ack failed, data: %s\n", tagCN, msg.Data)
			return
		}
		sendCNMap.Trick(ack.Key)
	}
}

func SendCN(sender string, msg []*model.MSG) {
	for _, v := range msg {
		go sendCN(sender, v)
	}
}

var sendCNMap = NewSyncMap()

func sendCN(sender string, msg *model.MSG) {
	_, ok := cnSentCache.Get(msg.SMS.Phone + msg.SMS.Message)
	if ok {
		msg.SMS.Time = "D:" + msg.SMS.Time
		msg.GenerateMd5()
	} else {
		cnSentCache.Set(msg.SMS.Phone+msg.SMS.Message, struct{}{}, 5*time.Minute)
	}
	id := db.InsertHistory("CN", sender, msg.SMS)
	if !ok {
		c := sendCNMap.Put(msg.Md5)
		send := func() {
			err := protCN.Write(msg.Bytes())
			for err != nil {
				time.Sleep(3 * time.Second)
				err = protCN.Write(msg.Bytes())
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
				if retry >= 2 {
					return
				}
				retry += 1
			}
		}()
		sendCNMap.Delete(msg.Md5)
	}
	db.UpdateHistorySent(id)
	glog.Trace("[%s] [Send] send Sender:[%s] Message:[%s]", tagCN, sender, msg.String())
}

func KillCN() {
	protCN.Kill()
}
