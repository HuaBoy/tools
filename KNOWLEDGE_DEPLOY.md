# 内部知识库 + 问题库 + RAG + AI 对话 部署指南

> 内网服务器部署，数据不出内网，纯本地模型推理。

## 一、服务器要求

| 配置项 | 最低要求 | 推荐 |
|--------|---------|------|
| CPU | 4 核 | **8 核** |
| 内存 | 16 GB | **32 GB** |
| 磁盘 | 200 GB SSD | 500 GB SSD |
| 系统 | Ubuntu 22.04 / CentOS 7+ | 同左 |
| Docker | 20.10+ | 最新 |

## 二、系统架构

```
用户浏览器
    │
    ▼
Nginx (frontend:80)
    │
    ▼
Go 后端 (backend:8080)
    ├── PostgreSQL + pgvector (postgres:5432)  ← 文档/向量/问题/对话
    └── Ollama (ollama:11434)                  ← 本地模型推理
          ├── deepseek-r1:7b       对话模型
          └── bge-m3               嵌入模型 (1.2GB, 中文效果最佳, 1024维)
```

## 三、快速部署

### 1. 安装 Docker + Compose 插件

```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com | sh
systemctl enable --now docker

# 安装 compose 插件
apt install docker-compose-plugin -y
```

### 2. 拉取模型（关键！必须先执行）

```bash
# 启动 ollama 容器
docker compose up -d ollama

# 拉取模型（在容器内执行）
docker exec ai-kb-ollama ollama pull deepseek-r1:7b
docker exec ai-kb-ollama ollama pull bge-small-zh-v1.5

# 验证
docker exec ai-kb-ollama ollama list
```

> **注意**：模型约 4.7GB，首次拉取需要较长时间。如服务器无法访问外网，
> 需预先在有网络的机器上下载模型后拷贝到 `./data/ollama/models` 目录。

### 3. 启动全部服务

```bash
cd /path/to/ai-blasting-toolset
docker compose up -d --build
```

### 4. 验证

```bash
# 健康检查
curl http://localhost:8080/health

# 查看容器状态
docker compose ps
```

### 5. 访问

- 前端: `http://<服务器IP>/`
- 后端: `http://<服务器IP>:8080/health`
- Ollama: `http://<服务器IP>:11434/api/tags`

## 四、配置说明

### 环境变量（backend/.env）

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `OLLAMA_URL` | `http://localhost:11434` | Ollama 地址（docker 内为 `http://ollama:11434`） |
| `OLLAMA_EMBEDDING_MODEL` | `bge-m3` | 嵌入模型 |
| `OLLAMA_CHAT_MODEL` | `deepseek-r1:7b` | 对话模型 |
| `KNOWLEDGE_STORE_DIR` | `./data/knowledge` | 文档存储目录 |
| `DB_DSN` | - | PostgreSQL 连接串 |

### 切换对话模型

有 GPU 时可升级模型：

```bash
docker exec ai-kb-ollama ollama pull deepseek-r1:14b
```

修改 `backend/.env`:
```
OLLAMA_CHAT_MODEL=deepseek-r1:14b
```

重启后端：
```bash
docker compose restart backend
```

## 五、功能清单

| 功能 | 页面 | 说明 |
|------|------|------|
| 文档知识库 | `/knowledge/base` | 上传 TXT/MD/JSON/CSV/LOG/DOCX/PDF，自动分块+向量化 |
| 问题库 | `/knowledge/base`（问题库 Tab） | 手动录入常见问题，自动向量化 |
| 智能搜索 | `/knowledge/base`（搜索 Tab） | 语义检索文档+问题库 |
| AI 对话 | `/knowledge/ai-chat` | RAG 增强对话，流式输出，带会话历史 |
| 数据统计 | 知识库首页 | 文档/分块/问题数、模型状态 |

## 六、数据备份

```bash
# 备份 PostgreSQL
docker exec ai-kb-postgres pg_dump -U kbuser knowledge > knowledge_backup.sql

# 备份知识文件
cp -r ./data/knowledge ./knowledge_backup

# 备份向量库（同上，knowledge_backup.sql 已包含向量）
```

## 七、常见问题

### Q1: Ollama 无法拉取模型（内网无法访问外网）
需要先在有网络的机器上执行 `ollama pull deepseek-r1:7b`，然后把
`~/.ollama/models` 目录拷贝到服务器 `./data/ollama/models`。

### Q2: 上传文档提示「Ollama 离线」
```bash
docker compose ps | grep ollama  # 检查 ollama 是否运行
docker exec ai-kb-ollama ollama list  # 检查模型是否已拉取
```

### Q3: 知识库页面提示「向量索引不存在」
pgvector 扩展需在首次启动自动创建，若失败可手动执行：
```bash
docker exec -i ai-kb-postgres psql -U kbuser -d knowledge -c "CREATE EXTENSION IF NOT EXISTS vector;"
```

### Q4: 对话很慢
CPU 推理 5-8 token/s 属正常。可减小 `top_k` 检索范围，或在有 GPU 的
服务器上运行 ollama。

## 八、端口规划

| 端口 | 服务 | 说明 |
|------|------|------|
| 80 | 前端 Nginx | 用户访问入口 |
| 8080 | Go 后端 | API 服务 |
| 11434 | Ollama | 本地模型（仅内网） |
| 5432 | PostgreSQL | 数据库（建议仅内网访问） |
