# 定价模型设计

**文档版本**: v1.0  
**设计日期**: 2026-01-17

---

## 定价类型

Happy Billing 支持完整的价格体系:

| 定价类型 | 说明 | 适用场景 |
|---------|------|---------|
| **固定价格** | 统一单价 | 标准产品 |
| **阶梯价格** | 用量越大单价越低 | 鼓励大量使用 |
| **时段价格** | 峰谷分时定价 | 削峰填谷 |
| **资源包** | 预购买优惠包 | 预付费优惠 |
| **折扣规则** | 用户级/促销折扣 | 营销活动 |

---

## 核心表设计

### 定价规则表

```sql
CREATE TABLE price_rules (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  rule_id         VARCHAR(64) UNIQUE NOT NULL,        -- PRICE20240117001
  rule_code       VARCHAR(64) UNIQUE NOT NULL,
  rule_name       VARCHAR(255) NOT NULL,
  
  -- 关联产品
  spu_code        VARCHAR(64),                        -- 关联 SPU
  sku_code        VARCHAR(64),                        -- 关联 SKU (优先级更高)
  
  rule_type       VARCHAR(32) NOT NULL,               -- FIXED/TIERED/TIME_BASED/PACKAGE
  pricing_detail  JSON NOT NULL,                      -- 定价详情
  
  currency        VARCHAR(8) DEFAULT 'CNY',
  
  effective_start TIMESTAMP NOT NULL,
  effective_end   TIMESTAMP,
  
  region          VARCHAR(64),                        -- 地域限定
  priority        INT DEFAULT 0,                      -- 优先级
  status          TINYINT DEFAULT 1,
  
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  INDEX idx_spu (spu_code),
  INDEX idx_sku (sku_code),
  INDEX idx_region (region),
  INDEX idx_priority (priority)
);
```

---

## 1. 固定价格 (FIXED)

**pricing_detail 示例:**
```json
{
  "unit_price": 10.00
}
```

**示例:**
```sql
INSERT INTO price_rules VALUES (
  rule_code: 'PRICE_GPU_A100_FIXED',
  rule_name: 'A100 GPU 固定价格',
  sku_code: 'SKU_GPU_A100_40GB_BJ',
  rule_type: 'FIXED',
  pricing_detail: '{"unit_price":10.00}',
  currency: 'CNY',
  effective_start: '2024-01-01'
);
```

**计费逻辑:**
```
费用 = 用量 × 单价
例如: 3小时 × ¥10/小时 = ¥30
```

---

## 2. 阶梯价格 (TIERED)

**pricing_detail 示例:**
```json
{
  "tiers": [
    {"from": 0, "to": 100, "unit_price": 1.00},
    {"from": 100, "to": 500, "unit_price": 0.80},
    {"from": 500, "to": null, "unit_price": 0.60}
  ]
}
```

**示例:**
```sql
INSERT INTO price_rules VALUES (
  rule_code: 'PRICE_STORAGE_TIERED',
  rule_name: '存储阶梯价格',
  spu_code: 'SPU_STORAGE_SSD',
  rule_type: 'TIERED',
  pricing_detail: '{
    "tiers": [
      {"from":0, "to":100, "unit_price":1.00},
      {"from":100, "to":500, "unit_price":0.80},
      {"from":500, "to":null, "unit_price":0.60}
    ]
  }'
);
```

**计费逻辑:**
```
用量 200GB:
  0-100GB:   100 × ¥1.00 = ¥100
  100-200GB: 100 × ¥0.80 = ¥80
  总费用:    ¥180
```

---

## 3. 时段价格 (TIME_BASED)

**pricing_detail 示例:**
```json
{
  "periods": [
    {"hour_from": 0, "hour_to": 8, "unit_price": 5.00},
    {"hour_from": 8, "hour_to": 20, "unit_price": 10.00},
    {"hour_from": 20, "hour_to": 24, "unit_price": 5.00}
  ]
}
```

**示例:**
```sql
INSERT INTO price_rules VALUES (
  rule_code: 'PRICE_GPU_TIME_BASED',
  rule_name: 'GPU 峰谷价格',
  spu_code: 'SPU_GPU_A100',
  rule_type: 'TIME_BASED',
  pricing_detail: '{
    "periods": [
      {"hour_from":0, "hour_to":8, "unit_price":5.00},
      {"hour_from":8, "hour_to":20, "unit_price":10.00},
      {"hour_from":20, "hour_to":24, "unit_price":5.00}
    ]
  }'
);
```

