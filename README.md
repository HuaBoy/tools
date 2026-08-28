# AI 起爆一体化工具集（AI Blasting Toolset）

> 一套系统管全部，AI 替代人工，数据驱动决策。

## 📌 项目简介

**AI 起爆一体化工具集** 是一站式智能化业务平台，面向工业起爆领域，集成 **AI 智能诊断、知识库、数据追溯、海外业务管理、通用工具集** 等能力，帮助企业摆脱多系统割裂、经验依赖、效率低下的现状。

平台基于 **Vue 3 前端 + Go 后端 + PostgreSQL** 全自研构建，**All-in-One 部署**，无需依赖任何第三方云服务，数据完全自主可控。

---

## 📢 最新更新（2026-08-28）

- **Python RAG 知识库服务**：基于 LlamaIndex + pgvector 的语义检索引擎，支持文档分块、向量化、语义搜索
- **AI 固件版本查询**：AI 助手新增设备固件版本查询能力，询问"查看起爆器固件版本"即可获取结构化数据
- **Ollama 本地部署**：AI 助手默认使用本地 Ollama 模型（deepseek-r1:7b + bge-m3 向量模型），数据零泄密
- **一键启动脚本**：`start-all.bat` 一键启动全部服务（Ollama → Python RAG → Go 后端 → Vue 前端）
- **开机自启**：`autostart-services.bat` 静默模式开机自启动，适合服务器部署
- **README 更新**：新增 Python RAG 详细说明、故障排查、监控健康检查、开发指南

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
| 智能知识库（RAG） | 基于 LlamaIndex + pgvector 的语义检索，支持文档分块、向量化、问答式检索 |
| 版本手册 | 产品手册多版本管理，全文搜索 |
| 版本履历 | 各版本迭代记录 |
| 产线履历 | 生产批次历史记录 |
| 录音管理 | 现场录音采集与归档 |

**知识库架构详解：**
```
用户提问
    ↓
Go 后端 (8080) → 调用 Python RAG (8001)
    ↓
LlamaIndex SentenceSplitter 分块 (512 token/块, 80 token 重叠)
    ↓
Ollama bge-m3 向量化 (本地, 零延迟)
    ↓
pgvector 余弦相似度检索 (PostgreSQL + pgvector 扩展)
    ↓
返回 Top-K 结果 → Go 后端构建 RAG 上下文 → LLM 生成回答
```

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

### 5. Python RAG 知识库服务详解

> 基于 LlamaIndex + pgvector 的本地向量检索引擎，为 AI 智能助手提供语义级文档检索能力。

#### 5.1 架构总览

```
┌──────────────────────────────────────────────────────────────────┐
│                        用户在浏览器提问                            │
└──────────────────────────┬───────────────────────────────────────┘
                           ↓
┌──────────────────────────────────────────────────────────────────┐
│              Go 后端 (gin, 端口 8080)                             │
│  ├── 接收用户问题                                                 │
│  ├── 调用 Python RAG 构建上下文 (/api/rag/context)                 │
│  ├── 将上下文 + 问题拼接为 Prompt                                   │
│  └── 调用 Ollama LLM (deepseek-r1:7b) 生成最终回答                  │
└──────┬──────────────────────────────┬────────────────────────────┘
       │                              │
       ↓                              ↓
┌────────────────────┐    ┌──────────────────────────────────────┐
│ Python RAG (8001)  │    │ Ollama LLM (11434)                   │
│ LlamaIndex 引擎     │    │ deepseek-r1:7b (文本生成)              │
│ ├── 文档分块        │    └──────────────────────────────────────┘
│ ├── 向量化 (bge-m3) │
│ ├── pgvector 检索   │
│ └── 上下文构建       │
└────────┬───────────┘
         ↓
┌──────────────────────────────────────────────────────────────────┐
│ PostgreSQL + pgvector (端口 5432)                                 │
│ ├── knowledge_documents      — 知识库文档                         │
│ ├── knowledge_chunks         — 文档分块 (embedding 向量)          │
│ ├── knowledge_faqs           — FAQ 条目 (embedding 向量)          │
│ └── knowledge_categories     — 知识分类                           │
└──────────────────────────────────────────────────────────────────┘
```

