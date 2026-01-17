package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// DiscountType 折扣类型
type DiscountType string

const (
	DiscountTypePercentage DiscountType = "PERCENTAGE" // 百分比折扣
	DiscountTypeAmount     DiscountType = "AMOUNT"     // 金额折扣
)

// DiscountTargetType 折扣目标类型
type DiscountTargetType string

const (
	DiscountTargetTypeTenant DiscountTargetType = "TENANT" // 租户级别
	DiscountTargetTypeSPU    DiscountTargetType = "SPU"    // SPU级别
	DiscountTargetTypeSKU    DiscountTargetType = "SKU"    // SKU级别
)

// ============================================================================
// 定价详情 JSON 结构体
// ============================================================================

// FixedPricingDetail 固定价格详情
type FixedPricingDetail struct {
	UnitPrice float64 `json:"unit_price"` // 单价
}

// TierPricing 阶梯价格层级
type TierPricing struct {
	From      float64  `json:"from"`       // 起始用量
	To        *float64 `json:"to"`         // 结束用量 (null 表示无限)
	UnitPrice float64  `json:"unit_price"` // 该阶梯单价
}

// TieredPricingDetail 阶梯价格详情
type TieredPricingDetail struct {
	Tiers []TierPricing `json:"tiers"` // 价格阶梯
}

// TimePeriodPricing 时段价格
type TimePeriodPricing struct {
	HourFrom  int     `json:"hour_from"`  // 起始小时 (0-23)
	HourTo    int     `json:"hour_to"`    // 结束小时 (0-24)
	UnitPrice float64 `json:"unit_price"` // 该时段单价
}

// TimeBasedPricingDetail 时段价格详情
type TimeBasedPricingDetail struct {
	Periods []TimePeriodPricing `json:"periods"` // 时段列表
}

// PackagePricingDetail 资源包详情
type PackagePricingDetail struct {
	PackageSize  float64 `json:"package_size"`  // 资源包大小
	PackagePrice float64 `json:"package_price"` // 资源包价格
	ValidityDays int     `json:"validity_days"` // 有效天数
}

// PricingDetail 定价详情（统一包装）
type PricingDetail map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (p *PricingDetail) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), p)
	}
	return json.Unmarshal(bytes, p)
}

// Value 实现 driver.Valuer 接口
func (p PricingDetail) Value() (driver.Value, error) {
	if len(p) == 0 {
		return nil, nil
	}
	return json.Marshal(p)
}

// ============================================================================
// PriceRule - 定价规则
// ============================================================================

