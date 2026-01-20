-- ============================================================================
-- Happy Billing 调试和初始化数据脚本
-- ============================================================================

-- ============================================================================
-- 1. 清理现有测试数据（支持重复执行）
-- ============================================================================
-- 注意：多币种字段已在 20240117_add_multi_currency_fields.sql 中添加，这里不再重复

-- ============================================================================
-- 2. 用户账号初始化
-- ============================================================================

-- 更新测试用户密码（testuser: 123456）
-- bcrypt hash: $2a$10$l1w0hTFJIKRye287UTa5ie4CO.LcjvYWAbwOKm.wdkAI5fxbdvZb.
UPDATE users
SET username = 'testuser',
    password_hash = '$2a$10$l1w0hTFJIKRye287UTa5ie4CO.LcjvYWAbwOKm.wdkAI5fxbdvZb.'
WHERE user_id = 'user_d6e3b5f8c2';

-- 更新管理员账号密码（admin: 123456）
UPDATE users
SET username = 'admin',
    password_hash = '$2a$10$l1w0hTFJIKRye287UTa5ie4CO.LcjvYWAbwOKm.wdkAI5fxbdvZb.'
WHERE user_id = 'user_a1b2c3d4e5';

-- ============================================================================
-- 3. Demo 租户及相关数据
-- ============================================================================

-- 创建 Demo 租户
INSERT INTO tenants (tenant_id, tenant_code, name, tenant_type, preferred_currency, verified, status)
VALUES ('tenant_demo_001', 'demo_tenant', 'Demo 演示租户', 'ENTERPRISE', 'CNY', 1, 1)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  tenant_type = VALUES(tenant_type);

-- 创建 Demo 组织
INSERT INTO organizations (organization_id, tenant_id, org_code, name, org_type, level, status)
VALUES ('org_demo_001', 'tenant_demo_001', 'demo_org', 'Demo 演示组织', 'DEPARTMENT', 1, 1)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  org_type = VALUES(org_type);

-- 创建 Demo 项目
INSERT INTO projects (project_id, tenant_id, organization_id, project_code, name, description, status)
VALUES ('proj_demo_001', 'tenant_demo_001', 'org_demo_001', 'demo_project', 'Demo 演示项目', 'Demo 演示项目，用于测试订单创建', 1)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  description = VALUES(description);

-- 创建 Demo 用户（密码：123456）
INSERT INTO users (user_id, tenant_id, is_primary, username, real_name, email, phone, password_hash, status)
VALUES ('user_demo_001', 'tenant_demo_001', 1, 'demouser', 'Demo 用户', 'demo@example.com', '13800138000', '$2a$10$l1w0hTFJIKRye287UTa5ie4CO.LcjvYWAbwOKm.wdkAI5fxbdvZb.', 1)
ON DUPLICATE KEY UPDATE
  real_name = VALUES(real_name),
  email = VALUES(email);

-- 创建 Demo 租户的账户余额
INSERT INTO account_balances (tenant_id, balance, frozen_balance, credit_limit, currency)
VALUES ('tenant_demo_001', 0, 0, 0, 'CNY')
ON DUPLICATE KEY UPDATE
  tenant_id = VALUES(tenant_id);

-- ============================================================================
-- 4. 汇率数据初始化
-- ============================================================================

-- 创建汇率表（如果不存在）
CREATE TABLE IF NOT EXISTS exchange_rates (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  from_currency VARCHAR(8) NOT NULL COMMENT '源货币',
  to_currency VARCHAR(8) NOT NULL COMMENT '目标货币',
  rate DECIMAL(18,8) NOT NULL COMMENT '汇率',
  effective_date DATE NOT NULL COMMENT '生效日期',
  source VARCHAR(64) NULL COMMENT '汇率来源',
  created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_from_to_date (from_currency, to_currency, effective_date),
  KEY idx_effective_date (effective_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='汇率表';

-- 插入初始汇率数据（CNY 到其他货币）
INSERT INTO exchange_rates (from_currency, to_currency, rate, effective_date, source)
VALUES
  ('CNY', 'USD', 0.1385, '2026-01-17', 'BANK'),
  ('CNY', 'EUR', 0.1275, '2026-01-17', 'BANK'),
  ('CNY', 'JPY', 20.35, '2026-01-17', 'BANK'),
  ('CNY', 'GBP', 0.1098, '2026-01-17', 'BANK'),
  ('CNY', 'HKD', 1.085, '2026-01-17', 'BANK')
ON DUPLICATE KEY UPDATE
  rate = VALUES(rate);

-- 添加反向汇率（其他货币到 CNY）
INSERT INTO exchange_rates (from_currency, to_currency, rate, effective_date, source)
VALUES
  ('USD', 'CNY', 7.22, '2026-01-19', 'BANK'),
  ('EUR', 'CNY', 7.84, '2026-01-19', 'BANK'),
  ('JPY', 'CNY', 0.049, '2026-01-19', 'BANK'),
  ('GBP', 'CNY', 9.11, '2026-01-19', 'BANK'),
  ('HKD', 'CNY', 0.92, '2026-01-19', 'BANK')
ON DUPLICATE KEY UPDATE
  rate = VALUES(rate);

-- ============================================================================
-- 5. 测试数据
-- ============================================================================

-- 插入余额变动测试记录（仅用于测试余额变动历史功能）
-- 注意：此记录与实际业务数据不完全对应，仅用于前端展示测试
INSERT INTO balance_transactions (
  transaction_id,
  tenant_id,
  transaction_type,
  amount,
  balance_before,
  balance_after,
  related_order_id,
  remark,
  created_at
)
VALUES (
  'trans_test001',
  'tenant_demo_001',
  'DEDUCT',
  -100,
  1000,
  900,
  'order_test001',
  '测试支付 - 订单 order_test001',
  NOW()
)
ON DUPLICATE KEY UPDATE
  transaction_id = VALUES(transaction_id);

-- ============================================================================
-- 说明
-- ============================================================================

/*
1. 用户密码：
   - testuser / 123456
   - admin / 123456
   - demouser / 123456
   - 密码使用 bcrypt 算法加密，cost=10

2. Demo 租户体系：
   - tenant_demo_001: Demo 演示租户（企业类型）
   - org_demo_001: Demo 演示组织
   - proj_demo_001: Demo 演示项目
   - user_demo_001: Demo 用户
   - 用于前端订单创建等功能的测试

3. 汇率数据：
   - 支持 CNY、USD、EUR、JPY、GBP、HKD 相互转换
   - 汇率数据需要定期更新以保持准确性
   - 生效日期用于查询特定日期的汇率

4. 数据库表修改：
   - orders 表添加了汇率相关字段（exchange_rate, base_currency, base_currency_amount）
   - payments 表添加了汇率相关字段（exchange_rate, base_currency, base_currency_amount）
   - 这些字段用于多币种订单和支付的本位币换算

5. 测试数据：
   - balance_transactions 中的测试记录可以在生产环境中删除
   - 实际的余额变动记录由支付流程自动创建
*/
