-- ========================================
-- Supabase 数据库初始化脚本
-- ========================================
-- 在 Supabase SQL Editor 中执行

-- ========================================
-- 1. 用户配置表 (user_profiles)
-- ========================================
CREATE TABLE IF NOT EXISTS public.user_profiles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  auth_id UUID REFERENCES auth.users(id) ON DELETE SET NULL,
  username TEXT UNIQUE NOT NULL,
  email TEXT UNIQUE,
  password TEXT,
  role TEXT DEFAULT 'user',
  permissions JSONB DEFAULT '[]'::jsonb,
  avatar_url TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

ALTER TABLE public.user_profiles ENABLE ROW LEVEL SECURITY;

-- RLS 策略
CREATE POLICY "Allow public read user_profiles"
  ON public.user_profiles FOR SELECT
  USING (true);

CREATE POLICY "Allow users to update own profile"
  ON public.user_profiles FOR UPDATE
  USING (auth.uid() = auth_id);

CREATE POLICY "Allow admins to update all users"
  ON public.user_profiles FOR UPDATE
  USING (
    EXISTS (
      SELECT 1 FROM public.user_profiles
      WHERE auth_id = auth.uid() AND role = 'admin'
    )
    OR auth.jwt() ->> 'role' = 'service_role'
  );

CREATE POLICY "Allow admins to delete users"
  ON public.user_profiles FOR DELETE
  USING (
    EXISTS (
      SELECT 1 FROM public.user_profiles
      WHERE auth_id = auth.uid() AND role = 'admin'
    )
    OR auth.jwt() ->> 'role' = 'service_role'
  );

CREATE POLICY "Allow admins to insert users"
  ON public.user_profiles FOR INSERT
  WITH CHECK (
    EXISTS (
      SELECT 1 FROM public.user_profiles
      WHERE auth_id = auth.uid() AND role = 'admin'
    )
    OR auth.jwt() ->> 'role' = 'service_role'
  );

-- 自动创建用户配置
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.user_profiles (auth_id, username, email, role, permissions)
  VALUES (
    NEW.id,
    COALESCE(NEW.raw_user_meta_data->>'username', split_part(NEW.email, '@', 1)),
    NEW.email,
    COALESCE(NEW.raw_user_meta_data->>'role', 'user'),
    COALESCE(NEW.raw_user_meta_data->>'permissions', '[]')::jsonb
  )
  ON CONFLICT (auth_id) DO NOTHING;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- ========================================
-- 2. 角色表 (roles)
-- ========================================
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

CREATE POLICY "Allow admins to manage roles"
  ON public.roles FOR ALL
  USING (
    EXISTS (
      SELECT 1 FROM public.user_profiles
      WHERE id = auth.uid() AND role = 'admin'
    )
    OR auth.jwt() ->> 'role' = 'service_role'
  );

-- 插入默认角色
INSERT INTO public.roles (name, code, description, permissions) VALUES
  (
    '超级管理员',
    'admin',
    '系统最高权限',
    '["user:view", "user:create", "user:edit", "user:delete", "role:assign", "permission:manage", "feature:toggle", "feature:view", "data:query", "data:export", "log:view"]'::jsonb
  ),
  (
    '普通用户',
    'user',
    '普通用户权限',
    '["user:view", "data:query", "data:export", "log:view"]'::jsonb
  ),
  (
    '操作员',
    'operator',
    '业务操作权限',
    '["data:query", "data:export", "log:view"]'::jsonb
  )
ON CONFLICT (code) DO NOTHING;

-- ========================================
-- 3. 功能管理表 (features)
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

ALTER TABLE public.features ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Allow public read features"
  ON public.features FOR SELECT
  USING (true);

CREATE POLICY "Allow admins to manage features"
  ON public.features FOR ALL
  USING (
    EXISTS (
      SELECT 1 FROM public.user_profiles
      WHERE id = auth.uid() AND role = 'admin'
    )
    OR auth.jwt() ->> 'role' = 'service_role'
  );

