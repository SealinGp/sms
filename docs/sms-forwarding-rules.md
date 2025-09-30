# SMS Forwarding Rules System

## Overview

基于规则的短信处理和转发系统，支持多种条件匹配和动作执行，实现智能短信分发。

## Rule Engine Architecture

### Rule Definition
```go
type ForwardingRule struct {
    ID          int       `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null"`
    Description string    `json:"description"`
    SerialPort  string    `json:"serial_port"`  // 绑定的串口名称
    Priority    int       `json:"priority"`     // 规则优先级 (数字越小优先级越高)
    Enabled     bool      `json:"enabled"`      // 规则启用状态
    Conditions  []RuleCondition `json:"conditions" gorm:"foreignKey:RuleID"`
    Actions     []RuleAction    `json:"actions" gorm:"foreignKey:RuleID"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type RuleCondition struct {
    ID        int    `json:"id" gorm:"primaryKey"`
    RuleID    int    `json:"rule_id"`
    Type      string `json:"type"`      // title, content, sender, time_range
    Operator  string `json:"operator"`  // contains, equals, regex, between
    Value     string `json:"value"`     // 匹配值
    Logic     string `json:"logic"`     // AND, OR (与下一个条件的逻辑关系)
}

type RuleAction struct {
    ID       int                    `json:"id" gorm:"primaryKey"`
    RuleID   int                    `json:"rule_id"`
    Type     string                 `json:"type"`     // feishu, dingtalk, forward_sms, email, webhook
    Config   map[string]interface{} `json:"config"`   // 动作配置参数
    Order    int                    `json:"order"`    // 执行顺序
    Enabled  bool                   `json:"enabled"`  // 动作启用状态
}
```

### Rule Processing Engine
```go
type RuleEngine interface {
    ProcessMessage(message *IncomingMessage) error
    AddRule(rule *ForwardingRule) error
    UpdateRule(ruleID int, rule *ForwardingRule) error
    DeleteRule(ruleID int) error
    GetRules(serialPort string) ([]*ForwardingRule, error)
    TestRule(ruleID int, testMessage *IncomingMessage) (*TestResult, error)
}

type IncomingMessage struct {
    SerialPort  string    `json:"serial_port"`
    Sender      string    `json:"sender"`
    Title       string    `json:"title"`
    Content     string    `json:"content"`
    ReceivedAt  time.Time `json:"received_at"`
    MessageID   string    `json:"message_id"`
}
```

## Condition Types

### 1. Title Matching (标题匹配)
```json
{
    "type": "title",
    "operator": "contains",
    "value": "验证码"
}
```

### 2. Content Matching (内容匹配)
```json
{
    "type": "content",
    "operator": "regex",
    "value": "\\d{6}"
}
```

### 3. Sender Matching (发送方匹配)
```json
{
    "type": "sender",
    "operator": "equals",
    "value": "+8610086"
}
```

### 4. Time Range (时间范围)
```json
{
    "type": "time_range",
    "operator": "between",
    "value": "09:00-18:00"
}
```

### 5. Composite Conditions (复合条件)
```json
{
    "conditions": [
        {
            "type": "title",
            "operator": "contains",
            "value": "告警",
            "logic": "AND"
        },
        {
            "type": "sender",
            "operator": "contains",
            "value": "监控",
            "logic": "OR"
        },
        {
            "type": "content",
            "operator": "regex",
            "value": "(严重|紧急)"
        }
    ]
}
```

## Action Types

### 1. 飞书通知 (Feishu Notification)
```go
type FeishuConfig struct {
    WebhookURL string `json:"webhook_url"`
    Secret     string `json:"secret"`
    Template   string `json:"template"`
}

// Template variables: {{.Sender}}, {{.Title}}, {{.Content}}, {{.ReceivedAt}}
// Example template:
{
    "type": "feishu",
    "config": {
        "webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
        "secret": "your_secret",
        "template": "🚨 SMS Alert\n发送方: {{.Sender}}\n标题: {{.Title}}\n内容: {{.Content}}\n时间: {{.ReceivedAt}}"
    }
}
```

