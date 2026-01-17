package models

import (
	"time"
)

// BaseModel 基础模型（所有表的公共字段）
type BaseModel struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// SoftDeleteModel 软删除模型
type SoftDeleteModel struct {
	BaseModel
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

// Status 通用状态
type Status int8

const (
	StatusDisabled Status = 0 // 禁用
	StatusEnabled  Status = 1 // 启用
)

// TenantType 租户类型
type TenantType string

const (
	TenantTypeEnterprise TenantType = "ENTERPRISE" // 企业
	TenantTypeIndividual TenantType = "INDIVIDUAL" // 个人
)

// VerifiedType 认证类型
type VerifiedType string

const (
	VerifiedTypeNone       VerifiedType = ""           // 未认证
	VerifiedTypePersonal   VerifiedType = "PERSONAL"   // 个人实名
	VerifiedTypeEnterprise VerifiedType = "ENTERPRISE" // 企业实名
)

// OrderType 订单类型
type OrderType string

const (
	OrderTypePrepaid  OrderType = "PREPAID"  // 预付费（包年包月）
	OrderTypePostpaid OrderType = "POSTPAID" // 后付费（按量计费）
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"   // 待支付
	OrderStatusPaid      OrderStatus = "PAID"      // 已支付
	OrderStatusActive    OrderStatus = "ACTIVE"    // 使用中
	OrderStatusExpired   OrderStatus = "EXPIRED"   // 已过期
	OrderStatusCancelled OrderStatus = "CANCELLED" // 已取消
	OrderStatusRefunded  OrderStatus = "REFUNDED"  // 已退款
)

// BillStatus 账单状态
type BillStatus string

const (
	BillStatusUnpaid    BillStatus = "UNPAID"    // 未支付
	BillStatusPaid      BillStatus = "PAID"      // 已支付
	BillStatusCancelled BillStatus = "CANCELLED" // 已取消（红冲）
)

// PaymentStatus 支付状态
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"   // 待支付
	PaymentStatusSuccess   PaymentStatus = "SUCCESS"   // 支付成功
	PaymentStatusFailed    PaymentStatus = "FAILED"    // 支付失败
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"  // 已退款
	PaymentStatusCancelled PaymentStatus = "CANCELLED" // 已取消
)

// PaymentMethod 支付方式
type PaymentMethod string

const (
	PaymentMethodBalance PaymentMethod = "BALANCE" // 余额支付
	PaymentMethodAlipay  PaymentMethod = "ALIPAY"  // 支付宝
	PaymentMethodWechat  PaymentMethod = "WECHAT"  // 微信支付
	PaymentMethodBank    PaymentMethod = "BANK"    // 银行转账
)

// Currency 货币类型
type Currency string

const (
	CurrencyCNY Currency = "CNY" // 人民币
	CurrencyUSD Currency = "USD" // 美元
	CurrencyEUR Currency = "EUR" // 欧元
	CurrencyJPY Currency = "JPY" // 日元
)

// PriceRuleType 定价规则类型
type PriceRuleType string

const (
	PriceRuleTypeFixed     PriceRuleType = "FIXED"      // 固定价格
	PriceRuleTypeTiered    PriceRuleType = "TIERED"     // 阶梯价格
	PriceRuleTypeTimeBased PriceRuleType = "TIME_BASED" // 时段价格
	PriceRuleTypePackage   PriceRuleType = "PACKAGE"    // 资源包
)

// BillingCycle 计费周期
type BillingCycle string

const (
	BillingCycleHourly  BillingCycle = "HOURLY"  // 按小时
	BillingCycleDaily   BillingCycle = "DAILY"   // 按天
	BillingCycleMonthly BillingCycle = "MONTHLY" // 按月
	BillingCycleYearly  BillingCycle = "YEARLY"  // 按年
)

// MeteringUnit 计量单位
type MeteringUnit string

const (
	MeteringUnitSecond      MeteringUnit = "SECOND"        // 秒
	MeteringUnitHour        MeteringUnit = "HOUR"          // 小时
	MeteringUnitGB          MeteringUnit = "GB"            // GB
	MeteringUnitGBHour      MeteringUnit = "GB_HOUR"       // GB·小时
	MeteringUnitToken       MeteringUnit = "TOKEN"         // Token
	MeteringUnitRequest     MeteringUnit = "REQUEST"       // 请求次数
	MeteringUnitGPUHour     MeteringUnit = "GPU_HOUR"      // GPU·小时
	MeteringUnitGPUCard     MeteringUnit = "GPU_CARD"      // GPU卡数
	MeteringUnitCPUCore     MeteringUnit = "CPU_CORE"      // CPU核心
	MeteringUnitCPUCoreHour MeteringUnit = "CPU_CORE_HOUR" // CPU核心·小时
)

// Pagination 分页参数
type Pagination struct {
	Page     int `json:"page" form:"page"`           // 页码，从1开始
	PageSize int `json:"page_size" form:"page_size"` // 每页数量
}

// GetOffset 计算偏移量
func (p *Pagination) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取限制数量
func (p *Pagination) GetLimit() int {
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return p.PageSize
}

// PageResult 分页结果
type PageResult struct {
	Total    int64       `json:"total"`     // 总记录数
	Page     int         `json:"page"`      // 当前页码
	PageSize int         `json:"page_size"` // 每页数量
	Data     interface{} `json:"data"`      // 数据列表
}
