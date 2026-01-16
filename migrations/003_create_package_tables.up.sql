-- 创建套餐包表
CREATE TABLE IF NOT EXISTS packages (
    id VARCHAR(64) PRIMARY KEY,
    package_no VARCHAR(64) UNIQUE NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    total_quota DECIMAL(20, 4) NOT NULL,
    used_quota DECIMAL(20, 4) NOT NULL DEFAULT 0,
    remaining_quota DECIMAL(20, 4) NOT NULL,
    quota_unit VARCHAR(20) NOT NULL,
    price DECIMAL(20, 4) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    purchased_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    valid_from TIMESTAMP NOT NULL,
    valid_to TIMESTAMP NOT NULL,
    exhausted_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建唯一索引：套餐包号
CREATE UNIQUE INDEX IF NOT EXISTS uk_package_no ON packages (package_no);

-- 创建索引：租户ID + 用户ID
CREATE INDEX IF NOT EXISTS idx_tenant_user ON packages (tenant_id, user_id);

-- 创建索引：套餐包类型
CREATE INDEX IF NOT EXISTS idx_type ON packages (type);

-- 创建索引：套餐包状态
CREATE INDEX IF NOT EXISTS idx_status ON packages (status);

-- 创建索引：过期时间
CREATE INDEX IF NOT EXISTS idx_valid_to ON packages (valid_to);

-- 创建复合索引：状态 + 过期时间（用于过期扫描）
CREATE INDEX IF NOT EXISTS idx_status_valid_to ON packages (status, valid_to);

-- 创建复合索引：租户 + 用户 + 类型 + 状态（用于可用套餐包查询）
CREATE INDEX IF NOT EXISTS idx_tenant_user_type_status ON packages (tenant_id, user_id, type, status);

-- 注释
COMMENT ON TABLE packages IS '套餐包表';
COMMENT ON COLUMN packages.id IS '套餐包ID';
COMMENT ON COLUMN packages.package_no IS '套餐包号';
COMMENT ON COLUMN packages.tenant_id IS '租户ID';
COMMENT ON COLUMN packages.user_id IS '用户ID';
COMMENT ON COLUMN packages.type IS '套餐包类型(GPU/STORAGE/TOKEN/TRAFFIC)';
COMMENT ON COLUMN packages.status IS '套餐包状态(ACTIVE/EXPIRED/EXHAUSTED/CANCELLED)';
COMMENT ON COLUMN packages.name IS '套餐包名称';
COMMENT ON COLUMN packages.description IS '描述';
COMMENT ON COLUMN packages.total_quota IS '总配额';
COMMENT ON COLUMN packages.used_quota IS '已使用配额';
COMMENT ON COLUMN packages.remaining_quota IS '剩余配额';
COMMENT ON COLUMN packages.quota_unit IS '配额单位';
COMMENT ON COLUMN packages.price IS '购买价格';
COMMENT ON COLUMN packages.currency IS '货币类型';
COMMENT ON COLUMN packages.purchased_at IS '购买时间';
COMMENT ON COLUMN packages.valid_from IS '生效时间';
COMMENT ON COLUMN packages.valid_to IS '过期时间';
COMMENT ON COLUMN packages.exhausted_at IS '耗尽时间';
COMMENT ON COLUMN packages.cancelled_at IS '取消时间';
COMMENT ON COLUMN packages.metadata IS '扩展元数据(JSONB)';
COMMENT ON COLUMN packages.created_at IS '创建时间';
COMMENT ON COLUMN packages.updated_at IS '更新时间';
