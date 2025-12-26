# Design Document

## Overview

本设计文档描述轻量级实时探针系统的技术架构和实现方案。系统采用 Go 语言开发，Core 和 Agent 通过 WebSocket 实时通信。

**设计原则：**
- **健壮性**：完善的错误处理、重试机制、优雅降级
- **安全性**：TLS 加密、Token 认证、输入验证、权限隔离
- **高数据密度前端**：现代化 UI，单屏展示更多信息，实时更新

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Core (主控)                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  WebSocket   │  │   REST API   │  │  Web Server  │          │
│  │   Handler    │  │   Handler    │  │  (Frontend)  │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
│         │                 │                 │                   │
│  ┌──────┴─────────────────┴─────────────────┴───────┐          │
│  │                  Service Layer                    │          │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │          │
│  │  │ Agent   │ │ Metric  │ │  Task   │ │ Alert   │ │          │
│  │  │ Service │ │ Service │ │ Service │ │ Service │ │          │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘ │          │
│  └──────────────────────┬───────────────────────────┘          │
│                         │                                       │
│  ┌──────────────────────┴───────────────────────────┐          │
│  │                  Data Layer                       │          │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────────────────┐ │          │
│  │  │ SQLite  │ │  Cache  │ │  IP Geolocation DB  │ │          │
│  │  │   DB    │ │ (Memory)│ │    (MaxMind/IP2)    │ │          │
│  │  └─────────┘ └─────────┘ └─────────────────────┘ │          │
│  └──────────────────────────────────────────────────┘          │
│                                                                 │
│  ┌──────────────────────────────────────────────────┐          │
│  │              Notification Layer                   │          │
│  │  ┌─────────────┐  ┌─────────────┐                │          │
│  │  │  Telegram   │  │    Email    │                │          │
│  │  │    Bot      │  │    SMTP     │                │          │
│  │  └─────────────┘  └─────────────┘                │          │
│  └──────────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────┘
                              │
                    WSS (TLS Encrypted)
                              │
┌─────────────────────────────┴───────────────────────────────────┐
│                        Agent (被控)                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  WebSocket   │  │   Metric     │  │    Task      │          │
│  │   Client     │  │  Collector   │  │   Executor   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│                                                                 │
│  ┌──────────────────────────────────────────────────┐          │
│  │              Local Buffer (SQLite)                │          │
│  │         (Offline data buffering)                  │          │
│  └──────────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### Core Components

#### 1. WebSocket Handler
负责管理所有 Agent 的 WebSocket 连接。

```go
type WSHandler struct {
    upgrader    websocket.Upgrader
    connections sync.Map // agentID -> *AgentConnection
    hub         *Hub
}

type AgentConnection struct {
    ID          string
    Conn        *websocket.Conn
    Send        chan []byte
    LastSeen    time.Time
    mu          sync.Mutex
}

// 消息类型
const (
    MsgTypeHeartbeat    = "heartbeat"
    MsgTypeRegister     = "register"
    MsgTypeMetrics      = "metrics"
    MsgTypeTaskAssign   = "task_assign"
    MsgTypeTaskResult   = "task_result"
    MsgTypeConfig       = "config"
)
```

#### 2. Agent Service
管理 Agent 注册、分组、状态。

```go
type AgentService interface {
    Register(ctx context.Context, req *RegisterRequest) (*Agent, error)
    UpdateStatus(ctx context.Context, agentID string, status AgentStatus) error
    GetByID(ctx context.Context, agentID string) (*Agent, error)
    List(ctx context.Context, filter *AgentFilter) ([]*Agent, error)
    UpdateRemark(ctx context.Context, agentID string, remark *AgentRemark) error
    AssignGroup(ctx context.Context, agentID, groupID string) error
    Delete(ctx context.Context, agentID string) error
}

type GroupService interface {
    Create(ctx context.Context, group *Group) error
    Update(ctx context.Context, group *Group) error
    Delete(ctx context.Context, groupID string) error
    List(ctx context.Context) ([]*Group, error)
}
```

#### 3. Metric Service
处理指标数据的接收、存储、查询。

