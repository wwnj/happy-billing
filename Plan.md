# Happy Billing - 企业级订单账单系统实施计划

## 一、项目概述

**目标**:构建一个高可用、可扩展、可插拔的订单账单系统,支持:
- 多维度计量模型(GPU算力、存储、LLM Token等)
- 混合计量策略(秒级/分钟级/小时级根据资源类型灵活配置)
- 混合付费模式(后付费、预付费、套餐包、组合模式)
- 高并发架构(事件驱动 + 消息队列)
- 双存储引擎(PostgreSQL + 时序数据库)

**技术栈**:
- 语言:Go 1.25.0
- 架构:DDD分层架构 + 事件驱动 + CQRS
- 消息队列:Kafka
- 数据库:PostgreSQL(订单账单) + InfluxDB(计量数据)
- 缓存:Redis(分布式锁 + 热点数据)
- 监控:Prometheus + Grafana

---

## 二、核心架构设计

### 2.1 DDD分层结构

```
happy-billing/
├── cmd/                    # 应用程序入口
│   ├── api/               # HTTP API服务
│   ├── worker/            # 异步处理Worker
│   └── migration/         # 数据库迁移工具
├── internal/              # 私有应用代码
│   ├── domain/           # 领域层(6个核心领域)
│   │   ├── meter/        # 计量领域
│   │   ├── pricing/      # 定价领域
│   │   ├── order/        # 订单领域
│   │   ├── billing/      # 账单领域
│   │   ├── balance/      # 余额领域
│   │   └── package/      # 套餐包领域
│   ├── application/       # 应用层(用例编排)
│   │   ├── command/      # 命令(写操作)
│   │   └── query/        # 查询(读操作 - CQRS)
│   ├── infrastructure/    # 基础设施层
│   │   ├── persistence/  # 持久化(PostgreSQL/InfluxDB/Redis)
│   │   ├── messaging/    # 消息队列(Kafka)
│   │   └── lock/         # 分布式锁(Redis Redlock)
│   └── interfaces/        # 接口层
│       ├── http/         # HTTP API
│       └── event/        # 事件处理器
├── plugins/               # 可插拔插件
│   ├── meters/           # 计量器插件(GPU/存储/Token)
│   └── pricers/          # 定价器插件(阶梯/折扣)
├── pkg/                   # 公共库
│   ├── errors/           # 错误处理
│   ├── money/            # 金额计算(decimal)
│   └── idempotent/       # 幂等性工具
└── migrations/            # 数据库迁移脚本
```

### 2.2 6大核心领域

| 领域 | 聚合根 | 核心职责 |
|------|--------|---------|
| **Meter(计量)** | MeterConfig | 资源计量采集、聚合(可插拔计量器) |
| **Pricing(定价)** | PriceRule | 定价规则、阶梯定价、折扣计算(可插拔定价器) |
| **Order(订单)** | Order | 订单生命周期管理、状态流转 |
| **Billing(账单)** | Bill | 账单生成、对账、结算 |
| **Balance(余额)** | Account | 余额管理、充值、扣费(带分布式锁和幂等性) |
| **Package(套餐包)** | Package | 资源包购买、配额抵扣、过期管理 |

### 2.3 核心数据流

```
计量采集 → Kafka(meter.collected) → Worker聚合 → InfluxDB
                                        ↓
                              定价计算 → 订单生成 → Kafka(order.created)
                                                        ↓
                              Worker扣费 → 分布式锁 → 余额扣除(幂等性)
                                                        ↓
                                      账单生成 → PostgreSQL
```

### 2.4 关键技术方案

| 技术场景 | 解决方案 |
|---------|---------|
| **高并发扣费** | Redis分布式锁 + 乐观锁(version) + 幂等性(transaction_id) |
| **计量高写入** | Kafka削峰 + InfluxDB批量写入(每批1000条) |
| **分布式事务** | Saga模式(事件驱动编排) + 补偿机制 |
| **数据一致性** | 事件溯源 + T+1对账 + 审计日志 |
| **缓存策略** | Cache-Aside模式,热点定价规则/用户余额缓存 |
| **金额精度** | 使用decimal.Decimal避免浮点误差 |

---

## 三、分阶段实施计划

### 【阶段1】基础设施与计量模块(2周)

**目标**:搭建项目骨架,实现计量数据采集与存储

#### 实施步骤:

1. **项目初始化**
   - 创建DDD目录结构
   - 配置管理(Viper + YAML)
   - 日志系统(Zap)
   - 错误处理封装(pkg/errors)

2. **基础设施层**
   - PostgreSQL连接池(GORM/sqlx混用)
   - InfluxDB客户端封装
   - Redis客户端封装
   - Kafka生产者/消费者封装

