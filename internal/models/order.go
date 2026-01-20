package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ResourceStatus 资源实例状态
type ResourceStatus string

const (
	ResourceStatusCreating ResourceStatus = "CREATING" // 创建中
	ResourceStatusRunning  ResourceStatus = "RUNNING"  // 运行中
	ResourceStatusStopped  ResourceStatus = "STOPPED"  // 已停止
	ResourceStatusDeleted  ResourceStatus = "DELETED"  // 已删除
)

// TransactionType 交易类型
type TransactionType string

const (
	TransactionTypeRecharge TransactionType = "RECHARGE" // 充值
	TransactionTypeDeduct   TransactionType = "DEDUCT"   // 扣减
	TransactionTypeRefund   TransactionType = "REFUND"   // 退款
	TransactionTypeFreeze   TransactionType = "FREEZE"   // 冻结
	TransactionTypeUnfreeze TransactionType = "UNFREEZE" // 解冻
)

// OrderDetail 订单详情 JSON 字段
type OrderDetail map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (o *OrderDetail) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), o)
	}
	return json.Unmarshal(bytes, o)
}

// Value 实现 driver.Valuer 接口
func (o OrderDetail) Value() (driver.Value, error) {
	if len(o) == 0 {
		return nil, nil
	}
	return json.Marshal(o)
}

// BillDetail 账单详情 JSON 字段
type BillDetail map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (b *BillDetail) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), b)
	}
	return json.Unmarshal(bytes, b)
}

// Value 实现 driver.Valuer 接口
func (b BillDetail) Value() (driver.Value, error) {
	if len(b) == 0 {
		return nil, nil
	}
	return json.Marshal(b)
}

// ExternalResponse 第三方支付响应 JSON 字段
type ExternalResponse map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (e *ExternalResponse) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), e)
	}
	return json.Unmarshal(bytes, e)
}

// Value 实现 driver.Valuer 接口
func (e ExternalResponse) Value() (driver.Value, error) {
	if len(e) == 0 {
		return nil, nil
	}
	return json.Marshal(e)
}

// ============================================================================
// Order - 订单
// ============================================================================

// Order 订单模型
type Order struct {
	ID                 int64        `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	OrderID            string       `gorm:"column:order_id;uniqueIndex;not null" json:"order_id"`
	OrderNo            string       `gorm:"column:order_no;uniqueIndex;not null" json:"order_no"`
	TenantID           string       `gorm:"column:tenant_id;not null;index" json:"tenant_id"`
	OrganizationID     *string      `gorm:"column:organization_id" json:"organization_id,omitempty"`
	ProjectID          string       `gorm:"column:project_id;not null;index" json:"project_id"`
	UserID             string       `gorm:"column:user_id;not null;index" json:"user_id"`
	OrderType          OrderType    `gorm:"column:order_type;not null" json:"order_type"`
	SpuCode            string       `gorm:"column:spu_code;not null" json:"spu_code"`
	SkuCode            string       `gorm:"column:sku_code;not null" json:"sku_code"`
	Currency           Currency     `gorm:"column:currency;default:CNY" json:"currency"`
	ExchangeRate       *float64     `gorm:"column:exchange_rate;type:decimal(18,8)" json:"exchange_rate,omitempty"`
	BaseCurrency       Currency     `gorm:"column:base_currency;default:CNY" json:"base_currency"`
	BaseCurrencyAmount *float64     `gorm:"column:base_currency_amount;type:decimal(18,4)" json:"base_currency_amount,omitempty"`
	OriginalAmount     float64      `gorm:"column:original_amount;not null" json:"original_amount"`
	DiscountAmount     float64      `gorm:"column:discount_amount;default:0" json:"discount_amount"`
	PayableAmount      float64      `gorm:"column:payable_amount;not null" json:"payable_amount"`
	PaidAmount         float64      `gorm:"column:paid_amount;default:0" json:"paid_amount"`
	PeriodStart        *time.Time   `gorm:"column:period_start" json:"period_start,omitempty"`
	PeriodEnd          *time.Time   `gorm:"column:period_end" json:"period_end,omitempty"`
	Status             OrderStatus  `gorm:"column:status;not null;index" json:"status"`
	OrderDetail        *OrderDetail `gorm:"column:order_detail;type:json" json:"order_detail,omitempty"`
	CreatedAt          time.Time    `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
	UpdatedAt          time.Time    `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// ============================================================================
// OrderItem - 订单明细
// ============================================================================

// OrderItem 订单明细模型
type OrderItem struct {
	ID          int64       `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	OrderID     string      `gorm:"column:order_id;not null;index" json:"order_id"`
	ItemNo      string      `gorm:"column:item_no;not null" json:"item_no"`
	SpuCode     string      `gorm:"column:spu_code;not null" json:"spu_code"`
	SkuCode     string      `gorm:"column:sku_code;not null;index" json:"sku_code"`
	SkuName     string      `gorm:"column:sku_name;not null" json:"sku_name"`
	SkuSpec     *SpecValues `gorm:"column:sku_spec;type:json" json:"sku_spec,omitempty"`
	Quantity    float64     `gorm:"column:quantity;not null" json:"quantity"`
	UnitPrice   float64     `gorm:"column:unit_price;not null" json:"unit_price"`
	Amount      float64     `gorm:"column:amount;not null" json:"amount"`
	PriceRuleID *string     `gorm:"column:price_rule_id" json:"price_rule_id,omitempty"`
	CreatedAt   time.Time   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}