#### 5.2 服务组件

| 组件 | 技术栈 | 端口 | 说明 |
|------|--------|------|------|
| Python RAG 引擎 | FastAPI + LlamaIndex + psycopg | 8001 | 文档分块、向量化、语义检索 |
| Ollama Embedding | bge-m3 (通过 Ollama) | 11434 | 文本 → 1024 维向量 |
| Ollama LLM | deepseek-r1:7b | 11434 | 根据检索结果生成回答 |
| PostgreSQL + pgvector | pgvector/pgvector:pg16 | 5432 | 向量存储与相似度检索 |

#### 5.3 核心 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查，返回 `{"status":"ok","engine":"llamaindex"}` |
| POST | `/api/rag/index` | 文档索引：分块 → 向量化 → 存入 pgvector |
| POST | `/api/rag/search` | 语义搜索：返回 Top-K 相关文档分块 |
| POST | `/api/rag/context` | 构建 RAG 上下文：搜索 + 格式化为 LLM Prompt 可用的文本 |
| POST | `/api/rag/faq` | FAQ 向量化并存入 |
| DELETE | `/api/rag/document/{id}` | 删除文档分块 |
| DELETE | `/api/rag/faq/{id}` | 删除 FAQ |

#### 5.4 数据库表结构

**knowledge_chunks（核心向量表）：**

| 列名 | 类型 | 说明 |
|------|------|------|
| `id` | SERIAL PRIMARY KEY | 分块 ID |
| `document_id` | INTEGER | 关联文档 ID |
| `chunk_index` | INTEGER | 分块序号（从 0 开始） |
| `content` | TEXT | 分块文本内容 |
| `embedding` | vector(1024) | bge-m3 生成的 1024 维向量 |
| `metadata_json` | JSONB | 元数据（文件名、页码等） |
| `created_at` | TIMESTAMP | 创建时间 |

**knowledge_faqs：**

| 列名 | 类型 | 说明 |
|------|------|------|
| `id` | SERIAL PRIMARY KEY | FAQ ID |
| `question` | TEXT | 问题 |
| `answer` | TEXT | 回答 |
| `embedding` | vector(1024) | 问题+回答的联合向量 |
| `category` | VARCHAR(100) | 分类 |
| `created_at` | TIMESTAMP | 创建时间 |

#### 5.5 向量化流程

1. **分块**：LlamaIndex `SentenceSplitter`，每块 512 token，相邻块重叠 80 token
2. **嵌入**：调用 Ollama `/api/embed` 接口，模型 `bge-m3`，输出 1024 维浮点向量
3. **存储**：`psycopg` 直接写入 `knowledge_chunks.embedding` 列（vector 类型）
4. **检索**：`pgvector` 的余弦距离运算符 `<=>`，返回相似度 ≥ 0.3 的 Top-K 结果

#### 5.6 配置说明（.env）

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `OLLAMA_BASE_URL` | `http://localhost:11434` | Ollama 服务地址 |
| `EMBED_MODEL` | `bge-m3` | 向量化模型 |
| `CHAT_MODEL` | `deepseek-r1:7b` | 对话模型 |
| `DATABASE_URL` | — | PostgreSQL 连接串 |
| `RAG_PORT` | `8001` | RAG 服务端口 |

#### 5.7 Python 依赖

```
fastapi                — Web 框架
uvicorn[standard]      — ASGI 服务器
llama-index-core>=0.12.0 — 文档分块引擎
llama-index-readers-file>=0.4.0 — 文件读取器
psycopg[binary]>=3.2.0 — PostgreSQL 驱动
httpx                  — HTTP 客户端（调用 Ollama）
python-dotenv          — 环境变量加载
openpyxl               — Excel 支持
```

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

## 📊 监控与健康检查

### 一键检查全部服务状态

