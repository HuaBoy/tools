# 本地跑通验证指南（Windows）

> ✅ **已在本地跑通验证（2026-08-12）**
> 上传文档 → bge-m3 向量化 → pgvector 存储 → 语义搜索 → deepseek-r1:7b 对话 全流程正常。

> 目标：在本地 Windows 机器上完整跑通「知识库上传 → 向量化 → 检索 → AI 对话」全流程。
> 服务器部署时使用相同软件栈（PostgreSQL 16 + pgvector + Ollama），只是容器化运行。

## ✅ 已验证清单

| 验证项 | 结果 |
|--------|------|
| PostgreSQL 16 + pgvector 0.8.0 | ✅ 可用（winget + 预编译 dll） |
| Ollama 服务 | ✅ 运行中 (localhost:11434) |
| 模型 deepseek-r1:7b | ✅ 已拉取 (4.7GB) |
| 模型 bge-m3 | ✅ 已拉取 (1.2GB, 1024维) |
| 后端启动 + 自动建表 | ✅ 13张表 |
| 登录 admin/800228 | ✅ 返回 JWT |
| 文档上传 + 向量化 | ✅ chunk_count=1, status=ready |
| 语义搜索 | ✅ score 0.68 命中 |
| AI 对话（非流式） | ✅ 基于知识库正确回答 |
| AI 对话（SSE 流式） | ✅ meta/chunk/done 帧正常 |
| FAQ 问题库 | ✅ 入库成功 |
| 前端页面 + API 代理 | ✅ localhost:5173 转发到本地后端 |

## 环境要求

| 组件 | 版本 | 安装方式 |
|------|------|---------|
| PostgreSQL | 16.x | winget / EDB 安装包 |
| pgvector | 0.8.x | 随 PostgreSQL 编译安装 |
| Ollama | 0.3.x | winget 便携版或官方安装包 |
| Node.js | 18+ | 已有 (v24) |
| Go | 1.22+ | 已有 |

---

## 一、安装 PostgreSQL 16 + pgvector

### 方式 A：winget 一键安装（推荐）

```powershell
# 1. 安装 PostgreSQL 16（安装时设置 postgres 用户密码，如 admin123）
winget install PostgreSQL.PostgreSQL.16 --accept-source-agreements --accept-package-agreements

# 2. 安装 pgvector 扩展（需要 Visual Studio Build Tools，先装依赖）
winget install Microsoft.VisualStudio.2022.BuildTools --accept-source-agreements --accept-package-agreements
# 在 Visual Studio Installer 中勾选 "C++ 生成工具"
```

### 方式 B：手动安装

1. 下载 PostgreSQL 16 安装包：https://www.enterprisedb.com/downloads/postgres-postgresql-downloads
2. 安装时记下端口（默认 5432）和 postgres 密码
3. pgvector 编译（需 VS Build Tools）：
   ```powershell
   git clone https://github.com/pgvector/pgvector.git
   cd pgvector
   # 打开 "x64 Native Tools Command Prompt for VS 2022"
   path\to\postgresql\bin\pg_config --version  # 确认环境
   nmake /F Makefile.win
   nmake /F Makefile.win install
   ```

### 验证

```sql
psql -U postgres -c "CREATE EXTENSION IF NOT EXISTS vector; SELECT vector '[1,2,3]'";
-- 应返回 [1,2,3]
```

---

## 二、安装 Ollama

```powershell
# 便携版（免安装，推荐）：
winget install Ollama.Ollama.Portable --accept-source-agreements --accept-package-agreements

# 或官方安装包（写入注册表，开机自启）：
winget install Ollama.Ollama --accept-source-agreements --accept-package-agreements
```

### 拉取模型

```powershell
# 启动 ollama 服务（便携版需手动启动）
ollama serve

# 新终端窗口：
ollama pull deepseek-r1:7b       # ~4.7GB，对话模型
ollama pull bge-m3               # ~1.2GB，嵌入模型（1024维）

# 验证
ollama list
```

---

## 三、初始化数据库

```powershell
# 创建知识库数据库（用 postgres 超管）
psql -U postgres -h localhost -c "CREATE DATABASE knowledge;"
psql -U postgres -h localhost -d knowledge -c "CREATE EXTENSION IF NOT EXISTS vector;"

# 创建应用用户
psql -U postgres -h localhost -d knowledge -c "CREATE USER kbuser WITH PASSWORD 'kbpass_2026';"
psql -U postgres -h localhost -d knowledge -c "GRANT ALL PRIVILEGES ON DATABASE knowledge TO kbuser;"
```

