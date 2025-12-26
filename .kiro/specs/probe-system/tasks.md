# Implementation Plan

- [x] 1. 项目初始化与基础架构




  - [ ] 1.1 创建 Go 项目结构
    - 初始化 Go modules (core 和 agent 两个模块)
    - 创建目录结构: cmd/, internal/, pkg/, web/


    - 配置 Makefile 用于构建
    - _Requirements: 12.3_
  - [ ] 1.2 定义核心数据模型和接口
    - 创建 pkg/models/ 下的所有数据结构 (Agent, Metrics, Task, Alert 等)
    - 定义 Service 接口




    - _Requirements: 1.1, 2.1, 3.1, 4.1, 7.1_
  - [x]* 1.3 编写数据模型属性测试


    - **Property 7: Metric Value Bounds**
    - **Validates: Requirements 3.1, 3.2, 3.3**

- [ ] 2. Core 数据层实现
  - [ ] 2.1 实现 SQLite 数据库初始化和迁移
    - 创建数据库 schema
    - 实现自动迁移逻辑

    - _Requirements: 8.1_

  - [ ] 2.2 实现 Agent Repository
    - CRUD 操作
    - 分组和筛选查询
    - _Requirements: 1.1, 1.7, 2.2, 2.4, 2.5_
  - [ ]* 2.3 编写 Agent Repository 属性测试
    - **Property 1: Agent Registration Data Integrity**

    - **Property 5: Group Assignment Consistency**

    - **Property 6: Agent Filter Correctness**
    - **Validates: Requirements 1.1, 1.2, 2.2, 2.4, 2.5**
  - [ ] 2.4 实现 Metrics Repository
    - 存储和查询指标数据
    - 历史数据查询
    - 数据清理
    - _Requirements: 3.5, 8.1, 8.3, 8.4_
  - [x]* 2.5 编写 Metrics Repository 属性测试




    - **Property 15: Data Retention Cleanup**
    - **Property 16: Historical Query Range**
    - **Validates: Requirements 8.3, 8.4**
  - [ ] 2.6 实现 Traffic/BillingCycle Repository
    - 流量累计计算

    - 周期归档和重置
    - _Requirements: 4.1, 4.2, 4.4, 4.5_

  - [ ]* 2.7 编写 Traffic Repository 属性测试
    - **Property 8: Traffic Accumulation Correctness**

    - **Property 9: Billing Cycle Reset**
    - **Validates: Requirements 4.2, 4.3, 4.4**


- [ ] 3. Checkpoint - 确保所有测试通过
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 4. Core 服务层实现
  - [x] 4.1 实现 IP Geolocation Service



    - 集成 MaxMind GeoLite2 或 IP2Location

    - 实现缓存机制
    - _Requirements: 1.3_

  - [ ]* 4.2 编写 Geolocation 属性测试
    - **Property 2: IP Geolocation Consistency**
    - **Validates: Requirements 1.3**
  - [ ] 4.3 实现 Agent Service
    - 注册、状态更新、分组管理



    - _Requirements: 1.1, 1.4, 1.6, 2.1, 2.2, 2.3_

  - [ ] 4.4 实现 Metric Service
    - 指标存储、查询、流量统计
    - _Requirements: 3.5, 4.2, 4.5_
  - [ ] 4.5 实现 Task Service
    - 任务创建、分发、结果收集

    - _Requirements: 5.1, 5.3_

  - [ ] 4.6 实现 Alert Service
    - 规则管理、阈值检测、告警触发
    - _Requirements: 7.1, 7.2, 7.6, 7.7_
  - [ ]* 4.7 编写 Alert Service 属性测试
    - **Property 13: Alert Threshold Triggering**



    - **Property 14: Alert Cooldown Enforcement**

    - **Validates: Requirements 7.2, 7.7**


- [ ] 5. Core 通知层实现
  - [ ] 5.1 实现 Telegram Notifier
    - Bot API 集成
    - 消息格式化
    - _Requirements: 7.3, 7.4_
  - [ ] 5.2 实现 Email Notifier
    - SMTP 发送

    - HTML 邮件模板
    - _Requirements: 7.3, 7.5_




- [-] 6. Checkpoint - 确保所有测试通过

  - Ensure all tests pass, ask the user if questions arise.

- [ ] 7. Core WebSocket 层实现
  - [ ] 7.1 实现 WebSocket Hub 和连接管理
    - 连接池管理

    - 心跳检测
    - 断线处理
    - _Requirements: 1.2, 1.4, 1.6_

  - [ ]* 7.2 编写 WebSocket 连接属性测试
    - **Property 4: Agent Status Transition**
    - **Validates: Requirements 1.4, 1.6**




  - [ ] 7.3 实现消息协议处理
    - 注册、心跳、指标、任务消息解析
    - Token 认证
    - _Requirements: 1.1, 11.2, 11.3_
  - [ ]* 7.4 编写认证属性测试
    - **Property 18: Authentication Enforcement**

    - **Property 19: Token Authentication**
    - **Validates: Requirements 10.1, 11.2, 11.3**

