# Probe System

轻量级实时探针系统，用于监控服务器状态、流量统计和告警通知。

## 架构

- **Core (主控)**: 接收 Agent 数据、提供 Web 界面、发送告警
- **Agent (被控)**: 采集系统指标、执行探测任务、上报数据

## 快速开始

### 构建

```bash
# 构建 Core 和 Agent
make all

# 或分别构建
make build-core
make build-agent

# 跨平台编译 Agent
make build-agent-linux
make build-agent-windows
```

### 运行 Core

```bash
# 使用默认配置
./bin/core

# 指定配置文件
./bin/core -config config.json
```

配置文件示例 (`config.json`):
```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "database": {
    "path": "probe.db"
  },
  "auth": {
    "jwt_secret": "your-secret-key"
  },
  "agent": {
    "token": "your-agent-token"
  }
}
```

### 运行 Agent

```bash
# 命令行参数
./bin/agent -server wss://your-core-server:8080/ws/agent -token your-agent-token

# 或使用配置文件
./bin/agent -config agent.json
```

配置文件示例 (`agent.json`):
```json
{
  "server_url": "wss://your-core-server:8080/ws/agent",
  "token": "your-agent-token",
  "metric_interval": 10
}
```

## 功能

### 系统监控
- CPU、内存、磁盘使用率
- 网络带宽和流量统计
- 自定义流量计费周期

### 探测任务
- Ping 网络连通性检测
- 预定义脚本执行

### 告警通知
- Telegram 机器人通知
- 邮件通知
- 可配置阈值和冷却期

### Web 界面
- 公开展示页面 (无需登录)
- 管理后台 (需要登录)
- 实时数据更新 (WebSocket)

## 默认账户

- 用户名: `admin`
- 密码: `admin`

**请在生产环境中修改默认密码！**

## API 端点

### 公开 API
- `GET /api/public/agents` - 获取公开 Agent 列表

### 管理 API (需要认证)
- `POST /api/auth/login` - 登录
- `GET /api/admin/agents` - Agent 列表
- `GET /api/admin/agents/:id` - Agent 详情
- `GET /api/admin/tasks` - 任务列表
- `GET /api/admin/alerts/rules` - 告警规则
- `GET /api/admin/settings` - 系统设置

### WebSocket
- `/ws/agent` - Agent 连接端点
- `/ws/dashboard` - 前端实时更新

## 技术栈

- **后端**: Go, Gin, SQLite, WebSocket
- **前端**: Vue 3, TypeScript, Tailwind CSS, Vite
- **通信**: WebSocket (WSS)

## 许可证

MIT
