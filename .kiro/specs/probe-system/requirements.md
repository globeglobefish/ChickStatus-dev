# Requirements Document

## Introduction

本文档定义了一个轻量级实时探针系统的需求。该系统采用 Core（主控）和 Agent（被控）架构，通过 WebSocket 实现实时双向通信。Core 负责任务分发、数据接收、流量统计协同计算和 Web 展示；Agent 以极轻量化方式运行在被监控节点上，执行探测任务并上报数据。

系统具有双重用途：内部监控管理和对外公开展示。前台展示页面需要精美的视觉设计，适合向外部用户展示服务器状态。系统使用 Go 语言实现。

## Glossary

- **Core**: 主控模块，负责管理 Agent、分发任务、接收数据、协同计算流量、提供 Web 界面
- **Agent**: 被控模块，部署在被监控节点上，执行探测任务并上报数据
- **Probe Task**: 探测任务，由 Core 分发给 Agent 执行的具体监控指令
- **Heartbeat**: 心跳，Agent 定期向 Core 发送的存活信号
- **Metric**: 指标数据，Agent 采集并上报的监控数据
- **Billing Cycle**: 流量计费周期，用户自定义的流量统计时间段
- **Predefined Script**: 预定义脚本，经过审核并存储在 Core 上的可执行脚本
- **Agent Group**: Agent 分组，用于组织和分类管理多个 Agent
- **Public Dashboard**: 公开展示页面，面向外部用户展示服务器状态的精美界面
- **Admin Panel**: 管理后台，供管理员进行系统配置和管理的界面

## Requirements

### Requirement 1: Agent 注册与连接管理

**User Story:** As a 系统管理员, I want to 管理所有 Agent 的注册和连接状态, so that 我能清楚了解整个监控网络的健康状况。

#### Acceptance Criteria

1. WHEN an Agent starts THEN the Agent SHALL establish a WebSocket connection to Core and register with metadata (hostname, IP, OS, version)
2. WHEN Core receives Agent connection THEN the Core SHALL extract and store the Agent's public IP address from the connection
3. WHEN Core stores Agent IP THEN the Core SHALL resolve the IP to geographic location (country, region, city) using IP geolocation
4. WHILE WebSocket connection is active THEN the Agent SHALL send heartbeat messages at 30-second intervals
5. IF WebSocket connection is lost THEN the Agent SHALL attempt reconnection with exponential backoff (starting at 1 second, max 60 seconds)
6. WHEN Core detects connection loss THEN the Core SHALL mark the Agent as offline within 60 seconds
7. WHEN a user views the Agent list in admin panel THEN the Core SHALL display IP address, location, online/offline status, and last seen time

### Requirement 2: Agent 分组与备注管理

**User Story:** As a 系统管理员, I want to 对 Agent 进行分组和添加备注, so that 我能更好地组织和识别大量服务器。

#### Acceptance Criteria

1. WHEN a user creates an Agent group THEN the Core SHALL store the group with name, description, and display order
2. WHEN a user assigns an Agent to a group THEN the Core SHALL update the Agent's group association
3. WHEN a user adds a remark to an Agent THEN the Core SHALL store custom name, description, and tags for the Agent
4. WHEN displaying Agents THEN the Core SHALL show Agents organized by groups with custom remarks visible
5. WHEN a user filters Agents THEN the Core SHALL support filtering by group, tags, status, and custom fields

### Requirement 3: 系统指标采集

**User Story:** As a 系统管理员, I want to 实时监控各节点的系统资源使用情况, so that 我能及时发现资源瓶颈。

#### Acceptance Criteria

1. WHEN Agent is running THEN the Agent SHALL collect CPU usage percentage at configurable intervals (default 10 seconds)
2. WHEN Agent is running THEN the Agent SHALL collect memory usage (total, used, available, percentage)
3. WHEN Agent is running THEN the Agent SHALL collect disk usage for all mounted partitions (total, used, available, percentage)
4. WHEN Agent is running THEN the Agent SHALL collect network bandwidth usage (bytes sent/received per second per interface)
5. WHEN metrics are collected THEN the Agent SHALL report metrics to Core via WebSocket immediately

### Requirement 4: 流量统计与计费周期

**User Story:** As a 系统管理员, I want to 按自定义周期统计各节点的流量使用量, so that 我能监控流量消耗并控制成本。

#### Acceptance Criteria

1. WHEN a user configures a billing cycle for an Agent THEN the Core SHALL store the cycle start date and duration (monthly, weekly, or custom days)
2. WHILE within a billing cycle THEN the Core SHALL accumulate traffic data from Agent reports
3. WHEN Agent reports network traffic THEN the Agent SHALL include cumulative bytes sent and received since last report
4. WHEN a billing cycle ends THEN the Core SHALL archive the cycle total and reset counters for the new cycle
5. WHEN a user views traffic statistics THEN the Core SHALL display current cycle usage, limit, and percentage with visual progress bar

### Requirement 5: 网络探测任务

**User Story:** As a 系统管理员, I want to 让 Agent 探测到指定目标的网络连通性, so that 我能监控各节点到不同区域的网络状况。

#### Acceptance Criteria

1. WHEN a user creates a ping task THEN the Core SHALL validate target address and push task to specified Agent(s)
2. WHEN Agent receives a ping task THEN the Agent SHALL execute ping to target and report results (latency, packet loss, status)
3. WHEN ping task is configured as recurring THEN the Agent SHALL execute at specified intervals until task is cancelled
4. WHEN ping fails THEN the Agent SHALL report failure status with error details to Core
5. WHEN a user views ping results THEN the Core SHALL display latency history and availability percentage

