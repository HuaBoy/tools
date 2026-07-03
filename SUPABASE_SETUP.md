# Supabase 配置完整指南

> 本指南将帮助您从零开始配置 Supabase，让用户管理、权限管理、功能管理页面能够正常使用。

---

## 一、问题分析

当前项目根目录的 `.env` 文件中：
```bash
VITE_SUPABASE_URL=https://your-project.supabase.co
VITE_SUPABASE_ANON_KEY=your-anon-key-from-supabase-dashboard
```

这是**占位符**，所以 `isSupabaseConfigured` 被判定为 `false`，项目运行在"本地模式"下：
- ✅ 登录可以用默认密码 `123456` 登录本地账号
- ✅ 本地账号默认有所有权限（包括 `user:view`）
- ❌ 但用户管理、权限管理、功能管理页面的数据来自 Supabase 数据库，没有配置时只能看到空数据

**"无访问权限"错误原因**：
- 如果您**未登录直接访问**用户管理页面，权限检查会失败
- 或者您登录的账号没有 `user:view` 权限

---

## 二、配置 Supabase 步骤

### 步骤 1：注册 Supabase 账号

1. 访问 [https://supabase.com](https://supabase.com)
2. 点击 "Start your project" 注册账号（推荐使用 GitHub 登录）
3. 完成邮箱验证

### 步骤 2：创建新项目

1. 登录后点击 "New Project"
2. 填写项目信息：
   - **Name**: `tester-platform`（或您喜欢的名字）
   - **Database Password**: 设置一个强密码（**请记住它**）
   - **Region**: 选择 `Singapore` 或 `Northeast Asia (Tokyo)`（离中国大陆近）
3. 点击 "Create new project"
4. 等待 1-2 分钟项目初始化完成

### 步骤 3：获取 API 密钥

1. 在项目左侧菜单点击 **⚙️ Project Settings** → **API**
2. 找到以下信息并复制：
   - **Project URL**: 类似 `https://abcdefghijk.supabase.co`
   - **anon public** key: 一长串以 `eyJ` 开头的 JWT token

### 步骤 4：修改项目 `.env` 文件

打开 `/Users/huaqin/Desktop/tools20260623/.env`，替换为：

```bash
VITE_SUPABASE_URL=https://您的项目URL.supabase.co
VITE_SUPABASE_ANON_KEY=您的anon密钥
```

**示例**：
```bash
VITE_SUPABASE_URL=https://abcdefghijk.supabase.co
VITE_SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImFiY2RlZmdoaWprIiwicm9sZSI6ImFub24iLCJpYXQiOjE2OTk5OTk5OTksImV4cCI6MjAxNTU3NTk5OX0.xxxxxxxxx
```

### 步骤 5：创建数据库表

在 Supabase 控制台，点击左侧 **🗄️ SQL Editor** → **New query**，复制并执行以下 SQL：

```sql
-- ========================================
-- 1. 用户配置表 (user_profiles)
-- ========================================
CREATE TABLE IF NOT EXISTS public.user_profiles (
  id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  username TEXT UNIQUE NOT NULL,
  email TEXT UNIQUE,
  role TEXT DEFAULT 'user',
  permissions JSONB DEFAULT '[]'::jsonb,
  avatar_url TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- 启用行级安全
ALTER TABLE public.user_profiles ENABLE ROW LEVEL SECURITY;

-- 允许所有人查看用户（可根据需要修改）
CREATE POLICY "Allow public read user_profiles"
  ON public.user_profiles FOR SELECT
  USING (true);

-- 允许登录用户更新自己的信息
CREATE POLICY "Allow users to update own profile"
  ON public.user_profiles FOR UPDATE
  USING (auth.uid() = id);

-- 允许 service_role 完全访问（用于后台管理）
CREATE POLICY "Allow service_role full access"
  ON public.user_profiles FOR ALL
  USING (auth.jwt() ->> 'role' = 'service_role');

-- 自动创建用户配置（注册时触发）
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.user_profiles (id, username, email, role, permissions)
  VALUES (
    NEW.id,
    COALESCE(NEW.raw_user_meta_data->>'username', split_part(NEW.email, '@', 1)),
    NEW.email,
    COALESCE(NEW.raw_user_meta_data->>'role', 'user'),
    COALESCE(NEW.raw_user_meta_data->>'permissions', '[]')::jsonb
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 绑定触发器
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- ========================================
-- 2. 功能管理表 (features)
-- ========================================
CREATE TABLE IF NOT EXISTS public.features (
  id BIGSERIAL PRIMARY KEY,
  code TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  category TEXT,
  enabled BOOLEAN DEFAULT true,
  "order" INTEGER DEFAULT 0,
  config JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- 启用行级安全
ALTER TABLE public.features ENABLE ROW LEVEL SECURITY;

-- 允许所有人查看功能
CREATE POLICY "Allow public read features"
  ON public.features FOR SELECT
  USING (true);

-- 允许 service_role 完全访问
CREATE POLICY "Allow service_role full access features"
  ON public.features FOR ALL
  USING (auth.jwt() ->> 'role' = 'service_role');

-- 插入初始功能数据
INSERT INTO public.features (code, name, description, category, enabled, "order") VALUES
  ('trace_analysis', '批次数据追溯', '查询和追溯批次数据', 'AI起爆数据专区', true, 1),
  ('factory_data', '智能制造系统', '智能制造数据查询', 'AI起爆数据专区', true, 2),
  ('log_decrypt', '日志解密', '解密设备日志', '日志处理专区', true, 3),
  ('log_analysis', 'AI日志分析', 'AI智能分析日志', '日志处理专区', true, 4),
  ('auth_converter', '授权码转换', '设备码转授权码', '授权管理', true, 5),
  ('third_party_auth', '三方账号授权', '第三方系统登录', '授权管理', true, 6),
  ('ai_translate', 'AI翻译', '多语言翻译工具', '通用辅助工具', true, 7),
  ('data_query', 'AI起爆数据', 'AI起爆数据查询', 'AI起爆数据专区', true, 8)
ON CONFLICT (code) DO NOTHING;

-- ========================================
-- 3. 操作日志表 (operation_logs) - 可选
-- ========================================
CREATE TABLE IF NOT EXISTS public.operation_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID REFERENCES auth.users(id),
  username TEXT,
  action TEXT NOT NULL,
  module TEXT,
  details TEXT,
  ip_address TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

ALTER TABLE public.operation_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Allow public read operation_logs"
  ON public.operation_logs FOR SELECT
  USING (true);

CREATE POLICY "Allow service_role full access logs"
  ON public.operation_logs FOR ALL
  USING (auth.jwt() ->> 'role' = 'service_role');
```

### 步骤 6：配置认证（关闭邮箱验证 - 可选）

为方便测试，可以在 **Authentication** → **Providers** → **Email** 中：
- 关闭 "Confirm email"（生产环境建议开启）
- 关闭 "Secure email change"

### 步骤 7：创建第一个用户

1. 在 Supabase 控制台点击 **🔐 Authentication** → **Users** → **Add user** → **Create new user**
2. 填写：
   - **Email**: `admin@example.com`
   - **Password**: `admin123456`（或您自己的密码）
   - **Auto Confirm User**: ✅ 勾选
3. 点击 "Create user"
4. 创建后，点击该用户，在 **User Metadata** 中添加：
   ```json
   {
     "username": "admin",
     "role": "admin",
     "permissions": ["user:view", "user:create", "user:edit", "user:delete", "role:assign", "permission:manage", "feature:toggle", "feature:view", "data:query", "data:export", "log:view"]
   }
   ```

### 步骤 8：重启项目

修改 `.env` 后必须重启开发服务器：

```bash
# 停止当前服务（Ctrl+C）
# 重新启动
cd /Users/huaqin/Desktop/tools20260623
npm run dev
```

### 步骤 9：测试登录

1. 访问 http://localhost:5173
2. 使用步骤 7 创建的账号登录（**不再使用 `123456` 默认密码**）
3. 进入"系统管理" → "用户管理"，应能看到用户列表

---

## 三、解决 "无访问权限" 错误

### 情况 1：本地模式登录后仍提示无权限

**原因**：本地账号 `123456` 默认有所有权限，但 `user_profiles` 表是空的。

**解决**：
- 本地模式下，用户管理页面将无法显示真实用户列表（因为没有 Supabase 数据源）
- 这是正常的，本地模式是降级方案
- 如果需要管理真实用户，请配置 Supabase（见上文）

### 情况 2：Supabase 模式下提示无权限

**原因**：登录的用户 `permissions` 字段为空。

**解决**：
- 在 Supabase 控制台打开 `user_profiles` 表
- 找到当前用户记录
- 在 `permissions` 字段填入：
  ```json
  ["user:view", "user:create", "user:edit", "user:delete", "role:assign", "permission:manage", "feature:toggle", "feature:view", "data:query", "data:export", "log:view"]
  ```
- 点击保存
- 重新登录

### 情况 3：未登录直接访问

**原因**：路由守卫会重定向到登录页，但如果直接通过 URL 访问，会触发权限检查。

**解决**：先登录再访问管理页面。

---

## 四、临时方案：纯本地模式

如果暂时不想配置 Supabase，可以继续使用本地模式（默认密码 `123456`）。

**注意**：本地模式下：
- ✅ 登录、授权码转换、批次追溯等核心功能可用
- ✅ 自动登录状态管理可用
- ❌ **用户管理、权限管理、功能管理页面数据为空**（因为需要 Supabase 数据库）
- ❌ **多设备同步不可用**（数据只存在本地 localStorage）

---

## 五、常见问题

### Q1：配置后还是提示 "Failed to fetch"
**A**：检查 `.env` 文件：
- URL 是否以 `https://` 开头
- URL 是否包含 `.supabase.co`
- 密钥是否以 `eyJ` 开头且长度 > 100
- 修改 `.env` 后必须重启 `npm run dev`

### Q2：注册用户时失败
**A**：检查 Supabase 控制台 → **Authentication** → **URL Configuration**：
- 添加 `http://localhost:5173` 到 Site URL
- 添加 `http://localhost:5173/**` 到 Redirect URLs

### Q3：数据库表创建失败
**A**：
- SQL 必须分步执行
- 检查是否已有同名表（如果已有，可加 `IF NOT EXISTS`）
- 在 SQL Editor 中查看错误信息

### Q4：需要支持 RLS（行级安全）
**A**：在 `user_profiles` 表的 RLS 策略中，可以限制只有管理员才能修改其他用户：

```sql
-- 只有 admin 角色的用户能更新所有用户
CREATE POLICY "Admins can update all users"
  ON public.user_profiles FOR UPDATE
  USING (
    EXISTS (
      SELECT 1 FROM public.user_profiles
      WHERE id = auth.uid() AND role = 'admin'
    )
  );
```

---

## 六、推荐的后续操作

1. **创建多个测试用户**：用于测试权限管理
2. **设置定期备份**：Supabase Pro 计划支持
3. **启用邮件模板**：自定义注册、找回密码邮件
4. **配置 OAuth 登录**：支持 Google、GitHub 等第三方登录
5. **监控使用量**：在 Supabase Dashboard 查看 API 调用次数

---

## 七、获取帮助

- Supabase 官方文档：https://supabase.com/docs
- Supabase Discord 社区：https://discord.supabase.com
- 项目中的 Supabase 配置文件：`src/utils/supabase.js`
- 项目中的 API 接口：`src/api/admin.js`