-- 插入初始功能数据
INSERT INTO public.features (code, name, description, category, enabled, "order") VALUES
  ('home', '首页', '智能搜索首页', '导航', true, 0),
  ('trace_analysis', '批次数据追溯', '查询和追溯批次数据', 'AI起爆数据专区', true, 1),
  ('factory_data', '智能制造系统', '智能制造数据查询', 'AI起爆数据专区', true, 2),
  ('log_decrypt', '日志解密', '解密设备日志', '日志处理专区', true, 3),
  ('log_analysis', 'AI日志分析', 'AI智能分析日志', '日志处理专区', true, 4),
  ('auth_converter', '授权码转换', '设备码转授权码', '授权管理', true, 5),
  ('third_party_auth', '三方账号授权', '第三方系统登录', '授权管理', true, 6),
  ('ai_translate', 'AI翻译', '多语言翻译工具', '通用辅助工具', true, 7),
  ('data_query', 'AI起爆数据', 'AI起爆数据查询', 'AI起爆数据专区', true, 8),
  ('admin', '后台管理', '系统后台管理', '管理', true, 99)
ON CONFLICT (code) DO NOTHING;

-- ========================================
-- 4. 操作日志表 (operation_logs)
-- ========================================
CREATE TABLE IF NOT EXISTS public.operation_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID REFERENCES auth.users(id),
  username TEXT,
  action TEXT NOT NULL,
  module TEXT,
  details TEXT,
  ip_address TEXT,
  user_agent TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_operation_logs_user_id ON public.operation_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_created_at ON public.operation_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_operation_logs_module ON public.operation_logs(module);

ALTER TABLE public.operation_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Allow admins to read logs"
  ON public.operation_logs FOR SELECT
  USING (
    EXISTS (
      SELECT 1 FROM public.user_profiles
      WHERE id = auth.uid() AND role = 'admin'
    )
    OR auth.jwt() ->> 'role' = 'service_role'
  );

CREATE POLICY "Allow service_role to insert logs"
  ON public.operation_logs FOR INSERT
  WITH CHECK (true);

-- ========================================
-- 5. 创建第一个管理员用户
-- ========================================
-- 注意：需要先在 Supabase Auth 中创建用户，然后更新其 role 和 permissions

-- 创建一个 admin 函数，方便设置第一个管理员
CREATE OR REPLACE FUNCTION public.set_admin(user_email TEXT)
RETURNS void AS $$
DECLARE
  target_user_id UUID;
BEGIN
  -- 查找用户
  SELECT id INTO target_user_id FROM auth.users WHERE email = user_email LIMIT 1;

  IF target_user_id IS NULL THEN
    RAISE EXCEPTION '用户 % 不存在，请先在 Supabase Auth 中创建', user_email;
  END IF;

  -- 更新 user_profiles
  INSERT INTO public.user_profiles (id, username, email, role, permissions)
  VALUES (
    target_user_id,
    split_part(user_email, '@', 1),
    user_email,
    'admin',
    '["user:view", "user:create", "user:edit", "user:delete", "role:assign", "permission:manage", "feature:toggle", "feature:view", "data:query", "data:export", "log:view"]'::jsonb
  )
  ON CONFLICT (id) DO UPDATE SET
    role = 'admin',
    permissions = EXCLUDED.permissions,
    updated_at = now();

  RAISE NOTICE '用户 % 已设为管理员', user_email;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 使用方法（将 your-email@example.com 替换为你的邮箱）:
-- SELECT public.set_admin('your-email@example.com');

-- ========================================
-- 6. 更新时间戳触发器
-- ========================================
CREATE OR REPLACE FUNCTION public.update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON public.user_profiles;
CREATE TRIGGER update_user_profiles_updated_at
  BEFORE UPDATE ON public.user_profiles
  FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();

DROP TRIGGER IF EXISTS update_roles_updated_at ON public.roles;
CREATE TRIGGER update_roles_updated_at
  BEFORE UPDATE ON public.roles
  FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();

DROP TRIGGER IF EXISTS update_features_updated_at ON public.features;
CREATE TRIGGER update_features_updated_at
  BEFORE UPDATE ON public.features
  FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();

-- ========================================
-- 完成！
-- ========================================
-- 接下来：
-- 1. 在 Supabase Authentication > Users 中创建第一个用户
-- 2. 执行: SELECT public.set_admin('your-email@example.com');
-- 3. 使用该账号登录系统
