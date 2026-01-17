# 多币种支持设计

**文档版本**: v1.0  
**设计日期**: 2026-01-17

---

## 设计原则

- **本位币**: CNY (人民币) - 所有内部核算统一使用本位币
- **展示币种**: 支持 USD、EUR、JPY 等多种展示货币
- **汇率管理**: 每日更新汇率,账单记录汇率快照

---

## 核心表设计

### 汇率表

```sql
CREATE TABLE exchange_rates (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  from_currency   VARCHAR(8) NOT NULL,                  -- CNY
  to_currency     VARCHAR(8) NOT NULL,                  -- USD
  
  rate            DECIMAL(18,8) NOT NULL,               -- 汇率 (1 CNY = ? USD)
  
  effective_date  DATE NOT NULL,
  
  source          VARCHAR(64),                          -- 汇率来源: BANK/API
  
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  UNIQUE KEY uk_currency_date (from_currency, to_currency, effective_date),
  INDEX idx_date (effective_date)
);
```

**汇率示例:**
```sql
-- 2024-01-17 汇率
INSERT INTO exchange_rates VALUES
('CNY', 'USD', 0.1394, '2024-01-17', 'BANK'),        -- 1 CNY = 0.1394 USD
('CNY', 'EUR', 0.1282, '2024-01-17', 'BANK'),        -- 1 CNY = 0.1282 EUR
('CNY', 'JPY', 20.1563, '2024-01-17', 'BANK');       -- 1 CNY = 20.1563 JPY
```

---

## 账单多币种设计

### 账单表增强

```sql
CREATE TABLE bills (
  ...
  
  -- 展示币种
  currency          VARCHAR(8) DEFAULT 'CNY',           -- 用户偏好币种
  exchange_rate     DECIMAL(18,8),                      -- 汇率快照
  
  -- 本位币 (核算)
  base_currency     VARCHAR(8) DEFAULT 'CNY',
  base_currency_amount DECIMAL(18,4),                   -- 本位币金额
  
  -- 展示币种金额
  payable_amount    DECIMAL(18,4) NOT NULL,             -- 应付金额(展示币种)
  
  ...
);
```

---

## 币种转换流程

### 下单时

```
1. 获取用户偏好币种 (从租户表)
   ↓
2. 查询当日汇率
   ↓
3. 计算本位币金额
   base_currency_amount = payable_amount / exchange_rate
   ↓
4. 生成账单时记录:
   - currency: USD
   - exchange_rate: 0.1394
   - payable_amount: 100.00 USD
   - base_currency_amount: 717.36 CNY
```

**代码示例:**
```go
func CreateBill(tenantID string, amount decimal.Decimal) (*Bill, error) {
    // 获取租户偏好币种
    tenant := db.GetTenant(tenantID)
    currency := tenant.PreferredCurrency  // 例如 "USD"
    
    // 查询当日汇率
    rate := db.GetExchangeRate("CNY", currency, time.Now())
    
    // 计算本位币金额
    baseCurrencyAmount := amount.Div(rate.Rate)
    
    // 创建账单
    bill := Bill{
        Currency:            currency,
        ExchangeRate:        rate.Rate,
        PayableAmount:       amount,
        BaseCurrency:        "CNY",
        BaseCurrencyAmount:  baseCurrencyAmount,
    }
    
    return &bill, nil
}
```

### 查询账单时

用户可以选择展示币种查看账单:

```go
func GetBillWithCurrency(billID, displayCurrency string) (*Bill, error) {
    bill := db.GetBill(billID)
    
    // 如果展示币种与账单币种相同,直接返回
    if bill.Currency == displayCurrency {
        return bill, nil
    }
    
    // 查询当前汇率
    rate := db.GetExchangeRate("CNY", displayCurrency, time.Now())
    
    // 从本位币转换到展示币种
    displayAmount := bill.BaseCurrencyAmount.Mul(rate.Rate)
    
    // 返回转换后的账单
    return &Bill{
        ...原账单信息...,
        DisplayCurrency: displayCurrency,
        DisplayAmount:   displayAmount,
        DisplayRate:     rate.Rate,
    }, nil
}
```

---

## 汇率更新策略

### 定时任务

```go
// 每天凌晨1点更新汇率
func UpdateDailyExchangeRates() {
    currencies := []string{"USD", "EUR", "JPY", "GBP", "HKD"}
    
    for _, currency := range currencies {
        // 调用第三方汇率API
        rate := fetchExchangeRate("CNY", currency)
        
        // 存储到数据库
        db.Insert(exchange_rates, {
            from_currency: "CNY",
            to_currency: currency,
            rate: rate,
            effective_date: time.Now().Format("2006-01-02"),
            source: "API",
        })
        
        // 更新Redis缓存
        redis.Set(fmt.Sprintf("rate:CNY:%s", currency), rate, 24*time.Hour)
    }
}
```

---

## 支付币种处理

### 在线支付

```
用户选择支付币种 (可能与账单币种不同)
  ↓
查询支付时汇率
  ↓
转换金额
  ↓
调用支付渠道
```

**示例:**
```go
func CreatePayment(billID string, paymentCurrency string) (*Payment, error) {
    bill := db.GetBill(billID)
    
    // 如果支付币种与账单币种不同,需要转换
    var paymentAmount decimal.Decimal
    if bill.Currency != paymentCurrency {
        // 先转回本位币
        baseCurrencyAmount := bill.BaseCurrencyAmount
        
        // 再转到支付币种
        rate := db.GetExchangeRate("CNY", paymentCurrency, time.Now())
        paymentAmount = baseCurrencyAmount.Mul(rate.Rate)
    } else {
        paymentAmount = bill.PayableAmount
    }
    
    // 创建支付记录
    payment := Payment{
        BillID:          billID,
        Amount:          paymentAmount,
        Currency:        paymentCurrency,
        ExchangeRate:    rate.Rate,
    }
    
    return &payment, nil
}
```

---

## 财务报表

所有财务报表统一使用本位币 (CNY) 核算:

```sql
-- 月度收入报表 (统一CNY核算)
SELECT 
  DATE_FORMAT(created_at, '%Y-%m') AS month,
  SUM(base_currency_amount) AS total_revenue_cny
FROM bills
WHERE status = 'PAID'
  AND base_currency = 'CNY'
GROUP BY month;
```

---

## 汇率风险控制

### 汇率波动保护

对于长期订单(包年包月),在下单时锁定汇率:

```sql
-- 订单表记录汇率快照
CREATE TABLE orders (
  ...
  currency          VARCHAR(8) DEFAULT 'CNY',
  exchange_rate     DECIMAL(18,8),
  locked_at         TIMESTAMP,                          -- 汇率锁定时间
  ...
);
```

### 汇率预警

```
当汇率波动超过5%时:
  → 发送预警通知
  → 财务人员审核
  → 考虑调整价格
```

---

## 示例场景

### 美国用户购买GPU

```
1. 用户偏好币种: USD
2. 产品定价: ¥10/小时 (本位币)
3. 当日汇率: 1 CNY = 0.14 USD
4. 展示价格: $1.40/小时
5. 用户下单: $1.40
6. 账单记录:
   - currency: USD
   - payable_amount: 1.40
   - exchange_rate: 0.14
   - base_currency_amount: 10.00 CNY
```

---

## 相关文档

- [账单模型](./06-billing-models.md)
- [支付结算](./07-payment-settlement.md)
