# AI 起爆一体化工具集（AI Blasting Toolset）

> 一套系统管全部，AI 替代人工，数据驱动决策。

## 📌 项目简介

**AI 起爆一体化工具集** 是一站式智能化业务平台，面向工业起爆领域，集成 **AI 智能诊断、知识库、数据追溯、海外业务管理、通用工具集** 等能力，帮助企业摆脱多系统割裂、经验依赖、效率低下的现状。

平台基于 **Vue 3 前端 + Go 后端 + PostgreSQL** 全自研构建，**All-in-One 部署**，无需依赖任何第三方云服务，数据完全自主可控。

---

## 📢 最新更新（2026-08-22）

- **AI 固件版本查询**：AI 助手新增设备固件版本查询能力，询问"查看起爆器固件版本"即可获取结构化数据
- **Ollama 本地部署**：AI 助手默认使用本地 Ollama 模型（deepseek-r1:7b），数据零泄密
- **一键部署**：支持本地构建后一键上传到服务器（`scp` + Caddy 自动热加载）
- **README 更新**：新增 Ollama 部署详细说明、更新日志

---

## ✨ 核心功能

### 🧠 智能诊断
| 模块 | 说明 |
|------|------|
| AI 诊断工作台 | 起爆器故障智能分析，AI 辅助定位问题 |
| 日志解密工具 | 起爆器日志一键解密，不再手工处理 |
| 日志分析 | 自动解析日志内容，快速定位异常 |
| 全链路追溯 | 从设备 → 批次 → 数据全程追溯 |

### 📚 AI 知识库
| 模块 | 说明 |
|------|------|
| 智能知识库（RAG） | 基于检索增强生成的问答式知识检索 |
| 版本手册 | 产品手册多版本管理，全文搜索 |
| 版本履历 | 各版本迭代记录 |
| 产线履历 | 生产批次历史记录 |
| 录音管理 | 现场录音采集与归档 |

### 🌐 海外业务
| 模块 | 说明 |
|------|------|
| 海外发货管理 | 发货记录完整 CRUD + 搜索/筛选/分页 |
| 操作手册 | 分类管理、搜索、在线下载 |
| 操作视频 | 视频分类、播放器、播放量统计 |
| 爆破设计软件 | 软件介绍 + 安装包下载 |

### 🔧 通用工具集
| 模块 | 说明 |
|------|------|
| AI 翻译 | 多语言智能翻译 |
| 文档格式转换 | Word/PDF/图片等格式互转 |
| 二维码生成 | 文本/链接转二维码 |
| AI 助手 | 通用 AI 问答 |
| 远程手机控制 | 远程操作手机（屏幕共享） |

### 📊 数据与分析
| 模块 | 说明 |
|------|------|
| 数据查询 | 起爆数据统一查询 |
| 数据追溯 | 批次/生产数据追溯 |
| 销售洞察 | 销售数据分析 |
| 数据大屏 | 可视化大屏展示 |

### ⚙️ 系统管理
| 模块 | 说明 |
|------|------|
| 用户管理 | 用户 CRUD + 密码重置 |
| 权限管理 | 角色/权限分配 |
| 功能管理 | 功能开关控制 |
| 数据库管理 | 后台数据库管理 |

### 📱 应用商店
| 模块 | 说明 |
|------|------|
| 盛景 App 管理 | 移动端应用（APK）发布管理：上传安装包/图标、版本管理、启停控制 |
| 推送管理 | 应用更新/消息推送：支持全部设备或指定设备（SN）定向推送 |
| 下载分发 | 内置应用详情页，扫码/链接下载 APK 安装包 |

### 🔌 平台对接
| 对接 | 说明 |
|------|------|
| 盛景平台（HolyView） | 第三方平台登录/数据对接 |
| 智能制造系统 | 工厂生产数据对接 |
| 硬件管理 | 设备/芯片/固件管理 |

---

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────┐
│                用户浏览器                     │
└──────────────────┬──────────────────────────┘
                   │ HTTPS
