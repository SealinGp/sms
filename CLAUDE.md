# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an SMS messaging platform built in Go that interfaces with Air780E hardware modules for sending and receiving SMS messages through serial communication on a Raspberry Pi.

## Development Commands

### Build and Run
```bash
# Build the application
go build -o sms

# Production build with optimizations
go build -trimpath -ldflags "-s -w" -o sms

# Initialize database (first time setup)
./sms -i -c config.ini

# Run the application
./sms -c config.ini

# Run with custom config
./sms -c /path/to/config.ini
```

### Dependencies
```bash
# Install/update dependencies
go mod download

# Tidy up dependencies
go mod tidy
```

## Architecture

The application follows a modular architecture with clear separation of concerns:

### Performance Highlights
- **High-Performance HTTP Server**: Uses CloudWeGo Hertz framework for better performance than standard Go HTTP server
- **Unified Serial Handler**: Single implementation reduces code duplication and improves maintainability
- **Asynchronous SMS Processing**: Non-blocking message sending with retry logic

### Core Components

1. **Serial Communication** (`serial/`)
   - **Unified Handler** (`serial_handler.go`): Single implementation for all serial devices
   - **Interface** (`interface.go`): Defines `SerialHandlerInterface` for abstraction
   - **Manager** (`manager.go`): Manages multiple serial devices dynamically
   - **Map** (`map.go`): Synchronization utilities for message acknowledgments
   - Supports unlimited Air780E modules on any serial device path
   - No distinction between USB/GPIO - just different device paths
   - Manages heartbeat signals, message queuing, and retry logic
   - Implements protocol communication with Air780E modules

2. **Web Server** (`app/`)
   - **Hertz Framework**: High-performance HTTP server using CloudWeGo Hertz
   - **Handler** (`handler_hertz.go`): API endpoints and web UI handlers
   - **Session Management** (`session_hertz.go`): Gorilla sessions integration with Hertz
   - **Global** (`global.go`): Server initialization and route registration
   - REST API endpoints for sending SMS and generating keys
   - Web UI for administration with template rendering
   - Authentication using username/password or API key

3. **Database** (`db/`)
   - SQLite database for SMS history
   - GORM ORM for database operations
   - Message persistence and retrieval

4. **Models** (`model/`)
   - SMS message structures
   - Acknowledgment handling
   - Protocol message definitions

5. **Configuration** (`config/`)
   - INI-based configuration system
   - Loads settings from config.ini

### API Endpoints

- `GET /random_key?range={value}&length={value}` - Generate random keys
- `POST /send_sms?key={key}&sender={sender}&phone={phone}&message={message}` - Send SMS
- Web interface requires authentication via session

### Key Configuration

The application uses `config.ini` for all configuration:
- Serial port settings (baud rate, timeouts)
- Server settings (port, HTTPS)
- Database path
- Authentication credentials
- Logging configuration

### Hardware Integration

- **Universal Serial Support**: Works with any serial device path (/dev/ttyUSB*, /dev/ttyAMA*, etc.)
- **Connection Types**: USB, GPIO, or any other serial interface - no code distinction needed
- Implements heartbeat mechanism for connection monitoring
- Air780E modules run Lua code:
  - `air780e/main_simplified.lua`: Simplified version that only forwards SMS to Raspberry Pi
  - `air780e/main.lua`: Original version with built-in logic (legacy)
- Supports unlimited Air780E modules (configured via `serial-device-*` sections in config.ini)

## Important Modifications Before Deployment

### Hardcoded Phone Numbers
The following files contain hardcoded phone numbers that MUST be changed:

1. **`air780e/main.lua`** - Replace `12345678900` with your daily-use phone number
   - This phone receives startup notifications and can send control commands

2. **`serial/serial_cn.go`** - Update `selfPhoneCN = "12345678900"`
   - Used for authentication and special command handling

### Special SMS Commands
The system responds to specific SMS commands:
- `help` - Returns system help information
- `hello` / `你好` - Basic response test
- `reboot` - Reboot the Air780E module (from authorized number)
- `status` / `cstatus` - Get module status
- Home Assistant integration commands (if configured):
  - `ha.help` - HA command help
  - `ha.op.reboot` - Reboot router via HA

### Configuration Notes
- Edit `config.ini` before running:
  - Configure `[serial-device-*]` sections for each Air780E module
  - Set `device_path` to the appropriate serial device (/dev/ttyUSB0, /dev/ttyAMA4, etc.)
  - Update `self_phone` for each device (your daily-use phone number)
  - Set appropriate `region` for each device (cn, us, etc.)
  - Update database path in `[database]` section
  - Set log file path in `[log]` section
  - Change default credentials in `[security]` section

## Production Deployment

### Systemd Service
Create `/etc/systemd/system/sms.service`:
```ini
[Unit]
Description=SMS Platform
After=network.target

[Service]
Type=simple
User=root
ExecStart=/path/to/sms -c /path/to/config.ini
WorkingDirectory=/path/to/data
Restart=always
RestartSec=15

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
systemctl enable sms
systemctl start sms
```

## Project Documentation

Full documentation available at: https://blog.akvicor.com/posts/project/sms/