// ============================================================================
// ResourceInstance - 资源实例
// ============================================================================

// InstanceSpec 实例规格 JSON 字段
type InstanceSpec map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (i *InstanceSpec) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), i)
	}
	return json.Unmarshal(bytes, i)
}

// Value 实现 driver.Valuer 接口
func (i InstanceSpec) Value() (driver.Value, error) {
	if len(i) == 0 {
		return nil, nil
	}
	return json.Marshal(i)
}

// ResourceInstance 资源实例模型
type ResourceInstance struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	InstanceID   string         `gorm:"column:instance_id;uniqueIndex;not null" json:"instance_id"`
	OrderID      string         `gorm:"column:order_id;not null;index" json:"order_id"`
	TenantID     string         `gorm:"column:tenant_id;not null;index" json:"tenant_id"`
	ProjectID    string         `gorm:"column:project_id;not null;index" json:"project_id"`
	ProductType  ProductType    `gorm:"column:product_type;not null" json:"product_type"`
	SpuCode      string         `gorm:"column:spu_code;not null" json:"spu_code"`
	SkuCode      string         `gorm:"column:sku_code;not null" json:"sku_code"`
	InstanceSpec *InstanceSpec  `gorm:"column:instance_spec;type:json" json:"instance_spec,omitempty"`
	Status       ResourceStatus `gorm:"column:status;not null;index" json:"status"`
	StartedAt    *time.Time     `gorm:"column:started_at" json:"started_at,omitempty"`
	StoppedAt    *time.Time     `gorm:"column:stopped_at" json:"stopped_at,omitempty"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (ResourceInstance) TableName() string {
	return "resource_instances"
}

// ============================================================================
// Bill - 账单
// ============================================================================

// Bill 账单模型
type Bill struct {
	ID                 int64       `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	BillID             string      `gorm:"column:bill_id;uniqueIndex;not null" json:"bill_id"`
	OrderID            string      `gorm:"column:order_id;not null;index" json:"order_id"`
	TenantID           string      `gorm:"column:tenant_id;not null;index" json:"tenant_id"`
	ProjectID          string      `gorm:"column:project_id;not null;index" json:"project_id"`
	BillType           string      `gorm:"column:bill_type;not null" json:"bill_type"`
	BillingPeriodStart *time.Time  `gorm:"column:billing_period_start;index" json:"billing_period_start,omitempty"`
	BillingPeriodEnd   *time.Time  `gorm:"column:billing_period_end;index" json:"billing_period_end,omitempty"`
	Currency           Currency    `gorm:"column:currency;default:CNY" json:"currency"`
	ExchangeRate       *float64    `gorm:"column:exchange_rate" json:"exchange_rate,omitempty"`
	BaseCurrency       *Currency   `gorm:"column:base_currency" json:"base_currency,omitempty"`
	BaseCurrencyAmount *float64    `gorm:"column:base_currency_amount" json:"base_currency_amount,omitempty"`
	OriginalAmount     float64     `gorm:"column:original_amount;not null" json:"original_amount"`
	DiscountAmount     float64     `gorm:"column:discount_amount;default:0" json:"discount_amount"`
	PayableAmount      float64     `gorm:"column:payable_amount;not null" json:"payable_amount"`
	Status             BillStatus  `gorm:"column:status;not null;index" json:"status"`
	BillDetail         *BillDetail `gorm:"column:bill_detail;type:json" json:"bill_detail,omitempty"`
	PaidAt             *time.Time  `gorm:"column:paid_at" json:"paid_at,omitempty"`
	CreatedAt          time.Time   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Bill) TableName() string {
	return "bills"
}

// ============================================================================
// Payment - 支付记录
// ============================================================================

