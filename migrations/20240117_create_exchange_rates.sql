-- 汇率表
CREATE TABLE IF NOT EXISTS exchange_rates (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  from_currency   VARCHAR(8) NOT NULL COMMENT '源币种',
  to_currency     VARCHAR(8) NOT NULL COMMENT '目标币种',

  rate            DECIMAL(18,8) NOT NULL COMMENT '汇率 (1 from_currency = ? to_currency)',

  effective_date  DATE NOT NULL COMMENT '生效日期',

  source          VARCHAR(64) COMMENT '汇率来源: BANK/API',

  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  UNIQUE KEY uk_currency_date (from_currency, to_currency, effective_date),
  INDEX idx_date (effective_date),
  INDEX idx_from_to (from_currency, to_currency)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='汇率表';

-- 插入初始汇率数据 (2024-01-17)
INSERT INTO exchange_rates (from_currency, to_currency, rate, effective_date, source) VALUES
('CNY', 'USD', 0.13940000, '2024-01-17', 'BANK'),
('CNY', 'EUR', 0.12820000, '2024-01-17', 'BANK'),
('CNY', 'JPY', 20.15630000, '2024-01-17', 'BANK'),
('CNY', 'GBP', 0.11050000, '2024-01-17', 'BANK'),
('CNY', 'HKD', 1.09120000, '2024-01-17', 'BANK');

-- 插入当日汇率 (2026-01-17)
INSERT INTO exchange_rates (from_currency, to_currency, rate, effective_date, source) VALUES
('CNY', 'USD', 0.13850000, '2026-01-17', 'BANK'),
('CNY', 'EUR', 0.12750000, '2026-01-17', 'BANK'),
('CNY', 'JPY', 20.35000000, '2026-01-17', 'BANK'),
('CNY', 'GBP', 0.10980000, '2026-01-17', 'BANK'),
('CNY', 'HKD', 1.08500000, '2026-01-17', 'BANK');
