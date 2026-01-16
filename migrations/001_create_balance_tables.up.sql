-- 创建账户表
CREATE TABLE IF NOT EXISTS accounts (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    balance DECIMAL(20, 4) NOT NULL DEFAULT 0,
    frozen_balance DECIMAL(20, 4) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    version BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建唯一索引：租户ID + 用户ID
CREATE UNIQUE INDEX IF NOT EXISTS uk_tenant_user ON accounts (tenant_id, user_id);

-- 创建索引：租户ID
CREATE INDEX IF NOT EXISTS idx_tenant_id ON accounts (tenant_id);

-- 创建索引：用户ID
CREATE INDEX IF NOT EXISTS idx_user_id ON accounts (user_id);

-- 注释
COMMENT ON TABLE accounts IS '账户表';
COMMENT ON COLUMN accounts.id IS '账户ID';
COMMENT ON COLUMN accounts.tenant_id IS '租户ID';
COMMENT ON COLUMN accounts.user_id IS '用户ID';
COMMENT ON COLUMN accounts.balance IS '可用余额';
COMMENT ON COLUMN accounts.frozen_balance IS '冻结余额';
COMMENT ON COLUMN accounts.currency IS '货币类型';
COMMENT ON COLUMN accounts.version IS '版本号(乐观锁)';
COMMENT ON COLUMN accounts.created_at IS '创建时间';
COMMENT ON COLUMN accounts.updated_at IS '更新时间';

-- 创建交易记录表
CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(64) PRIMARY KEY,
    transaction_id VARCHAR(64) NOT NULL,
    account_id VARCHAR(64) NOT NULL,
    type VARCHAR(20) NOT NULL,
    amount DECIMAL(20, 4) NOT NULL,
    balance_before DECIMAL(20, 4) NOT NULL,
    balance_after DECIMAL(20, 4) NOT NULL,
    order_id VARCHAR(64),
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建唯一索引：交易ID（幂等性保证）
CREATE UNIQUE INDEX IF NOT EXISTS uk_transaction_id ON transactions (transaction_id);

-- 创建索引：账户ID
CREATE INDEX IF NOT EXISTS idx_account_id ON transactions (account_id);

-- 创建索引：订单ID
CREATE INDEX IF NOT EXISTS idx_order_id ON transactions (order_id);

-- 创建索引：创建时间
CREATE INDEX IF NOT EXISTS idx_created_at ON transactions (created_at);

-- 创建索引：交易类型
CREATE INDEX IF NOT EXISTS idx_type ON transactions (type);

-- 注释
COMMENT ON TABLE transactions IS '交易记录表';
COMMENT ON COLUMN transactions.id IS '唯一ID';
COMMENT ON COLUMN transactions.transaction_id IS '业务交易ID(幂等性key)';
COMMENT ON COLUMN transactions.account_id IS '账户ID';
COMMENT ON COLUMN transactions.type IS '交易类型(CHARGE/DEDUCT/FREEZE/UNFREEZE/REFUND/ADJUST)';
COMMENT ON COLUMN transactions.amount IS '交易金额';
COMMENT ON COLUMN transactions.balance_before IS '交易前余额';
COMMENT ON COLUMN transactions.balance_after IS '交易后余额';
COMMENT ON COLUMN transactions.order_id IS '关联订单ID';
COMMENT ON COLUMN transactions.description IS '交易描述';
COMMENT ON COLUMN transactions.metadata IS '扩展元数据(JSONB)';
COMMENT ON COLUMN transactions.created_at IS '创建时间';
