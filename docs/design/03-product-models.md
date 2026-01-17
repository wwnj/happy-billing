# 产品模型设计 (SPU/SKU)

**文档版本**: v1.0  
**设计日期**: 2026-01-17

---

## 设计目标

采用电商行业成熟的 **SPU/SKU 模型**,实现产品的规格化管理和灵活定价。

### SPU vs SKU

- **SPU (Standard Product Unit)**: 标准产品单元,产品族
  - 例如: "NVIDIA A100 GPU"
- **SKU (Stock Keeping Unit)**: 库存保管单元,可售卖的具体规格
  - 例如: "A100 40GB 北京 25Gbps网络"

---

## 核心表设计

### 1. 产品分类表

```sql
CREATE TABLE product_categories (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  category_id     VARCHAR(64) UNIQUE NOT NULL,        -- CAT20240117001
  category_code   VARCHAR(64) UNIQUE NOT NULL,
  category_name   VARCHAR(255) NOT NULL,
  parent_id       BIGINT,                             -- 树形结构
  level           TINYINT,                            -- 层级 1,2,3
  sort_order      INT DEFAULT 0,
  icon            VARCHAR(512),
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  INDEX idx_parent (parent_id),
  INDEX idx_level (level)
);
```

**分类示例:**
```
计算服务 (L1)
  ├─ GPU计算 (L2)
  └─ CPU计算 (L2)
  
存储服务 (L1)
  ├─ 对象存储 (L2)
  └─ 块存储 (L2)
  
AI服务 (L1)
  └─ 大语言模型 (L2)
```

### 2. SPU 表 (产品族)

```sql
CREATE TABLE product_spu (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  spu_id          VARCHAR(64) UNIQUE NOT NULL,        -- SPU20240117001
  spu_code        VARCHAR(64) UNIQUE NOT NULL,        -- SPU_GPU_A100
  spu_name        VARCHAR(255) NOT NULL,              -- NVIDIA A100 GPU
  
  category_id     BIGINT NOT NULL,
  product_type    VARCHAR(32) NOT NULL,               -- GPU/STORAGE/LLM_TOKEN
  
  brand           VARCHAR(128),                       -- NVIDIA
  description     TEXT,
  
  billing_unit    VARCHAR(32) NOT NULL,               -- HOUR/GB/TOKEN
  
  spec_template   JSON,                               -- 规格模板
  
  status          TINYINT DEFAULT 1,
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  INDEX idx_category (category_id),
  INDEX idx_type (product_type),
  INDEX idx_status (status)
);
```

**规格模板示例:**
```json
{
  "memory": ["40GB", "80GB"],
  "network": ["25Gbps", "100Gbps"],
  "region": ["cn-beijing", "cn-shanghai", "us-west"]
}
```

### 3. SKU 表 (可售卖规格)

```sql
CREATE TABLE product_sku (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  sku_id          VARCHAR(64) UNIQUE NOT NULL,        -- SKU20240117000001
  sku_code        VARCHAR(64) UNIQUE NOT NULL,        -- SKU_GPU_A100_40GB_BJ
  
  spu_id          BIGINT NOT NULL,
  spu_code        VARCHAR(64) NOT NULL,               -- 冗余SPU编码
  
  sku_name        VARCHAR(255) NOT NULL,              -- A100 40GB 北京
  spec_values     JSON NOT NULL,                      -- 具体规格值
  
  region          VARCHAR(64),                        -- cn-beijing
  stock_type      VARCHAR(32),                        -- AVAILABLE/SOLD_OUT
  
  status          TINYINT DEFAULT 1,
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  INDEX idx_spu (spu_id),
  INDEX idx_region (region),
  INDEX idx_status (status)
);
```

**SKU 规格值示例:**
```json
{
  "memory": "40GB",
  "network": "25Gbps",
  "region": "cn-beijing"
}
```

---

## 数据示例

### 创建 A100 GPU 产品

