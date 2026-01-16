# Happy Billing - 企业级订单账单系统

[![Go Version](https://img.shields.io/badge/Go-1.25.0-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

基于**DDD(领域驱动设计)**和**事件驱动架构**的企业级订单账单系统,支持多维度计量、混合付费模式、高并发扣费。

---

## 🎯 项目特点

- **DDD分层架构**: 6大核心领域(计量、定价、订单、账单、余额、套餐包),业务逻辑高内聚
- **可插拔设计**: 计量器和定价器通过插件注册,灵活扩展新资源类型
- **事件驱动**: 基于Kafka的异步事件流,支持高并发场景
- **混合计量策略**: GPU用秒级、存储用小时级,根据资源类型灵活配置
- **金额精度保证**: 使用decimal避免浮点误差,确保财务准确
- **高并发扣费**: Redis分布式锁 + 乐观锁 + 幂等性三重保证

---

## 🏗️ 技术栈

| 组件 | 技术选型 | 用途 |
|------|---------|------|
| **语言** | Go 1.25.0 | 高性能并发编程 |
| **架构** | DDD + CQRS + 事件驱动 | 复杂业务领域建模 |
| **消息队列** | Kafka | 削峰填谷、异步处理 |
| **关系数据库** | PostgreSQL | 订单账单事务数据 |
| **时序数据库** | InfluxDB | 海量计量数据存储 |
| **缓存** | Redis | 热点数据、分布式锁 |
| **配置管理** | Viper | 多环境配置 |
| **日志** | Zap | 高性能结构化日志 |
| **金额计算** | shopspring/decimal | 避免浮点精度问题 |

---

## 📁 项目结构

```
happy-billing/
├── cmd/                      # 应用程序入口
│   ├── api/                 # HTTP API服务
│   ├── worker/              # 异步Worker
│   └── migration/           # 数据库迁移工具
│
├── internal/                 # 私有应用代码
│   ├── domain/              # 领域层(核心业务逻辑)
│   │   ├── meter/          # 计量领域
│   │   ├── pricing/        # 定价领域
│   │   ├── order/          # 订单领域
│   │   ├── billing/        # 账单领域
│   │   ├── balance/        # 余额领域
│   │   └── package/        # 套餐包领域
│   │
│   ├── application/         # 应用层(用例编排)
│   │   ├── command/        # 命令(写操作)
│   │   └── query/          # 查询(读操作-CQRS)
│   │
│   ├── infrastructure/      # 基础设施层
│   │   ├── config/         # 配置管理
│   │   ├── logger/         # 日志系统
│   │   ├── persistence/    # 持久化
│   │   ├── messaging/      # 消息队列
│   │   └── lock/           # 分布式锁
│   │
│   └── interfaces/          # 接口层
│       ├── http/           # HTTP API
│       └── event/          # 事件处理器
│
├── plugins/                  # 可插拔插件
│   ├── meters/             # 计量器插件
│   └── pricers/            # 定价器插件
│
├── pkg/                      # 公共库
│   ├── errors/             # 错误处理
│   ├── money/              # 金额计算
│   ├── idempotent/         # 幂等性工具
│   └── pagination/         # 分页工具
│
├── configs/                  # 配置文件
├── migrations/               # 数据库迁移脚本
├── docs/                     # 文档
├── test/                     # 测试
└── Plan.md                   # 实施计划
```

---

## 🚀 快速开始

### 1. 环境要求

- Go 1.25.0+
- PostgreSQL 15+
- Redis 7.0+
- Kafka 3.0+
- InfluxDB 2.0+

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置文件

复制配置模板并修改:

```bash
cp configs/config.yaml configs/config.dev.yaml
# 编辑config.dev.yaml,填入实际的数据库/Redis/Kafka配置
```

### 4. 运行服务

```bash
# 启动API服务
go run cmd/api/main.go

# 启动Worker
go run cmd/worker/main.go
```

---

## 📊 核心领域模型

### 1. 计量领域(Meter)

**聚合根**: `MeterConfig`
**实体**: `MeterRecord`
**值对象**: `ResourceType`, `Precision`
**插件接口**: `MeterPlugin`

**功能**:
- 支持多种资源类型(GPU/存储/Token等)
- 可插拔计量器,灵活扩展
- 混合计量精度(秒/分钟/小时级)

**示例代码**:

```go
// 注册GPU计量器
gpuMeter := meters.NewGPUMeter()
meters.Register(gpuMeter)

// 采集计量数据
record, err := gpuMeter.Collect(ctx, "gpu-001", map[string]interface{}{
    "tenant_id": "tenant-123",
    "user_id": "user-456",
    "start_time": startTime,
    "end_time": endTime,
})
```

### 2. 定价领域(Pricing)

**聚合根**: `PriceRule`
**实体**: `TierPrice`
**插件接口**: `PricerPlugin`

**功能**:
- 阶梯定价
- 折扣规则
- 促销定价

### 3. 订单领域(Order)

**聚合根**: `Order`
**实体**: `OrderItem`
**领域事件**: `OrderCreated`, `OrderCompleted`

**功能**:
- 订单生命周期管理
- 状态机流转
- 事件发布

---

## 🔌 可插拔设计示例

### 自定义计量器插件

```go
package meters

import (
    "context"
    "github.com/wwnj/happy-billing/internal/domain/meter"
    "github.com/wwnj/happy-billing/pkg/money"
)

// 实现MeterPlugin接口
type MyCustomMeter struct{}

func (m *MyCustomMeter) Name() string {
    return "my_custom_meter"
}

func (m *MyCustomMeter) Collect(ctx context.Context, resourceID string, metadata map[string]interface{}) (*meter.MeterRecord, error) {
    // 自定义计量逻辑
    // ...
}

func (m *MyCustomMeter) Aggregate(ctx context.Context, records []*meter.MeterRecord) (money.Decimal, error) {
    // 自定义聚合逻辑
    // ...
}

func (m *MyCustomMeter) Validate(config map[string]interface{}) error {
    // 验证配置
    // ...
}

// 注册插件
func init() {
    meters.Register(&MyCustomMeter{})
}
```

---

## 📋 实施进度

### ✅ 阶段1: 基础设施与计量模块 (已完成)

- [x] 创建DDD分层目录结构
- [x] 初始化Go模块依赖
- [x] 配置管理模块(Viper + YAML)
- [x] 日志系统(Zap)
- [x] 错误处理封装
- [x] 金额计算工具(decimal)
- [x] 计量领域模型(entity/value_object/repository)
- [x] 计量器插件系统(registry + GPU示例)
- [x] 配置文件模板

### ⏳ 阶段2: 定价与订单模块 (待实施)

- [ ] 定价领域实现
- [ ] 订单领域实现
- [ ] 订单API实现
- [ ] 订单事件处理

### ⏳ 阶段3: 余额与扣费模块 (待实施)

### ⏳ 阶段4: 账单与对账模块 (待实施)

### ⏳ 阶段5: 套餐包模块 (待实施)

### ⏳ 阶段6: 高可用与性能优化 (待实施)

---

## 🧪 测试

```bash
# 运行单元测试
go test ./...

# 运行测试(带覆盖率)
go test -cover ./...

# 运行集成测试
go test -tags=integration ./test/integration/...
```

---

## 📖 文档

- [实施计划](Plan.md) - 完整的分阶段实施计划
- [架构设计](docs/architecture/) - 系统架构文档
- [API文档](docs/api/) - API接口文档
- [数据库设计](docs/database/) - 数据库表结构

---

## 🤝 贡献指南

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交变更 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交Pull Request

---

## 📝 设计原则

项目严格遵循以下原则:

- **SOLID**: 单一职责、开闭原则、里氏替换、接口隔离、依赖倒置
- **DRY**: 避免重复代码
- **KISS**: 保持简单
- **YAGNI**: 只实现当前需要的功能

---

## 📄 License

MIT License - 详见 [LICENSE](LICENSE) 文件

---

## 🙏 致谢

感谢所有开源项目的贡献者,本项目使用了以下优秀的开源库:

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web框架
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Zap](https://github.com/uber-go/zap) - 高性能日志
- [GORM](https://gorm.io/) - ORM框架
- [Decimal](https://github.com/shopspring/decimal) - 精确decimal计算
- [Kafka-Go](https://github.com/segmentio/kafka-go) - Kafka客户端

---

**Happy Billing** - 让计费更简单 🚀