### 2. 钉钉通知 (DingTalk Notification)
```go
type DingTalkConfig struct {
    WebhookURL string `json:"webhook_url"`
    Secret     string `json:"secret"`
    AtMobiles  []string `json:"at_mobiles"`
    IsAtAll    bool   `json:"is_at_all"`
    Template   string `json:"template"`
}

{
    "type": "dingtalk",
    "config": {
        "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
        "secret": "your_secret",
        "at_mobiles": ["13800138000"],
        "template": "📱 短信转发\n{{.Title}}\n{{.Content}}"
    }
}
```

### 3. 短信转发 (SMS Forward)
```go
type SMSForwardConfig struct {
    TargetPort   string   `json:"target_port"`   // 目标串口
    TargetPhones []string `json:"target_phones"` // 目标手机号
    Template     string   `json:"template"`      // 转发模板
}

{
    "type": "forward_sms",
    "config": {
        "target_port": "Air780E-2",
        "target_phones": ["+8613800138000", "+8613800138001"],
        "template": "转发消息: {{.Content}}"
    }
}
```

### 4. 邮件通知 (Email Notification)
```go
type EmailConfig struct {
    SMTPHost     string   `json:"smtp_host"`
    SMTPPort     int      `json:"smtp_port"`
    Username     string   `json:"username"`
    Password     string   `json:"password"`
    FromEmail    string   `json:"from_email"`
    ToEmails     []string `json:"to_emails"`
    Subject      string   `json:"subject"`
    Template     string   `json:"template"`
    UseHTML      bool     `json:"use_html"`
}
```

### 5. Webhook 通知
```go
type WebhookConfig struct {
    URL         string            `json:"url"`
    Method      string            `json:"method"`      // GET, POST
    Headers     map[string]string `json:"headers"`
    Template    string            `json:"template"`    // JSON template
    Timeout     int               `json:"timeout"`     // seconds
    RetryCount  int               `json:"retry_count"`
}

{
    "type": "webhook",
    "config": {
        "url": "https://api.example.com/sms/notify",
        "method": "POST",
        "headers": {
            "Content-Type": "application/json",
            "Authorization": "Bearer token"
        },
        "template": "{\"message\": \"{{.Content}}\", \"from\": \"{{.Sender}}\", \"timestamp\": \"{{.ReceivedAt}}\"}"
    }
}
```

## API Endpoints

### 1. Rule Management
```
# Create rule
POST /api/rules
Request: {
    "name": "告警短信转发",
    "description": "将包含告警关键词的短信转发到飞书",
    "serial_port": "Air780E-1",
    "priority": 1,
    "conditions": [...],
    "actions": [...]
}

# List rules
GET /api/rules?serial_port=Air780E-1&enabled=true

# Update rule
PUT /api/rules/{id}

# Delete rule
DELETE /api/rules/{id}

# Enable/Disable rule
POST /api/rules/{id}/toggle
```

### 2. Rule Testing
```
POST /api/rules/{id}/test
Request: {
    "sender": "+8610086",
    "title": "系统告警",
    "content": "服务器CPU使用率超过90%",
    "received_at": "2024-01-01T12:00:00Z"
}

Response: {
    "code": 0,
    "msg": "success",
    "data": {
        "matched": true,
        "executed_actions": [
            {
                "type": "feishu",
                "status": "success",
                "message": "通知发送成功"
            }
        ]
    }
}
```

### 3. Rule Execution Logs
```
GET /api/rules/logs?rule_id=1&start_time=2024-01-01&end_time=2024-01-02

Response: {
    "code": 0,
    "msg": "success",
    "data": [
        {
            "id": 1,
            "rule_id": 1,
            "message_id": "msg_123",
            "matched": true,
            "executed_actions": 2,
            "success_actions": 2,
            "execution_time": "2024-01-01T12:00:00Z",
            "details": [...]
        }
    ]
}
```

