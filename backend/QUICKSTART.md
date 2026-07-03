# ========================================
# 部署到阿里云 ECS 完整流程
# ========================================

## 一键部署命令（在阿里云 ECS 上执行）

```bash
# 1. 初始化服务器（仅首次）
curl -fsSL https://raw.githubusercontent.com/your-repo/tester-platform/main/backend/scripts/init-server.sh | bash

# 2. 上传代码（在本机或 CI 中执行）
scp -r ./backend root@your-server-ip:/www/wwwroot/tester-platform/

# 3. 配置环境变量
ssh root@your-server-ip "cd /www/wwwroot/tester-platform/backend && cp .env.example .env && nano .env"

# 4. 执行部署
ssh root@your-server-ip "cd /www/wwwroot/tester-platform/backend && ./scripts/deploy.sh"

# 5. 配置 HTTPS
ssh root@your-server-ip "certbot --nginx -d api.your-domain.com"
```

## Docker 部署（推荐）

```bash
# 1. 上传代码
scp -r ./backend root@your-server-ip:/www/wwwroot/tester-platform/

# 2. 配置环境变量
cd /www/wwwroot/tester-platform/backend
cp .env.example .env
vim .env  # 填入 Supabase URL 和 Key

# 3. 一键启动
docker-compose up -d --build

# 4. 查看日志
docker-compose logs -f backend

# 5. 重启服务
docker-compose restart backend
```

## 验证部署

```bash
# 健康检查
curl https://api.your-domain.com/health

# 预期输出
# {"code":200,"message":"success","data":{"status":"ok","version":"1.0.0","supabase":true}}

# 测试登录
curl -X POST https://api.your-domain.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"your-password"}'
```
