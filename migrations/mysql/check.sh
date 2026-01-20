#!/bin/bash
# ============================================================================
# Happy Billing 数据库状态检查脚本
# 说明: 检查数据库表和测试数据是否正确加载
# ============================================================================

set -e

# 配置
MYSQL_USER="${MYSQL_USER:-billing_user}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-billing_pass_2024}"
MYSQL_DATABASE="${MYSQL_DATABASE:-happy_billing}"

# MySQL 命令包装
mysql_query() {
    docker compose exec -T mysql mysql -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" -sN -e "$1" 2>&1 | grep -v "Using a password"
}

echo "========================================="
echo "  Happy Billing 数据库状态检查"
echo "========================================="
echo ""

# 检查数据库是否存在
DB_EXISTS=$(mysql_query "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = '${MYSQL_DATABASE}'")
if [ -z "$DB_EXISTS" ]; then
    echo "❌ 数据库 ${MYSQL_DATABASE} 不存在"
    exit 1
fi
echo "✅ 数据库: ${MYSQL_DATABASE}"

# 检查表数量
TABLE_COUNT=$(mysql_query "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '${MYSQL_DATABASE}'")
echo "📊 数据表数量: ${TABLE_COUNT}"

# 检查各模块的表
echo ""
echo "📋 模块表统计:"
echo "   租户模块:"
echo "     - tenants: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.tenants") 条记录"
echo "     - organizations: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.organizations") 条记录"
echo "     - projects: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.projects") 条记录"
echo "     - users: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.users") 条记录"
echo ""
echo "   产品模块:"
echo "     - product_categories: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.product_categories") 条记录"
echo "     - product_spu: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.product_spu") 条记录"
echo "     - product_sku: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.product_sku") 条记录"
echo ""
echo "   定价模块:"
echo "     - price_rules: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.price_rules") 条记录"
echo "     - discount_rules: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.discount_rules") 条记录"
echo ""
echo "   订单模块:"
echo "     - orders: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.orders") 条记录"
echo "     - order_items: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.order_items") 条记录"
echo "     - bills: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.bills") 条记录"
echo ""
echo "   支付模块:"
echo "     - payments: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.payments") 条记录"
echo "     - balance_accounts: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.balance_accounts") 条记录"
echo ""
echo "   汇率模块:"
echo "     - exchange_rates: $(mysql_query "SELECT COUNT(*) FROM ${MYSQL_DATABASE}.exchange_rates") 条记录"

echo ""
echo "========================================="
echo "  测试数据示例"
echo "========================================="
echo ""
echo "租户列表:"
mysql_query "USE ${MYSQL_DATABASE}; SELECT tenant_id, name, tenant_type, status FROM tenants LIMIT 5" | column -t -s $'\t'
echo ""
echo "产品SKU列表:"
mysql_query "USE ${MYSQL_DATABASE}; SELECT sku_code, sku_name, status FROM product_sku LIMIT 5" | column -t -s $'\t'
echo ""
echo "最近订单:"
mysql_query "USE ${MYSQL_DATABASE}; SELECT order_no, tenant_id, status, created_at FROM orders ORDER BY created_at DESC LIMIT 5" | column -t -s $'\t'

echo ""
echo "========================================="
echo "✅ 数据库状态正常！"
echo "========================================="
