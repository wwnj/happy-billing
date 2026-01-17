# 支付与结算模型

**文档版本**: v1.0  
**设计日期**: 2026-01-17

---

## 支付方式

| 支付方式 | 适用用户 | 说明 |
|---------|---------|------|
| **账户余额** | 全部用户 | 预充值余额扣减 |
| **在线支付** | 全部用户 | 支付宝/微信/银行卡 |
| **月结** | 企业用户 | 月度对账后支付 |
| **授信** | 大客户 | 信用额度先使用后付款 |

---

## 核心表设计

### 1. 账户表

```sql
CREATE TABLE accounts (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  account_id      VARCHAR(64) UNIQUE NOT NULL,          -- ACC20240117001
  tenant_id       VARCHAR(64) NOT NULL,
  
  -- 余额
  balance         DECIMAL(18,4) DEFAULT 0,              -- 可用余额
  frozen_balance  DECIMAL(18,4) DEFAULT 0,              -- 冻结余额
  
  -- 授信
  credit_limit    DECIMAL(18,4) DEFAULT 0,              -- 授信额度
  credit_used     DECIMAL(18,4) DEFAULT 0,              -- 已用授信
  
  currency        VARCHAR(8) DEFAULT 'CNY',
  
  -- 乐观锁
  version         INT DEFAULT 0,
  
  status          TINYINT DEFAULT 1,
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  UNIQUE KEY uk_tenant (tenant_id),
  INDEX idx_account_id (account_id)
);
```

### 2. 账户流水表

```sql
CREATE TABLE account_transactions (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  transaction_id  VARCHAR(64) UNIQUE NOT NULL,          -- TXN20240117000001
  account_id      VARCHAR(64) NOT NULL,
  tenant_id       VARCHAR(64) NOT NULL,
  
  transaction_type VARCHAR(32) NOT NULL,                -- RECHARGE/DEDUCT/REFUND/FREEZE/UNFREEZE
  
  amount          DECIMAL(18,4) NOT NULL,               -- 正数或负数
  balance_before  DECIMAL(18,4) NOT NULL,
  balance_after   DECIMAL(18,4) NOT NULL,
  
  related_id      VARCHAR(64),                          -- 关联订单/账单ID
  remark          VARCHAR(512),
  
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  INDEX idx_account (account_id),
  INDEX idx_tenant (tenant_id),
  INDEX idx_type (transaction_type)
);
```

### 3. 支付记录表

```sql
CREATE TABLE payments (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  payment_id      VARCHAR(64) UNIQUE NOT NULL,          -- PAY20240117000001
  
  bill_id         VARCHAR(64) NOT NULL,
  order_id        VARCHAR(64) NOT NULL,
  tenant_id       VARCHAR(64) NOT NULL,
  
  payment_method  VARCHAR(32) NOT NULL,                 -- BALANCE/ALIPAY/WECHAT/BANK
  payment_channel VARCHAR(64),                          -- 支付渠道
  
  amount          DECIMAL(18,4) NOT NULL,
  currency        VARCHAR(8) DEFAULT 'CNY',
  
  status          VARCHAR(32) NOT NULL,                 -- PENDING/SUCCESS/FAILED/CANCELLED
  
  third_party_no  VARCHAR(128),                         -- 第三方支付流水号
  paid_at         TIMESTAMP,
  
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  INDEX idx_payment_id (payment_id),
  INDEX idx_bill (bill_id),
  INDEX idx_tenant (tenant_id),
  INDEX idx_third_party (third_party_no)
);
```

### 4. 月结对账表

```sql
CREATE TABLE settlements (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  settlement_id   VARCHAR(64) UNIQUE NOT NULL,          -- STL20240117001
  
  tenant_id       VARCHAR(64) NOT NULL,
  
  settlement_month VARCHAR(7) NOT NULL,                 -- 2024-01
  
  total_amount    DECIMAL(18,4) NOT NULL,
  paid_amount     DECIMAL(18,4) DEFAULT 0,
  
  status          VARCHAR(32) NOT NULL,                 -- PENDING/CONFIRMED/PAID
  
  confirmed_at    TIMESTAMP,
  paid_at         TIMESTAMP,
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  INDEX idx_tenant_month (tenant_id, settlement_month)
);
```