3. **计量领域实现**
   ```
   关键文件:
   - internal/domain/meter/entity.go          # MeterRecord/MeterConfig实体
   - internal/domain/meter/value_object.go    # ResourceType/Precision值对象
   - internal/domain/meter/meter_plugin.go    # 计量器插件接口
   - internal/domain/meter/repository.go      # 仓储接口
   - internal/infrastructure/persistence/timeseries/influxdb/meter_repo.go  # InfluxDB实现
   - plugins/meters/gpu_meter.go              # GPU计量器插件示例
   - plugins/meters/registry.go               # 插件注册中心
   ```

4. **计量采集API**
   ```
   接口:POST /api/v1/meters/collect
   请求体:
   {
     "resource_id": "gpu_123",
     "resource_type": "GPU",
     "quantity": 3600,
     "unit": "seconds",
     "start_time": "2025-01-01T00:00:00Z",
     "end_time": "2025-01-01T01:00:00Z"
   }
   ```

5. **计量聚合Worker**
   - 消费Kafka topic: `meter.collected`
   - 批量写入InfluxDB

#### 验收标准:
- ✅ 通过API采集计量数据
- ✅ 数据正确存入InfluxDB
- ✅ 可按时间范围查询聚合数据
- ✅ 单元测试覆盖率 > 80%

---

### 【阶段2】定价与订单模块(2周)

**目标**:实现定价规则引擎和订单生命周期管理

#### 实施步骤:

1. **定价领域实现**
   ```
   关键文件:
   - internal/domain/pricing/entity.go        # PriceRule/TierPrice实体
   - internal/domain/pricing/pricer_plugin.go # 定价器插件接口
   - plugins/pricers/tier_pricer.go           # 阶梯定价器实现
   - plugins/pricers/discount_pricer.go       # 折扣定价器实现
   ```

2. **订单领域实现**
   ```
   关键文件:
   - internal/domain/order/aggregate.go       # Order聚合根(状态机)
   - internal/domain/order/entity.go          # OrderItem实体
   - internal/domain/order/events.go          # 领域事件(Created/Completed/Cancelled)
   - internal/infrastructure/persistence/postgres/order_repo.go  # PostgreSQL实现
   ```

3. **订单API实现**
   ```
   接口:
   - POST   /api/v1/orders              # 创建订单
   - GET    /api/v1/orders/:id          # 查询订单详情
   - GET    /api/v1/orders              # 订单列表(分页)
   - PUT    /api/v1/orders/:id/cancel   # 取消订单
   ```

4. **订单事件处理**
   - 发布`order.created`事件到Kafka
   - OrderEventHandler消费处理

#### 验收标准:
- ✅ 根据计量数据自动计算费用(阶梯定价)
- ✅ 创建订单并持久化到PostgreSQL
- ✅ 订单状态机流转正确
- ✅ 领域事件正确发布

---

### 【阶段3】余额与扣费模块(2周)

**目标**:实现预付费余额管理和高并发扣费逻辑

#### 实施步骤:

1. **余额领域实现**
   ```
   关键文件:
   - internal/domain/balance/aggregate.go     # Account聚合根(Charge/Deduct方法)
   - internal/domain/balance/entity.go        # Transaction交易记录
   - internal/infrastructure/persistence/postgres/balance_repo.go  # 带乐观锁实现
   ```

2. **分布式锁实现**
   ```
   关键文件:
   - internal/infrastructure/lock/redis_lock.go  # Redlock算法实现
   ```

3. **幂等性设计**
   ```
   关键文件:
   - pkg/idempotent/idempotent.go            # 幂等性中间件
   - 数据库:transactions表transaction_id唯一索引
   ```

4. **余额API实现**
   ```
   接口:
   - POST /api/v1/balance/charge              # 充值
   - GET  /api/v1/balance/transactions        # 交易记录
   - GET  /api/v1/balance                     # 查询余额
   ```

5. **扣费Worker实现**
   ```
   流程:
   1. 消费order.created事件
   2. 获取Redis分布式锁(key: account:{accountID})
   3. 幂等性检查(transaction_id)
   4. 余额扣费(乐观锁version)
   5. 释放锁
   6. 发布balance.deducted事件
   ```

#### 验收标准:
- ✅ 余额充值正确记录
- ✅ 订单创建自动触发扣费
- ✅ 余额不足时订单失败
- ✅ 并发扣费无重复(压测1000并发)
- ✅ 幂等性保证

---

### 【阶段4】账单与对账模块(2周)

**目标**:实现账单生成、对账、结算

#### 实施步骤:

1. **账单领域实现**
   ```
   关键文件:
   - internal/domain/billing/aggregate.go     # Bill聚合根(Settle方法)
   - internal/domain/billing/entity.go        # BillItem账单明细
   - internal/infrastructure/persistence/postgres/billing_repo.go
   ```

