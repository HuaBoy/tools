# 阿里云 CentOS 7.9 部署指南
# 服务器：172.24.249.108
# Supabase: osjidnablnorytfzjjgb

## 一、Supabase 信息

根据您提供的 ProjectId，您的 Supabase 配置如下：

```bash
SUPABASE_URL=https://osjidnablnorytfzjjgb.supabase.co
SUPABASE_ANON_KEY=<需要从 Supabase Dashboard 获取>
SUPABASE_SERVICE_KEY=<需要从 Supabase Dashboard 获取>
```

**获取密钥步骤**：
1. 访问 https://supabase.com/dashboard/project/osjidnablnorytfzjjgb
2. 左侧菜单 **⚙️ Project Settings** → **API**
3. 复制 **Project URL** 确认是 `https://osjidnablnorytfzjjgb.supabase.co`
4. 复制以下密钥：
   - **anon public** key（公开密钥）
   - **service_role** key（服务端密钥，**保密**）

---

## 二、阿里云安全组配置（必须先做）

1. 登录阿里云控制台：https://ecs.console.aliyun.com
2. 找到您的 ECS 实例（公网 IP：172.24.249.108）
3. 点击实例 → **安全组** → **配置规则** → **入方向** → **手动添加**

添加以下规则：

| 端口范围 | 协议 | 授权对象 | 描述 |
|---------|------|---------|------|
| 22/22 | TCP | 0.0.0.0/0 | SSH |
| 80/80 | TCP | 0.0.0.0/0 | HTTP |
| 443/443 | TCP | 0.0.0.0/0 | HTTPS |
| 8080/8080 | TCP | 0.0.0.0/0 | 后端服务（仅测试用） |

---

## 三、SSH 连接到服务器

```bash
ssh root@172.24.249.108
```

输入服务器密码登录。

---

## 四、服务器初始化（首次部署）

### 1. 禁用 SELinux（重要）

CentOS 7 默认开启了 SELinux，会拦截很多操作。开发环境建议关闭：

```bash
# 临时关闭
setenforce 0

# 永久关闭
sed -i 's/SELINUX=enforcing/SELINUX=disabled/' /etc/selinux/config
```

### 2. 上传代码

**方式 A：使用 SCP 上传（在本机执行）**
```bash
cd /Users/huaqin/Desktop/tools20260623
scp -r ./backend root@172.24.249.108:/www/wwwroot/tester-platform/
```

**方式 B：在服务器上使用 Git 拉取**
```bash
ssh root@172.24.249.108
yum install -y git
mkdir -p /www/wwwroot
cd /www/wwwroot
# 假设您把代码推到了 Git 仓库
git clone https://github.com/your-username/tester-platform.git
```

### 3. 配置环境变量

```bash
ssh root@172.24.249.108
cd /www/wwwroot/tester-platform/backend
cp .env.example .env
vim .env
```

**修改以下内容**：

```bash
# ========================================
# Supabase 配置
# ========================================
SUPABASE_URL=https://osjidnablnorytfzjjgb.supabase.co
SUPABASE_ANON_KEY=粘贴您的anon密钥
SUPABASE_SERVICE_KEY=粘贴您的service_role密钥

# ========================================
# 服务配置
# ========================================
SERVER_PORT=8080
SERVER_HOST=127.0.0.1
SERVER_MODE=release

# ========================================
# JWT 配置
# ========================================
JWT_SECRET=$(openssl rand -hex 32)
JWT_EXPIRE_HOURS=720

# ========================================
# CORS 配置
# ========================================
CORS_ORIGINS=https://your-frontend-domain.com,http://localhost:5173,http://172.24.249.108:5173

# ========================================
# 日志配置
# ========================================
LOG_LEVEL=info
LOG_FILE=logs/app.log
```

保存退出（`:wq`）。

### 4. 初始化服务器环境

```bash
cd /www/wwwroot/tester-platform/backend
chmod +x scripts/*.sh
./scripts/init-server.sh
```