```go
type MetricService interface {
    Store(ctx context.Context, agentID string, metrics *Metrics) error
    GetLatest(ctx context.Context, agentID string) (*Metrics, error)
    GetHistory(ctx context.Context, agentID string, from, to time.Time) ([]*Metrics, error)
    GetTrafficStats(ctx context.Context, agentID string, cycleID string) (*TrafficStats, error)
    Cleanup(ctx context.Context, retentionDays int) error
}
```

#### 4. Task Service
管理探测任务的创建、分发、结果收集。

```go
type TaskService interface {
    Create(ctx context.Context, task *Task) error
    Assign(ctx context.Context, taskID, agentID string) error
    Cancel(ctx context.Context, taskID string) error
    GetResults(ctx context.Context, taskID string) ([]*TaskResult, error)
    ListByAgent(ctx context.Context, agentID string) ([]*Task, error)
}
```

#### 5. Alert Service
告警规则管理和通知发送。

```go
type AlertService interface {
    CreateRule(ctx context.Context, rule *AlertRule) error
    UpdateRule(ctx context.Context, rule *AlertRule) error
    DeleteRule(ctx context.Context, ruleID string) error
    CheckAndTrigger(ctx context.Context, agentID string, metrics *Metrics) error
    GetActiveAlerts(ctx context.Context) ([]*Alert, error)
}

type Notifier interface {
    Send(ctx context.Context, alert *Alert) error
}

type TelegramNotifier struct {
    botToken string
    chatID   string
    client   *http.Client
}

type EmailNotifier struct {
    smtpHost string
    smtpPort int
    username string
    password string
    from     string
}
```

#### 6. IP Geolocation Service
IP 地址解析为地理位置。

```go
type GeoService interface {
    Lookup(ip string) (*GeoLocation, error)
}

type GeoLocation struct {
    Country     string
    CountryCode string
    Region      string
    City        string
    Latitude    float64
    Longitude   float64
}
```

### Agent Components

#### 1. WebSocket Client
维护与 Core 的连接，处理重连。

```go
type WSClient struct {
    serverURL   string
    token       string
    conn        *websocket.Conn
    sendChan    chan []byte
    reconnectCh chan struct{}
    mu          sync.Mutex
}

func (c *WSClient) Connect() error
func (c *WSClient) Send(msg *Message) error
func (c *WSClient) reconnectWithBackoff()
```

#### 2. Metric Collector
采集系统指标。

```go
type MetricCollector struct {
    interval time.Duration
    stopCh   chan struct{}
}

func (c *MetricCollector) CollectCPU() (float64, error)
func (c *MetricCollector) CollectMemory() (*MemoryStats, error)
func (c *MetricCollector) CollectDisk() ([]*DiskStats, error)
func (c *MetricCollector) CollectNetwork() (*NetworkStats, error)
```

#### 3. Task Executor
执行探测任务。

```go
type TaskExecutor struct {
    scriptDir   string
    timeout     time.Duration
    restrictUID int
}

func (e *TaskExecutor) ExecutePing(target string, count int) (*PingResult, error)
func (e *TaskExecutor) ExecuteScript(scriptID string, params map[string]string) (*ScriptResult, error)
```

### WebSocket Message Protocol

```go
type Message struct {
    Type      string          `json:"type"`
    ID        string          `json:"id"`
    Timestamp int64           `json:"ts"`
    Payload   json.RawMessage `json:"payload"`
}

// Register Payload
type RegisterPayload struct {
    Hostname string `json:"hostname"`
    OS       string `json:"os"`
    Arch     string `json:"arch"`
    Version  string `json:"version"`
    Token    string `json:"token"`
}

// Metrics Payload
type MetricsPayload struct {
    CPU     float64        `json:"cpu"`
    Memory  MemoryStats    `json:"memory"`
    Disks   []DiskStats    `json:"disks"`
    Network NetworkStats   `json:"network"`
}

// Task Assign Payload
type TaskAssignPayload struct {
    TaskID   string            `json:"task_id"`
    Type     string            `json:"type"` // ping, script
    Target   string            `json:"target,omitempty"`
    ScriptID string            `json:"script_id,omitempty"`
    Params   map[string]string `json:"params,omitempty"`
    Interval int               `json:"interval,omitempty"` // seconds, 0 = one-time
}
```

## Data Models

### Agent