// PriceRule 定价规则模型
type PriceRule struct {
	ID             int64         `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	RuleID         string        `gorm:"column:rule_id;uniqueIndex;not null" json:"rule_id"`
	RuleCode       string        `gorm:"column:rule_code;uniqueIndex;not null" json:"rule_code"`
	RuleName       string        `gorm:"column:rule_name;not null" json:"rule_name"`
	SpuCode        *string       `gorm:"column:spu_code;index" json:"spu_code,omitempty"`
	SkuCode        *string       `gorm:"column:sku_code;index" json:"sku_code,omitempty"`
	RuleType       PriceRuleType `gorm:"column:rule_type;not null" json:"rule_type"`
	PricingDetail  PricingDetail `gorm:"column:pricing_detail;type:json;not null" json:"pricing_detail"`
	Currency       Currency      `gorm:"column:currency;default:CNY" json:"currency"`
	EffectiveStart time.Time     `gorm:"column:effective_start;not null" json:"effective_start"`
	EffectiveEnd   *time.Time    `gorm:"column:effective_end" json:"effective_end,omitempty"`
	Region         *string       `gorm:"column:region;index" json:"region,omitempty"`
	Priority       int           `gorm:"column:priority;default:0;index" json:"priority"`
	Status         Status        `gorm:"column:status;default:1;index" json:"status"`
	CreatedAt      time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (PriceRule) TableName() string {
	return "price_rules"
}

// ============================================================================
// DiscountRule - 折扣规则
// ============================================================================

// DiscountRule 折扣规则模型
type DiscountRule struct {
	ID             int64              `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	DiscountID     string             `gorm:"column:discount_id;uniqueIndex;not null" json:"discount_id"`
	DiscountName   string             `gorm:"column:discount_name;not null" json:"discount_name"`
	DiscountType   DiscountType       `gorm:"column:discount_type;not null" json:"discount_type"`
	DiscountValue  float64            `gorm:"column:discount_value;not null" json:"discount_value"`
	TargetType     DiscountTargetType `gorm:"column:target_type;not null;index" json:"target_type"`
	TargetID       *string            `gorm:"column:target_id;index" json:"target_id,omitempty"`
	EffectiveStart time.Time          `gorm:"column:effective_start;not null" json:"effective_start"`
	EffectiveEnd   *time.Time         `gorm:"column:effective_end" json:"effective_end,omitempty"`
	MaxDiscount    *float64           `gorm:"column:max_discount" json:"max_discount,omitempty"`
	MinAmount      *float64           `gorm:"column:min_amount" json:"min_amount,omitempty"`
	Status         Status             `gorm:"column:status;default:1;index" json:"status"`
	CreatedAt      time.Time          `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time          `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (DiscountRule) TableName() string {
	return "discount_rules"
}

// ============================================================================
// 请求/响应 DTO
// ============================================================================

// CreatePriceRuleRequest 创建定价规则请求
type CreatePriceRuleRequest struct {
	RuleCode       string                 `json:"rule_code" binding:"required"`
	RuleName       string                 `json:"rule_name" binding:"required"`
	SpuCode        *string                `json:"spu_code,omitempty"`
	SkuCode        *string                `json:"sku_code,omitempty"`
	RuleType       PriceRuleType          `json:"rule_type" binding:"required"`
	PricingDetail  map[string]interface{} `json:"pricing_detail" binding:"required"`
	Currency       Currency               `json:"currency"`
	EffectiveStart time.Time              `json:"effective_start" binding:"required"`
	EffectiveEnd   *time.Time             `json:"effective_end,omitempty"`
	Region         *string                `json:"region,omitempty"`
	Priority       int                    `json:"priority"`
}

// CreateDiscountRuleRequest 创建折扣规则请求
type CreateDiscountRuleRequest struct {
	DiscountName   string             `json:"discount_name" binding:"required"`
	DiscountType   DiscountType       `json:"discount_type" binding:"required"`
	DiscountValue  float64            `json:"discount_value" binding:"required"`
	TargetType     DiscountTargetType `json:"target_type" binding:"required"`
	TargetID       *string            `json:"target_id,omitempty"`
	EffectiveStart time.Time          `json:"effective_start" binding:"required"`
	EffectiveEnd   *time.Time         `json:"effective_end,omitempty"`
	MaxDiscount    *float64           `json:"max_discount,omitempty"`
	MinAmount      *float64           `json:"min_amount,omitempty"`
}

// PriceQueryRequest 价格查询请求
type PriceQueryRequest struct {
	SkuCode  string  `json:"sku_code" binding:"required"`
	Region   *string `json:"region,omitempty"`
	TenantID *string `json:"tenant_id,omitempty"` // 用于计算折扣
}

// PriceCalculateRequest 价格计算请求
type PriceCalculateRequest struct {
	SkuCode   string    `json:"sku_code" binding:"required"`
	Region    *string   `json:"region,omitempty"`
	TenantID  string    `json:"tenant_id" binding:"required"`
	Quantity  float64   `json:"quantity" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

// PriceCalculateResponse 价格计算响应
type PriceCalculateResponse struct {
	OriginalPrice  float64        `json:"original_price"`  // 原价
	DiscountAmount float64        `json:"discount_amount"` // 折扣金额
	FinalPrice     float64        `json:"final_price"`     // 最终价格
	Currency       Currency       `json:"currency"`
	PriceRule      *PriceRule     `json:"price_rule,omitempty"`
	Discounts      []DiscountRule `json:"discounts,omitempty"`
}

// PriceRuleListQueryRequest 定价规则列表查询请求
type PriceRuleListQueryRequest struct {
	Pagination
	SpuCode  *string        `json:"spu_code" form:"spu_code"`
	SkuCode  *string        `json:"sku_code" form:"sku_code"`
	RuleType *PriceRuleType `json:"rule_type" form:"rule_type"`
	Region   *string        `json:"region" form:"region"`
	Status   *Status        `json:"status" form:"status"`
}

// DiscountRuleListQueryRequest 折扣规则列表查询请求
type DiscountRuleListQueryRequest struct {
	Pagination
	TargetType *DiscountTargetType `json:"target_type" form:"target_type"`
	TargetID   *string             `json:"target_id" form:"target_id"`
	Status     *Status             `json:"status" form:"status"`
}