这个脚本会安装 Go、Nginx、Certbot、配置防火墙等。

**注意**：CentOS 7 的防火墙是 `firewalld`，不是 `ufw`。

### 5. 安装 Go 1.21（如果脚本未自动安装）

```bash
cd /tmp
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
source /etc/profile
go version  # 验证
```

### 6. 编译和部署

```bash
cd /www/wwwroot/tester-platform/backend
./scripts/deploy.sh
```

### 7. 启动服务

```bash
systemctl status tester-platform
# 或
systemctl start tester-platform
systemctl enable tester-platform
```

---

## 五、配置 Nginx 反向代理

### 1. 修改配置

```bash
vim /etc/nginx/conf.d/tester-platform.conf
```

将 `api.your-domain.com` 替换为您的域名（或暂时用 IP）。

### 2. 测试和启动

```bash
nginx -t
systemctl restart nginx
systemctl enable nginx
```

---

## 六、配置 HTTPS（可选）

### 1. 申请 Let's Encrypt 证书

```bash
certbot --nginx -d api.your-domain.com
```

### 2. 自动续期

```bash
echo "0 3 * * * root certbot renew --quiet --post-hook 'systemctl reload nginx'" >> /etc/crontab
```

---

## 七、初始化 Supabase 数据库

### 1. 在 Supabase SQL Editor 执行脚本

1. 访问 https://supabase.com/dashboard/project/osjidnablnorytfzjjgb/sql
2. 点击 **New query**
3. 复制 [init-supabase.sql](file:///Users/huaqin/Desktop/tools20260623/backend/scripts/init-supabase.sql) 内容
4. 点击 **Run** 执行

### 2. 创建第一个管理员用户

**步骤 1**：在 Supabase Dashboard 创建用户
1. 访问 https://supabase.com/dashboard/project/osjidnablnorytfzjjgb/auth/users
2. 点击 **Add user** → **Create new user**
3. 填写：
   - Email: `admin@your-domain.com`（替换为您的邮箱）
   - Password: 设置一个强密码
   - Auto Confirm User: ✅ 勾选
4. 点击 **Create user**

**步骤 2**：设为管理员
1. 回到 SQL Editor
2. 执行：
   ```sql
   SELECT public.set_admin('admin@your-domain.com');
   ```
3. 应该看到 `用户 admin@your-domain.com 已设为管理员`

---

## 八、配置 Supabase CORS

1. 访问 https://supabase.com/dashboard/project/osjidnablnorytfzjjgb/auth/url-configuration
2. **Site URL**: 添加 `https://api.your-domain.com` 和 `http://172.24.249.108:8080`
3. **Redirect URLs**: 添加以下（每行一个）：
   ```
   https://api.your-domain.com/**
   http://172.24.249.108:8080/**
   http://localhost:5173/**
   ```

---

## 九、验证部署

### 1. 服务状态

```bash
systemctl status tester-platform
```

应该看到 `active (running)`。

### 2. 健康检查

```bash
curl http://127.0.0.1:8080/health
```

预期输出：
```json
{"code":200,"message":"success","data":{"status":"ok","version":"1.0.0","supabase":true}}
```

### 3. 通过 Nginx 访问

```bash
curl http://172.24.249.108/health
```

或如果有域名：
```bash
curl https://api.your-domain.com/health
```

### 4. 测试登录 API

```bash
curl -X POST http://172.24.249.108/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@your-domain.com","password":"your-password"}'
```

预期返回 token：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJ...",
    "expires_in": 2592000,
    "user_id": "...",
    "username": "admin",
    "email": "admin@your-domain.com",
    "role": "admin",
    "permissions": ["user:view", ...]
  }
}
```

---

## 十、配置前端项目

修改前端项目 `/Users/huaqin/Desktop/tools20260623/.env`：

```bash
# 后端 API 地址
VITE_API_BASE_URL=http://172.24.249.108:8080
# 或使用域名
# VITE_API_BASE_URL=https://api.your-domain.com

