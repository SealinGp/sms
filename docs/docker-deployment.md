# Docker Deployment Strategy

## Overview

SMS 平台的容器化部署策略，支持 host 网络模式以访问宿主机串口设备。

## Network Mode Requirement

### Why Host Network Mode?

SMS 平台需要直接访问宿主机的串口设备 (`/dev/ttyUSB*`, `/dev/ttyACM*`)，因此必须使用 `host` 网络模式：

1. **直接硬件访问**: 容器需要访问 `/dev` 目录下的串口设备
2. **设备权限**: 需要与宿主机共享设备权限
3. **热插拔支持**: 支持 USB 设备的动态插拔检测

## Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o sms-platform .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    sqlite \
    tzdata \
    ca-certificates

# Create app user
RUN addgroup -g 1001 appgroup && \
    adduser -D -u 1001 -G appgroup appuser

# Add appuser to dialout group for serial port access
RUN addgroup dialout && adduser appuser dialout

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/sms-platform .

# Create necessary directories
RUN mkdir -p /app/config /app/data /app/logs && \
    chown -R appuser:appgroup /app

# Copy default configuration
COPY config/config.ini.example /app/config/config.ini

# Expose port (only used for documentation, host mode ignores this)
EXPOSE 8080

# Run as non-root user
USER appuser

CMD ["./sms-platform", "-c", "/app/config/config.ini"]
```

## Docker Compose Configuration

### Basic Setup
```yaml
version: '3.8'

services:
  sms-platform:
    build: .
    container_name: sms-platform
    network_mode: host
    privileged: true
    restart: unless-stopped

    volumes:
      # Serial devices access
      - /dev:/dev

      # Application data
      - ./config:/app/config
      - ./data:/app/data
      - ./logs:/app/logs

      # System information (for device detection)
      - /sys:/sys:ro
      - /proc:/proc:ro

    environment:
      - SMS_CONFIG_PATH=/app/config/config.ini
      - TZ=Asia/Shanghai
      - LOG_LEVEL=info

    # Device access
    devices:
      - /dev/ttyUSB0:/dev/ttyUSB0
      - /dev/ttyUSB1:/dev/ttyUSB1
      - /dev/ttyACM0:/dev/ttyACM0
      - /dev/ttyACM1:/dev/ttyACM1

    # Healthcheck
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

### Production Setup with Monitoring
```yaml
version: '3.8'

services:
  sms-platform:
    build: .
    container_name: sms-platform
    network_mode: host
    privileged: true
    restart: unless-stopped

    volumes:
      - /dev:/dev
      - ./config:/app/config
      - ./data:/app/data
      - ./logs:/app/logs
      - /sys:/sys:ro
      - /proc:/proc:ro

    environment:
      - SMS_CONFIG_PATH=/app/config/config.ini
      - TZ=Asia/Shanghai
      - LOG_LEVEL=info
      - ENABLE_METRICS=true

    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"

    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # Optional: Log aggregation
  loki:
    image: grafana/loki:latest
    ports:
      - "3100:3100"
    volumes:
      - ./loki-config.yaml:/etc/loki/local-config.yaml
    command: -config.file=/etc/loki/local-config.yaml

  # Optional: Metrics collection
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
```

## Configuration Management

### Environment Variables
```bash
# Application settings
SMS_CONFIG_PATH=/app/config/config.ini
LOG_LEVEL=info
TZ=Asia/Shanghai

# Database settings
DB_PATH=/app/data/sms.db

# Server settings
HTTP_ADDR=0.0.0.0
HTTP_PORT=8080
ENABLE_HTTPS=false

# Security settings
SESSION_SECRET=your-secret-key
ACCESS_KEY=your-access-key

# Serial port settings
SERIAL_TIMEOUT=30s
HEARTBEAT_INTERVAL=30s

# Monitoring
ENABLE_METRICS=true
METRICS_PORT=9100
```

### Configuration File Template
```ini
# /app/config/config.ini
[server]
http_addr = 0.0.0.0
http_port = 8080
enable_https = false
ssl_cert = /app/config/cert.pem
ssl_key = /app/config/key.pem

[database]
path = /app/data/sms.db

[security]
username = admin
password = your-password
access_key = your-access-key

[session]
name = sms_session
domain = localhost
path = /
max_age = 7200

[log]
log_to_file = true
file_path = /app/logs/sms.log

[serial]
timeout = 30
heartbeat_interval = 30
```

## Deployment Commands

### Build and Deploy
```bash
# Build image
docker build -t sms-platform:latest .

# Run with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f sms-platform

# Check status
docker-compose ps

# Stop services
docker-compose down
```

### Production Deployment
```bash
# Pull latest code
git pull origin main

# Build production image
docker build -t sms-platform:$(git rev-parse --short HEAD) .
docker tag sms-platform:$(git rev-parse --short HEAD) sms-platform:latest

# Deploy with zero downtime
docker-compose up -d --force-recreate

# Cleanup old images
docker image prune -f
```

## Volume Mounts