- [ ] 8. Core REST API 实现
  - [ ] 8.1 实现管理后台 API
    - Agent CRUD, 分组管理, 任务管理, 告警配置, 系统设置
    - JWT 认证中间件



    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6_
  - [ ] 8.2 实现公开展示 API
    - 公开 Agent 列表和指标

    - 过滤敏感信息 (IP)
    - _Requirements: 9.1, 9.2, 9.4, 9.5, 9.6_
  - [ ]* 8.3 编写公开 API 属性测试
    - **Property 17: Public API Data Filtering**
    - **Validates: Requirements 9.4, 9.6**
  - [x] 8.4 实现 Dashboard WebSocket 端点




    - 实时推送指标更新
    - _Requirements: 9.3, 9.7_


- [ ] 9. Checkpoint - 确保所有测试通过
  - Ensure all tests pass, ask the user if questions arise.

- [x] 10. Agent 核心实现

  - [ ] 10.1 实现 WebSocket Client
    - 连接建立和认证
    - 指数退避重连
    - 消息收发
    - _Requirements: 1.1, 1.2, 1.5, 11.1, 11.2_


  - [x]* 10.2 编写重连属性测试

    - **Property 3: Reconnection Backoff Pattern**
    - **Validates: Requirements 1.5**

  - [ ] 10.3 实现 Metric Collector
    - CPU、内存、磁盘、网络指标采集
    - 使用 gopsutil 库

    - _Requirements: 3.1, 3.2, 3.3, 3.4, 4.3_
  - [ ] 10.4 实现本地数据缓冲
    - SQLite 离线缓冲
    - 重连后同步

    - _Requirements: 3.5_

- [ ] 11. Agent 任务执行实现
  - [x] 11.1 实现 Ping 任务执行器

    - ICMP ping 实现
    - 结果上报
    - _Requirements: 5.2, 5.4_

  - [ ]* 11.2 编写 Ping 结果属性测试
    - **Property 10: Ping Result Validity**
    - **Validates: Requirements 5.2, 5.4**

  - [ ] 11.3 实现脚本任务执行器
    - 脚本下载和校验
    - 受限执行环境
    - 超时控制





    - _Requirements: 6.2, 6.3, 6.4, 6.5, 6.6_
  - [ ]* 11.4 编写脚本执行属性测试
    - **Property 11: Script Checksum Verification**
    - **Property 12: Script Timeout Enforcement**

    - **Validates: Requirements 6.3, 6.5**

- [x] 12. Agent 主程序和配置

  - [ ] 12.1 实现 Agent 主程序
    - 配置加载
    - 优雅启动和关闭
    - _Requirements: 12.4, 12.5_
  - [ ] 12.2 实现静态编译配置
    - CGO_ENABLED=0 静态链接
    - 多平台交叉编译
    - _Requirements: 12.3_

- [ ] 13. Checkpoint - 确保所有测试通过
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 14. 前端公开展示页面
  - [ ] 14.1 初始化 Vue 3 + Vite 项目
    - 配置 TypeScript, Tailwind CSS
    - 安装依赖 (Pinia, Chart.js, Heroicons)
    - _Requirements: 9.1_
  - [ ] 14.2 实现公开 Dashboard 布局
    - 响应式卡片网格
    - 深色主题
    - _Requirements: 9.2, 9.3_
  - [ ] 14.3 实现 Agent 状态卡片组件
    - 国旗图标显示
    - 实时指标展示 (CPU, 内存, 流量进度条)
    - 状态指示器
    - _Requirements: 9.5, 9.6_
  - [ ] 14.4 实现 WebSocket 实时更新
    - 连接管理
    - 数据更新处理
    - _Requirements: 9.7_

- [ ] 15. 前端管理后台
  - [ ] 15.1 实现登录页面和认证
    - JWT token 管理
    - _Requirements: 10.1_
  - [ ] 15.2 实现 Dashboard 概览页
    - Agent 统计
    - 告警摘要
    - _Requirements: 10.2_
  - [ ] 15.3 实现 Agent 管理页面
    - Agent 列表 (含 IP、位置)
    - 分组管理
    - 备注和标签编辑
    - _Requirements: 1.7, 10.3_
  - [ ] 15.4 实现 Agent 详情页
    - 实时指标图表
    - 流量统计
    - 任务列表
    - _Requirements: 4.5, 10.3_
  - [ ] 15.5 实现任务管理页面
    - 创建 Ping/脚本任务
    - 任务结果查看
    - _Requirements: 10.4_
  - [ ] 15.6 实现告警管理页面
    - 规则配置
    - 告警历史
    - _Requirements: 10.5_
  - [ ] 15.7 实现系统设置页面
    - 数据保留配置
    - 通知渠道配置
    - 公开可见性配置
    - _Requirements: 8.2, 10.6_

- [ ] 16. Core 主程序集成
  - [ ] 16.1 实现 Core 主程序
    - 配置加载
    - 服务初始化和依赖注入
    - HTTP/WebSocket 服务器启动
    - _Requirements: 11.1, 11.5_
  - [ ] 16.2 实现前端静态文件嵌入
    - 使用 go:embed 嵌入前端构建产物
    - _Requirements: 9.1_
  - [ ] 16.3 实现定时任务
    - 数据清理任务
    - 流量周期重置检查
    - _Requirements: 8.3, 4.4_

- [ ] 17. Final Checkpoint - 确保所有测试通过
  - Ensure all tests pass, ask the user if questions arise.