```powershell
# Windows PowerShell - 检查所有端口
netstat -ano | Select-String ":(8001|8080|5432|11434)\s" | Select-String "LISTENING"
```

### 各服务健康检查端点

| 服务 | 健康检查 URL | 预期响应 |
|------|-------------|---------|
| PostgreSQL | `psql -h 127.0.0.1 -p 5432 -U kbuser -d knowledge -c "SELECT 1"` | `1` |
| Ollama | `curl http://127.0.0.1:11434/api/tags` | JSON 模型列表 |
| Python RAG | `curl http://127.0.0.1:8001/health` | `{"status":"ok","engine":"llamaindex"}` |
| Go 后端 | `curl http://127.0.0.1:8080/health` | JSON 健康信息 |

### 一键验证脚本

```powershell
# check-services.ps1 - 验证全部服务
Write-Host "=== 知识库服务状态 ===" -ForegroundColor Cyan

# PostgreSQL
try {
    $r = python -c "import psycopg; c=psycopg.connect('postgresql://kbuser:kbpass_2026@127.0.0.1:5432/knowledge'); cur=c.cursor(); cur.execute('SELECT count(*) FROM knowledge_documents'); print('documents:', cur.fetchone()[0]); c.close()"
    Write-Host "[OK] PostgreSQL: $r" -ForegroundColor Green
} catch {
    Write-Host "[FAIL] PostgreSQL 未运行" -ForegroundColor Red
}

# Ollama
try {
    $r = (Invoke-WebRequest -Uri "http://127.0.0.1:11434/api/tags" -TimeoutSec 3).Content | ConvertFrom-Json
    Write-Host "[OK] Ollama: $($r.models.Count) 个模型" -ForegroundColor Green
} catch {
    Write-Host "[FAIL] Ollama 未运行" -ForegroundColor Red
}

# Python RAG
try {
    $r = (Invoke-WebRequest -Uri "http://127.0.0.1:8001/health" -TimeoutSec 3).Content
    Write-Host "[OK] Python RAG: $r" -ForegroundColor Green
} catch {
    Write-Host "[FAIL] Python RAG 未运行" -ForegroundColor Red
}

# Go 后端
try {
    $r = (Invoke-WebRequest -Uri "http://127.0.0.1:8080/health" -TimeoutSec 3).Content
    Write-Host "[OK] Go 后端: $r" -ForegroundColor Green
} catch {
    Write-Host "[FAIL] Go 后端未运行" -ForegroundColor Red
}
```

### 日志文件位置

| 服务 | 日志路径 | 说明 |
|------|---------|------|
| Go 后端 | `backend/logs/app.log` | API 请求日志 |
| Python RAG | `logs/python-rag.log` | RAG 引擎日志 |
| 开机自启 | `logs/go-backend.log` | 后端静默启动日志 |

### 关键进程管理

```powershell
# 查看关键进程
Get-Process ollama, python, server -ErrorAction SilentlyContinue | Select-Object Name, Id, CPU, WorkingSet

# 终止特定进程
taskkill /PID <进程ID> /F

# 查看端口占用
netstat -ano | findstr ":8080"
netstat -ano | findstr ":8001"
netstat -ano | findstr ":5432"
netstat -ano | findstr ":11434"
```

---

## 🔧 故障排查

### 1. Python RAG 服务启动失败

**现象**：端口 8001 无响应

```powershell
# 排查步骤
cd E:\tools20260623\python-rag
python run.py
# 观察终端输出的错误信息
```

**常见原因与解决方案**：

| 原因 | 错误信息 | 解决方案 |
|------|---------|---------|
| 缺少依赖 | `ModuleNotFoundError: No module named 'xxx'` | `pip install -r requirements.txt` |
| PostgreSQL 连不上 | `connection refused` | 确认 PostgreSQL 运行中：`sc query postgresql-x64-16` |
| 数据库不存在 | `database "knowledge" does not exist` | `createdb -U kbuser knowledge` |
| pgvector 未安装 | `type "vector" does not exist` | 安装 pgvector 扩展：`CREATE EXTENSION IF NOT EXISTS vector;` |
| Ollama 未运行 | `Connection refused` | 启动 Ollama：`ollama serve` |
| 端口被占用 | `Address already in use` | `netstat -ano | findstr ":8001"` 找到占用进程并终止 |

