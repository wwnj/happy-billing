-- Happy Billing 数据库完整初始化脚本
-- 自动执行：容器首次启动时
-- 版本: v1.0
-- 日期: 2026-01-17

USE happy_billing;

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- Happy Billing 租户模块数据库表
-- 创建时间: 2026-01-17
-- 说明: 租户、组织、项目、用户、认证表

-- ============================================================================
-- 1. 租户表 (tenants)
-- 说明: 租户是系统的顶层隔离单位，支持企业和个人用户
-- ============================================================================
CREATE TABLE IF NOT EXISTS `tenants` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `tenant_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: tenant_a3f9b2c4d5',
  `tenant_code` VARCHAR(64) UNIQUE NOT NULL COMMENT '租户编码',
  `name` VARCHAR(255) NOT NULL COMMENT '租户名称',
  `tenant_type` VARCHAR(32) NOT NULL COMMENT '租户类型: ENTERPRISE/INDIVIDUAL',
  `preferred_currency` VARCHAR(8) DEFAULT 'CNY' COMMENT '首选币种',

  `verified` TINYINT DEFAULT 0 COMMENT '是否已认证: 0-未认证, 1-已认证',
  `verified_type` VARCHAR(32) COMMENT '认证类型: INDIVIDUAL/ENTERPRISE',
  `verified_at` TIMESTAMP NULL COMMENT '认证时间',
  `verified_info` JSON COMMENT '认证信息',

  `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_tenant_id` (`tenant_id`),
  INDEX `idx_type` (`tenant_type`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户表';

-- ============================================================================
-- 2. 组织表 (organizations)
-- 说明: 组织是企业主体，支持树形结构（部门层级）
-- ============================================================================
CREATE TABLE IF NOT EXISTS `organizations` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `organization_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: org_b4e1c3f2a6',

  `tenant_id` VARCHAR(64) NOT NULL COMMENT '所属租户ID',
  `parent_organization_id` VARCHAR(64) COMMENT '父组织ID（树形结构）',

  `org_code` VARCHAR(64) UNIQUE NOT NULL COMMENT '组织编码',
  `name` VARCHAR(255) NOT NULL COMMENT '组织名称',
  `org_type` VARCHAR(32) COMMENT '组织类型: 总部/分公司/部门等',
  `level` INT DEFAULT 1 COMMENT '组织层级',

  `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_organization_id` (`organization_id`),
  INDEX `idx_tenant` (`tenant_id`),
  INDEX `idx_parent` (`parent_organization_id`),
  FOREIGN KEY (`tenant_id`) REFERENCES `tenants`(`tenant_id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组织表';

-- ============================================================================
-- 3. 项目表 (projects)
-- 说明: 项目是成本中心/工作区，用于资源隔离和成本核算
-- ============================================================================
CREATE TABLE IF NOT EXISTS `projects` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `project_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: proj_c5d2a4e7b1',

  `tenant_id` VARCHAR(64) NOT NULL COMMENT '所属租户ID',
  `organization_id` VARCHAR(64) NOT NULL COMMENT '所属组织ID',

  `project_code` VARCHAR(64) UNIQUE NOT NULL COMMENT '项目编码',
  `name` VARCHAR(255) NOT NULL COMMENT '项目名称',
  `description` TEXT COMMENT '项目描述',

  `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_project_id` (`project_id`),
  INDEX `idx_tenant` (`tenant_id`),
  INDEX `idx_org` (`organization_id`),
  INDEX `idx_status` (`status`),
  FOREIGN KEY (`tenant_id`) REFERENCES `tenants`(`tenant_id`) ON DELETE RESTRICT,
  FOREIGN KEY (`organization_id`) REFERENCES `organizations`(`organization_id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目表';

-- ============================================================================
-- 4. 用户表 (users)
-- 说明: 用户是实际使用者，个人用户有主账号，企业用户可以有多个员工账号
-- ============================================================================
CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `user_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: user_d6e3b5f8c2',

  `tenant_id` VARCHAR(64) NOT NULL COMMENT '所属租户ID',

  `is_primary` TINYINT DEFAULT 0 COMMENT '是否主账号: 0-否, 1-是',
  `username` VARCHAR(128) COMMENT '用户名',
  `real_name` VARCHAR(128) COMMENT '真实姓名',
  `id_card` VARCHAR(64) COMMENT '身份证号',

  `email` VARCHAR(255) COMMENT '邮箱',
  `phone` VARCHAR(32) COMMENT '手机号',
  `password_hash` VARCHAR(255) COMMENT '密码哈希',

  `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_tenant` (`tenant_id`),
  INDEX `idx_email` (`email`),
  INDEX `idx_phone` (`phone`),
  FOREIGN KEY (`tenant_id`) REFERENCES `tenants`(`tenant_id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ============================================================================
-- 5. 认证表 (verifications)
-- 说明: 实名认证记录，支持个人认证和企业认证
-- ============================================================================
CREATE TABLE IF NOT EXISTS `verifications` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `verification_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: verify_e7f4c6a9d3',

  `tenant_id` VARCHAR(64) NOT NULL COMMENT '所属租户ID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '申请用户ID',

  `verify_type` VARCHAR(32) NOT NULL COMMENT '认证类型: INDIVIDUAL/ENTERPRISE',

  -- 个人认证字段
  `real_name` VARCHAR(128) COMMENT '真实姓名',
  `id_card` VARCHAR(64) COMMENT '身份证号',
  `id_card_front` VARCHAR(512) COMMENT '身份证正面照URL',
  `id_card_back` VARCHAR(512) COMMENT '身份证背面照URL',

  -- 企业认证字段
  `company_name` VARCHAR(255) COMMENT '公司名称',
  `credit_code` VARCHAR(64) COMMENT '统一社会信用代码',
  `license_url` VARCHAR(512) COMMENT '营业执照URL',
  `legal_person` VARCHAR(128) COMMENT '法人代表',

  `status` VARCHAR(32) NOT NULL DEFAULT 'PENDING' COMMENT '状态: PENDING/APPROVED/REJECTED',
  `reject_reason` TEXT COMMENT '拒绝原因',
  `verified_at` TIMESTAMP NULL COMMENT '审核通过时间',
  `verified_by` VARCHAR(64) COMMENT '审核人',

  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_verification_id` (`verification_id`),
  INDEX `idx_tenant` (`tenant_id`),
  INDEX `idx_user` (`user_id`),
  INDEX `idx_status` (`status`),
  FOREIGN KEY (`tenant_id`) REFERENCES `tenants`(`tenant_id`) ON DELETE RESTRICT,
  FOREIGN KEY (`user_id`) REFERENCES `users`(`user_id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='认证表';

-- ============================================================================
-- 初始化数据
-- ============================================================================

-- 插入测试租户（个人开发者）
INSERT INTO `tenants` (`tenant_id`, `tenant_code`, `name`, `tenant_type`, `verified`)
VALUES ('tenant_a3f9b2c4d5', 'test_individual', '测试个人开发者', 'INDIVIDUAL', 1);

-- 插入测试租户（企业用户）
INSERT INTO `tenants` (`tenant_id`, `tenant_code`, `name`, `tenant_type`, `verified`)
VALUES ('tenant_e8d7c2a1f4', 'test_enterprise', '测试企业用户', 'ENTERPRISE', 1);

-- 插入测试组织（个人工作区）
INSERT INTO `organizations` (`organization_id`, `tenant_id`, `org_code`, `name`, `org_type`)
VALUES ('org_b4e1c3f2a6', 'tenant_a3f9b2c4d5', 'personal_workspace', '个人工作区', 'PERSONAL');

-- 插入测试组织（企业总部）
INSERT INTO `organizations` (`organization_id`, `tenant_id`, `org_code`, `name`, `org_type`)
VALUES ('org_f7a2d5e8c1', 'tenant_e8d7c2a1f4', 'headquarters', '总部', 'HEADQUARTERS');

-- 插入测试项目
INSERT INTO `projects` (`project_id`, `tenant_id`, `organization_id`, `project_code`, `name`)
VALUES ('proj_c5d2a4e7b1', 'tenant_a3f9b2c4d5', 'org_b4e1c3f2a6', 'default_project', '默认项目');

INSERT INTO `projects` (`project_id`, `tenant_id`, `organization_id`, `project_code`, `name`)
VALUES ('proj_a9f3e6b2c8', 'tenant_e8d7c2a1f4', 'org_f7a2d5e8c1', 'ai_platform', 'AI平台项目');

-- 插入测试用户
INSERT INTO `users` (`user_id`, `tenant_id`, `is_primary`, `real_name`, `email`, `phone`)
VALUES ('user_d6e3b5f8c2', 'tenant_a3f9b2c4d5', 1, '张三', 'zhangsan@example.com', '13800138000');

INSERT INTO `users` (`user_id`, `tenant_id`, `is_primary`, `real_name`, `email`, `phone`)
VALUES ('user_b1f4a7d9e3', 'tenant_e8d7c2a1f4', 1, '李四', 'lisi@example.com', '13900139000');

-- Happy Billing 产品模块数据库表
-- 创建时间: 2026-01-17
-- 说明: 产品分类、SPU、SKU表 (基于电商行业成熟的SPU/SKU模型)

-- ============================================================================
-- 1. 产品分类表 (product_categories)
-- 说明: 产品分类支持树形结构,用于产品的层级分类管理
-- ============================================================================
CREATE TABLE IF NOT EXISTS `product_categories` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `category_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: CAT20240117001',
  `category_code` VARCHAR(64) UNIQUE NOT NULL COMMENT '分类编码',
  `category_name` VARCHAR(255) NOT NULL COMMENT '分类名称',

  `parent_id` BIGINT COMMENT '父分类ID(树形结构)',
  `level` TINYINT COMMENT '层级: 1,2,3',
  `sort_order` INT DEFAULT 0 COMMENT '排序顺序',
  `icon` VARCHAR(512) COMMENT '分类图标URL',

  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

  INDEX `idx_category_id` (`category_id`),
  INDEX `idx_parent` (`parent_id`),
  INDEX `idx_level` (`level`),
  INDEX `idx_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品分类表';

-- ============================================================================
-- 2. SPU 表 (product_spu) - 标准产品单元/产品族
-- 说明: SPU是产品族,统一管理产品的通用信息和规格模板
-- ============================================================================
CREATE TABLE IF NOT EXISTS `product_spu` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `spu_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: SPU20240117001',
  `spu_code` VARCHAR(64) UNIQUE NOT NULL COMMENT 'SPU编码: SPU_GPU_A100',
  `spu_name` VARCHAR(255) NOT NULL COMMENT 'SPU名称: NVIDIA A100 GPU',

  `category_id` BIGINT NOT NULL COMMENT '所属分类ID',
  `product_type` VARCHAR(32) NOT NULL COMMENT '产品类型: GPU/STORAGE/LLM_TOKEN',

  `brand` VARCHAR(128) COMMENT '品牌: NVIDIA',
  `description` TEXT COMMENT '产品描述',

  `billing_unit` VARCHAR(32) NOT NULL COMMENT '计费单位: HOUR/GB/TOKEN',

  `spec_template` JSON COMMENT '规格模板',

  `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_spu_id` (`spu_id`),
  INDEX `idx_category` (`category_id`),
  INDEX `idx_type` (`product_type`),
  INDEX `idx_status` (`status`),
  FOREIGN KEY (`category_id`) REFERENCES `product_categories`(`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SPU表(产品族)';

-- ============================================================================
-- 3. SKU 表 (product_sku) - 库存保管单元/可售卖规格
-- 说明: SKU是具体的可售卖规格,包含完整的规格参数和地域信息
-- ============================================================================
CREATE TABLE IF NOT EXISTS `product_sku` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '物理主键',
  `sku_id` VARCHAR(64) UNIQUE NOT NULL COMMENT '业务主键: SKU20240117000001',
  `sku_code` VARCHAR(64) UNIQUE NOT NULL COMMENT 'SKU编码: SKU_GPU_A100_40GB_BJ_25G',

  `spu_id` BIGINT NOT NULL COMMENT '所属SPU ID',
  `spu_code` VARCHAR(64) NOT NULL COMMENT 'SPU编码(冗余)',

  `sku_name` VARCHAR(255) NOT NULL COMMENT 'SKU名称: A100 40GB 北京 25Gbps',
  `spec_values` JSON NOT NULL COMMENT '具体规格值',

  `region` VARCHAR(64) COMMENT '地域: cn-beijing',
  `stock_type` VARCHAR(32) COMMENT '库存类型: AVAILABLE/SOLD_OUT',

  `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX `idx_sku_id` (`sku_id`),
  INDEX `idx_spu` (`spu_id`),
  INDEX `idx_spu_code` (`spu_code`),
  INDEX `idx_region` (`region`),
  INDEX `idx_status` (`status`),
  FOREIGN KEY (`spu_id`) REFERENCES `product_spu`(`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SKU表(可售卖规格)';

-- ============================================================================
-- 初始化数据
-- ============================================================================

-- 插入产品分类 (计算服务)
INSERT INTO `product_categories` (`category_id`, `category_code`, `category_name`, `parent_id`, `level`, `sort_order`)
VALUES ('CAT20240117001', 'COMPUTE', '计算服务', NULL, 1, 1);

-- 插入子分类 (GPU计算)
INSERT INTO `product_categories` (`category_id`, `category_code`, `category_name`, `parent_id`, `level`, `sort_order`)
VALUES ('CAT20240117002', 'GPU_COMPUTE', 'GPU计算', 1, 2, 1);

-- 插入子分类 (CPU计算)
INSERT INTO `product_categories` (`category_id`, `category_code`, `category_name`, `parent_id`, `level`, `sort_order`)
VALUES ('CAT20240117003', 'CPU_COMPUTE', 'CPU计算', 1, 2, 2);

-- 插入产品分类 (存储服务)
INSERT INTO `product_categories` (`category_id`, `category_code`, `category_name`, `parent_id`, `level`, `sort_order`)
VALUES ('CAT20240117004', 'STORAGE', '存储服务', NULL, 1, 2);

-- 插入子分类 (对象存储)
INSERT INTO `product_categories` (`category_id`, `category_code`, `category_name`, `parent_id`, `level`, `sort_order`)
VALUES ('CAT20240117005', 'OBJECT_STORAGE', '对象存储', 4, 2, 1);

-- 插入子分类 (块存储)
INSERT INTO `product_categories` (`category_id`, `category_code`, `category_name`, `parent_id`, `level`, `sort_order`)
VALUES ('CAT20240117006', 'BLOCK_STORAGE', '块存储', 4, 2, 2);

-- 插入产品分类 (AI服务)
INSERT INTO `product_categories` (`category_id`, `category_code`, `category_name`, `parent_id`, `level`, `sort_order`)
VALUES ('CAT20240117007', 'AI_SERVICE', 'AI服务', NULL, 1, 3);

-- 插入子分类 (大语言模型)
INSERT INTO `product_categories` (`category_id`, `category_code`, `category_name`, `parent_id`, `level`, `sort_order`)
VALUES ('CAT20240117008', 'LLM', '大语言模型', 7, 2, 1);

-- ============================================================================
-- 插入测试 SPU: NVIDIA A100 GPU
-- ============================================================================
INSERT INTO `product_spu` (
  `spu_id`, `spu_code`, `spu_name`, `category_id`, `product_type`,
  `brand`, `billing_unit`, `spec_template`, `description`
)
VALUES (
  'SPU20240117001',
  'SPU_GPU_A100',
  'NVIDIA A100 GPU',
  2, -- GPU计算分类
  'GPU',
  'NVIDIA',
  'HOUR',
  '{
    "memory": ["40GB", "80GB"],
    "network": ["25Gbps", "100Gbps"],
    "region": ["cn-beijing", "cn-shanghai", "us-west"]
  }',
  'NVIDIA A100 Tensor Core GPU 提供卓越的加速性能,适用于 AI、数据分析和高性能计算应用'
);

-- ============================================================================
-- 插入测试 SKU: A100 不同规格
-- ============================================================================

-- SKU1: A100 40GB 北京 25Gbps
INSERT INTO `product_sku` (
  `sku_id`, `sku_code`, `spu_id`, `spu_code`, `sku_name`,
  `spec_values`, `region`, `stock_type`
)
VALUES (
  'SKU20240117000001',
  'SKU_GPU_A100_40GB_BJ_25G',
  1,
  'SPU_GPU_A100',
  'A100 40GB 北京 25Gbps',
  '{"memory":"40GB","network":"25Gbps","region":"cn-beijing"}',
  'cn-beijing',
  'AVAILABLE'
);

-- SKU2: A100 80GB 北京 100Gbps
INSERT INTO `product_sku` (
  `sku_id`, `sku_code`, `spu_id`, `spu_code`, `sku_name`,
  `spec_values`, `region`, `stock_type`
)
VALUES (
  'SKU20240117000002',
  'SKU_GPU_A100_80GB_BJ_100G',
  1,
  'SPU_GPU_A100',
  'A100 80GB 北京 100Gbps',
  '{"memory":"80GB","network":"100Gbps","region":"cn-beijing"}',
  'cn-beijing',
  'AVAILABLE'
);

-- SKU3: A100 40GB 上海 25Gbps
INSERT INTO `product_sku` (
  `sku_id`, `sku_code`, `spu_id`, `spu_code`, `sku_name`,
  `spec_values`, `region`, `stock_type`
)
VALUES (
  'SKU20240117000003',
  'SKU_GPU_A100_40GB_SH_25G',
  1,
  'SPU_GPU_A100',
  'A100 40GB 上海 25Gbps',
  '{"memory":"40GB","network":"25Gbps","region":"cn-shanghai"}',
  'cn-shanghai',
  'AVAILABLE'
);

-- SKU4: A100 80GB 上海 100Gbps
INSERT INTO `product_sku` (
  `sku_id`, `sku_code`, `spu_id`, `spu_code`, `sku_name`,
  `spec_values`, `region`, `stock_type`
)
VALUES (
  'SKU20240117000004',
  'SKU_GPU_A100_80GB_SH_100G',
  1,
  'SPU_GPU_A100',
  'A100 80GB 上海 100Gbps',
  '{"memory":"80GB","network":"100Gbps","region":"cn-shanghai"}',
  'cn-shanghai',
  'AVAILABLE'
);

-- ============================================================================
-- 插入测试 SPU: NVIDIA V100 GPU
-- ============================================================================
INSERT INTO `product_spu` (
  `spu_id`, `spu_code`, `spu_name`, `category_id`, `product_type`,
  `brand`, `billing_unit`, `spec_template`, `description`
)
VALUES (
  'SPU20240117002',
  'SPU_GPU_V100',
  'NVIDIA V100 GPU',
  2, -- GPU计算分类
  'GPU',
  'NVIDIA',
  'HOUR',
  '{
    "memory": ["16GB", "32GB"],
    "network": ["25Gbps"],
    "region": ["cn-beijing", "cn-shanghai"]
  }',
  'NVIDIA V100 Tensor Core GPU 提供高性能计算能力,适用于深度学习训练和推理'
);

-- SKU5: V100 32GB 北京
INSERT INTO `product_sku` (
  `sku_id`, `sku_code`, `spu_id`, `spu_code`, `sku_name`,
  `spec_values`, `region`, `stock_type`
)
VALUES (
  'SKU20240117000005',
  'SKU_GPU_V100_32GB_BJ',
  2,
  'SPU_GPU_V100',
  'V100 32GB 北京',
  '{"memory":"32GB","network":"25Gbps","region":"cn-beijing"}',
  'cn-beijing',
  'AVAILABLE'
);

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

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- 显示完成信息
SELECT 'Happy Billing Database Initialization Completed!' AS Status;
SELECT 'Total Tables:' AS Info, COUNT(*) AS Count FROM information_schema.tables WHERE table_schema = 'happy_billing';

