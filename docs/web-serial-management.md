# Web-based Serial Port Management

## Overview

Web 界面管理串口设备，支持树莓派和 immortalwrt x86 主机平台的动态串口配置。

## Architecture

### Serial Port Discovery
```go
type SerialPortInfo struct {
    Path        string `json:"path"`         // /dev/ttyUSB0, /dev/ttyACM0
    Name        string `json:"name"`         // Device name
    Description string `json:"description"`  // Device description
    VendorID    string `json:"vendor_id"`    // USB Vendor ID
    ProductID   string `json:"product_id"`   // USB Product ID
    SerialNum   string `json:"serial_num"`   // Device serial number
    IsAvailable bool   `json:"is_available"` // Device availability
}

type SerialPortManager interface {
    DiscoverPorts() ([]SerialPortInfo, error)
    AddPort(config SerialConfig) error
    RemovePort(portID string) error
    UpdatePort(portID string, config SerialConfig) error
    GetPortStatus(portID string) (PortStatus, error)
    StartPort(portID string) error
    StopPort(portID string) error
}
```

### Database Schema
```sql
-- Serial ports configuration table
CREATE TABLE serial_ports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    device_path VARCHAR(255) NOT NULL,
    baud_rate INTEGER DEFAULT 115200,
    data_bits INTEGER DEFAULT 8,
    stop_bits INTEGER DEFAULT 1,
    parity VARCHAR(10) DEFAULT 'none',
    phone_number VARCHAR(20),
    enabled BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Serial port status tracking
CREATE TABLE serial_port_status (
    port_id INTEGER PRIMARY KEY,
    status VARCHAR(20) NOT NULL, -- online, offline, error
    last_heartbeat DATETIME,
    error_message TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (port_id) REFERENCES serial_ports(id)
);
```

## API Endpoints

### 1. Serial Port Discovery
```
GET /api/serial/discover
Response: {
    "code": 0,
    "msg": "success",
    "data": [
        {
            "path": "/dev/ttyUSB0",
            "name": "Air780E Module",
            "description": "USB Serial Device",
            "vendor_id": "19d1",
            "product_id": "0001",
            "serial_num": "1234567890",
            "is_available": true
        }
    ]
}
```

### 2. Add Serial Port
```
POST /api/serial/ports
Request: {
    "name": "Air780E-1",
    "device_path": "/dev/ttyUSB0",
    "baud_rate": 115200,
    "phone_number": "+8613800138000"
}
Response: {
    "code": 0,
    "msg": "success",
    "data": {"port_id": 1}
}
```

### 3. List Serial Ports
```
GET /api/serial/ports
Response: {
    "code": 0,
    "msg": "success",
    "data": [
        {
            "id": 1,
            "name": "Air780E-1",
            "device_path": "/dev/ttyUSB0",
            "baud_rate": 115200,
            "phone_number": "+8613800138000",
            "enabled": true,
            "status": "online",
            "last_heartbeat": "2024-01-01T12:00:00Z"
        }
    ]
}
```

### 4. Update Serial Port
```
PUT /api/serial/ports/{id}
Request: {
    "name": "Air780E-Updated",
    "baud_rate": 9600,
    "enabled": false
}
```

### 5. Delete Serial Port
```
DELETE /api/serial/ports/{id}
```

### 6. Start/Stop Serial Port
```
POST /api/serial/ports/{id}/start
POST /api/serial/ports/{id}/stop
```

### 7. Port Status Monitor
```
GET /api/serial/ports/{id}/status
Response: {
    "code": 0,
    "msg": "success",
    "data": {
        "status": "online",
        "last_heartbeat": "2024-01-01T12:00:00Z",
        "uptime": "2h 30m 15s",
        "messages_sent": 150,
        "messages_received": 75
    }
}
```

## Frontend Interface

### 1. Serial Port Discovery Page
- 扫描按钮触发设备发现
- 显示可用串口设备列表
- 每个设备显示路径、描述、状态
- 添加按钮配置新设备

### 2. Serial Port Management Page
- 表格显示所有配置的串口
- 实时状态指示器 (在线/离线/错误)
- 启用/禁用开关
- 编辑/删除操作按钮
- 心跳状态和最后活动时间

### 3. Port Configuration Modal
- 设备名称输入
- 串口路径选择 (从发现列表)
- 波特率配置
- 数据位/停止位/校验位设置
- 关联电话号码
- 启用状态开关

## Implementation Details

### 1. Serial Port Discovery
```go
func DiscoverSerialPorts() ([]SerialPortInfo, error) {
    ports, err := serial.GetPortsList()
    if err != nil {
        return nil, err
    }

    var result []SerialPortInfo
    for _, port := range ports {
        info := SerialPortInfo{
            Path: port,
            Name: getDeviceName(port),
            Description: getDeviceDescription(port),
            IsAvailable: checkPortAvailability(port),
        }

        // Get USB device info if available
        if usbInfo := getUSBInfo(port); usbInfo != nil {
            info.VendorID = usbInfo.VendorID
            info.ProductID = usbInfo.ProductID
            info.SerialNum = usbInfo.SerialNumber
        }

        result = append(result, info)
    }

    return result, nil
}
```

### 2. Dynamic Port Management
```go
type DynamicSerialManager struct {
    ports    map[string]*SerialHandler
    configs  map[string]*SerialConfig
    mu       sync.RWMutex
}

func (m *DynamicSerialManager) AddPort(config *SerialConfig) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    handler := NewSerialHandler(config)
    if err := handler.Init(); err != nil {
        return err
    }

    m.ports[config.Name] = handler
    m.configs[config.Name] = config

    // Start handler in background
    go handler.Start()

    return nil
}

func (m *DynamicSerialManager) RemovePort(name string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if handler, exists := m.ports[name]; exists {
        handler.Stop()
        delete(m.ports, name)
        delete(m.configs, name)
    }

    return nil
}
```

### 3. Heartbeat Monitoring
```go
func (h *SerialHandler) startHeartbeatMonitor() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := h.sendHeartbeat(); err != nil {
                h.updateStatus("offline", err.Error())
            } else {
                h.updateStatus("online", "")
            }
        case <-h.stopChan:
            return
        }
    }
}

func (h *SerialHandler) updateStatus(status, errorMsg string) {
    db.UpdatePortStatus(h.config.Name, status, errorMsg, time.Now())
}
```

## Security Considerations

1. **Device Permission**: 确保 Web 服务进程有权限访问串口设备
2. **Authentication**: 串口管理功能需要管理员权限
3. **Input Validation**: 验证设备路径和配置参数
4. **Rate Limiting**: 限制 API 调用频率防止滥用

## Error Handling

1. **设备不存在**: 友好提示设备未连接
2. **权限不足**: 提示需要 root 权限或加入 dialout 组
3. **设备忙**: 提示设备正被其他程序使用
4. **通信错误**: 显示具体错误信息和解决建议