2. **账单生成定时任务**
   ```
   逻辑:
   1. 每月1日凌晨触发
   2. 查询上月所有订单
   3. 聚合账单明细
   4. 生成账单并持久化
   5. 发布bill.generated事件
   ```

3. **账单API实现**
   ```
   接口:
   - GET  /api/v1/bills                      # 账单列表
   - GET  /api/v1/bills/:id                  # 账单详情
   - POST /api/v1/bills/:id/settle           # 结算账单
   ```

4. **对账机制实现**
   ```
   对账维度:
   - 订单金额 vs 余额扣费金额
   - 账单金额 vs 订单金额
   - 套餐包抵扣记录 vs 实际使用量

   对账报告:
   - T+1日自动对账
   - 差异超过阈值触发告警
   ```

#### 验收标准:
- ✅ 每月自动生成账单
- ✅ 账单明细完整准确
- ✅ 对账无差异(误差 < 0.01元)
- ✅ 手动结算功能正常

---

### 【阶段5】套餐包模块(1周)

**目标**:实现资源包购买、抵扣、过期管理

#### 实施步骤:

1. **套餐包领域实现**
   ```
   关键文件:
   - internal/domain/package/aggregate.go     # Package聚合根(Consume方法)
   - internal/infrastructure/persistence/postgres/package_repo.go
   ```

2. **套餐包API实现**
   ```
   接口:
   - POST /api/v1/packages/purchase          # 购买套餐包
   - POST /api/v1/packages/consume           # 消费套餐包(内部接口)
   - GET  /api/v1/packages                   # 套餐包列表
   ```

3. **扣费优先级实现**
   ```
   逻辑:
   1. 查询用户可用套餐包(ACTIVE状态 + 未过期 + 有余额)
   2. 优先抵扣套餐包配额
   3. 不足部分走余额扣费
   ```

4. **过期处理定时任务**
   ```
   逻辑:
   1. 每小时扫描过期套餐包(valid_to < now)
   2. 更新状态为EXPIRED
   3. 发布package.expired事件
   ```

#### 验收标准:
- ✅ 购买套餐包成功
- ✅ 资源使用自动抵扣套餐包
- ✅ 套餐包耗尽自动切换余额扣费
- ✅ 过期套餐包无法使用

---

### 【阶段6】高可用与性能优化(2周)

**目标**:生产级部署准备

#### 实施步骤:

1. **缓存优化**
   ```
   缓存项:
   - 定价规则(TTL: 1小时)
   - 用户余额(TTL: 5分钟,写穿透)
   - 套餐包配额(TTL: 5分钟)

   一致性策略:Cache-Aside模式
   ```

2. **CQRS优化**
   ```
   读模型优化:
   - 账单列表:物化视图 + 分页索引
   - 用户消费统计:定时聚合到汇总表
   ```

3. **监控告警**
   ```
   Prometheus指标:
   - 计量采集TPS
   - 订单创建成功率
   - 扣费失败率
   - 账单生成耗时
   - Kafka消费延迟

   Grafana仪表盘:
   - 业务核心指标
   - 系统资源监控
   - 错误日志实时流
   ```

4. **压力测试**
   ```
   目标:
   - 计量采集:10,000 TPS
   - 订单创建:1,000 TPS
   - 扣费并发:1,000 TPS
   - 账单生成:1000万订单/小时
   ```

5. **容错与降级**
   ```
   策略:
   - Kafka消费失败:3次重试 + 死信队列
   - 熔断器:失败率 > 50% 触发熔断(Circuit Breaker)
   - 降级:缓存失败降级到数据库查询
   - 限流:API令牌桶限流(100 QPS/用户)
   ```

#### 验收标准:
- ✅ 通过压力测试达标
- ✅ 监控告警完善
- ✅ 模拟故障自动恢复
- ✅ 文档完善(架构文档、API文档、运维手册)

---

## 四、核心实现要点

### 4.1 计量器插件接口

```go
// internal/domain/meter/meter_plugin.go

type MeterPlugin interface {
    Name() string
    Collect(ctx context.Context, resourceID string, metadata map[string]interface{}) (*MeterRecord, error)
    Aggregate(ctx context.Context, records []*MeterRecord) (decimal.Decimal, error)
    Validate(config map[string]interface{}) error
}
```

### 4.2 订单聚合根(核心业务逻辑)

```go
// internal/domain/order/aggregate.go

type Order struct {
    ID          string
    OrderNo     string
    Status      OrderStatus
    TotalAmount decimal.Decimal
    events      []interface{}
}

func (o *Order) Complete() error {
    if o.Status != OrderStatusPending {
        return errors.New("invalid order status")
    }
    o.Status = OrderStatusCompleted
    o.AddEvent(OrderCompletedEvent{OrderID: o.ID})
    return nil
}
```

### 4.3 余额扣费(分布式锁 + 幂等性)