---

## 四、配置并启动后端

### 1. 修改 backend/.env

```ini
# 数据库指向本地
DB_DSN=postgres://kbuser:kbpass_2026@localhost:5432/knowledge?sslmode=disable
DATABASE_URL=postgres://kbuser:kbpass_2026@localhost:5432/knowledge?sslmode=disable

# Ollama 本地地址
OLLAMA_URL=http://localhost:11434
OLLAMA_EMBEDDING_MODEL=bge-small-zh-v1.5
OLLAMA_CHAT_MODEL=deepseek-r1:7b
KNOWLEDGE_STORE_DIR=./data/knowledge
```

### 2. 启动后端

```powershell
cd E:\tools20260623\backend
go mod tidy
go run ./cmd/server
# 看到 "服务启动成功" 即为成功
# 首次启动会自动建表 + 初始化 admin 账号
```

---

## 五、验证 API 全流程

### 1. 登录获取 token

```powershell
$login = Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/auth/login -ContentType "application/json" -Body '{"username":"admin","password":"800228"}'
$token = $login.data.token
Write-Host $token
```

### 2. 查看知识库统计

```powershell
Invoke-RestMethod -Uri http://localhost:8080/api/v1/knowledge/stats -Headers @{Authorization="Bearer $token"} | ConvertTo-Json -Depth 5
```

### 3. 上传测试文档

```powershell
# 先创建测试文档
@"
# 起爆器故障排查手册

## 1. 设备无法连接
当起爆器无法与电脑连接时，请依次检查：
1. 确认 USB 数据线已正确插入，尝试更换数据线
2. 确认设备已开机，指示灯为绿色
3. 打开设备管理器，查看是否识别到未知设备
4. 安装最新版本的 USB 驱动

## 2. 授权码验证失败
如果授权码验证失败：
1. 确认授权码未过期
2. 确认输入的授权码无多余空格
3. 联系销售重新申请授权码
"@ | Out-File -Encoding utf8 "C:\temp\test_manual.md"

# 上传
curl.exe -X POST http://localhost:8080/api/v1/knowledge/documents/upload `
  -H "Authorization: Bearer $token" `
  -F "file=@C:\temp\test_manual.md" `
  -F "title=起爆器故障排查手册" `
  -F "category=log"
```

### 4. 语义搜索

```powershell
$q = [uri]::EscapeDataString("设备无法连接怎么排查")
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/knowledge/search?q=$q" -Headers @{Authorization="Bearer $token"} | ConvertTo-Json -Depth 5
```

### 5. AI 对话（非流式）

```powershell
$body = @{ question="设备无法连接怎么办？"; conversation_id=0; stream=$false } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/knowledge/chat `
  -ContentType "application/json" -Body $body `
  -Headers @{Authorization="Bearer $token"} | ConvertTo-Json -Depth 5
```

### 6. AI 对话（流式 SSE）

```powershell
# 流式返回，浏览器或 curl -N 可见逐字输出
curl.exe -N -X POST http://localhost:8080/api/v1/knowledge/chat `
  -H "Authorization: Bearer $token" `
  -H "Content-Type: application/json" `
  -d '{"question":"授权码验证失败怎么办？","conversation_id":0,"stream":true}'
```

---

## 六、启动前端

```powershell
cd E:\tools20260623
# 开发模式（vite 默认端口可能是 5173）
npm run dev
```

浏览器访问：http://localhost:5173

> 注意：如果 vite.config.js 的 `/api/v1` 代理仍指向 `111.231.16.110`，
> 本地测试时改为 `http://localhost:8080`。

---

## 七、常见问题

| 问题 | 解决 |
|------|------|
| `CREATE EXTENSION vector` 失败 | pgvector 未编译安装成功，重看第一步 |
| Ollama 拉模型超时 | 检查网络；或设置代理 `set HTTPS_PROXY=...` |
| 登录失败 | 确认数据库迁移完成、admin 账号已初始化（密码 800228） |
| 对话返回「模型未找到」 | 执行 `ollama list` 确认模型已拉取 |
| 上传 PDF 报错 | PDF 需为文字型（非扫描件），或先用 TXT/MD 测试 |

## 八、测试完毕清理

```powershell
# 停掉 ollama 进程
Get-Process ollama -ErrorAction SilentlyContinue | Stop-Process -Force

# 删除测试库（如需）
psql -U postgres -c "DROP DATABASE knowledge;"
```
