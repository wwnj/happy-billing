-- Happy Billing 订单、账单、支付、资源实例模块数据库表
-- 创建时间: 2026-01-17
-- 说明: 订单主表、订单明细、资源实例、账单、支付记录等核心交易表

-- ============================================================================
-- 1. 订单主表 (orders)
-- 说明: 包年包月和按量计费订单
-- ============================================================================
CREATE TABLE IF NOT EXISTS `orders` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `order_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: order_xxx',
  `order_no` VARCHAR(64) UNIQUE NOT NULL COMMENT '展示给用户的订单号',

  -- 租户信息
  `tenant_id` VARCHAR(64) NOT NULL COMMENT '所属租户ID',
  `organization_id` VARCHAR(64) COMMENT '所属组织ID',
  `project_id` VARCHAR(64) NOT NULL COMMENT '所属项目ID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '下单用户ID',

  -- 订单类型
  `order_type` VARCHAR(32) NOT NULL COMMENT '订单类型: PREPAID/POSTPAID',

  -- 关联产品
  `spu_code` VARCHAR(64) NOT NULL COMMENT 'SPU编码',
  `sku_code` VARCHAR(64) NOT NULL COMMENT 'SKU编码',

  -- 金额
  `currency` VARCHAR(8) DEFAULT 'CNY' COMMENT '货币类型',
  `original_amount` DECIMAL(18,4) NOT NULL COMMENT '原价',
  `discount_amount` DECIMAL(18,4) DEFAULT 0 COMMENT '折扣金额',
  `payable_amount` DECIMAL(18,4) NOT NULL COMMENT '应付金额',
  `paid_amount` DECIMAL(18,4) DEFAULT 0 COMMENT '已付金额',

  -- 服务期(预付费)
  `period_start` TIMESTAMP NULL COMMENT '服务开始时间',
  `period_end` TIMESTAMP NULL COMMENT '服务结束时间',

  -- 订单状态
  `status` VARCHAR(32) NOT NULL COMMENT '订单状态: PENDING/PAID/ACTIVE/EXPIRED/CANCELLED/REFUNDED',
  `order_detail` JSON COMMENT '订单详情',

  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_order_id` (`order_id`),
  INDEX `idx_tenant` (`tenant_id`),
  INDEX `idx_project` (`project_id`),
  INDEX `idx_user` (`user_id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单主表';

-- ============================================================================
-- 2. 订单明细表 (order_items)
-- 说明: 订单商品明细,支持一单多品
-- ============================================================================
CREATE TABLE IF NOT EXISTS `order_items` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `order_id` VARCHAR(64) NOT NULL COMMENT '订单ID',
  `item_no` VARCHAR(64) NOT NULL COMMENT '明细编号',

  `spu_code` VARCHAR(64) NOT NULL COMMENT 'SPU编码',
  `sku_code` VARCHAR(64) NOT NULL COMMENT 'SKU编码',
  `sku_name` VARCHAR(255) NOT NULL COMMENT 'SKU名称(快照)',
  `sku_spec` JSON COMMENT 'SKU规格(快照)',

  `quantity` DECIMAL(18,4) NOT NULL COMMENT '数量',
  `unit_price` DECIMAL(18,4) NOT NULL COMMENT '单价',
  `amount` DECIMAL(18,4) NOT NULL COMMENT '金额',

  `price_rule_id` VARCHAR(64) COMMENT '定价规则ID',

  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

  INDEX `idx_order` (`order_id`),
  INDEX `idx_sku` (`sku_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单明细表';

-- ============================================================================
-- 3. 资源实例表 (resource_instances)
-- 说明: 实际分配的资源实例
-- ============================================================================
CREATE TABLE IF NOT EXISTS `resource_instances` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `instance_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: resource_xxx',

  `order_id` VARCHAR(64) NOT NULL COMMENT '来源订单ID',
  `tenant_id` VARCHAR(64) NOT NULL COMMENT '所属租户ID',
  `project_id` VARCHAR(64) NOT NULL COMMENT '所属项目ID',

  `product_type` VARCHAR(32) NOT NULL COMMENT '产品类型: GPU/STORAGE/LLM_TOKEN',
  `spu_code` VARCHAR(64) NOT NULL COMMENT 'SPU编码',
  `sku_code` VARCHAR(64) NOT NULL COMMENT 'SKU编码',

  `instance_spec` JSON COMMENT '实例规格',

  `status` VARCHAR(32) NOT NULL COMMENT '状态: CREATING/RUNNING/STOPPED/DELETED',

  `started_at` TIMESTAMP NULL COMMENT '启动时间',
  `stopped_at` TIMESTAMP NULL COMMENT '停止时间',

  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_instance_id` (`instance_id`),
  INDEX `idx_order` (`order_id`),
  INDEX `idx_tenant_project` (`tenant_id`, `project_id`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='资源实例表';

-- ============================================================================
-- 4. 账单表 (bills)
-- 说明: 账单汇总数据
-- ============================================================================
CREATE TABLE IF NOT EXISTS `bills` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `bill_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: bill_xxx',

  `order_id` VARCHAR(64) NOT NULL COMMENT '订单ID',
  `tenant_id` VARCHAR(64) NOT NULL COMMENT '租户ID',
  `project_id` VARCHAR(64) NOT NULL COMMENT '项目ID',

  `bill_type` VARCHAR(32) NOT NULL COMMENT '账单类型: SUBSCRIPTION/HOURLY/REFUND',

  -- 计费周期(后付费)
  `billing_period_start` TIMESTAMP NULL COMMENT '计费周期开始',
  `billing_period_end` TIMESTAMP NULL COMMENT '计费周期结束',

  -- 金额
  `currency` VARCHAR(8) DEFAULT 'CNY' COMMENT '货币类型',
  `exchange_rate` DECIMAL(18,8) COMMENT '汇率',
  `base_currency` VARCHAR(8) DEFAULT 'CNY' COMMENT '基础货币',
  `base_currency_amount` DECIMAL(18,4) COMMENT '基础货币金额',

  `original_amount` DECIMAL(18,4) NOT NULL COMMENT '原价',
  `discount_amount` DECIMAL(18,4) DEFAULT 0 COMMENT '折扣金额',
  `payable_amount` DECIMAL(18,4) NOT NULL COMMENT '应付金额',

  -- 账单状态
  `status` VARCHAR(32) NOT NULL COMMENT '状态: UNPAID/PAID/OVERDUE/CANCELLED',

  `bill_detail` JSON COMMENT '账单详情',

  `paid_at` TIMESTAMP NULL COMMENT '支付时间',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_bill_id` (`bill_id`),
  INDEX `idx_order` (`order_id`),
  INDEX `idx_tenant` (`tenant_id`),
  INDEX `idx_project` (`project_id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_period` (`billing_period_start`, `billing_period_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='账单表';

-- ============================================================================
-- 5. 支付记录表 (payments)
-- 说明: 支付流水记录
-- ============================================================================
CREATE TABLE IF NOT EXISTS `payments` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `payment_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: pay_xxx',

  `order_id` VARCHAR(64) COMMENT '订单ID',
  `bill_id` VARCHAR(64) COMMENT '账单ID',
  `tenant_id` VARCHAR(64) NOT NULL COMMENT '租户ID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',

  `payment_method` VARCHAR(32) NOT NULL COMMENT '支付方式: BALANCE/ALIPAY/WECHAT/BANK',
  `payment_channel` VARCHAR(64) COMMENT '支付渠道',

  `currency` VARCHAR(8) DEFAULT 'CNY' COMMENT '货币类型',
  `amount` DECIMAL(18,4) NOT NULL COMMENT '支付金额',

  `status` VARCHAR(32) NOT NULL COMMENT '状态: PENDING/SUCCESS/FAILED/REFUNDED/CANCELLED',

  `external_order_id` VARCHAR(128) COMMENT '第三方订单号',
  `external_response` JSON COMMENT '第三方响应',

  `paid_at` TIMESTAMP NULL COMMENT '支付完成时间',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_payment_id` (`payment_id`),
  INDEX `idx_order` (`order_id`),
  INDEX `idx_bill` (`bill_id`),
  INDEX `idx_tenant` (`tenant_id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_external` (`external_order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付记录表';

-- ============================================================================
-- 6. 账户余额表 (account_balances)
-- 说明: 租户账户余额
-- ============================================================================
CREATE TABLE IF NOT EXISTS `account_balances` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `tenant_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '租户ID',

  `balance` DECIMAL(18,4) NOT NULL DEFAULT 0 COMMENT '可用余额',
  `frozen_balance` DECIMAL(18,4) NOT NULL DEFAULT 0 COMMENT '冻结余额',
  `credit_limit` DECIMAL(18,4) NOT NULL DEFAULT 0 COMMENT '信用额度',

  `currency` VARCHAR(8) DEFAULT 'CNY' COMMENT '货币类型',

  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

  INDEX `idx_tenant` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='账户余额表';

-- ============================================================================
-- 7. 余额变动记录表 (balance_transactions)
-- 说明: 余额变动流水
-- ============================================================================
CREATE TABLE IF NOT EXISTS `balance_transactions` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `transaction_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键',

  `tenant_id` VARCHAR(64) NOT NULL COMMENT '租户ID',

  `transaction_type` VARCHAR(32) NOT NULL COMMENT '交易类型: RECHARGE/DEDUCT/REFUND/FREEZE/UNFREEZE',
  `amount` DECIMAL(18,4) NOT NULL COMMENT '金额(正数充值,负数扣减)',

  `balance_before` DECIMAL(18,4) NOT NULL COMMENT '变动前余额',
  `balance_after` DECIMAL(18,4) NOT NULL COMMENT '变动后余额',

  `related_order_id` VARCHAR(64) COMMENT '关联订单ID',
  `related_bill_id` VARCHAR(64) COMMENT '关联账单ID',
  `related_payment_id` VARCHAR(64) COMMENT '关联支付ID',

  `remark` VARCHAR(512) COMMENT '备注',

  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

  INDEX `idx_transaction_id` (`transaction_id`),
  INDEX `idx_tenant` (`tenant_id`),
  INDEX `idx_type` (`transaction_type`),
  INDEX `idx_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='余额变动记录表';

-- ============================================================================
-- 初始化数据
-- ============================================================================

-- 插入测试订单
INSERT INTO `orders` (
  `order_id`, `order_no`, `tenant_id`, `project_id`, `user_id`,
  `order_type`, `spu_code`, `sku_code`,
  `original_amount`, `payable_amount`, `status`
)
VALUES (
  'order_test001',
  'ORD20240117000001',
  'tenant_a3f9b2c4d5',
  'proj_c5d2a4e7b1',
  'user_d6e3b5f8c2',
  'PREPAID',
  'SPU_GPU_A100',
  'SKU_GPU_A100_40GB_BJ_25G',
  12000.00,
  12000.00,
  'PENDING'
);

-- 插入测试账户余额
INSERT INTO `account_balances` (`tenant_id`, `balance`, `currency`)
VALUES ('tenant_a3f9b2c4d5', 50000.00, 'CNY');

INSERT INTO `account_balances` (`tenant_id`, `balance`, `currency`)
VALUES ('tenant_e8d7c2a1f4', 100000.00, 'CNY');
