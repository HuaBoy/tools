# Tester Platform Backend

基于 Go + Gin + Supabase 的后端服务，提供用户管理、权限管理、功能管理等 API。

## 技术栈

- **语言**：Go 1.21+
- **Web 框架**：Gin
- **数据库/认证**：Supabase (PostgreSQL + Auth)
- **认证方式**：JWT
- **日志**：Zerolog

## 项目结构

```
backend/
├── cmd/server/              # 入口
│   └── main.go
├── internal/
│   ├── config/             # 配置管理
│   ├── supabase/           # Supabase 客户端
│   ├── middleware/         # 中间件（认证、日志、CORS）
│   ├── handlers/           # HTTP 处理器
│   ├── models/             # 数据模型
│   └── router/             # 路由配置
├── pkg/
│   ├── response/           # 统一响应
│   ├── jwt/                # JWT 工具
│   └── logger/             # 日志工具
├── scripts/                # 部署脚本
│   ├── deploy.sh           # 部署脚本
│   ├── init-server.sh      # 服务器初始化
│   ├── nginx.conf.example  # Nginx 配置示例
│   └── tester-platform.service  # systemd 服务
├── go.mod
├── .env.example
└── README.md
```

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 填入 Supabase URL 和 Key
vim .env
```

### 3. 本地运行

```bash
go run cmd/server/main.go
```

服务将监听 `http://localhost:8080`。

### 4. 编译

```bash
go build -o bin/tester-platform ./cmd/server
./bin/tester-platform
```

## API 接口

### 认证

- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/register` - 注册
- `GET /api/v1/auth/profile` - 获取当前用户（需认证）
- `POST /api/v1/auth/logout` - 登出（需认证）

### 用户管理（需认证 + 权限）

- `GET /api/v1/users` - 用户列表
- `GET /api/v1/users/:id` - 用户详情
- `POST /api/v1/users` - 创建用户（需 user:create）
- `PATCH /api/v1/users/:id` - 更新用户（需 user:edit）
- `DELETE /api/v1/users/:id` - 删除用户（需 user:delete）
- `POST /api/v1/users/:id/reset-password` - 重置密码（需 user:edit）

### 角色管理（需认证 + 权限）

- `GET /api/v1/roles` - 角色列表
- `GET /api/v1/roles/:id` - 角色详情
- `POST /api/v1/roles` - 创建角色（需 role:assign）
- `PATCH /api/v1/roles/:id` - 更新角色（需 role:assign）
- `DELETE /api/v1/roles/:id` - 删除角色（需 role:assign）
- `POST /api/v1/roles/assign` - 分配角色（需 role:assign）

### 功能管理（需认证 + 权限）

- `GET /api/v1/features` - 功能列表
- `POST /api/v1/features` - 创建功能（需 feature:toggle）
- `PATCH /api/v1/features/:id` - 更新功能（需 feature:toggle）
- `DELETE /api/v1/features/:id` - 删除功能（需 feature:toggle）
- `POST /api/v1/features/:id/toggle` - 切换启用状态（需 feature:toggle）

## 部署

### 阿里云 ECS

请参阅 [DEPLOY_ALIYUN.md](DEPLOY_ALIYUN.md) 详细文档。

快速部署：
```bash
# 1. 上传代码到服务器
scp -r ./backend root@your-server:/www/wwwroot/tester-platform/

# 2. SSH 到服务器
ssh root@your-server

# 3. 初始化（首次）
cd /www/wwwroot/tester-platform/backend
./scripts/init-server.sh

# 4. 配置 .env
cp .env.example .env
vim .env

# 5. 部署
./scripts/deploy.sh
```

## 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SUPABASE_URL` | Supabase 项目 URL | - |
| `SUPABASE_ANON_KEY` | Supabase anon 密钥 | - |
| `SUPABASE_SERVICE_KEY` | Supabase service_role 密钥（可选） | - |
| `SERVER_PORT` | 服务端口 | 8080 |
| `SERVER_HOST` | 监听地址 | 0.0.0.0 |
| `SERVER_MODE` | Gin 模式 | debug |
| `JWT_SECRET` | JWT 签名密钥 | - |
| `JWT_EXPIRE_HOURS` | Token 过期时间（小时） | 720 |
| `CORS_ORIGINS` | 允许的来源（逗号分隔） | * |
| `LOG_LEVEL` | 日志级别 | debug |
| `LOG_FILE` | 日志文件路径 | - |

## 权限码

| 权限 | 说明 |
|------|------|
| `user:view` | 查看用户 |
| `user:create` | 创建用户 |
| `user:edit` | 编辑用户 |
| `user:delete` | 删除用户 |
| `role:assign` | 分配角色 |
| `permission:manage` | 管理权限 |
| `feature:toggle` | 切换功能 |
| `feature:view` | 查看功能 |
| `data:query` | 数据查询 |
| `data:export` | 数据导出 |
| `log:view` | 查看日志 |

## 开发

```bash
# 运行测试
go test ./...

# 代码格式化
go fmt ./...

# 代码检查
go vet ./...
```

## License

MIT
