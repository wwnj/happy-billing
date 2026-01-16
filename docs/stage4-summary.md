# Stage 4: 账单与对账模块实施总结（核心部分完成）

## 📋 实施概述

**目标**: 实现账单生成、对账、结算功能

**实施进度**: ✅ 核心领域模型和基础设施完成，应用服务部分完成

**完成度**: 约70%（核心功能已完成，需补充应用服务和API）

---

## ✅ 已完成的工作

### 1. 账单领域模型实现

#### 值对象（Value Objects）
- **BillStatus**: 账单状态
  - `PENDING` - 待结算
  - `SETTLED` - 已结算
  - `CANCELLED` - 已取消
  - `OVERDUE` - 已逾期

- **BillCycle**: 账单周期
  - `MONTHLY` - 月度账单
  - `QUARTERLY` - 季度账单
  - `YEARLY` - 年度账单
  - `CUSTOM` - 自定义周期

- **BillItemType**: 账单明细类型
  - `ORDER` - 订单费用
  - `CHARGE` - 充值
  - `REFUND` - 退款
  - `ADJUST` - 调整
  - `DISCOUNT` - 折扣

#### 聚合根（Aggregate Root）
**Bill** - 账单聚合根
```go
type Bill struct {
    ID                string        // 唯一ID
    BillNo            string        // 账单号
    TenantID          string        // 租户ID
    UserID            string        // 用户ID
    Cycle             BillCycle     // 账单周期
    Status            BillStatus    // 账单状态
    PeriodStart       time.Time     // 账期开始时间
    PeriodEnd         time.Time     // 账期结束时间
    TotalAmount       money.Decimal // 总金额
    DiscountAmount    money.Decimal // 折扣金额
    TaxAmount         money.Decimal // 税额
    ActualAmount      money.Decimal // 实际应付金额
    PaidAmount        money.Decimal // 已支付金额
    OutstandingAmount money.Decimal // 未付金额
    Currency          string        // 货币类型
    Items             []*BillItem   // 账单明细
    DueDate           *time.Time    // 到期日期
    SettledAt         *time.Time    // 结算时间
    events            []interface{} // 领域事件
}
```

**核心方法:**
- `AddItem()` - 添加账单明细
- `RecalculateAmount()` - 重新计算账单金额
- `Settle()` - 结算账单
- `Cancel()` - 取消账单
- `MarkOverdue()` - 标记为逾期
- `SetDueDate()` - 设置到期日期
- `IsOverdue()` - 是否逾期

#### 实体（Entity）
**BillItem** - 账单明细实体
```go
type BillItem struct {
    ID          string        // 唯一ID
    BillID      string        // 账单ID
    Type        BillItemType  // 明细类型
    OrderID     *string       // 关联订单ID
    Description string        // 描述
    Amount      money.Decimal // 金额
    Quantity    money.Decimal // 数量
    UnitPrice   money.Decimal // 单价
    Discount    money.Decimal // 折扣金额
    TaxAmount   money.Decimal // 税额
    TotalAmount money.Decimal // 总金额（含税）
}
```

**核心方法:**
- `SetQuantityAndPrice()` - 设置数量和单价
- `SetDiscount()` - 设置折扣
- `SetTax()` - 设置税额

#### 领域事件
实现了4个核心领域事件：
1. `BillGeneratedEvent` - 账单生成
2. `BillSettledEvent` - 账单结算
3. `BillCancelledEvent` - 账单取消
4. `BillOverdueEvent` - 账单逾期

###  2. 基础设施层实现

#### PostgreSQL仓储实现

**BillRepository** - 账单仓储
- `Save()` - 保存新账单
- `Update()` - 更新账单
- `FindByID()` - 根据ID查询
- `FindByBillNo()` - 根据账单号查询
- `ListByUser()` - 查询用户账单列表（支持状态、时间范围过滤）
- `ListByTenant()` - 查询租户账单列表
- `ListOverdue()` - 查询逾期账单
- `SumAmountByUser()` - 统计用户账单总额
- `SumAmountByTenant()` - 统计租户账单总额