### 2. Ollama 模型不可用

**现象**：AI 助手回复空白或报错

```powershell
# 检查模型列表
ollama list
# 确认输出包含 deepseek-r1:7b 和 bge-m3

# 如果模型不在列表中，检查 OLLAMA_MODELS 路径
echo $env:OLLAMA_MODELS
# 应输出: E:\ollama_models\models

# 重新拉取模型
ollama pull deepseek-r1:7b
ollama pull bge-m3
```

**Windows 环境变量设置**（永久生效）：

```powershell
# 以管理员身份运行 PowerShell
[System.Environment]::SetEnvironmentVariable("OLLAMA_MODELS", "E:\ollama_models\models", "Machine")
# 设置后重启终端/Ollama 服务
```

### 3. Go 后端无法连接数据库

**现象**：`/health` 返回错误

```powershell
# 检查 .env 配置
cat E:\tools20260623\backend\.env | findstr DB_DSN
# 应输出: DB_DSN=postgres://kbuser:kbpass_2026@127.0.0.1:5432/knowledge?sslmode=disable

# 测试数据库连接
psql -h 127.0.0.1 -p 5432 -U kbuser -d knowledge -c "\dt"
# 应列出所有表
```

### 4. AI 助手回复质量差

**排查**：

1. **确认知识库有数据**：
```sql
psql -U kbuser -d knowledge -c "SELECT count(*) FROM knowledge_documents;"
psql -U kbuser -d knowledge -c "SELECT count(*) FROM knowledge_chunks;"
```

2. **确认向量检索有效**：
```sql
-- 检查是否有 embedding 数据
psql -U kbuser -d knowledge -c "SELECT count(*) FROM knowledge_chunks WHERE embedding IS NOT NULL;"
```

3. **确认 Ollama 模型正常**：
```bash
curl http://127.0.0.1:11434/api/embed -d '{"model":"bge-m3","input":"测试文本"}'
# 应返回 embedding 向量
```

### 5. 前端显示异常

| 现象 | 原因 | 解决方案 |
|------|------|---------|
| 页面空白 | 后端未启动 | 启动 Go 后端 |
| 登录失败 | admin 密码被重置 | 使用 `admin / 800228` 登录 |
| AI 标签灰色 | Ollama 未配置 | 三方账号授权页配置 Ollama |
| 接口 502 | Caddy 未配置 | 检查 Caddyfile 反向代理规则 |

### 6. 服务器部署问题

```bash
# SSH 到服务器
ssh root@server

# 检查服务状态
systemctl status go-server
docker ps | grep ai-kb

# 查看后端日志
tail -f /opt/go-server/logs/app.log

# 重启服务
systemctl restart go-server
```

---

## 💻 开发指南

### 项目结构

