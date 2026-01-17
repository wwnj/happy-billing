# 账单模型设计 (MySQL + ClickHouse)

**文档版本**: v1.0  
**设计日期**: 2026-01-17

---

## 双存储架构

| 存储 | 数据类型 | 数据量级 | 保留策略 | 典型查询 |
|------|---------|---------|---------|---------|
| **MySQL** | 账单汇总数据 | 万-百万级 | 永久保留 | 用户账单列表、月度对账 |
| **ClickHouse** | 秒级计量明细 | 亿-百亿级 | 3-6个月 | 账单明细追溯、成本分析 |

**数据流转:**
```
资源使用 → 秒级计量(ClickHouse) → 小时聚合 → 生成账单(MySQL)
                                              ↓
                                         账单结算 → 支付记录
```

---

## MySQL 账单表设计

### 账单主表

```sql
CREATE TABLE bills (
  id                BIGINT PRIMARY KEY AUTO_INCREMENT,
  bill_id           VARCHAR(64) UNIQUE NOT NULL,        -- BILL20240117000001
  
  order_id          VARCHAR(64) NOT NULL,
  tenant_id         VARCHAR(64) NOT NULL,
  project_id        VARCHAR(64) NOT NULL,
  
  bill_type         VARCHAR(32) NOT NULL,               -- SUBSCRIPTION/HOURLY
  
  -- 计费周期(后付费)
  billing_period_start  TIMESTAMP,
  billing_period_end    TIMESTAMP,
  
  -- 金额
  currency          VARCHAR(8) DEFAULT 'CNY',
  exchange_rate     DECIMAL(18,8),
  base_currency     VARCHAR(8) DEFAULT 'CNY',
  base_currency_amount DECIMAL(18,4),
  
  original_amount   DECIMAL(18,4) NOT NULL,
  discount_amount   DECIMAL(18,4) DEFAULT 0,
  payable_amount    DECIMAL(18,4) NOT NULL,
  
  -- 账单状态
  status            VARCHAR(32) NOT NULL,               -- UNPAID/PAID/OVERDUE/CANCELLED
  
  bill_detail       JSON,
  
  paid_at           TIMESTAMP,
  created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  INDEX idx_bill_id (bill_id),
  INDEX idx_order (order_id),
  INDEX idx_tenant (tenant_id),
  INDEX idx_status (status),
  INDEX idx_period (billing_period_start, billing_period_end)
);
```

---

## ClickHouse 计量表设计

### 秒级计量记录表

```sql
CREATE TABLE metering_records (
  timestamp           DateTime,
  instance_id         String,
  tenant_id           String,
  project_id          String,
  
  metric_type         String,                           -- gpu_usage/storage_usage/token_count
  metric_value        Float64,
  
  unit_price          Decimal(18,4),
  cost                Decimal(18,4),
  
  region              String,
  extra_info          String                            -- JSON字符串
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, project_id, instance_id, timestamp)
TTL timestamp + INTERVAL 6 MONTH;
```

### 小时聚合物化视图

```sql
CREATE MATERIALIZED VIEW metering_hourly
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMMDD(hour_start)
ORDER BY (tenant_id, project_id, instance_id, hour_start)
AS SELECT
  toStartOfHour(timestamp) AS hour_start,
  instance_id,
  tenant_id,
  project_id,
  metric_type,
  count() AS record_count,
  sum(metric_value) AS total_value,
  sum(cost) AS total_cost
FROM metering_records
GROUP BY hour_start, instance_id, tenant_id, project_id, metric_type;
```

---

## 计量与出账流程

### 按量计费 GPU 示例

```
10:00:00 - 10:59:59
  → 每秒上报1条计量记录 (3600条写入ClickHouse)
  
11:00:00 (整点触发出账)
  → 查询物化视图获取10:00-11:00聚合数据
  → 计算费用: 3600秒 × ¥0.0028/秒 = ¥10.08
  → 生成账单写入MySQL
  → 从账户余额扣减
```

**聚合查询:**
```sql
SELECT 
  instance_id,
  sum(total_cost) AS hour_cost
FROM metering_hourly
WHERE hour_start = '2024-01-17 10:00:00'
  AND tenant_id = 'T20240117001'
GROUP BY instance_id;
```

---

## LLM Token 计费示例

### 计量记录

```sql
-- 每次API调用产生2条记录(输入+输出)
INSERT INTO metering_records VALUES
('2024-01-17 10:05:32', 'LLM-API-001', 'T20240117001', 'PRJ001', 'input_tokens', 150, 0.00001, 0.0015),
('2024-01-17 10:05:32', 'LLM-API-001', 'T20240117001', 'PRJ001', 'output_tokens', 300, 0.00002, 0.0060);
```

### 小时汇总出账

```sql
SELECT 
  metric_type,
  sum(total_value) AS tokens,
  sum(total_cost) AS cost
FROM metering_hourly
WHERE hour_start = '2024-01-17 10:00:00'
  AND tenant_id = 'T20240117001'
  AND metric_type IN ('input_tokens', 'output_tokens')
GROUP BY metric_type;

-- 结果:
-- input_tokens: 50000, ¥0.50
-- output_tokens: 80000, ¥1.60
-- 总计: ¥2.10
```

---

## 红冲账单 (退款)

当用户退款时,生成负数账单:

```sql
-- 原账单
INSERT INTO bills VALUES (
  bill_id: 'BILL20240117000001',
  payable_amount: 120000.00,
  status: 'PAID'
);

-- 退款红冲账单
INSERT INTO bills VALUES (
  bill_id: 'BILL20240117000002',
  order_id: '同一订单ID',
  bill_type: 'REFUND',
  payable_amount: -120000.00,
  status: 'PAID'
);
```

---

## 账单状态流转

```
UNPAID (未支付)
  ↓ 支付成功
PAID (已支付)
  ↓ 超期未支付
OVERDUE (逾期)
  ↓ 取消
CANCELLED (已取消)
```

---

## 性能优化

### ClickHouse 优化

1. **分区策略**: 按月分区,6个月TTL自动清理
2. **物化视图**: 自动聚合小时数据,查询快速
3. **压缩**: ZSTD压缩,节省存储空间

### MySQL 优化

1. **索引**: tenant_id, project_id, status, period 复合索引
2. **分库分表**: 按tenant_id哈希分表
3. **归档**: 超过1年的账单迁移到归档表

---

## 相关文档

- [订单模型](./05-order-models.md)
- [支付结算](./07-payment-settlement.md)
