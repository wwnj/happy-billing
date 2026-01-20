-- ============================================================================
-- Happy Billing MySQL 数据库初始化脚本
-- 说明: 按顺序执行所有 DDL 和 DML 脚本
-- 创建时间: 2026-01-20
-- ============================================================================

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================================
-- 第一步：创建数据库（如果不存在）
-- ============================================================================
CREATE DATABASE IF NOT EXISTS `happy_billing`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE `happy_billing`;

-- ============================================================================
-- 第二步：DDL - 创建表结构（按依赖顺序）
-- ============================================================================

-- 2.1 租户模块表（基础表，无外键依赖）
SOURCE /docker-entrypoint-initdb.d/20260117_create_tenant_tables.sql;

-- 2.2 产品模块表（依赖租户）
SOURCE /docker-entrypoint-initdb.d/20260117_create_product_tables.sql;

-- 2.3 定价模块表（依赖产品）
SOURCE /docker-entrypoint-initdb.d/20260117_create_pricing_tables.sql;

-- 2.4 订单/账单/支付模块表（依赖租户、产品、定价）
SOURCE /docker-entrypoint-initdb.d/20260117_create_order_billing_payment_tables.sql;

-- 2.5 汇率表（独立表）
SOURCE /docker-entrypoint-initdb.d/20240117_create_exchange_rates.sql;

-- 2.6 多币种字段（表字段扩展）
SOURCE /docker-entrypoint-initdb.d/20240117_add_multi_currency_fields.sql;

-- 2.7 外键修复（如果有）
SOURCE /docker-entrypoint-initdb.d/fix_foreign_keys.sql;

-- ============================================================================
-- 第三步：DML - 插入测试数据
-- ============================================================================

-- 3.1 插入测试用户凭证
SOURCE /docker-entrypoint-initdb.d/add_test_users_credentials.sql;

-- 3.2 插入完整测试数据
SOURCE /docker-entrypoint-initdb.d/debug_and_test_data.sql;

-- ============================================================================
-- 初始化完成
-- ============================================================================

SET FOREIGN_KEY_CHECKS = 1;

SELECT '✅ MySQL 数据库初始化完成！' AS status;
SELECT CONCAT('   - 数据库: ', DATABASE()) AS info;
SELECT CONCAT('   - 表数量: ', COUNT(*)) AS tables_count FROM information_schema.tables WHERE table_schema = 'happy_billing';