**计费逻辑:**
```
00:00-08:00 (谷时): ¥5/小时
08:00-20:00 (峰时): ¥10/小时
20:00-24:00 (谷时): ¥5/小时
```

---

## 4. 资源包 (PACKAGE)

**pricing_detail 示例:**
```json
{
  "package_size": 1000,
  "package_price": 800.00,
  "validity_days": 30
}
```

**示例:**
```sql
INSERT INTO price_rules VALUES (
  rule_code: 'PRICE_LLM_PACKAGE',
  rule_name: 'LLM Token 资源包',
  spu_code: 'SPU_LLM_GPT4',
  rule_type: 'PACKAGE',
  pricing_detail: '{
    "package_size": 1000000,
    "package_price": 800.00,
    "validity_days": 30
  }'
);
```

**计费逻辑:**
```
购买资源包: 100万 Token ¥800 (有效期30天)
单价: ¥0.0008/Token
按量: ¥0.001/Token
优惠: 20% OFF
```

---

## 5. 折扣规则 (DISCOUNT)

**discount 表设计:**
```sql
CREATE TABLE discount_rules (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  discount_id     VARCHAR(64) UNIQUE NOT NULL,
  discount_name   VARCHAR(255) NOT NULL,
  
  discount_type   VARCHAR(32) NOT NULL,               -- PERCENTAGE/AMOUNT
  discount_value  DECIMAL(18,4) NOT NULL,             -- 折扣值
  
  target_type     VARCHAR(32) NOT NULL,               -- TENANT/SPU/SKU
  target_id       VARCHAR(64),                        -- 目标ID
  
  effective_start TIMESTAMP NOT NULL,
  effective_end   TIMESTAMP,
  
  max_discount    DECIMAL(18,4),                      -- 最大折扣金额
  min_amount      DECIMAL(18,4),                      -- 最小消费金额
  
  status          TINYINT DEFAULT 1,
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**折扣示例:**
```sql
-- VIP 客户 8折
INSERT INTO discount_rules VALUES (
  discount_id: 'DISC20240117001',
  discount_name: 'VIP客户折扣',
  discount_type: 'PERCENTAGE',
  discount_value: 0.20,                               -- 20% OFF
  target_type: 'TENANT',
  target_id: 'T20240117001'
);

-- GPU 促销满1000减200
INSERT INTO discount_rules VALUES (
  discount_id: 'DISC20240117002',
  discount_name: 'GPU促销活动',
  discount_type: 'AMOUNT',
  discount_value: 200.00,
  target_type: 'SPU',
  target_id: 'SPU_GPU_A100',
  min_amount: 1000.00
);
```

---

## 价格查询优先级

查询 SKU 价格时按以下优先级:

```
1. SKU 级别定价 (最高优先级)
   ↓
2. SPU + 地域定价
   ↓
3. SPU 级别定价 (最低优先级)
```

**查询代码示例:**
```go
func GetSKUPrice(skuCode, region string) (*PriceRule, error) {
    sku := db.GetSKU(skuCode)
    
    query := `
        SELECT * FROM price_rules
        WHERE status = 1
          AND effective_start <= NOW()
          AND (effective_end IS NULL OR effective_end >= NOW())
          AND (
              sku_code = ?                            -- 优先级1
              OR (spu_code = ? AND region = ?)        -- 优先级2
              OR spu_code = ?                         -- 优先级3
          )
        ORDER BY 
          CASE 
            WHEN sku_code IS NOT NULL THEN 30
            WHEN region IS NOT NULL THEN 20
            ELSE 10
          END DESC,
          priority DESC
        LIMIT 1
    `
    
    var rule PriceRule
    db.QueryRow(query, 
        skuCode,
        sku.SPUCode, region,
        sku.SPUCode,
    ).Scan(&rule)
    
    return &rule, nil
}
```

---

## 价格计算流程

```
1. 查询基础价格 (price_rules)
   ↓
2. 应用折扣规则 (discount_rules)
   ↓
3. 计算最终价格
   ↓
4. 生成订单
```

---

## 相关文档

- [产品模型](./03-product-models.md)
- [订单模型](./05-order-models.md)
