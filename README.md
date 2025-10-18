# fsync

fsync 是一个基于 WebSocket 和 HTTP 的文件同步系统，支持实时文件监控与双向同步，适用于个人或团队在多设备间安全同步文件。

**注意**: 本项目仍在开发阶段，核心同步功能尚未完全实现。

## 功能特性

- 实时文件变更监听与同步
- 用户认证与 JWT 鉴权
- WebSocket 实时通信
- 对象存储集成（MinIO/本地存储）
- HTTPS 安全通信（TLS 1.3）
- 自动证书管理

## 技术架构

- **语言**: Go 1.25.1
- **Web框架**: Gin v1.11.0
- **数据库**: MySQL (通过 GORM)
- **对象存储**: MinIO 客户端
- **日志系统**: zap + lumberjack.v2
- **认证机制**: JWT Token
- **安全通信**: HTTPS/TLS 1.3

## 目录结构

```
.
├── client/           # 客户端代码
│   ├── cmd/          # 客户端主程序入口
│   ├── configs/      # 客户端配置文件
│   ├── models/       # 客户端数据模型
│   ├── services/     # 客户端服务层
│   ├── storage/      # 客户端存储相关
│   └── watcher/      # 文件监控
├── server/           # 服务端代码
│   ├── cmd/          # 服务端主程序入口
│   ├── configs/      # 服务端配置文件
│   ├── global/       # 全局变量
│   ├── handlers/     # HTTP处理函数
│   ├── middleware/   # 中间件
│   ├── minio/        # MinIO相关封装
│   ├── models/       # 数据模型
│   ├── routers/      # 路由配置
│   ├── services/     # 业务逻辑层
│   ├── storage/      # 存储接口和实现
│   ├── utils/        # 工具函数
│   └── websocket/    # WebSocket相关
├── deploy/           # 部署相关
├── docs/             # 文档
├── pkg/              # 共享包
└── scripts/          # 脚本
```

## 快速开始

### 环境要求

- Go 1.25.1 或更高版本
- MySQL 5.7 或更高版本
- MinIO 服务（可选）

### 构建

分别进入 client 和 server 目录进行构建：

```bash
# 构建客户端
cd client/cmd
go build -o ../../bin/fsync-client main.go

# 构建服务端
cd server/cmd
go build -o ../../bin/fsync-server main.go
```

### 运行

1. 首先运行服务端：
```bash
cd server/cmd
go run main.go
```

2. 然后运行客户端：
```bash
cd client/cmd
go run main.go --login  # 登录
# 或
go run main.go --register  # 注册
```

## 安全特性

### HTTPS/TLS 通信

fsync 使用 HTTPS/TLS 1.3 加密所有客户端与服务器之间的通信，确保数据传输的安全性。

### 证书管理

- 服务端自动生成自签名证书并存储在 `server/certs/` 目录中
- 客户端首次连接时自动从服务端下载证书并存储在用户主目录的 `.fsync/` 文件夹中
- 所有后续通信都使用严格的 TLS 验证，防止中间人攻击
- 证书文件不会被提交到 Git 仓库中，确保存储安全

### 认证与授权

- 使用 JWT Token 进行用户身份验证
- 实现双令牌机制（Access Token 和 Refresh Token）
- 管理员用户具有特殊权限，可以创建普通用户

## 客户端使用说明

客户端支持以下命令行参数：

- `--login`: 用户登录
- `--register`: 用户注册
- `--logout`: 用户登出

### 注册流程

1. 运行 `client --register`
2. 输入要注册的用户名和密码
3. 系统会提示输入管理员用户名和密码进行验证
4. 验证通过后完成注册
5. 注册成功后需要手动登录

### 登录流程

1. 运行 `client --login`
2. 输入用户名和密码
3. 登录成功后自动开始文件同步

## 服务端管理

### 管理员用户

系统启动时会检查是否存在管理员用户，如果不存在会提示创建。

### API 接口

详细 API 接口文档请参考 [server/README.md](server/README.md)

## 开发指南

### 代码规范

- 遵循 Go 语言标准编码规范
- 使用 `gofmt` 格式化代码
- 添加必要的注释和文档

### 添加新功能

1. 根据功能性质确定是在 client 还是 server 端实现
2. 在对应目录下创建相关模块
3. 遵循项目现有的代码结构和规范

## 免责声明

本项目仅供学习和参考用途。作者不对以下情况负责：

1. 使用本软件造成的任何数据丢失或损坏
2. 因软件漏洞或缺陷导致的安全问题
3. 由于不当使用或配置引起的任何损失
4. 与第三方服务集成时产生的任何问题
5. 在生产环境中使用本软件带来的风险

**重要提醒：**
- 本项目仍处于开发阶段，功能尚未完善
- 不建议在生产环境中使用
- 使用前请务必备份重要数据
- 如发现安全漏洞，请及时报告

作者保留随时修改或停止本项目的权利，恕不另行通知。

**使用本软件即表示您同意承担所有相关风险和责任。**