## Implementation

### 1. Rule Engine Core
```go
type DefaultRuleEngine struct {
    rules    []*ForwardingRule
    actions  map[string]ActionExecutor
    mu       sync.RWMutex
    logger   *glog.Logger
}

func (e *DefaultRuleEngine) ProcessMessage(msg *IncomingMessage) error {
    e.mu.RLock()
    defer e.mu.RUnlock()

    // Sort rules by priority
    sort.Slice(e.rules, func(i, j int) bool {
        return e.rules[i].Priority < e.rules[j].Priority
    })

    for _, rule := range e.rules {
        if !rule.Enabled || rule.SerialPort != msg.SerialPort {
            continue
        }

        if e.evaluateConditions(rule.Conditions, msg) {
            e.executeActions(rule.Actions, msg)
            // Optional: break here if you want only first match
        }
    }

    return nil
}

func (e *DefaultRuleEngine) evaluateConditions(conditions []RuleCondition, msg *IncomingMessage) bool {
    if len(conditions) == 0 {
        return true
    }

    result := true
    logic := "AND"

    for i, condition := range conditions {
        conditionResult := e.evaluateCondition(&condition, msg)

        if i == 0 {
            result = conditionResult
        } else {
            if logic == "AND" {
                result = result && conditionResult
            } else {
                result = result || conditionResult
            }
        }

        logic = condition.Logic
    }

    return result
}
```

### 2. Action Executors
```go
type ActionExecutor interface {
    Execute(config map[string]interface{}, msg *IncomingMessage) error
    Validate(config map[string]interface{}) error
}

type FeishuExecutor struct{}

func (f *FeishuExecutor) Execute(config map[string]interface{}, msg *IncomingMessage) error {
    webhookURL := config["webhook_url"].(string)
    template := config["template"].(string)

    content := f.renderTemplate(template, msg)

    payload := map[string]interface{}{
        "msg_type": "text",
        "content": map[string]string{
            "text": content,
        },
    }

    return f.sendToFeishu(webhookURL, payload)
}
```

## Database Schema

```sql
-- Forwarding rules
CREATE TABLE forwarding_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    serial_port VARCHAR(50),
    priority INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Rule conditions
CREATE TABLE rule_conditions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_id INTEGER NOT NULL,
    type VARCHAR(20) NOT NULL,
    operator VARCHAR(20) NOT NULL,
    value TEXT NOT NULL,
    logic VARCHAR(10) DEFAULT 'AND',
    FOREIGN KEY (rule_id) REFERENCES forwarding_rules(id) ON DELETE CASCADE
);

-- Rule actions
CREATE TABLE rule_actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_id INTEGER NOT NULL,
    type VARCHAR(20) NOT NULL,
    config TEXT NOT NULL, -- JSON string
    order_num INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT true,
    FOREIGN KEY (rule_id) REFERENCES forwarding_rules(id) ON DELETE CASCADE
);

-- Rule execution logs
CREATE TABLE rule_execution_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_id INTEGER NOT NULL,
    message_id VARCHAR(100),
    sender VARCHAR(50),
    title TEXT,
    content TEXT,
    matched BOOLEAN NOT NULL,
    executed_actions INTEGER DEFAULT 0,
    success_actions INTEGER DEFAULT 0,
    error_message TEXT,
    execution_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (rule_id) REFERENCES forwarding_rules(id)
);
```

## Frontend Interface

### 1. Rule List Page
- 规则列表表格 (名称、描述、串口、状态、优先级)
- 启用/禁用开关
- 编辑/删除/测试操作
- 创建新规则按钮

### 2. Rule Editor
- 基本信息 (名称、描述、串口、优先级)
- 条件构建器 (可视化条件编辑)
- 动作配置器 (支持多种动作类型)
- 规则测试功能

### 3. Execution Logs
- 执行历史记录
- 成功/失败统计
- 错误详情查看
- 日志导出功能