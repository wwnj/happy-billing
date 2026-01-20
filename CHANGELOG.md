# 更新日志

本文件记录 Happy Billing 项目的重要更新和变更。

## [2026-01-20] 数据库初始化和可观测性增强

### ✨ 新增功能

#### 一键启动系统
- 🚀 **新增** `scripts/start-docker.sh` - 一键启动脚本
  - 自动启动所有 Docker 服务
  - 智能检测数据库状态
  - 自动初始化数据库表结构和测试数据
  - 友好的终端输出和进度提示

#### 数据库迁移系统
- 📋 **新增** `migrations/mysql/init.sh` - 数据库初始化脚本
  - 支持 Docker 容器模式和本地 MySQL 模式
  - 按正确顺序执行 DDL 和 DML
  - 包含完整的错误处理和状态检查

- 🔍 **新增** `migrations/mysql/check.sh` - 数据库状态检查脚本
  - 显示表数量统计
  - 显示各模块数据量
  - 展示测试数据示例

- 📖 **新增** `migrations/README.md` - 完整的迁移文档
  - 目录结构说明
  - 使用方法详解
  - DDL/DML 执行顺序
  - 故障排查指南

#### 测试数据
- 🧪 **自动生成** 完整的测试数据集
  - 3 个测试租户（个人/企业/演示账户）
  - 8 个产品分类，5 个 SKU
  - 4 条价格规则，2 条折扣规则
  - 9 个测试订单，涵盖不同状态

#### 可观测性增强
- 🔍 **集成** OpenTelemetry 链路追踪
  - Jaeger UI：http://localhost:16686
  - 自动追踪 HTTP 请求和数据库操作

- 📊 **集成** Loki + Promtail + Grafana 日志系统
  - Grafana UI：http://localhost:3000
  - 结构化 JSON 日志
  - trace_id 和 span_id 关联
  - GORM SQL 日志采集

- 🎨 **自定义** GORM 日志记录器
  - JSON 格式输出
  - 自动包含 trace_id 和 span_id
  - SQL 执行时间统计
  - 慢查询检测（>200ms）

### 🔧 优化改进

#### 配置管理
- **更新** `.env.example` - 添加新服务配置
  - Grafana 端口配置
  - Loki 端口配置
  - Promtail 端口配置

#### Docker Compose
- **更新** `docker-compose.yml`
  - 添加 Grafana 服务
  - 添加 Loki 日志聚合服务
  - 添加 Promtail 日志采集服务
  - MySQL 容器映射 migrations 目录

#### Makefile
- **新增** `make migrate-init` - 初始化数据库
- **新增** `make migrate-check` - 检查数据库状态
- **新增** `make docker-start` - 一键启动所有服务
- **新增** `make docker-logs` - 查看 Docker 日志
- **更新** 帮助信息，标注推荐命令

### 📚 文档更新

#### 主 README.md
- ✅ 更新核心特性，添加可观测性说明
- ✅ 重构技术栈部分，分类更清晰
- ✅ 完全重写快速开始部分
  - 突出一键启动方式
  - 添加服务访问地址表格
  - 添加测试数据说明
  - 添加 API 测试示例
- ✅ 更新文档索引，添加快速启动指南链接

#### 新增文档
- ✅ `QUICKSTART.md` - 详细的快速启动指南
  - 一键启动流程
  - 服务访问地址
  - 测试数据说明
  - 故障排查指南

- ✅ `migrations/README.md` - 数据库迁移文档
  - 目录结构说明
  - 使用方法
  - 执行顺序
  - 维护指南

#### 前端 README.md
- ✅ 更新快速开始部分
- ✅ 添加后端启动步骤
- ✅ 添加测试数据说明
- ✅ 添加快速启动指南链接

### 🎯 应用的工程原则

#### KISS（简单至上）
- 一键启动脚本，用户体验极简
- 智能检测，减少手动配置

#### DRY（避免重复）
- 统一的 MySQL 命令封装函数
- 复用的日志字段构建逻辑

#### 单一职责
- `init.sh` 专注数据库初始化
- `check.sh` 专注状态验证
- `start-docker.sh` 专注服务编排

#### 幂等性设计
- 使用 `CREATE TABLE IF NOT EXISTS`
- 支持多次执行而不会破坏数据
- 智能检测是否需要初始化

### 🔄 迁移指南

对于现有开发者：

```bash
# 1. 拉取最新代码
git pull

# 2. 停止现有容器
docker compose down

# 3. 使用新的一键启动脚本
bash scripts/start-docker.sh

# 4. 启动后端和前端
make run-api
cd ../happy-billing-frontend && npm run dev
```

### 📊 数据库结构

初始化后的数据库包含：
- **18 张业务表**（完整的业务模型）
- **测试数据**：
  - 3 个租户，3 个组织，3 个项目
  - 8 个产品分类，2 个 SPU，5 个 SKU
  - 4 条价格规则，2 条折扣规则
  - 9 个订单，8 个订单项，7 张账单
  - 15 条汇率数据

### 🐛 问题修复

- 修复 Promtail 日志字段丢失问题
- 修复 GORM 日志缺少 trace_id 问题
- 修复 Loki 配置版本兼容性问题

### ⚠️ 破坏性变更

无破坏性变更。所有更新向后兼容。

---

## 贡献者

- [@wwnj](https://github.com/wwnj) - 数据库初始化系统、可观测性集成、文档更新
