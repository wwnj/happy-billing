.PHONY: help fmt lint test coverage build clean run-api run-worker migrate-up migrate-down docker-build docker-up docker-down

# 默认目标：显示帮助信息
help:
	@echo "Happy Billing - Makefile 命令"
	@echo ""
	@echo "开发命令:"
	@echo "  make fmt           - 格式化代码"
	@echo "  make lint          - 代码质量检查"
	@echo "  make test          - 运行所有测试"
	@echo "  make coverage      - 查看测试覆盖率"
	@echo "  make run-api       - 启动 API 服务"
	@echo "  make run-worker    - 启动 Worker 服务"
	@echo ""
	@echo "构建命令:"
	@echo "  make build         - 构建所有服务"
	@echo "  make build-api     - 构建 API 服务"
	@echo "  make build-worker  - 构建 Worker 服务"
	@echo "  make clean         - 清理构建文件"
	@echo ""
	@echo "数据库命令:"
	@echo "  make migrate-up    - 执行数据库迁移"
	@echo "  make migrate-down  - 回滚数据库迁移"
	@echo ""
	@echo "Docker 命令:"
	@echo "  make docker-build  - 构建 Docker 镜像"
	@echo "  make docker-up     - 启动 Docker 容器"
	@echo "  make docker-down   - 停止 Docker 容器"

# 格式化代码
fmt:
	@echo "==> 格式化 Go 代码..."
	@gofmt -w .
	@echo "==> 优化 import..."
	@goimports -w .
	@echo "✅ 代码格式化完成"

# 代码质量检查
lint:
	@echo "==> 运行 golangci-lint..."
	@golangci-lint run --timeout 5m
	@echo "✅ 代码检查通过"

# 运行所有测试
test:
	@echo "==> 运行所有测试..."
	@go test -v -race ./...
	@echo "✅ 测试完成"

# 运行单元测试
test-unit:
	@echo "==> 运行单元测试..."
	@go test -v -short ./internal/...
	@echo "✅ 单元测试完成"

# 运行集成测试
test-integration:
	@echo "==> 运行集成测试..."
	@go test -v -tags=integration ./test/integration/...
	@echo "✅ 集成测试完成"

# 查看测试覆盖率
coverage:
	@echo "==> 生成测试覆盖率报告..."
	@go test -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告已生成: coverage.html"
	@go tool cover -func=coverage.out | grep total

# 构建所有服务
build: build-api build-worker build-migrate
	@echo "✅ 所有服务构建完成"

# 构建 API 服务
build-api:
	@echo "==> 构建 API 服务..."
	@mkdir -p bin
	@go build -o bin/api cmd/api/main.go
	@echo "✅ API 服务构建完成: bin/api"

# 构建 Worker 服务
build-worker:
	@echo "==> 构建 Worker 服务..."
	@mkdir -p bin
	@go build -o bin/worker cmd/worker/main.go
	@echo "✅ Worker 服务构建完成: bin/worker"

# 构建迁移工具
build-migrate:
	@echo "==> 构建迁移工具..."
	@mkdir -p bin
	@go build -o bin/migrate cmd/migrate/main.go
	@echo "✅ 迁移工具构建完成: bin/migrate"

# 清理构建文件
clean:
	@echo "==> 清理构建文件..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean
	@echo "✅ 清理完成"

# 启动 API 服务
run-api:
	@echo "==> 启动 API 服务..."
	@go run cmd/api/main.go

# 启动 Worker 服务
run-worker:
	@echo "==> 启动 Worker 服务..."
	@go run cmd/worker/main.go

# 执行数据库迁移
migrate-up:
	@echo "==> 执行数据库迁移..."
	@go run cmd/migrate/main.go up
	@echo "✅ 数据库迁移完成"

# 回滚数据库迁移
migrate-down:
	@echo "==> 回滚数据库迁移..."
	@go run cmd/migrate/main.go down
	@echo "✅ 数据库回滚完成"

# 构建 Docker 镜像
docker-build:
	@echo "==> 构建 Docker 镜像..."
	@docker build -t happy-billing-api:latest -f Dockerfile.api .
	@docker build -t happy-billing-worker:latest -f Dockerfile.worker .
	@echo "✅ Docker 镜像构建完成"

# 启动 Docker 容器（使用 docker-compose）
docker-up:
	@echo "==> 启动 Docker 容器..."
	@docker-compose up -d
	@echo "✅ Docker 容器已启动"

# 停止 Docker 容器
docker-down:
	@echo "==> 停止 Docker 容器..."
	@docker-compose down
	@echo "✅ Docker 容器已停止"

# 安装开发工具
install-tools:
	@echo "==> 安装开发工具..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ 开发工具安装完成"

# 生成 Swagger 文档
swagger:
	@echo "==> 生成 Swagger 文档..."
	@swag init -g cmd/api/main.go -o docs/swagger
	@echo "✅ Swagger 文档生成完成"

# 检查代码风格和最佳实践
check: fmt lint test
	@echo "✅ 所有检查通过"
