# Happy Billing 代码规范

**版本**: v1.0  
**更新日期**: 2026-01-17  
**适用范围**: Happy Billing 订单账单系统所有开发工作

---

## 目录

1. [Go 代码规范](#go-代码规范)
2. [项目结构规范](#项目结构规范)
3. [命名规范](#命名规范)
4. [注释规范](#注释规范)
5. [错误处理规范](#错误处理规范)
6. [数据库规范](#数据库规范)
7. [API 设计规范](#api-设计规范)
8. [Git 提交规范](#git-提交规范)
9. [测试规范](#测试规范)

---

## Go 代码规范

### 基础规范

遵循 [Effective Go](https://go.dev/doc/effective_go) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)。

### 代码格式化

```bash
# 使用 gofmt 格式化代码
gofmt -w .

# 使用 goimports 优化 import
goimports -w .

# 使用 golangci-lint 检查代码质量
golangci-lint run
```

**强制要求**:
- 所有代码提交前必须运行 `gofmt`
- 使用 `goimports` 自动管理导入
- 通过 `golangci-lint` 检查

### 包导入顺序

```go
import (
    // 1. 标准库
    "context"
    "fmt"
    "time"
    
    // 2. 第三方库
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    // 3. 项目内部包
    "happy-billing/internal/models"
    "happy-billing/internal/service"
    "happy-billing/pkg/logger"
)
```

### 变量声明

```go
// ✅ 推荐：使用简短声明
func processOrder(orderID string) error {
    order, err := orderService.GetOrder(orderID)
    if err != nil {
        return err
    }
    // ...
}

// ❌ 避免：不必要的变量声明
func processOrder(orderID string) error {
    var order *Order
    var err error
    order, err = orderService.GetOrder(orderID)
    // ...
}

// ✅ 推荐：零值初始化
var (
    count int       // 0
    name  string    // ""
    items []Item    // nil
)

// ✅ 推荐：分组声明常量
const (
    StatusPending  = "PENDING"
    StatusPaid     = "PAID"
    StatusCanceled = "CANCELED"
)
```

### 错误处理

```go
// ✅ 推荐：立即返回错误
func CreateOrder(req *OrderRequest) (*Order, error) {
    if err := validateOrderRequest(req); err != nil {
        return nil, fmt.Errorf("验证订单失败: %w", err)
    }
    
    order, err := orderRepo.Create(req)
    if err != nil {
        return nil, fmt.Errorf("创建订单失败: %w", err)
    }
    
    return order, nil
}

// ❌ 避免：忽略错误
func CreateOrder(req *OrderRequest) *Order {
    order, _ := orderRepo.Create(req)  // 错误被忽略
    return order
}
```

### 上下文传递

```go
// ✅ 推荐：所有需要超时控制的函数都应接收 context
func (s *OrderService) CreateOrder(ctx context.Context, req *OrderRequest) (*Order, error) {
    // 设置超时
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // 使用 context 进行数据库操作
    order, err := s.repo.CreateWithContext(ctx, req)
    if err != nil {
        return nil, err
    }
    
    return order, nil
}
```

### 接口设计

```go
// ✅ 推荐：接口应该小而精
type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, orderID string) (*Order, error)
    Update(ctx context.Context, order *Order) error
    Delete(ctx context.Context, orderID string) error
}

// ❌ 避免：臃肿的接口
type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, orderID string) (*Order, error)
    Update(ctx context.Context, order *Order) error
    Delete(ctx context.Context, orderID string) error
    List(ctx context.Context, filter *Filter) ([]*Order, error)
    Count(ctx context.Context, filter *Filter) (int64, error)
    BatchCreate(ctx context.Context, orders []*Order) error
    BatchUpdate(ctx context.Context, orders []*Order) error
    // ... 更多方法
}
```

### 结构体设计

```go
// ✅ 推荐：字段按重要性和类型分组，添加 JSON tag
type Order struct {
    // 主键和业务ID
    ID      int64  `json:"id" gorm:"primaryKey"`
    OrderID string `json:"order_id" gorm:"uniqueIndex;type:varchar(64)"`
    
    // 关联ID
    TenantID  string `json:"tenant_id" gorm:"index;type:varchar(64)"`
    ProjectID string `json:"project_id" gorm:"index;type:varchar(64)"`
    UserID    string `json:"user_id" gorm:"index;type:varchar(64)"`
    
    // 订单信息
    OrderType      string          `json:"order_type" gorm:"type:varchar(32)"`
    SPUCode        string          `json:"spu_code" gorm:"type:varchar(64)"`
    SKUCode        string          `json:"sku_code" gorm:"type:varchar(64)"`
    PayableAmount  decimal.Decimal `json:"payable_amount" gorm:"type:decimal(18,4)"`
    Status         string          `json:"status" gorm:"type:varchar(32);index"`
    
    // 时间戳
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
```

### 并发控制

```go
// ✅ 推荐：使用 sync.WaitGroup 等待协程完成
func ProcessOrdersBatch(orderIDs []string) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(orderIDs))
    
    for _, orderID := range orderIDs {
        wg.Add(1)
        go func(id string) {
            defer wg.Done()
            if err := processOrder(id); err != nil {
                errChan <- err
            }
        }(orderID)
    }
    
    wg.Wait()
    close(errChan)
    
    // 收集错误
    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("批量处理失败: %v", errors)
    }
    
    return nil
}

// ✅ 推荐：使用互斥锁保护共享资源
type Cache struct {
    mu    sync.RWMutex
    data  map[string]interface{}
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.data[key]
    return val, ok
}

func (c *Cache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}
```

---

## 项目结构规范

### 标准项目结构

```
happy-billing/
├── cmd/                          # 主程序入口
│   ├── api/                      # API 服务入口
│   │   └── main.go
│   ├── worker/                   # 后台任务入口
│   │   └── main.go
│   └── migrate/                  # 数据库迁移工具
│       └── main.go
├── internal/                     # 私有应用代码
│   ├── api/                      # API 处理层
│   │   ├── handler/              # HTTP 处理器
│   │   │   ├── order_handler.go
│   │   │   ├── bill_handler.go
│   │   │   └── payment_handler.go
│   │   ├── middleware/           # 中间件
│   │   │   ├── auth.go
│   │   │   ├── logger.go
│   │   │   └── ratelimit.go
│   │   └── router/               # 路由配置
│   │       └── router.go
│   ├── service/                  # 业务逻辑层
│   │   ├── order_service.go
│   │   ├── bill_service.go
│   │   ├── payment_service.go
│   │   └── pricing_service.go
│   ├── repository/               # 数据访问层
│   │   ├── order_repo.go
│   │   ├── bill_repo.go
│   │   └── account_repo.go
│   ├── models/                   # 数据模型
│   │   ├── order.go
│   │   ├── bill.go
│   │   └── account.go
│   ├── domain/                   # 领域模型
│   │   ├── order/
│   │   ├── billing/
│   │   └── payment/
│   └── worker/                   # 后台任务
│       ├── bill_generator.go
│       └── settlement_worker.go
├── pkg/                          # 公共库代码
│   ├── logger/                   # 日志库
│   ├── database/                 # 数据库连接
│   ├── redis/                    # Redis 客户端
│   ├── kafka/                    # Kafka 客户端
│   ├── errors/                   # 错误处理
│   └── utils/                    # 工具函数
├── config/                       # 配置文件
│   ├── config.yaml
│   ├── config.dev.yaml
│   └── config.prod.yaml
├── migrations/                   # 数据库迁移文件
│   ├── 001_create_tenants.sql
│   ├── 002_create_orders.sql
│   └── 003_create_bills.sql
├── scripts/                      # 脚本文件
│   ├── build.sh
│   └── deploy.sh
├── docs/                         # 文档
│   ├── design/                   # 设计文档
│   └── development/              # 开发文档
├── test/                         # 测试文件
│   ├── integration/              # 集成测试
│   └── e2e/                      # 端到端测试
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 分层架构

```
┌─────────────────────────────────────┐
│         API Handler Layer           │  HTTP 请求处理
│  (internal/api/handler)             │
└─────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────┐
│        Service Layer                │  业务逻辑
│  (internal/service)                 │
└─────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────┐
│       Repository Layer              │  数据访问
│  (internal/repository)              │
└─────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────┐
│         Data Layer                  │  数据存储
│  (MySQL/ClickHouse/Redis)           │
└─────────────────────────────────────┘
```

**职责说明**:

| 层级 | 职责 | 禁止事项 |
|------|------|----------|
| **Handler** | 接收请求、参数验证、响应返回 | 不包含业务逻辑、不直接访问数据库 |
| **Service** | 业务逻辑、事务控制、领域模型操作 | 不处理 HTTP 细节、不直接操作数据库连接 |
| **Repository** | 数据访问、SQL 执行、缓存操作 | 不包含业务逻辑 |

### 文件命名规范

```
# ✅ 推荐：使用小写+下划线
order_service.go
bill_handler.go
account_repository.go

# ❌ 避免：使用驼峰
OrderService.go
BillHandler.go
```

---

## 命名规范

### 包命名

```go
// ✅ 推荐：简短、小写、单数
package order
package billing
package payment

// ❌ 避免：复数、下划线、驼峰
package orders
package order_service
package orderService
```

### 变量命名

```go
// ✅ 推荐：驼峰命名
var orderID string
var totalAmount decimal.Decimal
var createdAt time.Time

// ✅ 推荐：常量使用大写+下划线
const (
    MaxRetryCount    = 3
    DefaultPageSize  = 20
    TimeoutSeconds   = 30
)

// ✅ 推荐：枚举使用前缀
const (
    OrderStatusPending   = "PENDING"
    OrderStatusPaid      = "PAID"
    OrderStatusCanceled  = "CANCELED"
    
    BillTypeSubscription = "SUBSCRIPTION"
    BillTypeHourly       = "HOURLY"
)
```

### 函数命名

```go
// ✅ 推荐：动词开头,清晰表达意图
func CreateOrder(req *OrderRequest) (*Order, error)
func GetOrderByID(orderID string) (*Order, error)
func UpdateOrderStatus(orderID string, status string) error
func DeleteOrder(orderID string) error
func ValidateOrderRequest(req *OrderRequest) error
func CalculateTotalAmount(items []OrderItem) decimal.Decimal

// ✅ 推荐：布尔函数使用 Is/Has/Can 前缀
func IsValidTenant(tenantID string) bool
func HasSufficientBalance(accountID string, amount decimal.Decimal) bool
func CanCreateOrder(tenantID string) bool
```

### 接口命名

```go
// ✅ 推荐：以 -er 结尾
type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, orderID string) (*Order, error)
}

type PriceCalculator interface {
    Calculate(ctx context.Context, req *PriceRequest) (*Price, error)
}

type BillGenerator interface {
    Generate(ctx context.Context, orderID string) (*Bill, error)
}
```

---

## 注释规范

### 包注释

```go
// Package order 提供订单管理的核心业务逻辑。
//
// 该包实现了订单的创建、查询、更新、删除等操作，
// 支持包年包月和按量计费两种订单类型。
//
// 基本用法:
//
//     service := order.NewService(repo, priceService)
//     order, err := service.CreateOrder(ctx, req)
//     if err != nil {
//         // 处理错误
//     }
package order
```

### 函数注释

```go
// CreateOrder 创建新订单。
//
// 该方法会执行以下步骤:
//   1. 验证订单请求参数
//   2. 查询商品定价
//   3. 计算订单金额
//   4. 创建订单记录
//   5. 发送订单创建事件
//
// 参数:
//   ctx: 上下文,用于超时控制和取消操作
//   req: 订单创建请求,包含商品、数量、租户等信息
//
// 返回值:
//   *Order: 创建成功的订单对象
//   error: 创建失败时的错误信息
//
// 示例:
//
//     req := &OrderRequest{
//         TenantID: "T20240117001",
//         SKUCode:  "SKU_GPU_A100_40GB",
//         Quantity: 2,
//     }
//     order, err := service.CreateOrder(ctx, req)
//     if err != nil {
//         return fmt.Errorf("创建订单失败: %w", err)
//     }
func (s *OrderService) CreateOrder(ctx context.Context, req *OrderRequest) (*Order, error) {
    // 实现...
}
```

### 结构体注释

```go
// Order 表示订单实体。
//
// 订单记录了用户的购买行为,包含商品信息、金额、状态等。
// 订单状态流转: PENDING -> PAID -> RUNNING -> COMPLETED
type Order struct {
    // ID 是订单的物理主键(数据库自增)
    ID int64 `json:"id" gorm:"primaryKey"`
    
    // OrderID 是订单的业务主键,格式为 ORD+YYYYMMDD+序号
    // 例如: ORD202401170000001
    OrderID string `json:"order_id" gorm:"uniqueIndex;type:varchar(64)"`
    
    // TenantID 是租户ID,用于多租户隔离
    TenantID string `json:"tenant_id" gorm:"index;type:varchar(64)"`
    
    // OrderType 订单类型: SUBSCRIPTION(包年包月) 或 PAY_AS_YOU_GO(按量计费)
    OrderType string `json:"order_type" gorm:"type:varchar(32)"`
    
    // PayableAmount 应付金额(CNY)
    PayableAmount decimal.Decimal `json:"payable_amount" gorm:"type:decimal(18,4)"`
    
    // Status 订单状态: PENDING/PAID/RUNNING/COMPLETED/CANCELED
    Status string `json:"status" gorm:"type:varchar(32);index"`
}
```

### TODO 注释

```go
// TODO(zhangsan): 添加重试机制处理网络超时
// TODO(lisi): 优化 SQL 查询性能,考虑添加索引
// FIXME(wangwu): 修复并发场景下的余额扣减问题
// HACK(zhaoliu): 临时方案,等待上游服务修复后移除
```

---

## 错误处理规范

### 错误定义

```go
// ✅ 推荐:使用 errors.New 或 fmt.Errorf 创建错误
var (
    ErrOrderNotFound     = errors.New("订单不存在")
    ErrInsufficientBalance = errors.New("余额不足")
    ErrInvalidOrderStatus  = errors.New("订单状态无效")
)

// ✅ 推荐:自定义错误类型
type BusinessError struct {
    Code    string
    Message string
    Cause   error
}

func (e *BusinessError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewBusinessError(code, message string) *BusinessError {
    return &BusinessError{
        Code:    code,
        Message: message,
    }
}
```

### 错误包装

```go
// ✅ 推荐:使用 %w 包装错误,保留错误链
func (s *OrderService) CreateOrder(ctx context.Context, req *OrderRequest) (*Order, error) {
    if err := s.validateRequest(req); err != nil {
        return nil, fmt.Errorf("验证订单请求失败: %w", err)
    }
    
    order, err := s.repo.Create(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("创建订单失败: %w", err)
    }
    
    return order, nil
}

// ✅ 推荐:使用 errors.Is 和 errors.As 判断错误
func HandleError(err error) {
    if errors.Is(err, ErrOrderNotFound) {
        // 处理订单不存在错误
    }
    
    var bizErr *BusinessError
    if errors.As(err, &bizErr) {
        // 处理业务错误
        log.Error("业务错误", "code", bizErr.Code, "message", bizErr.Message)
    }
}
```

### HTTP 错误响应

```go
// 定义标准错误响应
type ErrorResponse struct {
    Code    string `json:"code"`              // 错误码
    Message string `json:"message"`           // 错误信息
    Details string `json:"details,omitempty"` // 详细信息(可选)
}

// 错误码定义
const (
    ErrCodeInvalidParam    = "INVALID_PARAM"
    ErrCodeUnauthorized    = "UNAUTHORIZED"
    ErrCodeForbidden       = "FORBIDDEN"
    ErrCodeNotFound        = "NOT_FOUND"
    ErrCodeInternalError   = "INTERNAL_ERROR"
    ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// 统一错误处理
func HandleHTTPError(c *gin.Context, err error) {
    var statusCode int
    var errResp ErrorResponse
    
    switch {
    case errors.Is(err, ErrOrderNotFound):
        statusCode = http.StatusNotFound
        errResp = ErrorResponse{
            Code:    ErrCodeNotFound,
            Message: "订单不存在",
        }
    case errors.Is(err, ErrInsufficientBalance):
        statusCode = http.StatusBadRequest
        errResp = ErrorResponse{
            Code:    ErrCodeInvalidParam,
            Message: "余额不足",
        }
    default:
        statusCode = http.StatusInternalServerError
        errResp = ErrorResponse{
            Code:    ErrCodeInternalError,
            Message: "服务器内部错误",
        }
        // 记录详细错误日志
        log.Error("未处理的错误", "error", err)
    }
    
    c.JSON(statusCode, errResp)
}
```

---

## 数据库规范

### SQL 命名规范

```sql
-- ✅ 推荐:表名使用小写+下划线,复数形式
CREATE TABLE orders (...)
CREATE TABLE order_items (...)
CREATE TABLE account_transactions (...)

-- ✅ 推荐:字段名使用小写+下划线
CREATE TABLE orders (
    id              BIGINT,
    order_id        VARCHAR(64),
    tenant_id       VARCHAR(64),
    created_at      TIMESTAMP
);

-- ✅ 推荐:索引命名: idx_表名_字段名
CREATE INDEX idx_orders_tenant_id ON orders(tenant_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE UNIQUE INDEX uk_orders_order_id ON orders(order_id);

-- ✅ 推荐:外键命名: fk_表名_引用表名
ALTER TABLE orders ADD CONSTRAINT fk_orders_tenants 
    FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id);
```

### GORM 使用规范

```go
// ✅ 推荐:使用 GORM tag 定义表结构
type Order struct {
    ID        int64  `gorm:"primaryKey;autoIncrement"`
    OrderID   string `gorm:"uniqueIndex;type:varchar(64);not null"`
    TenantID  string `gorm:"index;type:varchar(64);not null"`
    Status    string `gorm:"type:varchar(32);index;not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// ✅ 推荐:使用事务
func (r *OrderRepository) CreateOrder(ctx context.Context, order *Order, bill *Bill) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 创建订单
        if err := tx.Create(order).Error; err != nil {
            return fmt.Errorf("创建订单失败: %w", err)
        }
        
        // 创建账单
        if err := tx.Create(bill).Error; err != nil {
            return fmt.Errorf("创建账单失败: %w", err)
        }
        
        return nil
    })
}

// ✅ 推荐:使用预加载避免 N+1 查询
func (r *OrderRepository) GetOrdersWithItems(ctx context.Context, tenantID string) ([]*Order, error) {
    var orders []*Order
    err := r.db.WithContext(ctx).
        Preload("Items").
        Where("tenant_id = ?", tenantID).
        Find(&orders).Error
    
    return orders, err
}

// ✅ 推荐:使用原生 SQL 优化复杂查询
func (r *BillRepository) GetMonthlySummary(ctx context.Context, tenantID string, month time.Time) (*Summary, error) {
    var summary Summary
    err := r.db.WithContext(ctx).Raw(`
        SELECT 
            COUNT(*) as bill_count,
            SUM(payable_amount) as total_amount
        FROM bills
        WHERE tenant_id = ?
          AND created_at >= ?
          AND created_at < ?
    `, tenantID, month, month.AddDate(0, 1, 0)).Scan(&summary).Error
    
    return &summary, err
}
```

### 数据库迁移规范

```sql
-- migrations/001_create_tenants.sql
-- 描述: 创建租户表
-- 作者: zhangsan
-- 日期: 2026-01-17

-- UP
CREATE TABLE tenants (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id       VARCHAR(64) UNIQUE NOT NULL COMMENT '租户业务ID',
    tenant_code     VARCHAR(64) UNIQUE NOT NULL COMMENT '租户编码',
    name            VARCHAR(255) NOT NULL COMMENT '租户名称',
    tenant_type     VARCHAR(32) NOT NULL COMMENT '租户类型: ENTERPRISE/INDIVIDUAL',
    status          TINYINT DEFAULT 1 COMMENT '状态: 1-启用 0-禁用',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_type (tenant_type),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户表';

-- DOWN
DROP TABLE IF EXISTS tenants;
```

---

## API 设计规范

### RESTful API 规范

```
# ✅ 推荐:使用标准 HTTP 方法
GET     /api/v1/orders              # 获取订单列表
GET     /api/v1/orders/:id          # 获取订单详情
POST    /api/v1/orders              # 创建订单
PUT     /api/v1/orders/:id          # 更新订单
PATCH   /api/v1/orders/:id/status   # 部分更新(更新状态)
DELETE  /api/v1/orders/:id          # 删除订单

# ✅ 推荐:使用资源嵌套表达关系
GET     /api/v1/tenants/:tenant_id/orders              # 获取租户的订单
GET     /api/v1/orders/:order_id/bills                 # 获取订单的账单
GET     /api/v1/accounts/:account_id/transactions      # 获取账户流水

# ✅ 推荐:使用查询参数进行过滤、分页、排序
GET /api/v1/orders?tenant_id=T001&status=PAID&page=1&page_size=20&sort=created_at:desc
```

### 请求响应格式

```go
// 统一请求格式
type OrderCreateRequest struct {
    TenantID  string          `json:"tenant_id" binding:"required"`
    ProjectID string          `json:"project_id" binding:"required"`
    SKUCode   string          `json:"sku_code" binding:"required"`
    Quantity  int             `json:"quantity" binding:"required,gt=0"`
    OrderType string          `json:"order_type" binding:"required,oneof=SUBSCRIPTION PAY_AS_YOU_GO"`
}

// 统一响应格式
type Response struct {
    Code    int         `json:"code"`              // 业务状态码
    Message string      `json:"message"`           // 响应信息
    Data    interface{} `json:"data,omitempty"`    // 响应数据
}

// 分页响应
type PageResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Meta    PageMeta    `json:"meta"`
}

type PageMeta struct {
    Page      int   `json:"page"`         // 当前页码
    PageSize  int   `json:"page_size"`    // 每页数量
    Total     int64 `json:"total"`        // 总记录数
    TotalPage int   `json:"total_page"`   // 总页数
}

// 使用示例
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    var req OrderCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, Response{
            Code:    400,
            Message: "参数验证失败",
        })
        return
    }
    
    order, err := h.service.CreateOrder(c.Request.Context(), &req)
    if err != nil {
        HandleHTTPError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, Response{
        Code:    200,
        Message: "订单创建成功",
        Data:    order,
    })
}
```

### API 版本控制

```go
// ✅ 推荐:URL 路径版本控制
r := gin.Default()

v1 := r.Group("/api/v1")
{
    v1.POST("/orders", orderHandler.CreateOrder)
    v1.GET("/orders/:id", orderHandler.GetOrder)
}

v2 := r.Group("/api/v2")
{
    v2.POST("/orders", orderHandlerV2.CreateOrder)
    v2.GET("/orders/:id", orderHandlerV2.GetOrder)
}
```

---

## Git 提交规范

### Commit Message 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型**:
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式调整(不影响功能)
- `refactor`: 代码重构
- `perf`: 性能优化
- `test`: 测试代码
- `chore`: 构建工具或辅助工具变动
- `revert`: 回滚代码

**Scope 范围**:
- `order`: 订单模块
- `bill`: 账单模块
- `payment`: 支付模块
- `pricing`: 定价模块
- `account`: 账户模块

**示例**:

```
feat(order): 添加包年包月订单创建功能

- 新增 CreateSubscriptionOrder 方法
- 支持 SKU 选择和数量配置
- 自动生成账单并关联订单

Closes #123
```

```
fix(payment): 修复余额扣减并发问题

使用乐观锁(version 字段)解决并发场景下的余额扣减问题

Fixes #456
```

```
docs(readme): 更新项目文档结构说明

添加项目结构图和开发规范链接
```

### 分支管理

```
main                  # 主分支,生产环境代码
  ├─ develop          # 开发分支
  │   ├─ feature/order-subscription    # 功能分支
  │   ├─ feature/bill-generator        # 功能分支
  │   └─ bugfix/payment-concurrency    # 修复分支
  └─ hotfix/critical-bug               # 紧急修复分支
```

**分支命名规范**:
- `feature/模块名-功能名`: 新功能开发
- `bugfix/模块名-问题描述`: Bug 修复
- `hotfix/问题描述`: 紧急修复
- `refactor/模块名`: 重构代码

---

## 测试规范

### 单元测试

```go
// ✅ 推荐:测试文件命名为 _test.go
// order_service_test.go

func TestOrderService_CreateOrder(t *testing.T) {
    // 使用 table-driven tests
    tests := []struct {
        name    string
        req     *OrderRequest
        want    *Order
        wantErr bool
    }{
        {
            name: "创建包年订单成功",
            req: &OrderRequest{
                TenantID:  "T001",
                SKUCode:   "SKU_GPU_A100",
                OrderType: "SUBSCRIPTION",
            },
            want: &Order{
                OrderID:   "ORD001",
                OrderType: "SUBSCRIPTION",
                Status:    "PENDING",
            },
            wantErr: false,
        },
        {
            name: "租户ID为空返回错误",
            req: &OrderRequest{
                TenantID: "",
                SKUCode:  "SKU_GPU_A100",
            },
            want:    nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 创建 mock 对象
            mockRepo := NewMockOrderRepository()
            service := NewOrderService(mockRepo)
            
            // 执行测试
            got, err := service.CreateOrder(context.Background(), tt.req)
            
            // 断言
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateOrder() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr && got.OrderType != tt.want.OrderType {
                t.Errorf("CreateOrder() got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 集成测试

```go
// 使用 testcontainers 启动真实数据库
func TestOrderRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("跳过集成测试")
    }
    
    // 启动 MySQL 容器
    ctx := context.Background()
    mysqlC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "mysql:8.0",
            ExposedPorts: []string{"3306/tcp"},
            Env: map[string]string{
                "MYSQL_ROOT_PASSWORD": "password",
                "MYSQL_DATABASE":      "test_db",
            },
        },
        Started: true,
    })
    defer mysqlC.Terminate(ctx)
    
    // 连接数据库
    db := setupTestDB(t, mysqlC)
    
    // 执行测试
    repo := NewOrderRepository(db)
    order := &Order{
        OrderID:  "ORD001",
        TenantID: "T001",
        Status:   "PENDING",
    }
    
    err = repo.Create(ctx, order)
    assert.NoError(t, err)
    
    // 验证结果
    got, err := repo.GetByID(ctx, "ORD001")
    assert.NoError(t, err)
    assert.Equal(t, order.OrderID, got.OrderID)
}
```

### 测试覆盖率

```bash
# 运行测试并生成覆盖率报告
go test -cover -coverprofile=coverage.out ./...

# 查看覆盖率详情
go tool cover -html=coverage.out

# 要求:核心业务代码测试覆盖率 >= 80%
```

---

## 附录：工具配置

### .golangci.yml

```yaml
linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck
    - golint
    - stylecheck
    - gosec
    - dupl
    - goconst
    - gocyclo
    - gocognit

linters-settings:
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  goconst:
    min-len: 3
    min-occurrences: 3
```

### Makefile

```makefile
.PHONY: fmt lint test build clean

# 格式化代码
fmt:
	gofmt -w .
	goimports -w .

# 代码检查
lint:
	golangci-lint run

# 运行测试
test:
	go test -v -cover ./...

# 构建
build:
	go build -o bin/api cmd/api/main.go

# 清理
clean:
	rm -rf bin/
	go clean
```

---

**更新记录**:
- 2026-01-17: 初始版本发布
