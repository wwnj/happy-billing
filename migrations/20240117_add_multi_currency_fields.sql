-- 修改订单表，增加多币种支持
-- 使用存储过程检查字段是否存在，支持幂等执行
DELIMITER $$

DROP PROCEDURE IF EXISTS add_multi_currency_to_orders$$
CREATE PROCEDURE add_multi_currency_to_orders()
BEGIN
    IF NOT EXISTS (SELECT * FROM information_schema.COLUMNS
                   WHERE TABLE_SCHEMA = DATABASE()
                   AND TABLE_NAME = 'orders'
                   AND COLUMN_NAME = 'exchange_rate') THEN
        ALTER TABLE orders
          ADD COLUMN exchange_rate DECIMAL(18,8) COMMENT '汇率快照' AFTER currency,
          ADD COLUMN base_currency VARCHAR(8) DEFAULT 'CNY' COMMENT '本位币' AFTER exchange_rate,
          ADD COLUMN base_currency_amount DECIMAL(18,4) COMMENT '本位币金额' AFTER base_currency;
    END IF;
END$$

CALL add_multi_currency_to_orders()$$
DROP PROCEDURE add_multi_currency_to_orders$$

-- 修改账单表，增加多币种支持
DROP PROCEDURE IF EXISTS add_multi_currency_to_bills$$
CREATE PROCEDURE add_multi_currency_to_bills()
BEGIN
    IF NOT EXISTS (SELECT * FROM information_schema.COLUMNS
                   WHERE TABLE_SCHEMA = DATABASE()
                   AND TABLE_NAME = 'bills'
                   AND COLUMN_NAME = 'exchange_rate') THEN
        ALTER TABLE bills
          ADD COLUMN exchange_rate DECIMAL(18,8) COMMENT '汇率快照' AFTER currency,
          ADD COLUMN base_currency VARCHAR(8) DEFAULT 'CNY' COMMENT '本位币' AFTER exchange_rate,
          ADD COLUMN base_currency_amount DECIMAL(18,4) COMMENT '本位币应付金额' AFTER base_currency;
    END IF;
END$$

CALL add_multi_currency_to_bills()$$
DROP PROCEDURE add_multi_currency_to_bills$$

-- 修改支付表，增加多币种支持
DROP PROCEDURE IF EXISTS add_multi_currency_to_payments$$
CREATE PROCEDURE add_multi_currency_to_payments()
BEGIN
    IF NOT EXISTS (SELECT * FROM information_schema.COLUMNS
                   WHERE TABLE_SCHEMA = DATABASE()
                   AND TABLE_NAME = 'payments'
                   AND COLUMN_NAME = 'exchange_rate') THEN
        ALTER TABLE payments
          ADD COLUMN exchange_rate DECIMAL(18,8) COMMENT '汇率快照' AFTER currency,
          ADD COLUMN base_currency VARCHAR(8) DEFAULT 'CNY' COMMENT '本位币' AFTER exchange_rate,
          ADD COLUMN base_currency_amount DECIMAL(18,4) COMMENT '本位币金额' AFTER base_currency;
    END IF;
END$$

CALL add_multi_currency_to_payments()$$
DROP PROCEDURE add_multi_currency_to_payments$$

DELIMITER ;

-- 更新已有数据（假设都是CNY，汇率1:1）
UPDATE orders SET exchange_rate = 1.0, base_currency = 'CNY', base_currency_amount = payable_amount WHERE exchange_rate IS NULL;
UPDATE bills SET exchange_rate = 1.0, base_currency = 'CNY', base_currency_amount = payable_amount WHERE exchange_rate IS NULL;
UPDATE payments SET exchange_rate = 1.0, base_currency = 'CNY', base_currency_amount = amount WHERE exchange_rate IS NULL;