**BillItemRepository** - 账单明细仓储
- `Save()` - 保存账单明细
- `BatchSave()` - 批量保存（支持100条/批）
- `FindByID()` - 根据ID查询
- `ListByBill()` - 查询账单的明细列表
- `ListByOrder()` - 查询订单关联的账单明细
- `SumAmountByBill()` - 统计账单明细总额

**数据库模型转换**:
- DO（Data Object）↔ 领域对象的双向转换
- 支持JSONB元数据字段
- 使用sql.NullTime处理可空时间字段

### 3. 应用层查询服务实现

**BillQuery** - 账单查询服务
- `GetBill()` - 查询账单基本信息
- `GetBillDetail()` - 查询账单详情（包含明细）
- `ListBills()` - 查询账单列表（分页、过滤）
- `ListOverdueBills()` - 查询逾期账单
- `SumBillAmount()` - 统计账单总额

**DTO设计:**
- `BillDTO` - 账单视图
- `BillItemDTO` - 账单明细视图
- `BillDetailDTO` - 账单详情视图
- `ListBillsResult` - 分页查询结果

---

## 🎯 核心架构设计

### 账单生成流程
```
1. 定时任务触发（每月1日凌晨）
   ↓
2. 查询上月所有订单
   ↓
3. 按用户聚合订单明细
   ↓
4. 创建账单聚合根
   ↓
5. 添加账单明细
   ↓
6. 计算总金额、折扣、税额
   ↓
7. 持久化账单和明细
   ↓
8. 发布bill.generated事件
```

### 账单结算流程
```
1. 接收结算请求
   ↓
2. 查询账单（验证状态）
   ↓
3. 执行领域逻辑（Settle方法）
   ↓
4. 更新账单状态
   ↓
5. 持久化
   ↓
6. 发布bill.settled事件
```

### 逾期检测流程
```
1. 定时任务扫描（每天凌晨）
   ↓
2. 查询待结算账单（due_date < now）
   ↓
3. 批量标记为逾期
   ↓
4. 发布bill.overdue事件
   ↓
5. 触发告警通知
```

---

## 📊 数据库设计（已实现）

### bills表（账单表）
```sql
CREATE TABLE bills (
    id VARCHAR(64) PRIMARY KEY,
    bill_no VARCHAR(64) UNIQUE NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    cycle VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    total_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    discount_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    actual_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    paid_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    outstanding_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    due_date TIMESTAMP,
    settled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_status (status),
    INDEX idx_period (period_start, period_end),
    INDEX idx_due_date (due_date)
);
```

### bill_items表（账单明细表）
```sql
CREATE TABLE bill_items (
    id VARCHAR(64) PRIMARY KEY,
    bill_id VARCHAR(64) NOT NULL,
    type VARCHAR(20) NOT NULL,
    order_id VARCHAR(64),
    description TEXT,
    amount DECIMAL(20, 4) NOT NULL,
    quantity DECIMAL(20, 4) NOT NULL DEFAULT 0,
    unit_price DECIMAL(20, 4) NOT NULL DEFAULT 0,
    discount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    total_amount DECIMAL(20, 4) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_bill_id (bill_id),
    INDEX idx_order_id (order_id)
);
```

---

## 📦 已实现文件清单

```
internal/domain/billing/
├── value_object.go       # 值对象（BillStatus/BillCycle/BillItemType）
├── aggregate.go          # 账单聚合根 + 账单明细实体
├── events.go             # 4个领域事件
└── repository.go         # 仓储接口

internal/infrastructure/persistence/postgres/
├── model/
│   └── billing.go        # DO模型 + 转换器
├── bill_repository.go    # 账单仓储实现
└── bill_item_repository.go  # 账单明细仓储实现

internal/application/query/billing/
└── bill_query.go         # 账单查询服务
```

---

## 🚧 待完成工作

### 1. 应用层命令服务（高优先级）
- [ ] `GenerateBillService` - 账单生成服务
  - 聚合订单数据
  - 创建账单和明细
  - 发布领域事件

