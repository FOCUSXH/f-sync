```text
backend/
├── cmd/
│   └── server/
│       └── main.go                     # 程序入口：初始化 Gin、配置、数据库、日志等
├── configs/
│   ├── config.yaml                     # 配置文件（DB、JWT、Ollama、Qwen API Key 等）
│   └── config.go                       # 使用 viper 加载配置
├── pkg/
│   └── logger/
│       └── logger.go                   # 基于 zap 的日志封装
├── internal/
│   ├── middleware/                     # Gin 自定义中间件
│   │   ├── jwt_auth.go                 # JWT 验证中间件
│   │   ├── cors.go                     # 跨域处理
│   │   └── request_id.go               # 请求 ID 追踪
│   ├── auth/                           # 认证服务（独立于 MVC 模块）
│   │   ├── jwt.go                      # Token 生成/解析（支持双 Token）
│   │   └── service.go                  # 登录/刷新 Token 等逻辑
│   ├── llm/                            # 大模型抽象层（被 chat service 调用）
│   │   ├── provider.go                 # 统一接口
│   │   ├── ollama.go
│   │   ├── qwen.go
│   │   └── factory.go
│   └── modules/                        # 【核心】按功能划分的 MVC 模块
│       ├── user/                       # 用户模块（注册/登录/信息）
│       │   ├── model/                  # Model：GORM 结构体 + 数据库操作
│       │   │   └── user.go
│       │   ├── service/                # Service：业务逻辑（bcrypt、唯一性检查等）
│       │   │   └── user_service.go
│       │   └── handler/                # Handler（即 Controller）：Gin HTTP 处理函数
│       │       └── user_handler.go
│       └── chat/                       # 对话模块
│           ├── model/
│           │   └── message.go          # 对话记录模型（可选）
│           ├── service/
│           │   └── chat_service.go     # 调用 llm.Provider 生成回复
│           └── handler/
│               └── chat_handler.go     # 处理 /chat/completions 等请求
├── routers/
│   └── router.go                       # 【Gin 核心】注册所有路由，应用中间件
├── migrations/                         # （可选）数据库迁移脚本
├── go.mod
├── go.sum
└── README.md
```