#!/bin/bash
set -e

# Happy Billing 快速启动脚本

echo "========================================="
echo "  Happy Billing 一键启动"
echo "========================================="
echo ""

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Error: Docker 未安装"
    echo "请访问 https://docs.docker.com/get-docker/ 安装 Docker"
    exit 1
fi

echo "✅ Docker 环境检查通过"
echo ""

# 创建 .env 文件
if [ ! -f .env ]; then
    echo "📝 创建 .env 配置文件..."
    cp .env.example .env
    echo "✅ .env 文件已创建"
else
    echo "✅ .env 文件已存在"
fi
echo ""

# 启动 Docker 服务
echo "🚀 启动 Docker 服务..."
docker compose up -d

echo ""
echo "⏳ 等待服务就绪..."
echo "   等待 MySQL 启动..."
sleep 10

# 检查服务状态
echo ""
echo "📊 检查服务状态..."
docker compose ps

# 初始化数据库
echo ""
echo "========================================="
echo "  数据库初始化"
echo "========================================="
echo ""

# 检查是否需要初始化数据库
NEED_INIT=false

# 检查数据库中是否有表
TABLE_COUNT=$(docker compose exec -T mysql mysql -ubilling_user -pbilling_pass_2024 -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'happy_billing'" 2>/dev/null | tail -1 || echo "0")

if [ "$TABLE_COUNT" -eq "0" ]; then
    NEED_INIT=true
    echo "📋 检测到空数据库，将执行初始化..."
else
    echo "✅ 数据库已存在 $TABLE_COUNT 个表"
    read -p "是否重新初始化数据库? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        NEED_INIT=true
        echo "⚠️  将清空并重新初始化数据库..."
    fi
fi

if [ "$NEED_INIT" = true ]; then
    echo ""
    echo "🔧 执行数据库初始化脚本..."

    # 如果是重新初始化，设置 RESET_DB 环境变量
    if [ "$TABLE_COUNT" -gt "0" ]; then
        export RESET_DB=true
    fi

    bash ./migrations/mysql/init.sh

    if [ $? -eq 0 ]; then
        echo "✅ 数据库初始化成功！"
    else
        echo "❌ 数据库初始化失败，请检查日志"
        exit 1
    fi
else
    echo "⏭️  跳过数据库初始化"
fi

echo ""
echo "✅ 所有服务已启动！"
echo ""
echo "========================================="
echo "  访问地址"
echo "========================================="
echo "前端 UI:       http://localhost:5173"
echo "API 服务:      http://localhost:8080"
echo "Jaeger UI:     http://localhost:16686"
echo "Grafana UI:    http://localhost:3000"
echo "               (默认账号: admin/admin)"
echo ""
echo "========================================="
echo "  下一步"
echo "========================================="
echo "1. 启动后端 API 服务:"
echo "   cd happy-billing && make run-api"
echo ""
echo "2. 启动前端开发服务:"
echo "   cd happy-billing-frontend && npm run dev"
echo ""
echo "3. 测试健康检查:"
echo "   curl http://localhost:8080/health"
echo ""
echo "4. 查看服务日志:"
echo "   docker compose logs -f [服务名]"
echo "   服务名: mysql, clickhouse, redis, kafka, jaeger, loki, promtail, grafana"
echo ""
echo "5. 停止所有服务:"
echo "   docker compose down"
echo ""
echo "========================================="
