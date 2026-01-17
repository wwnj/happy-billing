-- 修改订单表，增加多币种支持
ALTER TABLE orders
  ADD COLUMN exchange_rate DECIMAL(18,8) COMMENT '汇率快照' AFTER currency,
  ADD COLUMN base_currency VARCHAR(8) DEFAULT 'CNY' COMMENT '本位币' AFTER exchange_rate,
  ADD COLUMN base_currency_amount DECIMAL(18,4) COMMENT '本位币金额' AFTER base_currency;

-- 修改账单表，增加多币种支持
ALTER TABLE bills
  ADD COLUMN exchange_rate DECIMAL(18,8) COMMENT '汇率快照' AFTER currency,
  ADD COLUMN base_currency VARCHAR(8) DEFAULT 'CNY' COMMENT '本位币' AFTER exchange_rate,
  ADD COLUMN base_currency_amount DECIMAL(18,4) COMMENT '本位币应付金额' AFTER base_currency;

-- 修改支付表，增加多币种支持
ALTER TABLE payments
  ADD COLUMN exchange_rate DECIMAL(18,8) COMMENT '汇率快照' AFTER currency,
  ADD COLUMN base_currency VARCHAR(8) DEFAULT 'CNY' COMMENT '本位币' AFTER exchange_rate,
  ADD COLUMN base_currency_amount DECIMAL(18,4) COMMENT '本位币金额' AFTER base_currency;

-- 更新已有数据（假设都是CNY，汇率1:1）
UPDATE orders SET exchange_rate = 1.0, base_currency = 'CNY', base_currency_amount = payable_amount WHERE exchange_rate IS NULL;
UPDATE bills SET exchange_rate = 1.0, base_currency = 'CNY', base_currency_amount = payable_amount WHERE exchange_rate IS NULL;
UPDATE payments SET exchange_rate = 1.0, base_currency = 'CNY', base_currency_amount = amount WHERE exchange_rate IS NULL;