```sql
-- 1. 创建 SPU (产品族)
INSERT INTO product_spu VALUES (
  spu_id: 'SPU20240117001',
  spu_code: 'SPU_GPU_A100',
  spu_name: 'NVIDIA A100 GPU',
  category_id: 11,              -- GPU计算分类
  product_type: 'GPU',
  brand: 'NVIDIA',
  billing_unit: 'HOUR',
  spec_template: '{
    "memory": ["40GB", "80GB"],
    "network": ["25Gbps", "100Gbps"],
    "region": ["cn-beijing", "cn-shanghai"]
  }'
);

-- 2. 创建多个 SKU (具体规格)
-- SKU1: 40GB 北京 25Gbps
INSERT INTO product_sku VALUES (
  sku_id: 'SKU20240117000001',
  sku_code: 'SKU_GPU_A100_40GB_BJ_25G',
  spu_id: 1,
  spu_code: 'SPU_GPU_A100',
  sku_name: 'A100 40GB 北京 25Gbps',
  spec_values: '{"memory":"40GB","network":"25Gbps","region":"cn-beijing"}',
  region: 'cn-beijing',
  stock_type: 'AVAILABLE'
);

-- SKU2: 80GB 北京 100Gbps
INSERT INTO product_sku VALUES (
  sku_id: 'SKU20240117000002',
  sku_code: 'SKU_GPU_A100_80GB_BJ_100G',
  spu_id: 1,
  spu_code: 'SPU_GPU_A100',
  sku_name: 'A100 80GB 北京 100Gbps',
  spec_values: '{"memory":"80GB","network":"100Gbps","region":"cn-beijing"}',
  region: 'cn-beijing',
  stock_type: 'AVAILABLE'
);

-- SKU3: 40GB 上海 25Gbps
INSERT INTO product_sku VALUES (
  sku_id: 'SKU20240117000003',
  sku_code: 'SKU_GPU_A100_40GB_SH_25G',
  spu_id: 1,
  spu_code: 'SPU_GPU_A100',
  sku_name: 'A100 40GB 上海 25Gbps',
  spec_values: '{"memory":"40GB","network":"25Gbps","region":"cn-shanghai"}',
  region: 'cn-shanghai',
  stock_type: 'AVAILABLE'
);
```

---

## 用户选购流程

```
1. 用户浏览分类
   → 计算服务 → GPU计算

2. 展示 SPU 列表
   → NVIDIA A100 GPU
   → NVIDIA V100 GPU
   → NVIDIA T4 GPU

3. 点击 A100,展示规格选择器
   内存: [40GB] [80GB]
   网络: [25Gbps] [100Gbps]
   地域: [北京] [上海]

4. 用户选择: 40GB + 北京 + 25Gbps
   → 系统匹配 SKU: SKU_GPU_A100_40GB_BJ_25G

5. 查询价格并下单
```

---

## SPU/SKU vs 扁平设计对比

| 维度 | 扁平设计 | SPU/SKU 设计 |
|------|---------|-------------|
| **产品管理** | 每个规格独立维护 | SPU 统一管理通用信息 |
| **定价灵活性** | 仅按产品编码定价 | 按 SPU/SKU/地域多维定价 |
| **新增规格** | 复制完整产品信息 | 仅新增 SKU 即可 |
| **产品查询** | 无法按分类筛选 | 支持分类/品牌/规格组合查询 |
| **库存管理** | 无法区分地域库存 | SKU 级别精确库存控制 |
| **数据冗余** | 高 (重复产品信息) | 低 (SPU 层统一维护) |

---

## SKU 编码规则

推荐格式: `SKU_产品类型_型号_规格_地域_网络`

示例:
- `SKU_GPU_A100_40GB_BJ_25G` - A100 40GB 北京 25Gbps
- `SKU_GPU_A100_80GB_SH_100G` - A100 80GB 上海 100Gbps
- `SKU_STORAGE_SSD_1TB_BJ` - 1TB SSD 块存储 北京

---

## 关联设计

### 订单关联 SKU

```sql
CREATE TABLE orders (
  ...
  spu_code        VARCHAR(64) NOT NULL,   -- 冗余 SPU
  sku_code        VARCHAR(64) NOT NULL,   -- 关联 SKU
  ...
);
```

### 定价规则关联 SKU

详见 [定价模型文档](./04-pricing-models.md)

---

## 相关文档

- [定价模型](./04-pricing-models.md)
- [订单模型](./05-order-models.md)
