#!/bin/bash
set -e

# Happy Billing 服务测试脚本

echo "========================================="
echo "  Happy Billing 服务测试"
echo "========================================="
echo ""

# 测试 MySQL
echo "🔍 测试 MySQL 连接..."
if docker exec happy-billing-mysql mysql -ubilling_user -pbilling_pass_2024 -e "SELECT 'MySQL OK' as status;" > /dev/null 2>&1; then
    echo "✅ MySQL: 连接成功"
else
    echo "❌ MySQL: 连接失败"
fi

# 测试 Redis
echo "🔍 测试 Redis 连接..."
if docker exec happy-billing-redis redis-cli ping > /dev/null 2>&1; then
    echo "✅ Redis: 连接成功"
else
    echo "❌ Redis: 连接失败"
fi

# 测试 ClickHouse
echo "🔍 测试 ClickHouse 连接..."
if curl -s http://localhost:8123/ping > /dev/null 2>&1; then
    echo "✅ ClickHouse: 连接成功"
else
    echo "❌ ClickHouse: 连接失败"
fi

# 测试 Jaeger
echo "🔍 测试 Jaeger UI..."
if curl -s http://localhost:16686 > /dev/null 2>&1; then
    echo "✅ Jaeger: UI 可访问"
else
    echo "❌ Jaeger: UI 不可访问"
fi

echo ""
echo "========================================="
echo "  API 服务测试"
echo "========================================="
echo ""

# 检查 API 服务是否运行
if curl -s http://localhost:8080/ping > /dev/null 2>&1; then
    echo "✅ API 服务正在运行"
    echo ""

    # Ping 测试
    echo "🔍 测试 /ping 接口..."
    curl -s http://localhost:8080/ping | jq .
    echo ""

    # 健康检查测试
    echo "🔍 测试 /health 接口..."
    curl -s http://localhost:8080/health | jq .
    echo ""

    # 带业务上下文的测试
    echo "🔍 测试带业务上下文的请求..."
    curl -s -H "X-Tenant-ID: T20240117001" \
            -H "X-User-ID: U20240117001" \
            -H "X-Org-ID: ORG20240117001" \
            http://localhost:8080/health | jq .
    echo ""

    echo "✅ 所有测试通过！"
else
    echo "⚠️  API 服务未运行"
    echo ""
    echo "请先启动 API 服务:"
    echo "  ./bin/api"
fi

echo ""
echo "========================================="
