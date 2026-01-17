# 订单模型设计

**文档版本**: v1.0  
**设计日期**: 2026-01-17

---

## 订单类型

| 类型 | 说明 | 计费模式 | 账单生成时机 |
|-----|------|---------|------------|
| **SUBSCRIPTION** | 包年包月 | 预付费 | 下单立即生成 |
| **PAY_AS_YOU_GO** | 按量计费 | 后付费 | 每小时生成 |

---

## 核心表设计

### 1. 订单主表

```sql
CREATE TABLE orders (
  id                BIGINT PRIMARY KEY AUTO_INCREMENT,
  order_id          VARCHAR(64) UNIQUE NOT NULL,        -- ORD20240117000001
  order_no          VARCHAR(64) UNIQUE NOT NULL,        -- 展示给用户的订单号
  
  -- 租户信息
  tenant_id         VARCHAR(64) NOT NULL,
  organization_id   VARCHAR(64),
  project_id        VARCHAR(64) NOT NULL,
  user_id           VARCHAR(64) NOT NULL,
  
  -- 订单类型
  order_type        VARCHAR(32) NOT NULL,               -- SUBSCRIPTION/PAY_AS_YOU_GO
  
  -- 关联产品
  spu_code          VARCHAR(64) NOT NULL,
  sku_code          VARCHAR(64) NOT NULL,
  
  -- 金额
  original_amount   DECIMAL(18,4) NOT NULL,
  discount_amount   DECIMAL(18,4) DEFAULT 0,
  payable_amount    DECIMAL(18,4) NOT NULL,
  paid_amount       DECIMAL(18,4) DEFAULT 0,
  
  -- 服务期(预付费)
  period_start      TIMESTAMP,
  period_end        TIMESTAMP,
  
  -- 订单状态
  status            VARCHAR(32) NOT NULL,               -- PENDING/PAID/RUNNING/EXPIRED/CANCELLED
  order_detail      JSON,
  
  created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  INDEX idx_order_id (order_id),
  INDEX idx_tenant (tenant_id),
  INDEX idx_project (project_id),
  INDEX idx_status (status)
);
```

### 2. 订单明细表

```sql
CREATE TABLE order_items (
  id                BIGINT PRIMARY KEY AUTO_INCREMENT,
  order_id          VARCHAR(64) NOT NULL,
  item_no           VARCHAR(64) NOT NULL,
  
  spu_code          VARCHAR(64) NOT NULL,
  sku_code          VARCHAR(64) NOT NULL,
  sku_name          VARCHAR(255) NOT NULL,              -- 订单快照
  sku_spec          JSON,                               -- 订单快照
  
  quantity          DECIMAL(18,4) NOT NULL,
  unit_price        DECIMAL(18,4) NOT NULL,
  amount            DECIMAL(18,4) NOT NULL,
  
  price_rule_id     VARCHAR(64),
  
  created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  INDEX idx_order (order_id),
  INDEX idx_sku (sku_code)
);
```

### 3. 资源实例表

```sql
CREATE TABLE resource_instances (
  id                BIGINT PRIMARY KEY AUTO_INCREMENT,
  instance_id       VARCHAR(64) UNIQUE NOT NULL,        -- INS20240117000001
  
  order_id          VARCHAR(64) NOT NULL,
  tenant_id         VARCHAR(64) NOT NULL,
  project_id        VARCHAR(64) NOT NULL,
  
  product_type      VARCHAR(32) NOT NULL,
  spu_code          VARCHAR(64) NOT NULL,
  sku_code          VARCHAR(64) NOT NULL,
  
  instance_spec     JSON,
  
  status            VARCHAR(32) NOT NULL,               -- CREATING/RUNNING/STOPPED/DELETED
  
  started_at        TIMESTAMP,
  stopped_at        TIMESTAMP,
  
  created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  INDEX idx_instance_id (instance_id),
  INDEX idx_order (order_id),
  INDEX idx_tenant_project (tenant_id, project_id),
  INDEX idx_status (status)
);
```

---

## 业务流程

### 场景1: 购买包年GPU (预付费)

```
1. 用户下单
   → 创建订单 (PENDING)
   → 查询定价规则
   → 计算金额

2. 生成账单 (立即)
   → 创建账单 (UNPAID)
   → 金额 = 订单应付金额

3. 用户支付
   → 扣减余额/在线支付
   → 更新账单 (PAID)
   → 更新订单 (PAID)

4. 创建资源实例
   → 分配GPU资源
   → 实例状态 (RUNNING)
   → 订单状态 (RUNNING)
```

**SQL示例:**
```sql
-- Step1: 创建订单
INSERT INTO orders VALUES (
  order_id: 'ORD20240117000001',
  tenant_id: 'T20240117001',
  project_id: 'PRJ20240117001',
  order_type: 'SUBSCRIPTION',
  spu_code: 'SPU_GPU_A100',
  sku_code: 'SKU_GPU_A100_40GB_BJ',
  payable_amount: 120000.00,
  period_start: '2024-01-17',
  period_end: '2025-01-16',
  status: 'PENDING'
);

-- Step2: 生成账单
INSERT INTO bills VALUES (
  bill_id: 'BILL20240117000001',
  order_id: 'ORD20240117000001',
  bill_type: 'SUBSCRIPTION',
  amount: 120000.00,
  status: 'UNPAID'
);

-- Step3: 支付成功
UPDATE bills SET status='PAID' WHERE bill_id='BILL20240117000001';
UPDATE orders SET status='PAID' WHERE order_id='ORD20240117000001';

-- Step4: 创建实例
INSERT INTO resource_instances VALUES (
  instance_id: 'INS20240117000001',
  order_id: 'ORD20240117000001',
  tenant_id: 'T20240117001',
  sku_code: 'SKU_GPU_A100_40GB_BJ',
  status: 'RUNNING',
  started_at: NOW()
);

UPDATE orders SET status='RUNNING' WHERE order_id='ORD20240117000001';
```

### 场景2: 按量计费GPU (后付费)

```
1. 用户下单
   → 创建订单 (RUNNING)
   → 初始金额 = 0

2. 创建资源实例 (立即)
   → 分配GPU资源
   → 实例状态 (RUNNING)

3. 持续计量
   → 每秒上报计量数据
   → 写入 ClickHouse

4. 每小时出账
   → 聚合计量数据
   → 生成账单
   → 从余额扣减
```

---

## 订单状态流转

```
预付费流程:
PENDING → PAID → RUNNING → EXPIRED

后付费流程:
RUNNING → (持续运行) → STOPPED

取消流程:
PENDING → CANCELLED
PAID → REFUNDING → REFUNDED
```

---

## 关键设计点

### 1. 为什么订单要冗余tenant_id/project_id?

**原因**: 快速查询和成本分析

```sql
-- 查询项目总费用 (冗余后)
SELECT SUM(payable_amount) 
FROM orders 
WHERE project_id = 'PRJ20240117001';

-- 不冗余需要JOIN资源实例表
SELECT SUM(o.payable_amount)
FROM orders o
JOIN resource_instances ri ON o.id = ri.order_id
WHERE ri.project_id = 'PRJ20240117001';
```

### 2. 订单快照设计

订单明细表存储SKU名称和规格的快照,防止后续产品信息变更影响历史订单。

---

## 相关文档

- [产品模型](./03-product-models.md)
- [定价模型](./04-pricing-models.md)
- [账单模型](./06-billing-models.md)
