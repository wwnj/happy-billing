# Happy Billing - 企业级订单账单系统

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-latest-brightgreen.svg)](docs/design/README.md)

Happy Billing 是一个专为 AI 云 PaaS 平台设计的企业级订单账单系统，支持 GPU 算力、存储、LLM Token 等多种产品类型的计费计量。

## ✨ 核心特性

- 🚀 **高可扩展性**: 支持新产品类型快速接入，无需修改核心代码
- 🔌 **易接入**: 提供标准化的计量上报和定价配置接口
- 🎯 **可插拔**: 定价规则、支付方式、结算模式均支持灵活配置
- 💰 **混合计费**: 支持预付费（包年包月）和后付费（按量计费）两种模式
- 🌍 **多币种**: 支持 CNY、USD、EUR、JPY 等多币种计费和结算
- 📊 **双存储**: MySQL 存储业务数据 + ClickHouse 存储海量计量数据
- 🔐 **多租户**: 完整的租户体系（Tenant → Organization → Project → User）
- ⚡ **高性能**: 单节点支持 10 万 TPS 计量数据写入

## 🛠️ 技术栈

- **应用框架**: Go 1.25 + Gin/Echo
- **业务数据库**: MySQL 8.0（订单、账单、用户）
- **计量数据库**: ClickHouse 23.x+（海量时序数据）
- **缓存**: Redis 7.x+（定价规则、汇率、会话）
- **消息队列**: Kafka 3.x+（事件驱动）
- **容器化**: Docker + Kubernetes
- **监控**: Prometheus + Grafana
- **日志**: ELK Stack

## 📁 项目结构

```
happy-billing/
├── cmd/                      # 主程序入口
│   ├── api/                  # API 服务
│   ├── worker/               # 后台任务
│   └── migrate/              # 数据库迁移
├── internal/                 # 私有应用代码
│   ├── api/                  # API 处理层
│   ├── service/              # 业务逻辑层
│   ├── repository/           # 数据访问层
│   ├── models/               # 数据模型
│   └── worker/               # 后台任务
├── pkg/                      # 公共库代码
│   ├── logger/               # 日志库
│   ├── database/             # 数据库连接
│   └── utils/                # 工具函数
├── config/                   # 配置文件
├── migrations/               # 数据库迁移文件
├── docs/                     # 文档
│   ├── design/               # 设计文档
│   └── development/          # 开发文档
└── test/                     # 测试文件
```

详细项目结构说明请参考：[代码规范 - 项目结构](docs/development/code-standards.md#项目结构规范)

## 🚀 快速开始

### 前置条件

- Go 1.25+
- MySQL 8.0+
- ClickHouse 23.x+
- Redis 7.x+
- Kafka 3.x+（可选，开发环境可跳过）

### 安装依赖

```bash
# 克隆项目
git clone https://github.com/wwnj/happy-billing.git
cd happy-billing

# 安装 Go 依赖
go mod download
```

### 配置环境

```bash
# 复制配置文件
cp config/config.example.yaml config/config.yaml

# 编辑配置文件，填入数据库连接信息
vim config/config.yaml
```

### 数据库迁移

```bash
# 运行数据库迁移
make migrate-up

# 或者手动执行
go run cmd/migrate/main.go up
```

### 启动服务

```bash
# 启动 API 服务
make run-api

# 或者
go run cmd/api/main.go
```

### 验证服务

```bash
# 健康检查
curl http://localhost:8080/health

# 查看 API 文档
open http://localhost:8080/swagger/index.html
```

## 📚 文档

### 设计文档

完整的系统设计文档请访问：[docs/design/](docs/design/)

- [系统概述](docs/design/00-overview.md) - 业务背景和核心特性
- [系统架构](docs/design/01-architecture.md) - 整体架构和服务划分
- [租户模型](docs/design/02-tenant-models.md) - 租户、组织、项目、用户体系
- [产品模型](docs/design/03-product-models.md) - SPU/SKU 产品抽象
- [定价模型](docs/design/04-pricing-models.md) - 定价引擎和规则配置
- [订单模型](docs/design/05-order-models.md) - 订单和资源实例
- [账单模型](docs/design/06-billing-models.md) - 账单生成和计量聚合
- [支付结算](docs/design/07-payment-settlement.md) - 支付流程和结算管理
- [多币种支持](docs/design/08-multi-currency.md) - 汇率管理和币种转换
- [架构图](docs/design/09-architecture-diagrams.md) - 系统架构可视化

### 开发文档

- **[📖 代码规范](docs/development/code-standards.md)** ⭐ **新成员必读**
  - Go 代码规范
  - 项目结构规范
  - 命名规范
  - 注释规范
  - 错误处理规范
  - 数据库规范
  - API 设计规范
  - Git 提交规范
  - 测试规范

- [API 文档](docs/api/) - RESTful API 接口文档（待补充）
- [部署指南](docs/design/10-implementation-guide.md) - 生产环境部署指南

## 🔨 开发指南

### 代码格式化

```bash
# 格式化代码
make fmt

# 代码检查
make lint
```

### 运行测试

```bash
# 运行所有测试
make test

# 运行单元测试
go test -v ./internal/...

# 运行集成测试
go test -v -tags=integration ./test/integration/...

# 查看测试覆盖率
make coverage
```

### 构建项目

```bash
# 构建所有服务
make build

# 构建 API 服务
make build-api

# 构建 Worker 服务
make build-worker
```

### Git 提交规范

本项目遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型**：
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `perf`: 性能优化
- `test`: 测试代码
- `chore`: 构建工具或辅助工具变动

**示例**：
```
feat(order): 添加包年包月订单创建功能

- 新增 CreateSubscriptionOrder 方法
- 支持 SKU 选择和数量配置
- 自动生成账单并关联订单

Closes #123
```

详细规范请参考：[Git 提交规范](docs/development/code-standards.md#git-提交规范)

## 🤝 贡献指南

我们欢迎所有形式的贡献，包括但不限于：

- 🐛 提交 Bug 报告
- 💡 提出新功能建议
- 📝 改进文档
- 🔧 提交代码修复或新功能

### 贡献流程

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat(module): add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 代码审查

所有 Pull Request 都需要通过以下检查：

- ✅ 代码格式化检查 (gofmt)
- ✅ 代码质量检查 (golangci-lint)
- ✅ 单元测试通过
- ✅ 测试覆盖率 ≥ 80%
- ✅ 至少 1 位 Maintainer 审查通过

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

感谢所有为本项目做出贡献的开发者！

## 📞 联系方式

- **项目主页**: https://github.com/wwnj/happy-billing
- **问题反馈**: https://github.com/wwnj/happy-billing/issues
- **邮件**: billing-support@example.com

---

**⭐ 如果这个项目对你有帮助，请给我们一个 Star！**
