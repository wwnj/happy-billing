-- Happy Billing ClickHouse 初始化脚本

-- 创建数据库
CREATE DATABASE IF NOT EXISTS happy_billing;

-- 创建健康检查表
CREATE TABLE IF NOT EXISTS happy_billing.health_check
(
    `id` UInt32,
    `status` String,
    `timestamp` DateTime DEFAULT now()
)
ENGINE = MergeTree()
ORDER BY id
COMMENT '健康检查表';

-- 插入测试数据
INSERT INTO happy_billing.health_check (id, status) VALUES (1, 'ok');