# Supabase 配置（前端直连 Supabase 时使用）
VITE_SUPABASE_URL=https://osjidnablnorytfzjjgb.supabase.co
VITE_SUPABASE_ANON_KEY=您的anon密钥
```

---

## 十一、常用运维命令

```bash
# 查看服务状态
systemctl status tester-platform

# 查看实时日志
journalctl -u tester-platform -f

# 查看应用日志
tail -f /www/wwwroot/tester-platform/backend/logs/app.log

# 重启服务
systemctl restart tester-platform

# 重新部署（拉取新代码后）
cd /www/wwwroot/tester-platform/backend
./scripts/deploy.sh

# 查看 Nginx 日志
tail -f /var/log/nginx/tester-platform.access.log
tail -f /var/log/nginx/tester-platform.error.log
```

---

## 十二、故障排查

### 问题 1：服务启动失败

```bash
# 查看详细错误
journalctl -u tester-platform -n 50 --no-pager

# 常见原因：
# 1. .env 配置错误（URL 格式不对）
# 2. Supabase Key 错误
# 3. 端口被占用
netstat -tlnp | grep 8080
```

### 问题 2：Nginx 502 Bad Gateway

```bash
# 检查后端是否运行
systemctl status tester-platform
curl http://127.0.0.1:8080/health

# 检查 Nginx 配置
nginx -t
tail -f /var/log/nginx/tester-platform.error.log
```

### 问题 3：Supabase 调用失败

```bash
# 测试 Supabase 连通性
curl https://osjidnablnorytfzjjgb.supabase.co/auth/v1/health

# 检查 .env 文件
cat /www/wwwroot/tester-platform/backend/.env
```

### 问题 4：CentOS 7 yum 报错

```bash
# yum 源可能已停止维护，可更换为阿里云源
curl -o /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-7.repo
yum clean all
yum makecache
```

---

## 十三、成本估算（CentOS 7 ECS）

| 项目 | 配置 | 月费用（预估） |
|------|------|--------------|
| ECS | 2核4G, 2Mbps, CentOS 7.9 | ~150 元 |
| 域名（可选） | .com | ~60 元/年 |
| SSL 证书 | Let's Encrypt | 免费 |
| Supabase | Free Plan | 免费 |
| **合计** | | **~150 元/月** |

---

## 完整部署命令（一键复制）

```bash
# 1. SSH 到服务器
ssh root@172.24.249.108

# 2. 禁用 SELinux
setenforce 0
sed -i 's/SELINUX=enforcing/SELINUX=disabled/' /etc/selinux/config

# 3. 准备目录
mkdir -p /www/wwwroot
cd /www/wwwroot

# 4. 上传代码（在本地执行）
# scp -r ./backend root@172.24.249.108:/www/wwwroot/tester-platform/

# 5. 配置
cd /www/wwwroot/tester-platform/backend
cp .env.example .env
vim .env  # 填入 Supabase URL 和 Key

# 6. 部署
chmod +x scripts/*.sh
./scripts/init-server.sh
./scripts/deploy.sh

# 7. 启动
systemctl start tester-platform
systemctl enable tester-platform

# 8. 配置 Nginx
cp scripts/nginx.conf.example /etc/nginx/conf.d/tester-platform.conf
sed -i 's/api.your-domain.com/YOUR_DOMAIN_OR_IP/g' /etc/nginx/conf.d/tester-platform.conf
nginx -t && systemctl restart nginx

# 9. 验证
curl http://127.0.0.1:8080/health
```

---

**完成部署后**，您的 API 将可以通过以下地址访问：
- `http://172.24.249.108:8080`（直连）
- `http://172.24.249.108`（通过 Nginx）
- `https://api.your-domain.com`（配置域名 + HTTPS 后）

如需任何步骤的详细说明，请告诉我！