```
E:\tools20260623/
├── frontend/                 # Vue3 前端
│   ├── src/
│   │   ├── views/            # 页面组件
│   │   │   ├── Home.vue              # 首页仪表盘（含 AI 助手）
│   │   │   ├── AiAssistant.vue       # 独立 AI 助手页
│   │   │   ├── knowledge/            # 知识库管理
│   │   │   ├── overseas-shipping/    # 海外发货
│   │   │   ├── operation-manual/     # 操作手册
│   │   │   ├── app-store/            # 应用商店
│   │   │   └── system/               # 系统管理
│   │   ├── services/
│   │   │   └── aiService.js          # AI 服务（支持 Ollama/DeepSeek/混元）
│   │   ├── components/layout/        # 布局组件（含侧边栏 AI 助手）
│   │   └── router/                   # 路由配置
│   ├── vite.config.js                # Vite 配置（含代理规则）
│   └── package.json
│
├── backend/                  # Go 后端
│   ├── cmd/server/main.go            # 入口
│   ├── internal/
│   │   ├── handler/          # HTTP 处理器
│   │   ├── model/            # 数据模型
│   │   ├── repository/       # 数据库操作
│   │   ├── service/          # 业务逻辑
│   │   ├── middleware/       # 中间件（JWT、CORS 等）
│   │   └── migration/       # 数据库迁移
│   ├── .env                          # 环境变量
│   └── server.exe                    # 编译后的可执行文件
│
├── python-rag/               # Python RAG 知识库服务
│   ├── main.py               # FastAPI 入口
│   ├── rag_engine.py         # LlamaIndex 引擎
│   ├── config.py             # 配置管理
│   ├── run.py                # 启动脚本
│   ├── .env                  # 环境变量
│   └── requirements.txt      # Python 依赖
│
├── knowledge/                # 知识库文件存储
├── logs/                     # 日志文件
├── data/                     # 运行时数据
│
├── start-all.bat             # 一键启动全部服务
├── autostart-services.bat    # 开机自启动（静默模式）
├── local-setup.bat           # 本地开发环境初始化
├── docker-compose.yml        # Docker 编排（PostgreSQL + pgvector）
├── deploy.sh                 # 服务器部署脚本
└── README.md                 # 本文档
```

### 开发环境搭建

#### 前端开发

```bash
cd E:\tools20260623\frontend

# 安装依赖
npm install
# 或
pnpm install

# 启动开发服务器（热重载，端口 3000）
npm run dev

# 生产构建
npm run build
# 产物在 dist/ 目录
```

#### 后端开发

```bash
cd E:\tools20260623\backend

# 安装 Go 依赖
go mod download

# 本地运行
go run cmd/server/main.go

# 编译生产版本
go build -o server.exe ./cmd/server

# 交叉编译 Linux 版本
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/server-linux-amd64 ./cmd/server
```

#### Python RAG 开发

```bash
cd E:\tools20260623\python-rag

# 创建虚拟环境（首次）
python -m venv venv
venv\Scripts\activate

# 安装依赖
pip install -r requirements.txt

# 启动开发服务器（热重载）
python run.py
# 或手动：
uvicorn main:app --host 0.0.0.0 --port 8001 --reload
```

### 数据库管理

```sql
-- 连接数据库
psql -h 127.0.0.1 -p 5432 -U kbuser -d knowledge

-- 查看所有表
\dt

-- 查看知识库文档数
SELECT count(*) FROM knowledge_documents;

-- 查看向量分块数
SELECT count(*) FROM knowledge_chunks WHERE embedding IS NOT NULL;

-- 查看 FAQ 数
SELECT count(*) FROM knowledge_faqs;

-- 查看文档分块详情
SELECT dc.id, dc.title, count(ch.id) as chunk_count
FROM knowledge_documents dc
LEFT JOIN knowledge_chunks ch ON ch.document_id = dc.id
GROUP BY dc.id, dc.title;

-- 手动重建索引（向量检索性能优化）
REINDEX TABLE knowledge_chunks;

-- 查看 pgvector 索引状态
SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'knowledge_chunks';
```

### 环境变量速查

#### 前端

| 变量 | 文件 | 说明 |
|------|------|------|
| `VITE_API_BASE_URL` | `.env` | 后端 API 地址 |
| 代理规则 | `vite.config.js` | 开发环境代理配置 |

#### Go 后端

| 变量 | 文件 | 说明 |
|------|------|------|
| `DB_DSN` | `backend/.env` | PostgreSQL 连接串 |
| `JWT_SECRET` | `backend/.env` | JWT 签名密钥 |
| `OLLAMA_URL` | `backend/.env` | Ollama 服务地址 |
| `OLLAMA_CHAT_MODEL` | `backend/.env` | 对话模型名 |
| `OLLAMA_EMBEDDING_MODEL` | `backend/.env` | 向量模型名 |
| `KNOWLEDGE_STORE_DIR` | `backend/.env` | 知识库存储路径 |
| `SERVER_PORT` | `backend/.env` | 服务端口 |

#### Python RAG