- [ ] `SettleBillService` - 账单结算服务
  - 验证支付信息
  - 执行结算逻辑
  - 更新账单状态

- [ ] `CancelBillService` - 账单取消服务

### 2. HTTP API接口（高优先级）
```
GET    /api/v1/bills                          # 账单列表
GET    /api/v1/bills/:id                      # 账单详情
POST   /api/v1/bills/:id/settle               # 结算账单
POST   /api/v1/bills/:id/cancel               # 取消账单
GET    /api/v1/bills/overdue                  # 逾期账单列表
GET    /api/v1/bills/statistics               # 账单统计
```

### 3. 定时任务（高优先级）
- [ ] 账单生成定时任务（每月1日凌晨）
- [ ] 逾期检测定时任务（每天凌晨）

### 4. 对账机制（中优先级）
- [ ] 订单金额 vs 账单明细金额对账
- [ ] 账单金额 vs 余额扣费金额对账
- [ ] 对账差异报告生成
- [ ] 差异告警机制

### 5. 数据库迁移脚本（高优先级）
- [ ] `002_create_billing_tables.up.sql`
- [ ] `002_create_billing_tables.down.sql`

### 6. 单元测试
- [ ] 领域逻辑测试
- [ ] 仓储集成测试
- [ ] 应用服务测试

---

## 🎓 架构亮点

### 领域驱动设计
- 账单聚合根管理一致性边界
- 账单明细作为实体从属于聚合根
- 完整的领域事件支持事件溯源

### 业务规则封装
- 账单金额自动计算（总额、折扣、税额）
- 状态机管理账单生命周期
- 逾期检测自动化

### 查询优化
- 支持多维度过滤（状态、时间范围、用户）
- 分页查询性能优化
- 聚合查询使用数据库原生SUM

---

## ✅ 验收标准（部分完成）

| 标准 | 状态 | 说明 |
|------|------|------|
| 账单领域模型 | ✅ | 完整实现聚合根、实体、值对象 |
| 账单仓储 | ✅ | PostgreSQL实现完成 |
| 账单查询服务 | ✅ | CQRS查询端完成 |
| 账单生成服务 | ⏳ | 待实现 |
| 账单结算服务 | ⏳ | 待实现 |
| HTTP API | ⏳ | 待实现 |
| 定时任务 | ⏳ | 待实现 |
| 对账机制 | ⏳ | 待实现 |
| 数据库迁移 | ⏳ | 待实现 |

---

## 🚀 下一步建议

### Option 1: 完成Stage 4剩余部分
1. 实现账单生成和结算服务
2. 实现HTTP API接口
3. 创建数据库迁移脚本
4. 实现定时任务

### Option 2: 进入Stage 5（套餐包模块）
在当前进度基础上继续实施套餐包功能

### Option 3: 完善测试和文档
- 编写单元测试和集成测试
- 完善API文档
- 编写运维手册

---

## 📚 核心代码示例

### 账单生成示例
```go
// 创建账单
bill, err := billing.NewBill(
    uuid.New().String(),
    "BILL_202501",
    "tenant_001",
    "user_123",
    billing.BillCycleMonthly,
    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
    time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
    "CNY",
)

// 添加账单明细
for _, order := range orders {
    item, _ := billing.NewBillItem(
        uuid.New().String(),
        bill.ID,
        billing.BillItemTypeOrder,
        &order.ID,
        order.Description,
        order.Amount,
        nil,
    )
    bill.AddItem(item)
}

// 保存账单
billRepo.Save(ctx, bill)
billItemRepo.BatchSave(ctx, bill.Items)
```

### 账单结算示例
```go
// 查询账单
bill, _ := billRepo.FindByID(ctx, billID)

// 执行结算
err := bill.Settle(paidAmount)

// 更新账单
billRepo.Update(ctx, bill)

// 发布事件
for _, event := range bill.GetEvents() {
    eventBus.Publish(ctx, event)
}
```

---

**实施完成时间**: 2025-01-16
**实施进度**: Stage 4 核心部分完成（约70%）
**代码质量**: ✅ 编译通过、架构清晰、注释完整