### Required Mounts
```yaml
volumes:
  # Device access
  - /dev:/dev                    # Serial devices

  # Application data
  - ./config:/app/config         # Configuration files
  - ./data:/app/data             # Database and data files
  - ./logs:/app/logs             # Log files

  # System info (read-only)
  - /sys:/sys:ro                 # System information
  - /proc:/proc:ro               # Process information
```

### Directory Structure
```
project/
├── docker-compose.yml
├── Dockerfile
├── config/
│   ├── config.ini
│   ├── cert.pem (if HTTPS)
│   └── key.pem (if HTTPS)
├── data/
│   └── sms.db (created on first run)
└── logs/
    └── sms.log
```

## Device Access and Permissions

### Udev Rules (Host System)
```bash
# /etc/udev/rules.d/99-sms-devices.rules
# Air780E devices
SUBSYSTEM=="tty", ATTRS{idVendor}=="19d1", ATTRS{idProduct}=="0001", GROUP="dialout", MODE="0664"

# Reload udev rules
sudo udevadm control --reload-rules
sudo udevadm trigger
```

### Container Permissions
```dockerfile
# Add user to dialout group
RUN addgroup dialout && adduser appuser dialout

# Or run as root (less secure)
USER root
```

### Runtime Device Access
```yaml
# Static device mapping
devices:
  - /dev/ttyUSB0:/dev/ttyUSB0
  - /dev/ttyUSB1:/dev/ttyUSB1

# Or full /dev access (more flexible)
volumes:
  - /dev:/dev
```

## Health Checks and Monitoring

### Health Check Endpoint
```go
// Add to your application
func healthCheck(ctx context.Context, c *app.RequestContext) {
    status := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now(),
        "version": version,
        "serial_ports": getSerialPortStatus(),
    }

    c.JSON(consts.StatusOK, status)
}
```

### Docker Health Check
```dockerfile
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/api/health || exit 1
```

### Monitoring Scripts
```bash
#!/bin/bash
# monitor.sh - Container monitoring script

CONTAINER_NAME="sms-platform"

# Check container status
if ! docker ps | grep -q $CONTAINER_NAME; then
    echo "Container $CONTAINER_NAME is not running"
    docker-compose up -d
fi

# Check application health
if ! curl -f http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "Application health check failed"
    docker-compose restart $CONTAINER_NAME
fi

# Check disk space
DISK_USAGE=$(df /var/lib/docker | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 90 ]; then
    echo "Disk usage is high: $DISK_USAGE%"
    docker system prune -f
fi
```

## Security Considerations

### 1. Container Security
```yaml
# Run as non-root user
user: "1001:1001"

# Read-only root filesystem
read_only: true
tmpfs:
  - /tmp
  - /var/tmp

# Drop capabilities
cap_drop:
  - ALL
cap_add:
  - CHOWN
  - SETGID
  - SETUID
```

### 2. Network Security
```yaml
# Use host network but restrict access
network_mode: host

# Alternative: Use custom network with port mapping
networks:
  sms_network:
    driver: bridge
ports:
  - "127.0.0.1:8080:8080"  # Bind to localhost only
```

### 3. Secret Management
```yaml
secrets:
  config:
    file: ./secrets/config.ini
  ssl_cert:
    file: ./secrets/cert.pem
  ssl_key:
    file: ./secrets/key.pem

services:
  sms-platform:
    secrets:
      - config
      - ssl_cert
      - ssl_key
```

## Backup and Recovery

### Database Backup
```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/backup/sms"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup database
docker exec sms-platform sqlite3 /app/data/sms.db ".backup /tmp/backup.db"
docker cp sms-platform:/tmp/backup.db $BACKUP_DIR/sms_$DATE.db

# Backup configuration
docker cp sms-platform:/app/config/config.ini $BACKUP_DIR/config_$DATE.ini

# Cleanup old backups (keep 30 days)
find $BACKUP_DIR -name "*.db" -mtime +30 -delete
```

### Restore Process
```bash
#!/bin/bash
# restore.sh
BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

# Stop container
docker-compose down

# Restore database
cp $BACKUP_FILE ./data/sms.db

# Start container
docker-compose up -d
```

## Troubleshooting

### Common Issues

1. **Serial Port Access Denied**
```bash
# Check device permissions
ls -la /dev/ttyUSB*

# Add user to dialout group
sudo usermod -a -G dialout $USER

# Or use privileged mode
privileged: true
```

2. **Container Cannot Access Devices**
```bash
# Verify device mapping
docker exec sms-platform ls -la /dev/ttyUSB*

# Check udev rules
sudo udevadm info -a -n /dev/ttyUSB0
```

3. **Port Already in Use**
```bash
# Check what's using the port
sudo lsof -i :8080

# Use different port
- HTTP_PORT=8081
```

### Debug Commands
```bash
# Enter container shell
docker exec -it sms-platform sh

# Check logs
docker-compose logs -f --tail=100 sms-platform

# Monitor resource usage
docker stats sms-platform

# Check device access
docker exec sms-platform ls -la /dev/tty*
```