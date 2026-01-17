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
