package config

type Model struct {
	BrandName string        `ini:"brand_name"`
	Prod      bool          `ini:"prod"`
	SerialCN  SerialModel   `ini:"serial-cn"`
	SerialUS  SerialModel   `ini:"serial-us"`
	Server    ServerModel   `ini:"server"`
	Session   SessionModel  `ini:"session"`
	Database  DatabaseModel `ini:"database"`
	Log       LogModel      `ini:"log"`
	Security  SecurityModel `ini:"security"`
}

type SerialModel struct {
	Name                    string `ini:"name"`
	Baud                    int    `ini:"baud"`
	SendQueueSize           int    `ini:"send_queue_size"`
	HeartbeatSendInterval   uint   `ini:"heartbeat_send_interval"`
	HeartbeatReceiveTimeout uint   `ini:"heartbeat_receive_timeout"`
}

type ServerModel struct {
	HTTPAddr    string `ini:"http_addr"`
	HTTPPort    int    `ini:"http_port"`
	EnableHTTPS bool   `ini:"enable_https"`
	SSLCert     string `ini:"ssl_cert"`
	SSLKey      string `ini:"ssl_key"`
}

type SessionModel struct {
	Domain string `ini:"domain"`
	Path   string `ini:"path"`
	Name   string `ini:"name"`
	MaxAge int    `ini:"max_age"`
}

type DatabaseModel struct {
	Path string `ini:"path"`
}

type LogModel struct {
	LogToFile bool   `ini:"log_to_file"`
	FilePath  string `ini:"file_path"`
}

type SecurityModel struct {
	Username  string `ini:"username"`
	Password  string `ini:"password"`
	AccessKey string `ini:"access_key"`
}
