-- Happy Billing 定价模块数据库表
-- 创建时间: 2026-01-17
-- 说明: 定价规则、折扣规则表

-- ============================================================================
-- 1. 定价规则表 (price_rules)
-- 说明: 支持固定价格、阶梯价格、时段价格、资源包等多种定价模式
-- ============================================================================
CREATE TABLE IF NOT EXISTS `price_rules` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `rule_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: PRICE20240117001',
  `rule_code` VARCHAR(64) UNIQUE NOT NULL COMMENT '规则编码',
  `rule_name` VARCHAR(255) NOT NULL COMMENT '规则名称',

  -- 关联产品
  `spu_code` VARCHAR(64) COMMENT '关联 SPU',
  `sku_code` VARCHAR(64) COMMENT '关联 SKU (优先级更高)',

  `rule_type` VARCHAR(32) NOT NULL COMMENT '定价类型: FIXED/TIERED/TIME_BASED/PACKAGE',
  `pricing_detail` JSON NOT NULL COMMENT '定价详情',

  `currency` VARCHAR(8) DEFAULT 'CNY' COMMENT '货币类型',

  `effective_start` TIMESTAMP NOT NULL COMMENT '生效开始时间',
  `effective_end` TIMESTAMP NULL COMMENT '生效结束时间',

  `region` VARCHAR(64) COMMENT '地域限定',
  `priority` INT DEFAULT 0 COMMENT '优先级',
  `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',

  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_rule_id` (`rule_id`),
  INDEX `idx_spu` (`spu_code`),
  INDEX `idx_sku` (`sku_code`),
  INDEX `idx_region` (`region`),
  INDEX `idx_priority` (`priority`),
  INDEX `idx_status` (`status`),
  INDEX `idx_effective` (`effective_start`, `effective_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='定价规则表';

-- ============================================================================
-- 2. 折扣规则表 (discount_rules)
-- 说明: 支持百分比折扣和金额折扣，可针对租户/SPU/SKU级别设置
-- ============================================================================
CREATE TABLE IF NOT EXISTS `discount_rules` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `discount_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: DISC20240117001',
  `discount_name` VARCHAR(255) NOT NULL COMMENT '折扣名称',

  `discount_type` VARCHAR(32) NOT NULL COMMENT '折扣类型: PERCENTAGE/AMOUNT',
  `discount_value` DECIMAL(18,4) NOT NULL COMMENT '折扣值',

  `target_type` VARCHAR(32) NOT NULL COMMENT '目标类型: TENANT/SPU/SKU',
  `target_id` VARCHAR(64) COMMENT '目标ID',

  `effective_start` TIMESTAMP NOT NULL COMMENT '生效开始时间',
  `effective_end` TIMESTAMP NULL COMMENT '生效结束时间',

  `max_discount` DECIMAL(18,4) COMMENT '最大折扣金额',
  `min_amount` DECIMAL(18,4) COMMENT '最小消费金额',

  `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_discount_id` (`discount_id`),
  INDEX `idx_target` (`target_type`, `target_id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_effective` (`effective_start`, `effective_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='折扣规则表';

-- ============================================================================
-- 初始化数据
-- ============================================================================

-- 插入固定价格规则: A100 GPU
INSERT INTO `price_rules` (
  `rule_id`, `rule_code`, `rule_name`, `sku_code`, `rule_type`,
  `pricing_detail`, `currency`, `effective_start`
)
VALUES (
  'PRICE20240117001',
  'PRICE_GPU_A100_FIXED',
  'A100 GPU 固定价格',
  'SKU_GPU_A100_40GB_BJ_25G',
  'FIXED',
  '{"unit_price":10.00}',
  'CNY',
  '2024-01-01 00:00:00'
);

-- 插入阶梯价格规则: 对象存储
INSERT INTO `price_rules` (
  `rule_id`, `rule_code`, `rule_name`, `spu_code`, `rule_type`,
  `pricing_detail`, `currency`, `effective_start`
)
VALUES (
  'PRICE20240117002',
  'PRICE_STORAGE_TIERED',
  '存储阶梯价格',
  'SPU_STORAGE_OBJ',
  'TIERED',
  '{
    "tiers": [
      {"from":0, "to":100, "unit_price":1.00},
      {"from":100, "to":500, "unit_price":0.80},
      {"from":500, "to":null, "unit_price":0.60}
    ]
  }',
  'CNY',
  '2024-01-01 00:00:00'
);

-- 插入时段价格规则: GPU 峰谷价格
INSERT INTO `price_rules` (
  `rule_id`, `rule_code`, `rule_name`, `spu_code`, `rule_type`,
  `pricing_detail`, `currency`, `effective_start`
)
VALUES (
  'PRICE20240117003',
  'PRICE_GPU_TIME_BASED',
  'GPU 峰谷价格',
  'SPU_GPU_A100',
  'TIME_BASED',
  '{
    "periods": [
      {"hour_from":0, "hour_to":8, "unit_price":5.00},
      {"hour_from":8, "hour_to":20, "unit_price":10.00},
      {"hour_from":20, "hour_to":24, "unit_price":5.00}
    ]
  }',
  'CNY',
  '2024-01-01 00:00:00'
);

-- 插入资源包价格规则: LLM Token
INSERT INTO `price_rules` (
  `rule_id`, `rule_code`, `rule_name`, `spu_code`, `rule_type`,
  `pricing_detail`, `currency`, `effective_start`
)
VALUES (
  'PRICE20240117004',
  'PRICE_LLM_PACKAGE',
  'LLM Token 资源包',
  'SPU_LLM_GPT4',
  'PACKAGE',
  '{
    "package_size": 1000000,
    "package_price": 800.00,
    "validity_days": 30
  }',
  'CNY',
  '2024-01-01 00:00:00'
);

-- 插入折扣规则: VIP客户折扣
INSERT INTO `discount_rules` (
  `discount_id`, `discount_name`, `discount_type`, `discount_value`,
  `target_type`, `target_id`, `effective_start`
)
VALUES (
  'DISC20240117001',
  'VIP客户折扣',
  'PERCENTAGE',
  0.20,
  'TENANT',
  'tenant_a3f9b2c4d5',
  '2024-01-01 00:00:00'
);

-- 插入折扣规则: GPU促销活动
INSERT INTO `discount_rules` (
  `discount_id`, `discount_name`, `discount_type`, `discount_value`,
  `target_type`, `target_id`, `effective_start`, `min_amount`
)
VALUES (
  'DISC20240117002',
  'GPU促销活动',
  'AMOUNT',
  200.00,
  'SPU',
  'SPU_GPU_A100',
  '2024-01-01 00:00:00',
  1000.00
);
