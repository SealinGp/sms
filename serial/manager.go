package serial

import (
	"fmt"
	"github.com/Akvicor/glog"
	"sms/config"
	"sms/model"
	"time"
)

var Manager *SerialManager

func InitSerialManager() {
	Manager = NewSerialManager()

	glog.Info("Initializing serial manager with %d devices", len(config.Global.SerialDevices))

	for _, deviceConfig := range config.Global.SerialDevices {
		cfg := &SerialConfig{
			Name:                    deviceConfig.Name,
			DevicePath:              deviceConfig.DevicePath,
			Baud:                    deviceConfig.Baud,
			SendQueueSize:           deviceConfig.SendQueueSize,
			HeartbeatSendInterval:   time.Duration(deviceConfig.HeartbeatSendInterval) * time.Second,
			HeartbeatReceiveTimeout: time.Duration(deviceConfig.HeartbeatReceiveTimeout) * time.Second,
			SelfPhone:               deviceConfig.SelfPhone,
			Region:                  deviceConfig.Region,
		}

		handler := NewSerialHandler(cfg)
		Manager.AddHandler(deviceConfig.Name, handler)
		glog.Info("Added serial device: %s on %s", deviceConfig.Name, deviceConfig.DevicePath)
	}
}

func EnableSerial() {
	InitSerialManager()

	if err := Manager.StartAll(); err != nil {
		glog.Fatal("Failed to start serial handlers: %v", err)
	}

	glog.Info("All serial handlers started successfully")
}

func KillSerial() {
	if Manager != nil {
		if err := Manager.StopAll(); err != nil {
			glog.Error("Failed to stop serial handlers: %v", err)
		}
	}
}

func Send(deviceName string, sender string, msg []*model.MSG) error {
	handler := Manager.GetHandler(deviceName)
	if handler == nil {
		return fmt.Errorf("device %s not found", deviceName)
	}

	return handler.Send(sender, msg)
}

func SendToAll(sender string, msg []*model.MSG) {
	for name, handler := range Manager.GetAllHandlers() {
		if err := handler.Send(sender, msg); err != nil {
			glog.Error("Failed to send to %s: %v", name, err)
		}
	}
}

func GetDeviceStatus(deviceName string) (bool, error) {
	handler := Manager.GetHandler(deviceName)
	if handler == nil {
		return false, fmt.Errorf("device %s not found", deviceName)
	}

	return handler.IsAlive(), nil
}

func GetAllDeviceStatus() map[string]bool {
	status := make(map[string]bool)
	for name, handler := range Manager.GetAllHandlers() {
		status[name] = handler.IsAlive()
	}
	return status
}