---

## 支付流程

### 余额支付

```
1. 查询账单金额
   ↓
2. 检查余额是否充足
   ↓
3. 扣减余额 (乐观锁)
   ↓
4. 记录支付记录
   ↓
5. 更新账单状态
```

**扣减余额代码:**
```go
func DeductBalance(accountID string, amount decimal.Decimal, billID string) error {
    tx.Begin()
    
    // 查询账户(加乐观锁)
    account := db.QueryRow(`
        SELECT balance, version 
        FROM accounts 
        WHERE account_id = ? 
        FOR UPDATE
    `, accountID)
    
    // 检查余额
    if account.Balance < amount {
        return errors.New("余额不足")
    }
    
    // 扣减余额(乐观锁)
    result := db.Exec(`
        UPDATE accounts 
        SET balance = balance - ?,
            version = version + 1
        WHERE account_id = ? 
          AND version = ?
    `, amount, accountID, account.Version)
    
    if result.RowsAffected == 0 {
        return errors.New("并发冲突,请重试")
    }
    
    // 记录流水
    db.Insert(account_transactions, {
        transaction_type: "DEDUCT",
        amount: -amount,
        balance_before: account.Balance,
        balance_after: account.Balance - amount,
        related_id: billID,
    })
    
    // 记录支付
    db.Insert(payments, {
        bill_id: billID,
        payment_method: "BALANCE",
        amount: amount,
        status: "SUCCESS",
    })
    
    // 更新账单
    db.Update(bills, {status: "PAID"}, billID)
    
    tx.Commit()
}
```

### 在线支付 (第三方)

```
1. 创建支付订单
   ↓
2. 调用支付宝/微信API
   ↓
3. 用户扫码支付
   ↓
4. 接收支付回调 (幂等性)
   ↓
5. 更新支付记录和账单
```

**支付回调处理:**
```go
func HandlePaymentCallback(paymentID, thirdPartyNo string) error {
    // 分布式锁防止重复处理
    lock := redis.Lock("payment:" + paymentID)
    defer lock.Unlock()
    
    // 查询支付记录
    payment := db.GetPayment(paymentID)
    
    // 幂等性检查
    if payment.Status == "SUCCESS" {
        return nil  // 已处理,直接返回
    }
    
    tx.Begin()
    
    // 更新支付记录
    db.Update(payments, {
        status: "SUCCESS",
        third_party_no: thirdPartyNo,
        paid_at: time.Now(),
    }, paymentID)
    
    // 更新账单
    db.Update(bills, {
        status: "PAID",
        paid_at: time.Now(),
    }, payment.BillID)
    
    // 更新订单
    db.Update(orders, {
        status: "PAID",
        paid_amount: payment.Amount,
    }, payment.OrderID)
    
    // 发送 Kafka 事件
    kafka.Send("payment.success", payment)
    
    tx.Commit()
}
```

---

## 月结流程

```
每月1日
  ↓
生成上月对账单
  ↓
汇总所有未结算账单
  ↓
发送对账邮件
  ↓
企业确认
  ↓
企业付款
  ↓
更新结算单状态
```

---

## 授信管理

### 授信额度

```sql
-- 设置授信额度
UPDATE accounts 
SET credit_limit = 100000.00 
WHERE tenant_id = 'T20240117001';
```

### 授信扣减

```
可用授信 = credit_limit - credit_used

当余额不足时:
  → 检查可用授信
  → 扣减授信额度
  → 记录授信使用
```

---

## 退款流程

```
1. 创建退款申请
   ↓
2. 审核通过
   ↓
3. 原路退款
   ↓
4. 生成红冲账单
   ↓
5. 余额退款 → 增加余额
   在线支付退款 → 调用第三方退款API
```

---

## 相关文档

- [账单模型](./06-billing-models.md)
- [多币种支持](./08-multi-currency.md)
