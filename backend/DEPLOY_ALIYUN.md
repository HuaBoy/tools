# ========================================
# 阿里云部署指南
# ========================================

本文档介绍如何将 Go + Supabase 后端部署到阿里云 ECS 服务器。

---

## 一、准备工作

### 1. 阿里云账号和资源
- ✅ 阿里云账号
- ✅ ECS 服务器（推荐配置：2 核 4G，2Mbps 带宽）
- ✅ 已备案的域名（用于 HTTPS）
- ✅ Supabase 项目（已创建）

### 2. 域名解析
在阿里云 DNS 解析中：
- 添加 A 记录：`api.your-domain.com` → ECS 公网 IP

### 3. 阿里云安全组配置
在 ECS 控制台 → 安全组 → 配置规则 → 入方向：

| 端口范围 | 协议 | 授权对象 | 说明 |
|---------|------|---------|------|
| 22/22 | TCP | 0.0.0.0/0 | SSH |
| 80/80 | TCP | 0.0.0.0/0 | HTTP |
| 443/443 | TCP | 0.0.0.0/0 | HTTPS |

---

## 二、服务器初始化

### 1. SSH 连接服务器

```bash
ssh root@your-server-ip
```

### 2. 上传代码

**方式 A：使用 Git**（推荐）
```bash
cd /www/wwwroot
git clone https://github.com/your-username/tester-platform.git
cd tester-platform
```

**方式 B：使用 SCP**
```bash
# 在本地执行
scp -r ./backend root@your-server-ip:/www/wwwroot/tester-platform/
```

### 3. 运行初始化脚本

```bash
cd /www/wwwroot/tester-platform/backend
chmod +x scripts/*.sh
./scripts/init-server.sh
```

这个脚本会：
- 更新系统
- 安装基础工具
- 创建 www 用户
- 安装 Nginx
- 安装 Certbot（用于 HTTPS）
- 配置防火墙

---

## 三、配置环境变量

```bash
cd /www/wwwroot/tester-platform/backend
cp .env.example .env
vim .env
```

修改以下关键配置：

```bash
# Supabase
SUPABASE_URL=https://你的项目.supabase.co
SUPABASE_ANON_KEY=你的anon_key
SUPABASE_SERVICE_KEY=你的service_role_key

# JWT
JWT_SECRET=$(openssl rand -hex 32)  # 生成强随机密钥
JWT_EXPIRE_HOURS=720

# CORS
CORS_ORIGINS=https://your-frontend-domain.com,http://localhost:5173

# 服务器
SERVER_PORT=8080
SERVER_HOST=127.0.0.1  # 仅监听本地，由 Nginx 反向代理

# 日志
LOG_LEVEL=info
LOG_FILE=logs/app.log
```

---

## 四、部署应用

```bash
chmod +x scripts/deploy.sh
./scripts/deploy.sh
```

部署脚本会：
1. 安装 Go 1.21
2. 编译 Go 程序
3. 配置 systemd 服务
4. 启动服务
5. 配置 Nginx 反向代理

---

## 五、配置 HTTPS

### 1. 申请 SSL 证书（Let's Encrypt 免费）

```bash
# 先确保域名解析正确
certbot --nginx -d api.your-domain.com
```

按提示输入邮箱、同意条款即可。

### 2. 配置自动续期

```bash
# 添加定时任务
crontab -e
# 添加以下行
0 3 * * * certbot renew --quiet --post-hook "systemctl reload nginx"
```

### 3. 替换 Nginx 配置

```bash
cp /www/wwwroot/tester-platform/backend/scripts/nginx.conf.example /etc/nginx/conf.d/tester-platform.conf

# 修改 server_name 为你的域名
sed -i 's/api.your-domain.com/api.your-actual-domain.com/g' /etc/nginx/conf.d/tester-platform.conf

# 重载 Nginx
nginx -t && nginx -s reload
```

---

## 六、配置 Supabase CORS

在 Supabase 控制台：
1. **Authentication** → **URL Configuration**
2. 添加到 **Site URL**：`https://api.your-domain.com`
3. 添加到 **Redirect URLs**：
   - `https://api.your-domain.com/**`
   - `https://your-frontend-domain.com/**`

---

## 七、配置 Supabase 数据库

在 Supabase 控制台 **SQL Editor** 中执行以下 SQL：

```sql
-- 创建 roles 表
CREATE TABLE IF NOT EXISTS public.roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  code TEXT UNIQUE NOT NULL,
  description TEXT,
  permissions JSONB DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

ALTER TABLE public.roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Allow public read roles"
  ON public.roles FOR SELECT
  USING (true);

CREATE POLICY "Allow service_role full access roles"
  ON public.roles FOR ALL
  USING (auth.jwt() ->> 'role' = 'service_role');

-- 插入默认角色
INSERT INTO public.roles (name, code, description, permissions) VALUES
  ('超级管理员', 'admin', '系统最高权限', '["user:view", "user:create", "user:edit", "user:delete", "role:assign", "permission:manage", "feature:toggle", "feature:view", "data:query", "data:export", "log:view"]'),
  ('普通用户', 'user', '普通用户权限', '["user:view", "data:query", "data:export", "log:view"]')
ON CONFLICT (code) DO NOTHING;
```

---

## 八、前端配置

修改前端的 `.env` 文件：

