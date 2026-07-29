-- ============================================================
-- 海外发货管理 overseas_shipping
-- 在 Supabase SQL Editor 中执行（需要 service_role 权限）
-- 后端 Go 服务通过 service_role key 访问该表
-- ============================================================

CREATE TABLE IF NOT EXISTS overseas_shipping (
  id                          BIGSERIAL PRIMARY KEY,
  product_type                TEXT,
  sales_engineer             TEXT,
  country                    TEXT,
  customer_name              TEXT,
  controller_sn              TEXT,
  handheld_sn                TEXT,
  controller_hw_version      TEXT,
  controller_upgrade_history TEXT,
  handheld_upgrade_history   TEXT,
  last_upgrade_date          DATE,
  assistant                  TEXT,
  remark                     TEXT,
  created_at                 TIMESTAMPTZ DEFAULT now(),
  updated_at                 TIMESTAMPTZ DEFAULT now()
);

-- 行级安全
ALTER TABLE overseas_shipping ENABLE ROW LEVEL SECURITY;

-- service_role 拥有全部权限（后端使用 service_role key 访问）
DROP POLICY IF EXISTS "overseas_shipping_service_all" ON overseas_shipping;
CREATE POLICY "overseas_shipping_service_all" ON overseas_shipping
  FOR ALL TO service_role USING (true) WITH CHECK (true);

-- 登录用户可读取
DROP POLICY IF EXISTS "overseas_shipping_select" ON overseas_shipping;
CREATE POLICY "overseas_shipping_select" ON overseas_shipping
  FOR SELECT TO authenticated USING (true);

-- updated_at 自动更新（若触发器函数不存在则创建，已存在则覆盖，安全幂等）
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_overseas_shipping_updated_at ON overseas_shipping;
CREATE TRIGGER trg_overseas_shipping_updated_at
  BEFORE UPDATE ON overseas_shipping
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- 常用查询索引（可选，提升按国家/客户/控制器编号检索速度）
CREATE INDEX IF NOT EXISTS idx_overseas_shipping_country ON overseas_shipping(country);
CREATE INDEX IF NOT EXISTS idx_overseas_shipping_customer ON overseas_shipping(customer_name);
CREATE INDEX IF NOT EXISTS idx_overseas_shipping_controller ON overseas_shipping(controller_sn);