### Requirement 6: 预定义脚本执行

**User Story:** As a 系统管理员, I want to 在 Agent 上执行预先审核的脚本, so that 我能进行自定义监控而不引入安全风险。

#### Acceptance Criteria

1. WHEN a user uploads a script THEN the Core SHALL store the script with a unique identifier and metadata
2. WHEN a user assigns a script task to an Agent THEN the Core SHALL send only the script identifier and parameters
3. WHEN Agent receives a script task THEN the Agent SHALL download the script from Core, verify checksum, and execute
4. WHEN script execution completes THEN the Agent SHALL report exit code, stdout, and stderr to Core
5. IF script execution exceeds timeout (default 60 seconds) THEN the Agent SHALL terminate the process and report timeout error
6. WHEN executing scripts THEN the Agent SHALL run scripts with restricted permissions (non-root user, limited filesystem access)

### Requirement 7: 告警通知

**User Story:** As a 系统管理员, I want to 在资源异常或流量接近阈值时收到通知, so that 我能及时响应问题。

#### Acceptance Criteria

1. WHEN a user configures alert rules THEN the Core SHALL store threshold conditions for CPU, memory, disk, and traffic
2. WHEN metric value exceeds configured threshold THEN the Core SHALL trigger an alert
3. WHEN an alert is triggered THEN the Core SHALL send notification via configured channels (Telegram bot and/or email)
4. WHEN configuring Telegram notification THEN the Core SHALL allow setting bot token and chat ID
5. WHEN configuring email notification THEN the Core SHALL allow setting SMTP server, sender, and recipient addresses
6. WHEN an alert condition is resolved THEN the Core SHALL send recovery notification
7. WHEN multiple alerts occur for same condition THEN the Core SHALL implement cooldown period to prevent notification spam

### Requirement 8: 数据持久化

**User Story:** As a 系统管理员, I want to 持久化存储监控数据并配置保留策略, so that 我能查看历史数据并控制存储空间。

#### Acceptance Criteria

1. WHEN Core receives metric data THEN the Core SHALL persist data to database immediately
2. WHEN a user configures data retention period THEN the Core SHALL store the setting (default 7 days)
3. WHEN data exceeds retention period THEN the Core SHALL delete expired data during daily cleanup
4. WHEN a user queries historical data THEN the Core SHALL retrieve data within the specified time range
5. WHEN database storage approaches capacity THEN the Core SHALL alert administrators

### Requirement 9: 公开展示页面

**User Story:** As a 系统管理员, I want to 提供精美的公开展示页面, so that 我能向外部用户展示服务器状态和服务质量。

#### Acceptance Criteria

1. WHEN an external user visits the public dashboard THEN the Core SHALL display a visually appealing status page without requiring login
2. WHEN displaying public dashboard THEN the Core SHALL show Agent status cards with real-time metrics in modern UI design
3. WHEN displaying metrics THEN the Core SHALL use animated charts and visual indicators (gauges, progress bars, status dots)
4. WHEN a user configures public visibility THEN the Core SHALL allow selecting which Agents and metrics are publicly visible
5. WHEN displaying Agent cards THEN the Core SHALL show custom name, group, geographic location (country/region flag), and key metrics (uptime, CPU, memory, traffic)
6. WHEN displaying geographic location THEN the Core SHALL show country/region flag icon based on IP geolocation (IP address itself SHALL NOT be displayed on public page)
7. WHEN public page loads THEN the Core SHALL establish WebSocket connection for real-time updates without refresh

### Requirement 10: 管理后台

**User Story:** As a 系统管理员, I want to 通过管理后台配置和管理整个系统, so that 我能完全控制监控系统的各项设置。

#### Acceptance Criteria

1. WHEN a user accesses admin panel THEN the Core SHALL require authentication (username/password)
2. WHEN authenticated THEN the Core SHALL display comprehensive dashboard with all Agents and system health
3. WHEN managing Agents THEN the Core SHALL provide interface to group, tag, configure, and remove Agents
4. WHEN managing tasks THEN the Core SHALL provide interface to create, assign, monitor, and cancel probe tasks
5. WHEN managing alerts THEN the Core SHALL provide interface to configure thresholds and notification channels
6. WHEN managing settings THEN the Core SHALL provide interface to configure data retention, billing cycles, and public visibility

### Requirement 11: 安全通信

**User Story:** As a 系统管理员, I want to 确保 Core 和 Agent 之间的通信安全, so that 监控数据不被窃取或篡改。

#### Acceptance Criteria

1. WHEN Agent connects to Core THEN the connection SHALL use WSS (WebSocket Secure) with TLS encryption
2. WHEN Agent registers THEN the Agent SHALL authenticate using a pre-shared token configured during deployment
3. IF authentication fails THEN the Core SHALL reject the connection and log the attempt
4. WHEN transmitting sensitive data THEN the system SHALL ensure data integrity through TLS
5. WHEN a user accesses admin panel THEN the Core SHALL require authentication and use HTTPS

### Requirement 12: Agent 轻量化

**User Story:** As a 系统管理员, I want Agent 占用极少的系统资源, so that 它不会影响被监控节点的正常运行。

#### Acceptance Criteria

1. WHILE Agent is running THEN the Agent SHALL consume less than 30MB of memory under normal operation
2. WHILE Agent is idle THEN the Agent SHALL consume less than 1% CPU on average
3. WHEN Agent binary is built THEN the build process SHALL produce a single statically-linked executable
4. WHEN Agent starts THEN the Agent SHALL complete initialization and connect to Core within 5 seconds
5. WHEN Agent configuration changes THEN the Agent SHALL apply changes without restart where possible