```go
// internal/application/command/balance/deduct.go

func (s *DeductService) Execute(ctx context.Context, cmd DeductCommand) error {
    // 1. 获取分布式锁
    lock, err := s.lockClient.Acquire(ctx, "account:"+cmd.AccountID, 10*time.Second)
    if err != nil {
        return err
    }
    defer s.lockClient.Release(ctx, lock)

    // 2. 幂等性检查
    if s.txRepo.Exists(ctx, cmd.TransactionID) {
        return nil // 已处理,直接返回
    }

    // 3. 加载聚合根
    account, err := s.accountRepo.FindByID(ctx, cmd.AccountID)
    if err != nil {
        return err
    }

    // 4. 执行领域逻辑
    if err := account.Deduct(cmd.Amount, cmd.TransactionID); err != nil {
        return err
    }

    // 5. 持久化(乐观锁)
    if err := s.accountRepo.Save(ctx, account); err != nil {
        return err
    }

    // 6. 发布领域事件
    for _, event := range account.GetEvents() {
        s.eventBus.Publish(ctx, "balance.deducted", event)
    }

    return nil
}
```

### 4.4 数据库表(关键字段)

**订单表(orders)**
```sql
CREATE TABLE orders (
    id VARCHAR(64) PRIMARY KEY,
    order_no VARCHAR(64) UNIQUE NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    order_type VARCHAR(20) NOT NULL,
    payment_mode VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    total_amount DECIMAL(20, 4) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_status (status)
);
```

**账户表(accounts)**
```sql
CREATE TABLE accounts (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    balance DECIMAL(20, 4) NOT NULL DEFAULT 0,
    frozen_balance DECIMAL(20, 4) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    version BIGINT NOT NULL DEFAULT 0,  -- 乐观锁
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE INDEX uk_tenant_user (tenant_id, user_id)
);
```

**交易记录表(transactions)**
```sql
CREATE TABLE transactions (
    id VARCHAR(64) PRIMARY KEY,
    transaction_id VARCHAR(64) UNIQUE NOT NULL,  -- 幂等性key
    account_id VARCHAR(64) NOT NULL,
    type VARCHAR(20) NOT NULL,
    amount DECIMAL(20, 4) NOT NULL,
    balance_before DECIMAL(20, 4) NOT NULL,
    balance_after DECIMAL(20, 4) NOT NULL,
    order_id VARCHAR(64),
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_account_id (account_id),
    INDEX idx_transaction_id (transaction_id)
);
```

---

## 五、架构原则体现

### 5.1 SOLID原则
- **单一职责(SRP)**:每个领域模型只负责自己的业务逻辑
- **开闭原则(OCP)**:通过插件接口扩展计量器和定价器,无需修改核心代码
- **里氏替换(LSP)**:所有插件实现可相互替换
- **接口隔离(ISP)**:仓储接口按聚合根隔离,避免大而全
- **依赖倒置(DIP)**:领域层不依赖基础设施层,通过接口解耦

### 5.2 DRY原则
- 金额计算统一使用`pkg/money/decimal.go`
- 错误处理统一封装在`pkg/errors`
- 消息队列、缓存、锁统一接口

### 5.3 KISS原则
- 领域模型简洁清晰
- 避免过度设计

### 5.4 YAGNI原则
- 只实现当前需要的功能
- 预留扩展点但不提前实现未使用的插件

---

## 六、风险与应对

| 风险 | 应对措施 |
|------|---------|
| **分布式事务失败** | Saga补偿机制 + 人工复核工具 |
| **计量数据丢失** | Kafka持久化 + 消费offset管理 + 重试机制 |
| **高并发扣费冲突** | 分布式锁 + 乐观锁双重保证 |
| **金额精度误差** | 使用decimal.Decimal + 单元测试全覆盖 |
| **性能瓶颈** | 压力测试 + 缓存优化 + 数据库索引优化 |

---

## 七、交付物清单

- [ ] 完整代码仓库(符合DDD分层架构)
- [ ] 数据库迁移脚本(migrations/)
- [ ] API文档(Swagger/OpenAPI)
- [ ] 架构设计文档(docs/architecture/)
- [ ] 部署文档(Dockerfile + K8s YAML)
- [ ] 监控配置(Prometheus + Grafana)
- [ ] 压力测试报告
- [ ] 单元测试覆盖率报告(目标 > 80%)

---

## 八、后续扩展点

1. **多币种支持**:扩展Currency值对象
2. **发票管理**:新增Invoice领域
3. **优惠券系统**:新增Coupon领域
4. **分账结算**:支持多商户分账
5. **国际化支持**:多语言账单
6. **机器学习预测**:用量预测与成本优化建议

---

**实施原则**:分阶段迭代,每阶段独立可验证,持续集成持续交付(CI/CD),代码质量优先。
