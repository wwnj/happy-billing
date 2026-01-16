-- 创建账单表
CREATE TABLE IF NOT EXISTS bills (
    id VARCHAR(64) PRIMARY KEY,
    bill_no VARCHAR(64) UNIQUE NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    cycle VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    total_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    discount_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    actual_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    paid_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    outstanding_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    due_date TIMESTAMP,
    settled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建唯一索引：账单号
CREATE UNIQUE INDEX IF NOT EXISTS uk_bill_no ON bills (bill_no);

-- 创建索引：租户ID + 用户ID
CREATE INDEX IF NOT EXISTS idx_tenant_user ON bills (tenant_id, user_id);

-- 创建索引：状态
CREATE INDEX IF NOT EXISTS idx_status ON bills (status);

-- 创建索引：账期时间
CREATE INDEX IF NOT EXISTS idx_period ON bills (period_start, period_end);

-- 创建索引：到期日期
CREATE INDEX IF NOT EXISTS idx_due_date ON bills (due_date);

-- 注释
COMMENT ON TABLE bills IS '账单表';
COMMENT ON COLUMN bills.id IS '账单ID';
COMMENT ON COLUMN bills.bill_no IS '账单号';
COMMENT ON COLUMN bills.tenant_id IS '租户ID';
COMMENT ON COLUMN bills.user_id IS '用户ID';
COMMENT ON COLUMN bills.cycle IS '账单周期(MONTHLY/QUARTERLY/YEARLY/CUSTOM)';
COMMENT ON COLUMN bills.status IS '账单状态(PENDING/SETTLED/CANCELLED/OVERDUE)';
COMMENT ON COLUMN bills.period_start IS '账期开始时间';
COMMENT ON COLUMN bills.period_end IS '账期结束时间';
COMMENT ON COLUMN bills.total_amount IS '总金额';
COMMENT ON COLUMN bills.discount_amount IS '折扣金额';
COMMENT ON COLUMN bills.tax_amount IS '税额';
COMMENT ON COLUMN bills.actual_amount IS '实际应付金额';
COMMENT ON COLUMN bills.paid_amount IS '已支付金额';
COMMENT ON COLUMN bills.outstanding_amount IS '未付金额';
COMMENT ON COLUMN bills.currency IS '货币类型';
COMMENT ON COLUMN bills.due_date IS '到期日期';
COMMENT ON COLUMN bills.settled_at IS '结算时间';
COMMENT ON COLUMN bills.created_at IS '创建时间';
COMMENT ON COLUMN bills.updated_at IS '更新时间';

-- 创建账单明细表
CREATE TABLE IF NOT EXISTS bill_items (
    id VARCHAR(64) PRIMARY KEY,
    bill_id VARCHAR(64) NOT NULL,
    type VARCHAR(20) NOT NULL,
    order_id VARCHAR(64),
    description TEXT,
    amount DECIMAL(20, 4) NOT NULL,
    quantity DECIMAL(20, 4) NOT NULL DEFAULT 0,
    unit_price DECIMAL(20, 4) NOT NULL DEFAULT 0,
    discount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    total_amount DECIMAL(20, 4) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引：账单ID
CREATE INDEX IF NOT EXISTS idx_bill_id ON bill_items (bill_id);

-- 创建索引：订单ID
CREATE INDEX IF NOT EXISTS idx_order_id ON bill_items (order_id);

-- 创建索引：明细类型
CREATE INDEX IF NOT EXISTS idx_type ON bill_items (type);

-- 注释
COMMENT ON TABLE bill_items IS '账单明细表';
COMMENT ON COLUMN bill_items.id IS '明细ID';
COMMENT ON COLUMN bill_items.bill_id IS '账单ID';
COMMENT ON COLUMN bill_items.type IS '明细类型(ORDER/CHARGE/REFUND/ADJUST/DISCOUNT)';
COMMENT ON COLUMN bill_items.order_id IS '关联订单ID';
COMMENT ON COLUMN bill_items.description IS '描述';
COMMENT ON COLUMN bill_items.amount IS '金额';
COMMENT ON COLUMN bill_items.quantity IS '数量';
COMMENT ON COLUMN bill_items.unit_price IS '单价';
COMMENT ON COLUMN bill_items.discount IS '折扣金额';
COMMENT ON COLUMN bill_items.tax_amount IS '税额';
COMMENT ON COLUMN bill_items.total_amount IS '总金额（含税）';
COMMENT ON COLUMN bill_items.metadata IS '扩展元数据(JSONB)';
COMMENT ON COLUMN bill_items.created_at IS '创建时间';