┌──────────────────▼──────────────────────────┐
│          Caddy（反向代理 + 静态资源）          │
│   /api/v1/*      → Go 后端 (localhost:8080)  │
│   /api/blade-auth → 盛景平台 (mp.holyview.cn)│
│   /iot-api/*     → 智能制造 (218.90.146.230) │
│   静态文件        → /opt/ai-blasting/dist    │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│            Go 后端 (Gin)                     │
│   认证 / 用户 / 权限 / 业务 CRUD / AI 接口     │
└──────────────────┬──────────────────────────┘
                   │ 直连
┌──────────────────▼──────────────────────────┐
│        PostgreSQL（本地 Supabase）            │
│   sys_user / overseas_shipping / 知识库等    │
└─────────────────────────────────────────────┘
```

**架构特点：**
- ✅ **All-in-One**：一台服务器承载全部服务
- ✅ **零云依赖**：不依赖任何第三方云服务
- ✅ **直连数据库**：绕过 REST 层限制，性能大幅提升
- ✅ **自动建表**：后端启动自动迁移数据库表，零运维
- ✅ **JWT 自主认证**：不依赖外部 Auth 服务

---

## 🛠️ 技术栈

### 前端
| 技术 | 说明 |
|------|------|
| Vue 3 | 渐进式前端框架（Composition API） |
| Vite 5 | 现代化构建工具 |
| Element Plus | UI 组件库 |
| Pinia | 状态管理 |
| Vue Router | 路由管理 |
| ECharts | 数据可视化 |
| Axios | HTTP 客户端 |

### 后端
| 技术 | 说明 |
|------|------|
| Go 1.21+ | 高性能后端语言 |
| Gin | Web 框架 |
| lib/pq | PostgreSQL 驱动（直连） |
| golang-jwt | JWT 认证 |
| Zerolog | 结构化日志 |

### 数据库
| 技术 | 说明 |
|------|------|
| PostgreSQL | 关系型数据库 |
| Supabase（本地） | 自建 Supabase 实例 |

---

## 📂 目录结构

```
ai-blasting-toolset/
├── src/                     # 前端源码（Vue 3）
│   ├── api/                 # API 封装
│   ├── components/          # 公共组件
│   ├── composables/         # 组合式函数
│   ├── router/              # 路由配置
│   ├── services/            # HTTP 服务（axios 封装）
│   ├── stores/              # Pinia 状态管理
│   ├── utils/               # 工具函数
│   └── views/               # 页面视图
│       ├── ai/              # AI 智能诊断
│       ├── admin/           # 系统管理
│       ├── appstore/        # 应用商店
│       ├── data/            # 数据查询/追溯
│       ├── knowledge/       # AI 知识库
│       ├── log/             # 日志工具
│       ├── overseas/        # 海外业务
│       ├── sales/           # 销售洞察
│       ├── stats/           # 数据大屏
│       ├── tools/           # 通用工具集
│       └── trace/           # 全链路追溯
├── backend/                 # 后端源码（Go）
│   ├── cmd/server/          # 程序入口
│   ├── internal/
│   │   ├── config/          # 配置管理
│   │   ├── handlers/        # HTTP 处理器
│   │   ├── middleware/      # 中间件
│   │   ├── models/          # 数据模型
│   │   ├── router/          # 路由配置
│   │   └── supabase/        # Supabase 客户端
│   └── pkg/                 # 公共包（response/jwt/logger）
├── scripts/                 # 部署脚本
├── mobile/                  # 移动端（UniApp 跨端）
│   └── src/
│       ├── pages/           # 页面（工作台/诊断/数据/硬件/知识库/工具/管理）
│       ├── services/        # API 服务
│       ├── stores/          # 状态管理
│       ├── components/      # 公共组件
│       ├── pages.json       # 路由与 tabBar 配置
│       └── manifest.json    # 应用配置（可编译 APK）
├── Dockerfile               # Docker 构建
├── docker-compose.yml       # Docker 编排
├── vite.config.js           # Vite 配置
├── package.json             # 前端依赖
└── README.md                # 本文档
```

---

## 📱 移动端（UniApp）

移动端基于 **UniApp（Vue 3）** 开发，一套代码可编译为 Android APK / iOS App / 小程序 / H5。

```bash
cd mobile
npm install

# H5 开发调试
npm run dev:h5

# 构建 Android APK（需 HBuilderX 或 CLI 环境）
npm run build:app
```

### 移动端页面

| 页面 | 说明 |
|------|------|
| AI 工作台（首页） | 功能总览入口 |
| 登录 | 账号密码登录（对接后端 JWT） |
| 智能诊断 | 起爆器故障诊断 |
| AI 起爆数据 | 数据查询 |
| 硬件管理 | 设备列表 |
| 芯片管理 | 芯片管理 |
| 固件升级 | 固件版本管理 |
| AI 知识库 | 知识检索 |
| 通用工具 | 工具集入口 |
| 系统管理 | 个人中心/系统设置 |

底部 TabBar：**AI工作台 / 诊断 / 硬件 / 工具 / 我的**

---

## 🚀 快速开始

### 环境要求
- Node.js 18+
- Go 1.21+
- PostgreSQL 12+（或 Supabase）

### 1. 前端启动

```bash
# 安装依赖
npm install

# 本地开发（默认端口 3000）
npm run dev

# 生产构建
npm run build
```

构建产物在 `dist/` 目录。

### 2. 后端启动

```bash
cd backend

# 安装依赖
go mod download

# 配置环境变量
cp .env.example .env
# 编辑 .env，填入 SUPABASE_URL / DB_DSN / JWT_SECRET

# 本地运行（默认端口 8080）
go run cmd/server/main.go

# 编译 Linux 二进制
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/server-linux-amd64 ./cmd/server
```

### 3. 默认账号

| 账号 | 密码 | 说明 |
|------|------|------|
| `admin` | `800228` | 系统管理员（启动时自动重置密码） |

> 后端启动时会自动重置 admin 密码并创建所需数据库表，开箱即用。

### 4. AI 智能助手 — 本地 Ollama 部署（推荐）

> **零泄密**：所有数据在内网流转，不经过任何外部 API，适合对数据安全要求极高的工业场景。

#### 4.1 架构说明

```
浏览器(前端) ──→ Ollama 本地服务(11434) ──→ deepseek-r1:7b 模型
                      ↑
              Docker 容器 (ai-kb-ollama)
```

- **模型文件存储**：`E:/tools20260623/ollama_data/` 目录（项目根目录下），Docker 持久化，重启不丢失
- **API 接口**：与 OpenAI 完全兼容（`/v1/chat/completions`），前端无需任何适配
- **模型大小**：`deepseek-r1:7b` 约 4.7GB，首次拉取需一定时间，后续启动秒级恢复

#### 4.2 启动步骤

```bash
# 1. 进入项目根目录
cd E:\tools20260623

# 2. 一键启动 Ollama 容器（含模型下载服务）
docker-compose up -d ollama

# 3. 首次拉取模型（约 4.7GB，需联网；后续无需重复）
docker exec ai-kb-ollama ollama pull deepseek-r1:7b

# 4. 确认模型已就绪
docker exec ai-kb-ollama ollama list
# 应输出: deepseek-r1:7b  xxx  4.7 GB
```

#### 4.3 前端配置

1. 启动前端：`pnpm dev` → 访问 `http://localhost:9000`
2. 登录后进入 **系统管理 → 三方账号授权** 页面
3. 找到 **🤖 Ollama 本地模型（推荐 · 数据零泄密）** 绿色卡片
4. 默认地址 `http://localhost:11434`，一般无需修改
5. 点击 **🔌 检测连接**，通过后自动获取模型列表
6. 从下拉列表选择 `deepseek-r1:7b`
7. 点击 **💾 保存 Ollama 配置**

> 配置完成后，AI 标签会从灰色变为绿色，并显示 **"Ollama 本地"** 徽章，表示当前使用的是本地模型。

#### 4.4 验证是否生效

在首页（仪表盘）的 AI 助手对话框中发送任意消息（如"你好"）：

- ✅ 正常：收到模型回复，且页面右上角 AI 标签显示 "Ollama 本地"
- ❌ 失败：提示连接错误 → 检查 Ollama 容器是否运行、端口是否可达

#### 4.5 常见问题

| 问题 | 解决方案 |
|------|---------|
| `检测连接` 失败 | 执行 `docker ps \| grep ollama` 确认容器运行中；检查 11434 端口未被占用 |
| 模型回复很慢 | 首次加载模型需 10-30 秒（取决于机器配置），后续会缓存加速 |
| 磁盘空间不足 | `deepseek-r1:7b` 占 4.7GB；可改用 `deepseek-r1:1.5b`（约 1.1GB，精度略低） |
| 想用云端 API | 在同一页面切换到 DeepSeek 或混元配置即可，配置互不影响 |
| 模型拉取超时 | 检查网络代理设置；Docker 容器内可能需要配置镜像加速器 |
| 启动后端/前端一体化工具 | 已在 `docker-compose.yml` 中配置好 ollama 服务，和后端、数据库、Redis 一键启动 |

#### 4.6 与其他 Provider 切换

系统支持多 Provider 无缝切换，配置互不干扰：

| Provider | 数据去向 | 需要 API Key | 推荐场景 |
|----------|---------|-------------|---------|
| **Ollama 本地** | 本机 Docker 容器内 | ❌ 不需要 | 生产环境、涉密场景（推荐） |
| DeepSeek 云端 | DeepSeek 服务器 | ✅ 需要 | 快速体验、模型能力更强时 |
| 混元 云端 | 腾讯云服务器 | ✅ 需要 | 腾讯云生态用户 |

切换方法：在 **系统管理 → 三方账号授权** 页面修改对应配置即可，AI 标签实时反映当前 Provider。

---

## 📡 API 概览

> 基础路径：`/api/v1`，除登录/健康检查外均需 `Authorization: Bearer <token>`。

### 认证
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/login` | 登录（username + password） |
| POST | `/auth/register` | 注册 |
| GET | `/auth/profile` | 当前用户信息 |
| POST | `/auth/logout` | 登出 |

### 用户管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/users` | 用户列表 |
| POST | `/users` | 创建用户 |
| PATCH | `/users/:id` | 更新用户 |
| DELETE | `/users/:id` | 删除用户 |
| POST | `/users/:id/reset-password` | 重置密码 |

### 角色/权限
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/roles` | 角色列表 |
| POST | `/roles` | 创建角色 |
| PATCH | `/roles/:id` | 更新角色 |
| DELETE | `/roles/:id` | 删除角色 |
| POST | `/roles/assign` | 分配角色 |

### 海外发货
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/overseas-shipping` | 列表（分页/搜索/筛选） |
| GET | `/overseas-shipping/:id` | 详情 |
| POST | `/overseas-shipping` | 新增 |
| PATCH | `/overseas-shipping/:id` | 更新 |
| DELETE | `/overseas-shipping/:id` | 删除 |

### 操作手册 / 视频
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/operation-manuals` | 手册列表 |
| POST | `/operation-manuals` | 新增手册 |
| PATCH | `/operation-manuals/:id` | 更新手册 |
| DELETE | `/operation-manuals/:id` | 删除手册 |
| GET | `/operation-videos` | 视频列表 |
| POST | `/operation-videos` | 新增视频 |
| PATCH | `/operation-videos/:id` | 更新视频 |
| DELETE | `/operation-videos/:id` | 删除视频 |

### 应用商店（App Store）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/apps` | 应用列表（支持 name 模糊搜索） |
| GET | `/apps/:id` | 应用详情 |
| POST | `/apps` | 新增应用 |
| PUT | `/apps/:id` | 更新应用 |
| DELETE | `/apps/:id` | 删除应用 |
| POST | `/apps/upload` | 上传 APK/图标（multipart） |
| GET | `/pushes` | 推送任务列表 |
| POST | `/pushes` | 创建推送任务（all / specified） |
| POST | `/pushes/:id/send` | 下发推送 |
| DELETE | `/pushes/:id` | 删除推送任务 |
| GET | `/devices/:sn/updates` | 设备按 SN 拉取可见版本 |

### 系统
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/config` | 配置信息 |
| GET | `/features` | 功能列表 |

---

## 🗄️ 数据库

后端启动时自动执行迁移，无需手动建表：

| 表名 | 说明 |
|------|------|
| `sys_user` | 系统用户（遗留表，登录认证用） |
| `user_profiles` | 用户资料 |
| `overseas_shipping` | 海外发货记录 |
| `operation_manuals` | 操作手册 |
| `operation_videos` | 操作视频 |
| `roles` / `user_roles` | 角色管理 |
| `features` | 功能管理 |

---

## 📦 部署

### 服务器部署

```bash
# 1. 上传前端
scp -r ./dist root@server:/opt/ai-blasting/dist

# 2. 上传后端二进制
scp backend/bin/server-linux-amd64 root@server:/opt/go-server/bin/

# 3. SSH 到服务器
ssh root@server

# 4. 重启后端
systemctl restart go-server

# 5. 配置 Caddy 反向代理
# 见 Caddyfile.frontend
```

### 反向代理（Caddy）

```caddyfile
:80 {
    root * /opt/ai-blasting/dist
    encode gzip

    # 后端 API
    handle /api/v1/* {
        reverse_proxy localhost:8080
    }

    # 盛景平台 blade-auth（第三方）
    handle /api/blade-auth/* {
        reverse_proxy https://mp.holyview.cn:9443
    }

    # 智能制造
    handle_path /iot-api/* {
        reverse_proxy http://218.90.146.230:20001
    }

    # SPA 路由回退
    handle {
        @spa not file
        rewrite @spa /index.html
        file_server
    }
}
```

---

## ⚙️ 配置说明

### 后端环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SUPABASE_URL` | Supabase 项目 URL | - |
| `SUPABASE_ANON_KEY` | Supabase anon 密钥 | - |
| `SUPABASE_SERVICE_KEY` | service_role 密钥 | - |
| `DB_DSN` | PostgreSQL 连接串 | - |
| `JWT_SECRET` | JWT 签名密钥 | - |
| `JWT_EXPIRE_HOURS` | Token 过期（小时） | 720 |
| `SERVER_PORT` | 服务端口 | 8080 |
| `CORS_ORIGINS` | 允许来源 | * |

### 前端代理（vite.config.js）

| 路径 | 目标 | 说明 |
|------|------|------|
| `/api/v1` | 本后端 | 业务接口 |
| `/api/blade-auth` | 盛景平台 | 第三方登录 |
| `/smart-api` | 智能制造 | 工厂数据 |
| `/iot-api` | 智能制造 | 设备数据 |
| `/big-screen` | 盛景平台 | 大屏数据 |

---

## 🔐 权限体系

| 权限码 | 说明 |
|--------|------|
| `user:view` / `user:create` / `user:edit` / `user:delete` | 用户管理 |
| `role:assign` | 角色分配 |
| `permission:manage` | 权限管理 |
| `feature:toggle` | 功能开关 |
| `data:query` / `data:export` | 数据查询/导出 |
| `log:view` | 日志查看 |

---

## 🤝 代码托管

项目代码自动同步到以下仓库（每日自动提交推送）：

- **GitHub**：`git@github.com:HuaBoy/tools.git`
- **Gitee（主仓库）**：`https://gitee.com/makersoft/app-store.git`
- **Gitee（历史仓库）**：`https://gitee.com/makersoft/ai-tracking-and-toolset.git`

---

## 📄 License

内部项目，未经授权不得对外分发。
