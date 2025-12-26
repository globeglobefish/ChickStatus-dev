# Probe System

轻量级实时探针系统，用于监控服务器状态、流量统计和告警通知。

## 架构

- **Core (主控)**: 接收 Agent 数据、提供 Web 界面、发送告警
- **Agent (被控)**: 采集系统指标、执行探测任务、上报数据

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- GCC (用于 SQLite CGO)

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

# 构建前端
cd core/web/frontend
npm install
npm run build
```

### Docker 构建

```bash
# 构建 Core 镜像 (在 core/ 目录下)
docker build -t probe-core ./core

# 构建 Agent 镜像 (在 agent/ 目录下)
docker build -t probe-agent ./agent

# 或使用 Makefile
make docker-build

# 使用 docker-compose 启动
docker-compose up -d
```

### 运行 Core

```bash
# 复制示例配置
cp core/config.example.json config.json

# 编辑配置文件
# 修改 jwt_secret 和 agent.token

# 运行
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
./bin/agent -server ws://your-core-server:8080/ws/agent -token your-agent-token

# 或使用配置文件
cp agent/agent.example.json agent.json
./bin/agent -config agent.json
```

配置文件示例 (`agent.json`):
```json
{
  "server_url": "ws://your-core-server:8080/ws/agent",
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
- 预定义脚本执行（安全校验）

### 告警通知
- Telegram 机器人通知
- 邮件通知
- 可配置阈值和冷却期

### Web 界面
- 公开展示页面 (无需登录，显示国旗，不暴露 IP)
- 管理后台 (需要登录，显示完整信息)
- 实时数据更新 (WebSocket)

### Agent 管理
- 分组管理
- 自定义名称和备注
- 标签系统
- 公开可见性控制

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
- `PATCH /api/admin/agents/:id/remark` - 更新备注
- `PATCH /api/admin/agents/:id/group` - 分配分组
- `PATCH /api/admin/agents/:id/visibility` - 设置公开可见性
- `GET /api/admin/groups` - 分组列表
- `GET /api/admin/tasks` - 任务列表
- `GET /api/admin/scripts` - 脚本列表
- `GET /api/admin/alerts/rules` - 告警规则
- `GET /api/admin/settings` - 系统设置

### WebSocket
- `/ws/agent` - Agent 连接端点
- `/ws/dashboard` - 前端实时更新

## 技术栈

- **后端**: Go, Gin, SQLite, WebSocket
- **前端**: Vue 3, TypeScript, Tailwind CSS, Vite
- **通信**: WebSocket (支持 WS/WSS)

## 目录结构

```
.
├── agent/                 # Agent 模块
│   ├── cmd/agent/        # 入口
│   ├── internal/         # 内部实现
│   │   ├── buffer/       # 离线数据缓冲
│   │   ├── collector/    # 指标采集
│   │   ├── executor/     # 任务执行
│   │   └── ws/           # WebSocket 客户端
│   └── pkg/protocol/     # 通信协议
├── core/                  # Core 模块
│   ├── cmd/core/         # 入口
│   ├── internal/         # 内部实现
│   │   ├── config/       # 配置
│   │   ├── handler/      # HTTP 处理器
│   │   ├── models/       # 数据模型
│   │   ├── notify/       # 通知器
│   │   ├── repository/   # 数据访问
│   │   ├── service/      # 业务逻辑
│   │   └── ws/           # WebSocket 服务
│   ├── pkg/protocol/     # 通信协议
│   └── web/frontend/     # 前端代码
├── Makefile
└── README.md
```

## 许可证

MIT