| 变量 | 文件 | 说明 |
|------|------|------|
| `OLLAMA_BASE_URL` | `python-rag/.env` | Ollama 服务地址 |
| `EMBED_MODEL` | `python-rag/.env` | 向量化模型名 |
| `CHAT_MODEL` | `python-rag/.env` | 对话模型名 |
| `DATABASE_URL` | `python-rag/.env` | PostgreSQL 连接串 |
| `RAG_PORT` | `python-rag/.env` | RAG 服务端口 |

---

## 📁 项目目录速览

```
E:\tools20260623\
├── README.md                        # 本文档（约 800 行）
├── docker-compose.yml               # Docker 编排（PostgreSQL + pgvector + 后端 + 前端）
├── start-all.bat                    # 一键启动（开发环境）
├── autostart-services.bat           # 开机自启（服务器/生产环境）
├── local-setup.bat                  # 本地环境初始化
├── deploy.sh                        # Linux 服务器部署脚本
├── deploy-knowledge.sh              # 知识库服务部署脚本
│
├── frontend/                        # Vue3 + Vite 前端
│   ├── src/views/                   # 页面（首页/知识库/发货/手册/系统管理）
│   ├── src/services/aiService.js    # AI 服务封装（Ollama/DeepSeek/混元）
│   ├── src/components/layout/       # 布局 + 侧边栏 AI 助手
│   └── vite.config.js              # 代理配置
│
├── backend/                         # Go 后端 (gin)
│   ├── cmd/server/main.go          # 入口
│   ├── internal/                   # 分层架构 (handler/model/service/repository)
│   ├── server.exe                  # 编译后二进制
│   └── .env                        # 环境变量
│
├── python-rag/                      # Python RAG 知识库服务
│   ├── main.py                     # FastAPI 入口
│   ├── rag_engine.py               # LlamaIndex 向量引擎
│   ├── config.py                   # 配置
│   ├── run.py                      # 启动脚本
│   └── requirements.txt            # Python 依赖
│
├── knowledge/                       # 知识库文件（上传的文档）
├── data/                            # 运行时数据
│   ├── postgres/                   # PostgreSQL 数据目录
│   └── knowledge/                  # 知识库存储
└── logs/                            # 日志
    ├── app.log                     # Go 后端日志
    └── python-rag.log              # RAG 服务日志
```

---

## 🤝 代码托管

项目代码自动同步到以下仓库（每日自动提交推送）：

- **GitHub**：`git@github.com:HuaBoy/tools.git`
- **Gitee（主仓库）**：`https://gitee.com/makersoft/app-store.git`
- **Gitee（历史仓库）**：`https://gitee.com/makersoft/ai-tracking-and-toolset.git`

---

## 📝 更新日志

### 2026-08-28

- 新增 Python RAG 知识库服务详细文档（架构图、API、数据库表结构、向量化流程）
- 新增监控与健康检查章节（一键检查脚本、日志文件位置、进程管理）
- 新增故障排查章节（6 类常见问题与解决方案）
- 新增开发指南章节（项目结构、开发环境搭建、数据库管理、环境变量速查）
- 新增项目目录速览（完整目录树）
- 新增更新日志
- 更新服务架构图（含 Python RAG 引擎）

### 2026-08-22

- AI 固件版本查询：AI 助手新增设备固件版本查询能力
- Ollama 本地部署：AI 助手默认使用本地 Ollama 模型，数据零泄密
- 一键部署：支持本地构建后一键上传到服务器
- 新增 Ollama 部署详细说明

### 2026-08-10

- Python RAG 引擎上线：LlamaIndex + pgvector 语义检索
- Go 后端 RAG 集成：自动调用 Python RAG 构建上下文
- 数据库迁移：新增 knowledge_chunks（pgvector 向量表）

### 2026-07-22

- 首版上线：Vue3 前端 + Go 后端 + PostgreSQL
- AI 智能助手（支持 DeepSeek / 混元 云端 API）
- 知识库管理、海外发货、操作手册、应用商店、系统管理

---

## 📄 License

内部项目，未经授权不得对外分发。