```go
type Agent struct {
    ID          string       `json:"id" db:"id"`
    Hostname    string       `json:"hostname" db:"hostname"`
    IP          string       `json:"ip" db:"ip"`
    OS          string       `json:"os" db:"os"`
    Arch        string       `json:"arch" db:"arch"`
    Version     string       `json:"version" db:"version"`
    Status      AgentStatus  `json:"status" db:"status"`
    GroupID     *string      `json:"group_id" db:"group_id"`
    CustomName  string       `json:"custom_name" db:"custom_name"`
    Description string       `json:"description" db:"description"`
    Tags        []string     `json:"tags" db:"tags"`
    Location    *GeoLocation `json:"location" db:"location"`
    LastSeenAt  time.Time    `json:"last_seen_at" db:"last_seen_at"`
    CreatedAt   time.Time    `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

type AgentStatus string

const (
    AgentStatusOnline  AgentStatus = "online"
    AgentStatusOffline AgentStatus = "offline"
)
```

### Group

```go
type Group struct {
    ID          string    `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Description string    `json:"description" db:"description"`
    DisplayOrder int      `json:"display_order" db:"display_order"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
```

### Metrics

```go
type Metrics struct {
    ID        string       `json:"id" db:"id"`
    AgentID   string       `json:"agent_id" db:"agent_id"`
    CPU       float64      `json:"cpu" db:"cpu"`
    Memory    MemoryStats  `json:"memory" db:"memory"`
    Disks     []DiskStats  `json:"disks" db:"disks"`
    Network   NetworkStats `json:"network" db:"network"`
    Timestamp time.Time    `json:"timestamp" db:"timestamp"`
}

type MemoryStats struct {
    Total     uint64  `json:"total"`
    Used      uint64  `json:"used"`
    Available uint64  `json:"available"`
    Percent   float64 `json:"percent"`
}

type DiskStats struct {
    Path      string  `json:"path"`
    Total     uint64  `json:"total"`
    Used      uint64  `json:"used"`
    Available uint64  `json:"available"`
    Percent   float64 `json:"percent"`
}

type NetworkStats struct {
    BytesSent     uint64 `json:"bytes_sent"`
    BytesRecv     uint64 `json:"bytes_recv"`
    BytesSentRate uint64 `json:"bytes_sent_rate"` // per second
    BytesRecvRate uint64 `json:"bytes_recv_rate"` // per second
}
```

### Traffic Billing

```go
type BillingCycle struct {
    ID        string    `json:"id" db:"id"`
    AgentID   string    `json:"agent_id" db:"agent_id"`
    StartDate time.Time `json:"start_date" db:"start_date"`
    Duration  int       `json:"duration" db:"duration"` // days
    Limit     uint64    `json:"limit" db:"limit"`       // bytes, 0 = unlimited
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type TrafficStats struct {
    CycleID     string    `json:"cycle_id"`
    AgentID     string    `json:"agent_id"`
    BytesSent   uint64    `json:"bytes_sent"`
    BytesRecv   uint64    `json:"bytes_recv"`
    TotalBytes  uint64    `json:"total_bytes"`
    Limit       uint64    `json:"limit"`
    Percent     float64   `json:"percent"`
    CycleStart  time.Time `json:"cycle_start"`
    CycleEnd    time.Time `json:"cycle_end"`
}
```

### Task

```go
type Task struct {
    ID        string            `json:"id" db:"id"`
    Type      TaskType          `json:"type" db:"type"`
    Target    string            `json:"target" db:"target"`
    ScriptID  string            `json:"script_id" db:"script_id"`
    Params    map[string]string `json:"params" db:"params"`
    Interval  int               `json:"interval" db:"interval"` // seconds
    Status    TaskStatus        `json:"status" db:"status"`
    CreatedAt time.Time         `json:"created_at" db:"created_at"`
}

type TaskType string

const (
    TaskTypePing   TaskType = "ping"
    TaskTypeScript TaskType = "script"
)

type TaskResult struct {
    ID        string    `json:"id" db:"id"`
    TaskID    string    `json:"task_id" db:"task_id"`
    AgentID   string    `json:"agent_id" db:"agent_id"`
    Success   bool      `json:"success" db:"success"`
    Output    string    `json:"output" db:"output"`
    Error     string    `json:"error" db:"error"`
    Duration  int64     `json:"duration" db:"duration"` // milliseconds
    Timestamp time.Time `json:"timestamp" db:"timestamp"`
}
```

### Alert

```go
type AlertRule struct {
    ID          string      `json:"id" db:"id"`
    Name        string      `json:"name" db:"name"`
    MetricType  string      `json:"metric_type" db:"metric_type"` // cpu, memory, disk, traffic
    Operator    string      `json:"operator" db:"operator"`       // gt, lt, eq
    Threshold   float64     `json:"threshold" db:"threshold"`
    Duration    int         `json:"duration" db:"duration"`       // seconds, consecutive
    Cooldown    int         `json:"cooldown" db:"cooldown"`       // seconds
    AgentIDs    []string    `json:"agent_ids" db:"agent_ids"`     // empty = all
    Enabled     bool        `json:"enabled" db:"enabled"`
    CreatedAt   time.Time   `json:"created_at" db:"created_at"`
}

type Alert struct {
    ID          string      `json:"id" db:"id"`
    RuleID      string      `json:"rule_id" db:"rule_id"`
    AgentID     string      `json:"agent_id" db:"agent_id"`
    Status      AlertStatus `json:"status" db:"status"`
    Value       float64     `json:"value" db:"value"`
    Message     string      `json:"message" db:"message"`
    TriggeredAt time.Time   `json:"triggered_at" db:"triggered_at"`
    ResolvedAt  *time.Time  `json:"resolved_at" db:"resolved_at"`
}

type AlertStatus string

const (
    AlertStatusFiring   AlertStatus = "firing"
    AlertStatusResolved AlertStatus = "resolved"
)
```

### Settings

```go
type Settings struct {
    DataRetentionDays int    `json:"data_retention_days"`
    TelegramBotToken  string `json:"telegram_bot_token"`
    TelegramChatID    string `json:"telegram_chat_id"`
    SMTPHost          string `json:"smtp_host"`
    SMTPPort          int    `json:"smtp_port"`
    SMTPUsername      string `json:"smtp_username"`
    SMTPPassword      string `json:"smtp_password"`
    SMTPFrom          string `json:"smtp_from"`
    AlertEmailTo      string `json:"alert_email_to"`
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Agent Registration Data Integrity
*For any* Agent registration request with valid token, the Core SHALL store all metadata (hostname, IP, OS, version) and the stored data SHALL be retrievable with identical values.
**Validates: Requirements 1.1, 1.2**

### Property 2: IP Geolocation Consistency
*For any* valid IP address, the geolocation lookup SHALL return a valid GeoLocation with non-empty CountryCode, and repeated lookups for the same IP SHALL return consistent results.
**Validates: Requirements 1.3**

### Property 3: Reconnection Backoff Pattern
*For any* sequence of connection failures, the Agent's reconnection delays SHALL follow exponential backoff pattern where delay(n) = min(2^n seconds, 60 seconds) for n >= 0.
**Validates: Requirements 1.5**

### Property 4: Agent Status Transition
*For any* Agent, when heartbeat timeout (60s) is exceeded, the status SHALL transition from "online" to "offline", and when connection is re-established, status SHALL transition back to "online".
**Validates: Requirements 1.4, 1.6**

### Property 5: Group Assignment Consistency
*For any* Agent assigned to a Group, the Agent's GroupID SHALL match the assigned Group's ID, and listing Agents by Group SHALL include that Agent.
**Validates: Requirements 2.2, 2.4**

### Property 6: Agent Filter Correctness
*For any* filter criteria (group, tags, status), the filtered Agent list SHALL contain only Agents matching ALL specified criteria, and SHALL contain ALL Agents matching those criteria.
**Validates: Requirements 2.5**

### Property 7: Metric Value Bounds
*For any* collected metrics: CPU percentage SHALL be in range [0, 100], Memory percent SHALL equal (Used / Total) * 100, Disk percent SHALL equal (Used / Total) * 100 for each partition.
**Validates: Requirements 3.1, 3.2, 3.3**

### Property 8: Traffic Accumulation Correctness
*For any* billing cycle, the accumulated traffic SHALL equal the sum of all reported traffic increments within that cycle period.
**Validates: Requirements 4.2, 4.3**

### Property 9: Billing Cycle Reset
*For any* billing cycle that ends, the archived total SHALL equal the final accumulated value, and the new cycle SHALL start with zero accumulated traffic.
**Validates: Requirements 4.4**

### Property 10: Ping Result Validity
*For any* ping task execution, the result SHALL contain latency >= 0 (or -1 for failure), packet_loss in range [0, 100], and success status consistent with packet_loss < 100.
**Validates: Requirements 5.2, 5.4**

### Property 11: Script Checksum Verification
*For any* script download, if the computed checksum does not match the expected checksum, the Agent SHALL reject execution and report error.
**Validates: Requirements 6.3**

### Property 12: Script Timeout Enforcement
*For any* script execution exceeding the configured timeout, the process SHALL be terminated and result SHALL indicate timeout error.
**Validates: Requirements 6.5**

### Property 13: Alert Threshold Triggering
*For any* metric value exceeding a configured threshold for the specified duration, an alert SHALL be triggered with status "firing".
**Validates: Requirements 7.2**

### Property 14: Alert Cooldown Enforcement
*For any* alert in "firing" status, no duplicate alert SHALL be created for the same rule and agent within the cooldown period.
**Validates: Requirements 7.7**

### Property 15: Data Retention Cleanup
*For any* metric data older than the configured retention period, the cleanup process SHALL delete that data, and queries SHALL not return deleted data.
**Validates: Requirements 8.3**

### Property 16: Historical Query Range
*For any* historical data query with time range [from, to], the returned data SHALL contain only records with timestamp >= from AND timestamp <= to.
**Validates: Requirements 8.4**

### Property 17: Public API Data Filtering
*For any* public dashboard request, the response SHALL NOT contain IP addresses, and SHALL only contain Agents and metrics marked as publicly visible.
**Validates: Requirements 9.4, 9.6**

### Property 18: Authentication Enforcement
*For any* request to admin endpoints without valid authentication, the Core SHALL return 401 Unauthorized and SHALL NOT return protected data.
**Validates: Requirements 10.1, 11.2, 11.3**

### Property 19: Token Authentication
*For any* Agent connection attempt with invalid token, the Core SHALL reject the connection and log the failed attempt.
**Validates: Requirements 11.2, 11.3**

## Error Handling

### Core Error Handling

| Error Type | Handling Strategy | Recovery Action |
|------------|-------------------|-----------------|
| WebSocket Connection Lost | Detect via heartbeat timeout | Mark agent offline, cleanup connection resources |
| Database Write Failure | Retry with exponential backoff (3 attempts) | Log error, return error to caller |
| IP Geolocation Failure | Use cached result or return unknown | Log warning, continue with empty location |
| Notification Send Failure | Retry 3 times with 5s interval | Log error, mark notification as failed |
| Invalid Message Format | Reject message, log warning | Send error response to agent |
| Authentication Failure | Reject immediately | Log attempt with IP and timestamp |

### Agent Error Handling

| Error Type | Handling Strategy | Recovery Action |
|------------|-------------------|-----------------|
| Connection Lost | Exponential backoff reconnection | Buffer metrics locally (SQLite) |
| Metric Collection Failure | Skip failed metric, continue others | Log error, report partial metrics |
| Script Execution Timeout | Kill process, report timeout | Cleanup temp files |
| Script Download Failure | Retry 3 times | Report failure to Core |
| Local Buffer Full | Drop oldest entries | Log warning |

### Graceful Degradation

1. **Network Instability**: Agent buffers data locally, syncs when connection restored
2. **Core Overload**: Rate limiting on WebSocket messages, queue overflow protection
3. **Database Unavailable**: In-memory cache for recent data, retry writes
4. **Notification Service Down**: Queue notifications, retry with backoff

## Testing Strategy

### Unit Testing

使用 Go 标准 `testing` 包进行单元测试：

- **Service Layer Tests**: 测试各 Service 的业务逻辑
- **Repository Tests**: 使用 SQLite 内存数据库测试数据访问
- **Message Parsing Tests**: 测试 WebSocket 消息序列化/反序列化
- **Validation Tests**: 测试输入验证逻辑

### Property-Based Testing

使用 **rapid** 库 (`pgregory.net/rapid`) 进行属性测试：

```go
import "pgregory.net/rapid"
```

每个属性测试配置运行 **100 次迭代**。

属性测试标注格式：
```go
// **Feature: probe-system, Property {number}: {property_text}**
// **Validates: Requirements X.Y**
```

### Integration Testing

- **WebSocket 连接测试**: 测试完整的连接、认证、消息收发流程
- **端到端数据流测试**: Agent 采集 -> 上报 -> Core 存储 -> API 查询
- **告警流程测试**: 指标超阈值 -> 触发告警 -> 发送通知

### Test Coverage Goals

- Unit Tests: > 80% code coverage
- Property Tests: 覆盖所有 Correctness Properties
- Integration Tests: 覆盖主要用户场景

## Frontend Design

### Design Principles

1. **高数据密度**: 单屏展示尽可能多的信息，减少滚动和点击
2. **实时更新**: WebSocket 推送，无需刷新
3. **现代化 UI**: 使用 Tailwind CSS，深色主题为主
4. **响应式**: 适配桌面和移动设备

### Public Dashboard Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  Logo    System Status: ● All Systems Operational    [Theme]   │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌───────────┐ │
│  │ 🇺🇸 US-East │ │ 🇩🇪 EU-West │ │ 🇯🇵 Asia    │ │ 🇸🇬 SG    │ │
│  │ ● Online   │ │ ● Online   │ │ ● Online   │ │ ○ Offline │ │
│  │ CPU: 23%   │ │ CPU: 45%   │ │ CPU: 12%   │ │ CPU: --   │ │
│  │ MEM: 67%   │ │ MEM: 82%   │ │ MEM: 34%   │ │ MEM: --   │ │
│  │ ▓▓▓▓░░░░░░ │ │ ▓▓▓▓▓▓▓░░░ │ │ ▓▓▓░░░░░░░ │ │ ░░░░░░░░░░│ │
│  │ Traffic:   │ │ Traffic:   │ │ Traffic:   │ │ Traffic:  │ │
│  │ 234GB/1TB  │ │ 567GB/2TB  │ │ 89GB/500GB │ │ --/--     │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └───────────┘ │
│                                                                 │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │ 🇬🇧 UK      │ │ 🇦🇺 AU      │ │ 🇨🇦 CA      │               │
│  │ ...        │ │ ...        │ │ ...        │               │
│  └─────────────┘ └─────────────┘ └─────────────┘               │
└─────────────────────────────────────────────────────────────────┘
```

### Admin Dashboard Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  [≡] Probe System    Dashboard | Agents | Tasks | Alerts | ⚙️  │
├──────────┬──────────────────────────────────────────────────────┤
│ Overview │  Total: 12  Online: 10  Offline: 2  Alerts: 3       │
├──────────┼──────────────────────────────────────────────────────┤
│ Groups   │  ┌─────────────────────────────────────────────────┐ │
│ ├ US     │  │ Agent Details: us-east-1                        │ │
│ ├ EU     │  │ IP: 192.168.1.100  Location: Virginia, US       │ │
│ └ Asia   │  │ Status: ● Online  Uptime: 15d 3h 22m            │ │
│          │  ├─────────────────────────────────────────────────┤ │
│ Agents   │  │ CPU ████████░░░░░░░░░░ 42%                      │ │
│ ├ us-e1  │  │ MEM ██████████████░░░░ 71%                      │ │
│ ├ us-w1  │  │ DSK ████░░░░░░░░░░░░░░ 23%                      │ │
│ ├ eu-c1  │  │ NET ↑ 12.3 MB/s  ↓ 45.6 MB/s                   │ │
│ └ ...    │  ├─────────────────────────────────────────────────┤ │
│          │  │ Traffic: 234.5 GB / 1 TB (23.4%)                │ │
│          │  │ ▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░ Resets: 5 days  │ │
│          │  └─────────────────────────────────────────────────┘ │
└──────────┴──────────────────────────────────────────────────────┘
```

### Technology Stack

- **Frontend Framework**: Vue 3 + TypeScript
- **UI Library**: Tailwind CSS + Headless UI
- **Charts**: Chart.js or ECharts
- **Icons**: Heroicons + Flag Icons
- **Build Tool**: Vite
- **State Management**: Pinia

### Real-time Updates

```typescript
// WebSocket connection for real-time updates
const ws = new WebSocket('wss://core.example.com/ws/dashboard');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  switch (data.type) {
    case 'metrics':
      updateAgentMetrics(data.agent_id, data.metrics);
      break;
    case 'status':
      updateAgentStatus(data.agent_id, data.status);
      break;
    case 'alert':
      showAlertNotification(data.alert);
      break;
  }
};
```

