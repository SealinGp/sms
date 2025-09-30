package serial

import (
	"fmt"
	"github.com/Akvicor/glog"
	"github.com/Akvicor/protocol"
	"github.com/Akvicor/util"
	"github.com/patrickmn/go-cache"
	"github.com/tarm/serial"
	"sms/db"
	"sms/model"
	"strings"
	"time"
)

// SerialHandler handles communication with an Air780E module via serial port
type SerialHandler struct {
	config      *SerialConfig
	conn        *serial.Port
	protocol    *protocol.Protocol
	sentCache   *cache.Cache
	sentMap     *SyncMap
	isRunning   bool
}

// NewSerialHandler creates a new serial handler
func NewSerialHandler(config *SerialConfig) *SerialHandler {
	return &SerialHandler{
		config:    config,
		sentCache: cache.New(3*time.Minute, 5*time.Minute),
		sentMap:   NewSyncMap(),
	}
}

// Init initializes the serial connection
func (h *SerialHandler) Init() error {
	conn, err := serial.OpenPort(&serial.Config{
		Name: h.config.DevicePath,
		Baud: h.config.Baud,
	})
	if err != nil {
		return fmt.Errorf("failed to open serial port %s: %v", h.config.DevicePath, err)
	}
	h.conn = conn

	tag := fmt.Sprintf("Air780E-%s", h.config.Name)
	h.protocol = protocol.New(tag, h.conn, h.conn, h.config.SendQueueSize,
		h.readCallback, h.heartbeatFailed, nil, nil, func() {
			time.Sleep(3 * time.Second)
		}, nil, func() {
			_ = h.conn.Close()
		})

	return nil
}

// Start starts the serial handler
func (h *SerialHandler) Start() error {
	if h.isRunning {
		return fmt.Errorf("serial handler %s is already running", h.config.Name)
	}

	if err := h.Init(); err != nil {
		return err
	}

	h.protocol.SetHeartbeatInterval(uint8(h.config.HeartbeatSendInterval))
	h.protocol.SetHeartbeatTimeout(uint8(h.config.HeartbeatReceiveTimeout))
	h.protocol.Connect(true)
	h.isRunning = true

	glog.Info("Serial handler %s started on %s", h.config.Name, h.config.DevicePath)
	return nil
}

// Stop stops the serial handler
func (h *SerialHandler) Stop() error {
	if !h.isRunning {
		return nil
	}

	h.protocol.Kill()
	if h.conn != nil {
		h.conn.Close()
	}
	h.isRunning = false

	glog.Info("Serial handler %s stopped", h.config.Name)
	return nil
}

// Send sends messages via serial port
func (h *SerialHandler) Send(sender string, msgs []*model.MSG) error {
	if !h.isRunning {
		return fmt.Errorf("serial handler %s is not running", h.config.Name)
	}

	for _, msg := range msgs {
		go h.sendSingle(sender, msg)
	}
	return nil
}

// sendSingle sends a single message with retry logic
func (h *SerialHandler) sendSingle(sender string, msg *model.MSG) {
	// Check for duplicate
	cacheKey := msg.SMS.Phone + msg.SMS.Message
	_, isDuplicate := h.sentCache.Get(cacheKey)
	if isDuplicate {
		msg.SMS.Time = "D:" + msg.SMS.Time
		msg.GenerateMd5()
	} else {
		h.sentCache.Set(cacheKey, struct{}{}, 5*time.Minute)
	}

	// Insert into history
	id := db.InsertHistory(h.config.Region, sender, msg.SMS)

	// Send with retry if not duplicate
	if !isDuplicate {
		c := h.sentMap.Put(msg.Md5)
		send := func() {
			err := h.protocol.Write(msg.Bytes())
			for err != nil {
				time.Sleep(3 * time.Second)
				err = h.protocol.Write(msg.Bytes())
			}
		}

		send()

		// Retry logic
		go func() {
			retry := 0
			for {
				select {
				case <-time.After(30 * time.Second):
					send()
				case <-c:
					return
				}
				retry++
				if retry > 10 {
					break
				}
			}
		}()
	}

	// Update history status
	db.UpdateHistorySent(id)
	glog.Trace("[%s] [Send] Sender:[%s] Message:[%s]", h.config.Name, sender, msg.String())
}