// Payment 支付记录模型
type Payment struct {
	ID                 int64             `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	PaymentID          string            `gorm:"column:payment_id;uniqueIndex;not null" json:"payment_id"`
	OrderID            *string           `gorm:"column:order_id;index" json:"order_id,omitempty"`
	BillID             *string           `gorm:"column:bill_id;index" json:"bill_id,omitempty"`
	TenantID           string            `gorm:"column:tenant_id;not null;index" json:"tenant_id"`
	UserID             string            `gorm:"column:user_id;not null" json:"user_id"`
	PaymentMethod      PaymentMethod     `gorm:"column:payment_method;not null" json:"payment_method"`
	PaymentChannel     *string           `gorm:"column:payment_channel" json:"payment_channel,omitempty"`
	Currency           Currency          `gorm:"column:currency;default:CNY" json:"currency"`
	ExchangeRate       *float64          `gorm:"column:exchange_rate;type:decimal(18,8)" json:"exchange_rate,omitempty"`
	BaseCurrency       Currency          `gorm:"column:base_currency;default:CNY" json:"base_currency"`
	BaseCurrencyAmount *float64          `gorm:"column:base_currency_amount;type:decimal(18,4)" json:"base_currency_amount,omitempty"`
	Amount             float64           `gorm:"column:amount;not null" json:"amount"`
	Status             PaymentStatus     `gorm:"column:status;not null;index" json:"status"`
	ExternalOrderID    *string           `gorm:"column:external_order_id;index" json:"external_order_id,omitempty"`
	ExternalResponse   *ExternalResponse `gorm:"column:external_response;type:json" json:"external_response,omitempty"`
	PaidAt             *time.Time        `gorm:"column:paid_at" json:"paid_at,omitempty"`
	CreatedAt          time.Time         `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time         `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Payment) TableName() string {
	return "payments"
}

// ============================================================================
// AccountBalance - 账户余额
// ============================================================================

// AccountBalance 账户余额模型
type AccountBalance struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	TenantID      string    `gorm:"column:tenant_id;uniqueIndex;not null" json:"tenant_id"`
	Balance       float64   `gorm:"column:balance;not null;default:0" json:"balance"`
	FrozenBalance float64   `gorm:"column:frozen_balance;not null;default:0" json:"frozen_balance"`
	CreditLimit   float64   `gorm:"column:credit_limit;not null;default:0" json:"credit_limit"`
	Currency      Currency  `gorm:"column:currency;default:CNY" json:"currency"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (AccountBalance) TableName() string {
	return "account_balances"
}

// ============================================================================
// BalanceTransaction - 余额变动记录
// ============================================================================

// BalanceTransaction 余额变动记录模型
type BalanceTransaction struct {
	ID               int64           `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	TransactionID    string          `gorm:"column:transaction_id;uniqueIndex;not null" json:"transaction_id"`
	TenantID         string          `gorm:"column:tenant_id;not null;index" json:"tenant_id"`
	TransactionType  TransactionType `gorm:"column:transaction_type;not null;index" json:"transaction_type"`
	Amount           float64         `gorm:"column:amount;not null" json:"amount"`
	BalanceBefore    float64         `gorm:"column:balance_before;not null" json:"balance_before"`
	BalanceAfter     float64         `gorm:"column:balance_after;not null" json:"balance_after"`
	RelatedOrderID   *string         `gorm:"column:related_order_id" json:"related_order_id,omitempty"`
	RelatedBillID    *string         `gorm:"column:related_bill_id" json:"related_bill_id,omitempty"`
	RelatedPaymentID *string         `gorm:"column:related_payment_id" json:"related_payment_id,omitempty"`
	Remark           *string         `gorm:"column:remark" json:"remark,omitempty"`
	CreatedAt        time.Time       `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

// TableName 指定表名
func (BalanceTransaction) TableName() string {
	return "balance_transactions"
}

// ============================================================================
// 请求/响应 DTO
// ============================================================================

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	TenantID       string                 `json:"tenant_id" binding:"required"`
	ProjectID      string                 `json:"project_id" binding:"required"`
	UserID         string                 `json:"user_id" binding:"required"`
	OrganizationID *string                `json:"organization_id,omitempty"`
	OrderType      OrderType              `json:"order_type" binding:"required"`
	SkuCode        string                 `json:"sku_code" binding:"required"`
	Quantity       float64                `json:"quantity" binding:"required"`
	PeriodStart    *time.Time             `json:"period_start,omitempty"`
	PeriodEnd      *time.Time             `json:"period_end,omitempty"`
	OrderDetail    map[string]interface{} `json:"order_detail,omitempty"`
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OrderID        *string       `json:"order_id,omitempty"`
	BillID         *string       `json:"bill_id,omitempty"`
	TenantID       string        `json:"tenant_id" binding:"required"`
	UserID         string        `json:"user_id" binding:"required"`
	PaymentMethod  PaymentMethod `json:"payment_method" binding:"required"`
	PaymentChannel *string       `json:"payment_channel,omitempty"`
	Amount         float64       `json:"amount" binding:"required"`
	Currency       Currency      `json:"currency"`
}

// OrderListQueryRequest 订单列表查询请求
type OrderListQueryRequest struct {
	Pagination
	TenantID  *string      `json:"tenant_id" form:"tenant_id"`
	ProjectID *string      `json:"project_id" form:"project_id"`
	UserID    *string      `json:"user_id" form:"user_id"`
	OrderType *OrderType   `json:"order_type" form:"order_type"`
	Status    *OrderStatus `json:"status" form:"status"`
	Currency  *string      `json:"currency" form:"currency"`
	Keyword   *string      `json:"keyword" form:"keyword"`
}

// BillListQueryRequest 账单列表查询请求
type BillListQueryRequest struct {
	Pagination
	TenantID  *string     `json:"tenant_id" form:"tenant_id"`
	ProjectID *string     `json:"project_id" form:"project_id"`
	OrderID   *string     `json:"order_id" form:"order_id"`
	BillType  *string     `json:"bill_type" form:"bill_type"`
	Status    *BillStatus `json:"status" form:"status"`
}
