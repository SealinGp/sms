package serial

import (
	"sms/model"
	"time"
)

type SerialHandlerInterface interface {
	Init() error
	Start() error
	Stop() error
	Send(sender string, msg []*model.MSG) error
	GetName() string
	GetPhone() string
	IsAlive() bool
}

type SerialConfig struct {
	Name                    string
	DevicePath              string
	Baud                    int
	SendQueueSize           int
	HeartbeatSendInterval   time.Duration
	HeartbeatReceiveTimeout time.Duration
	SelfPhone               string
	Region                  string
}

type SerialManager struct {
	handlers map[string]SerialHandlerInterface
}

func NewSerialManager() *SerialManager {
	return &SerialManager{
		handlers: make(map[string]SerialHandlerInterface),
	}
}

func (sm *SerialManager) AddHandler(name string, handler SerialHandlerInterface) {
	sm.handlers[name] = handler
}

func (sm *SerialManager) GetHandler(name string) SerialHandlerInterface {
	return sm.handlers[name]
}

func (sm *SerialManager) GetAllHandlers() map[string]SerialHandlerInterface {
	return sm.handlers
}

func (sm *SerialManager) StartAll() error {
	for _, handler := range sm.handlers {
		if err := handler.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (sm *SerialManager) StopAll() error {
	for _, handler := range sm.handlers {
		if err := handler.Stop(); err != nil {
			return err
		}
	}
	return nil
}