// GetName returns the handler name
func (h *SerialHandler) GetName() string {
	return h.config.Name
}

// GetPhone returns the self phone number
func (h *SerialHandler) GetPhone() string {
	return h.config.SelfPhone
}

// IsAlive returns whether the handler is running
func (h *SerialHandler) IsAlive() bool {
	return h.isRunning && h.protocol != nil
}

// heartbeatFailed handles heartbeat failure
func (h *SerialHandler) heartbeatFailed(p *protocol.Protocol) bool {
	glog.Trace("[%s] heartbeat failed", p.GetTag())
	return true
}

// readCallback handles incoming data
func (h *SerialHandler) readCallback(data []byte) {
	msg := model.UnmarshalMSG(data)
	if msg == nil {
		glog.Warning("[%s] unmarshal msg failed", h.config.Name)
		return
	}

	switch msg.Tag {
	case model.MsgTagSmsReceived:
		h.handleReceivedSMS(msg)
	case model.MsgTagSmsACK:
		h.handleACK(msg)
	default:
		glog.Debug("[%s] unknown message tag: %d", h.config.Name, msg.Tag)
	}
}

// handleReceivedSMS handles incoming SMS
func (h *SerialHandler) handleReceivedSMS(msg *model.MSG) {
	sms := model.UnmarshalSMS([]byte(msg.Data))
	if sms == nil {
		glog.Error("[%s] failed to parse SMS", h.config.Name)
		return
	}

	glog.Info("[%s] received SMS from %s: %s", h.config.Name, sms.Phone, sms.Message)
	db.InsertHistory(h.config.Region, h.config.Name, sms)

	// Process commands
	h.processCommands(sms)
}

// handleACK handles acknowledgment messages
func (h *SerialHandler) handleACK(msg *model.MSG) {
	ack := model.UnmarshalACK([]byte(msg.Data))
	if ack == nil {
		glog.Warning("[%s] unmarshal ack failed", h.config.Name)
		return
	}

	h.sentMap.Trick(ack.Key)
	glog.Info("[%s] SMS sent successfully: %s", h.config.Name, ack.Key)
}

// processCommands processes SMS commands
func (h *SerialHandler) processCommands(sms *model.SMS) {
	isSelfPhone := strings.Contains(sms.Phone, h.config.SelfPhone)

	switch sms.Message {
	case "hello":
		response := "Hello! This is SMS service."
		if isSelfPhone {
			response = fmt.Sprintf("Hello %s! This is SMS service on %s.", h.config.SelfPhone, h.config.Name)
		}
		h.Send("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, response)))

	case "你好":
		response := "你好！这里是SMS服务。"
		if isSelfPhone {
			response = fmt.Sprintf("你好 %s！这里是%s的SMS服务。", h.config.SelfPhone, h.config.Name)
		}
		h.Send("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, response)))

	case "status":
		status := fmt.Sprintf("[SMS][%s] Device: %s, Status: Active", h.config.Name, h.config.DevicePath)
		h.Send("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, status)))

	case "ha.help":
		if isSelfPhone {
			h.Send("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, "[HA][HELP]\nha.op.reboot - Reboot OP")))
		}

	case "ha.op.reboot":
		if isSelfPhone {
			_, _ = util.HttpPost("http://127.0.0.1/api/services/script/reboot_router", nil,
				util.HTTPContentTypeJson, map[string]string{"Authorization": "Bearer xxxxxxx"})
			h.Send("sms", model.NewMSG(model.MsgTagSmsSend, model.NewSMSLong(sms.Phone, "Reboot OP")))
		}
	}
}