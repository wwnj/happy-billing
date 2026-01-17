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

# 检查 Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Error: Docker Compose 未安装"
    echo "请访问 https://docs.docker.com/compose/install/ 安装 Docker Compose"
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
docker-compose up -d

echo ""
echo "⏳ 等待服务就绪（约30秒）..."
sleep 10

# 检查服务状态
echo ""
echo "📊 检查服务状态..."
docker-compose ps

echo ""
echo "✅ 所有服务已启动！"
echo ""
echo "========================================="
echo "  访问地址"
echo "========================================="
echo "Jaeger UI:     http://localhost:16686"
echo "API 服务:      http://localhost:8080"
echo ""
echo "========================================="
echo "  下一步"
echo "========================================="
echo "1. 启动 API 服务:"
echo "   ./bin/api"
echo ""
echo "2. 测试健康检查:"
echo "   curl http://localhost:8080/health"
echo ""
echo "3. 查看服务日志:"
echo "   docker-compose logs -f"
echo ""
echo "4. 停止所有服务:"
echo "   docker-compose down"
echo ""
echo "========================================="