```bash
# 前端项目的 .env
VITE_API_BASE_URL=https://api.your-domain.com
VITE_SUPABASE_URL=https://你的项目.supabase.co
VITE_SUPABASE_ANON_KEY=你的anon_key
```

---

## 九、日常运维

### 1. 查看服务状态
```bash
systemctl status tester-platform
```

### 2. 查看日志
```bash
# 实时日志
journalctl -u tester-platform -f

# 应用日志
tail -f /www/wwwroot/tester-platform/backend/logs/app.log
```

### 3. 重启服务
```bash
systemctl restart tester-platform
```

### 4. 重新部署
```bash
cd /www/wwwroot/tester-platform/backend
./scripts/deploy.sh
```

### 5. 备份
```bash
# 备份应用
tar -czf backup-$(date +%Y%m%d).tar.gz /www/wwwroot/tester-platform

# 备份到 OSS（需安装 ossutil）
ossutil cp backup-$(date +%Y%m%d).tar.gz oss://your-bucket/backups/
```

---

## 十、监控和告警

### 1. 阿里云监控
- 在 ECS 控制台配置云监控
- 设置 CPU、内存、磁盘使用率告警
- 建议告警阈值：CPU > 80%, 内存 > 85%, 磁盘 > 90%

### 2. 配置日志收集
```bash
# 安装阿里云日志服务 CLI
wget https://aliyun-log-cli.oss-cn-hangzhou.aliyuncs.com/aliyun-log-cli.linux-amd64.tar.gz
tar -xzf aliyun-log-cli.linux-amd64.tar.gz
```

### 3. 配置健康检查
在 Nginx 中添加健康检查端点：
```nginx
location /health {
    proxy_pass http://127.0.0.1:8080/health;
    access_log off;
}
```

可以使用阿里云的云拨测定期检查 `/health` 端点。

---

## 十一、安全建议

### 1. SSH 安全
```bash
# 修改 SSH 端口
sed -i 's/#Port 22/Port 2222/' /etc/ssh/sshd_config

# 禁用密码登录，使用密钥
sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
systemctl restart sshd
```

### 2. 配置 fail2ban
```bash
apt install -y fail2ban  # Debian/Ubuntu
yum install -y fail2ban  # CentOS
systemctl enable fail2ban
systemctl start fail2ban
```

### 3. 数据库密钥管理
- 不要将 `.env` 文件提交到 Git
- 定期轮换 `JWT_SECRET` 和 Supabase 密钥
- 使用阿里云 KMS 管理密钥（推荐）

### 4. 定期更新
```bash
# 设置自动更新（Debian/Ubuntu）
apt install -y unattended-upgrades
dpkg-reconfigure -plow unattended-upgrades
```

---

## 十二、常见问题

### Q1：部署后访问 502 Bad Gateway
**原因**：Go 服务未启动
**解决**：
```bash
systemctl status tester-platform
journalctl -u tester-platform -n 50
```

### Q2：Supabase 调用失败
**原因**：`.env` 配置错误或 Supabase 密钥错误
**解决**：
```bash
# 测试 Supabase 连通性
curl https://你的项目.supabase.co/auth/v1/health
```

### Q3：CORS 错误
**原因**：`.env` 中 `CORS_ORIGINS` 配置错误
**解决**：
```bash
# 修改 CORS_ORIGINS 后重启服务
systemctl restart tester-platform
```

### Q4：HTTPS 证书过期
**原因**：Let's Encrypt 证书 90 天过期
**解决**：
```bash
certbot renew
systemctl reload nginx
```

### Q5：磁盘空间不足
**解决**：
```bash
# 清理旧日志
journalctl --vacuum-time=7d
find /www/wwwroot/tester-platform/backend/logs -name "*.log.*" -mtime +7 -delete
```

---

## 十三、性能优化

### 1. 启用 Go 编译优化
在 `deploy.sh` 中已使用 `-ldflags="-s -w"` 减少二进制大小。

### 2. 启用 Nginx 缓存
```nginx
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=api_cache:10m max_size=1g inactive=60m;

location /api/v1/ {
    proxy_cache api_cache;
    proxy_cache_valid 200 5m;
    proxy_cache_key "$scheme$request_method$host$request_uri";
    proxy_pass http://127.0.0.1:8080;
}
```

### 3. 启用 Gzip 压缩
```nginx
gzip on;
gzip_types application/json text/plain text/css application/javascript;
gzip_min_length 1024;
```

### 4. 启用 HTTP/2
已在 nginx.conf.example 中配置 `listen 443 ssl http2;`

---

## 十四、成本估算

| 项目 | 配置 | 月费用（预估） |
|------|------|--------------|
| ECS | 2核4G, 2Mbps | ~150 元 |
| 域名 | .com | ~60 元/年 |
| SSL 证书 | Let's Encrypt | 免费 |
| Supabase | Free Plan | 免费 |
| 阿里云 DNS | 免费版 | 免费 |
| **合计** | | **~210 元/年** |

如需更高性能，建议：
- ECS: 4核8G, 5Mbps
- Supabase: Pro 计划 ($25/月)

---

## 联系支持

- 阿里云工单：https://workorder.console.aliyun.com
- Supabase Discord：https://discord.supabase.com
- 项目 Issues：https://github.com/your-username/tester-platform/issues
