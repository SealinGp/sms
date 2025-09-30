# SMS Platform Next Version Roadmap

## 1. Web-based Serial Port Management

### 1.1 Serial Port Selection and Management
- **Platform Support**: 树莓派 (Raspberry Pi) 和 immortalwrt x86 主机
- **功能特性**:
  - 通过 Web 界面扫描和选择可用串口设备
  - 动态添加串口配置到数据库
  - 实时启动/停止串口的读写操作
  - 串口状态监控（基于心跳机制判断设备在线状态）

### 1.2 Serial Port CRUD Operations
- **添加串口**: 通过 Web 界面配置新的串口设备
- **删除串口**: 移除不再使用的串口配置
- **更新串口**: 修改现有串口的配置参数
- **状态监控**: 实时显示串口连接状态和通信状态

### 1.3 Dynamic Serial Port Management
- 支持热插拔设备的自动检测
- 动态启动和停止串口服务
- 串口设备的健康状态检查

## 2. SMS Forwarding Rules System

### 2.1 Rule-based SMS Processing
- **基于串口的转发规则配置**
- **条件匹配机制**:
  - 短信标题内容匹配
  - 短信正文内容匹配
  - 发送方号码匹配
  - 时间范围匹配

### 2.2 Action Types
- **飞书通知**: 集成飞书 API 发送通知消息
- **短信转发**: 将收到的短信转发到指定号码
- **钉钉通知**: 集成钉钉 API 发送通知消息
- **邮件通知**: 发送邮件通知
- **Webhook**: 支持自定义 HTTP 回调

### 2.3 Rule Management
- Web 界面配置转发规则
- 规则优先级管理
- 规则启用/禁用开关
- 规则执行日志和统计

## 3. Docker Deployment Strategy

### 3.1 Container Configuration
- **网络模式**: 必须使用 `host` 网络模式
- **原因**: 需要访问宿主机的串口设备 (`/dev/ttyUSB*`, `/dev/ttyACM*`)

### 3.2 Volume Mounts
- 串口设备映射: `/dev:/dev`
- 配置文件挂载: `./config:/app/config`
- 数据库文件挂载: `./data:/app/data`
- 日志文件挂载: `./logs:/app/logs`

### 3.3 Docker Compose Example
```yaml
version: '3.8'
services:
  sms-platform:
    build: .
    network_mode: host
    privileged: true
    volumes:
      - /dev:/dev
      - ./config:/app/config
      - ./data:/app/data
      - ./logs:/app/logs
    environment:
      - SMS_CONFIG_PATH=/app/config/config.ini
    restart: unless-stopped
```

### 3.4 Device Permissions
- 容器内访问串口设备权限配置
- udev 规则配置 (如需要)
- 用户组权限管理

## 4. Implementation Priority

1. **Phase 1**: Web-based Serial Port Management
   - 串口扫描和配置界面
   - 动态串口管理 API
   - 状态监控面板

2. **Phase 2**: SMS Forwarding Rules
   - 规则配置界面
   - 消息处理引擎
   - 第三方集成 (飞书、钉钉)

3. **Phase 3**: Docker Deployment
   - Dockerfile 优化
   - Docker Compose 配置
   - 部署文档

## 5. Technical Considerations

- 保持现有 Hertz 框架架构
- 扩展现有 interface 设计
- 数据库 schema 升级脚本
- API 版本兼容性
- 性能优化和监控