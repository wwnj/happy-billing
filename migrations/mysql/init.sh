#!/bin/bash
# ============================================================================
# Happy Billing MySQL 数据库初始化脚本
# 说明: 初始化数据库表结构和测试数据
# ============================================================================

# 注意：不使用 set -e，手动检查关键步骤

# 配置
MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-billing_user}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-billing_pass_2024}"
MYSQL_DATABASE="${MYSQL_DATABASE:-happy_billing}"

MIGRATIONS_DIR="$(cd "$(dirname "$0")/.." && pwd)"
USE_DOCKER="${USE_DOCKER:-true}"  # 默认使用 Docker 容器执行
RESET_DB="${RESET_DB:-false}"      # 是否重置数据库

# MySQL 命令包装函数
mysql_exec() {
    if [ "$USE_DOCKER" = "true" ]; then
        docker compose exec -T mysql mysql -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" "$@" 2>&1 | grep -v "Using a password" || true
    else
        mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" "$@" 2>&1 || true
    fi
}

mysql_exec_file() {
    local file="$1"
    if [ "$USE_DOCKER" = "true" ]; then
        docker compose exec -T mysql bash -c "mysql -u${MYSQL_USER} -p${MYSQL_PASSWORD} ${MYSQL_DATABASE} < /migrations/$(basename ${file}) 2>&1" | grep -v "Using a password" || true
    else
        mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" "${MYSQL_DATABASE}" < "${file}" 2>&1 || true
    fi
}

echo "========================================="
echo "  MySQL 数据库初始化"
echo "========================================="
echo "数据库地址: ${MYSQL_HOST}:${MYSQL_PORT}"
echo "数据库名称: ${MYSQL_DATABASE}"
echo "用户名称:   ${MYSQL_USER}"
echo "执行方式:   $([ "$USE_DOCKER" = "true" ] && echo "Docker 容器" || echo "本地 MySQL")"
echo ""

# 等待 MySQL 就绪
echo "⏳ 等待 MySQL 服务就绪..."
for i in {1..30}; do
    if mysql_exec -e "SELECT 1" >/dev/null 2>&1; then
        echo "✅ MySQL 服务已就绪"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ 错误: MySQL 服务未就绪，超时退出"
        exit 1
    fi
    echo "   等待中... ($i/30)"
    sleep 2
done

# 如果需要重置数据库
if [ "$RESET_DB" = "true" ]; then
    echo ""
    echo "🗑️  重置数据库..."
    mysql_exec -e "DROP DATABASE IF EXISTS ${MYSQL_DATABASE};" > /dev/null
    mysql_exec -e "CREATE DATABASE ${MYSQL_DATABASE} DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" > /dev/null
    echo "✅ 数据库已重置"
else
    # 确保数据库存在（即使不是重置模式）
    echo ""
    echo "🔍 检查数据库是否存在..."
    DB_EXISTS=$(mysql_exec -sN -e "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = '${MYSQL_DATABASE}'" 2>/dev/null | head -1)
    if [ -z "$DB_EXISTS" ]; then
        echo "📝 数据库不存在，创建数据库..."
        CREATE_RESULT=$(mysql_exec -e "CREATE DATABASE IF NOT EXISTS ${MYSQL_DATABASE} DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>&1)

        # 验证数据库是否创建成功
        sleep 1
        DB_CHECK=$(mysql_exec -sN -e "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = '${MYSQL_DATABASE}'" 2>/dev/null | head -1)
        if [ -z "$DB_CHECK" ]; then
            echo "❌ 错误: 数据库创建失败"
            echo "$CREATE_RESULT"
            exit 1
        fi
        echo "✅ 数据库已创建"
    else
        echo "✅ 数据库已存在"
    fi
fi

echo ""
echo "📋 执行数据库初始化..."
echo ""

# 执行 DDL（按顺序）
echo "1️⃣  创建租户模块表..."
mysql_exec_file "${MIGRATIONS_DIR}/20260117_create_tenant_tables.sql"

echo "2️⃣  创建产品模块表..."
mysql_exec_file "${MIGRATIONS_DIR}/20260117_create_product_tables.sql"

echo "3️⃣  创建定价模块表..."
mysql_exec_file "${MIGRATIONS_DIR}/20260117_create_pricing_tables.sql"

echo "4️⃣  创建订单/账单/支付模块表..."
mysql_exec_file "${MIGRATIONS_DIR}/20260117_create_order_billing_payment_tables.sql"

echo "5️⃣  创建汇率表..."
mysql_exec_file "${MIGRATIONS_DIR}/20240117_create_exchange_rates.sql"

echo "6️⃣  添加多币种字段..."
mysql_exec_file "${MIGRATIONS_DIR}/20240117_add_multi_currency_fields.sql"

echo "7️⃣  修复外键约束..."
mysql_exec_file "${MIGRATIONS_DIR}/fix_foreign_keys.sql" 2>/dev/null || echo "   ⚠️  外键修复脚本可能已执行或不需要"

# 执行 DML（测试数据）
echo ""
echo "📝 插入测试数据..."
echo ""

echo "8️⃣  添加测试用户凭证..."
mysql_exec_file "${MIGRATIONS_DIR}/add_test_users_credentials.sql"

echo "9️⃣  插入完整测试数据..."
mysql_exec_file "${MIGRATIONS_DIR}/debug_and_test_data.sql"

echo ""
echo "========================================="
echo "  ✅ 数据库初始化完成！"
echo "========================================="

# 显示统计信息
TABLES_COUNT=$(mysql_exec -sN -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '${MYSQL_DATABASE}'" 2>/dev/null || echo "N/A")
TENANTS_COUNT=$(mysql_exec -sN -e "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.tenants" 2>/dev/null || echo "N/A")
PRODUCTS_COUNT=$(mysql_exec -sN -e "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.product_sku" 2>/dev/null || echo "N/A")
ORDERS_COUNT=$(mysql_exec -sN -e "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.orders" 2>/dev/null || echo "N/A")

echo "📊 数据统计:"
echo "   - 数据表数量: ${TABLES_COUNT}"
echo "   - 测试租户数: ${TENANTS_COUNT}"
echo "   - 测试产品数: ${PRODUCTS_COUNT}"
echo "   - 测试订单数: ${ORDERS_COUNT}"
echo ""
