package pricing

import (
	"fmt"

	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// PriceType 定价类型(值对象)
type PriceType string

// 定价类型常量
const (
	PriceTypeTier      PriceType = "TIER"      // 阶梯定价
	PriceTypeFlat      PriceType = "FLAT"      // 固定定价
	PriceTypePromotion PriceType = "PROMOTION" // 促销定价
	PriceTypeSpot      PriceType = "SPOT"      // 竞价定价
)

// String 返回字符串表示
func (p PriceType) String() string {
	return string(p)
}

// IsValid 验证定价类型是否有效
func (p PriceType) IsValid() bool {
	switch p {
	case PriceTypeTier, PriceTypeFlat, PriceTypePromotion, PriceTypeSpot:
		return true
	default:
		return false
	}
}

// Validate 验证并返回错误
func (p PriceType) Validate() error {
	if !p.IsValid() {
		return errors.New(errors.CodeInvalidParam,
			fmt.Sprintf("invalid price type: %s", p))
	}
	return nil
}

// DiscountType 折扣类型(值对象)
type DiscountType string

// 折扣类型常量
const (
	DiscountTypePercentage DiscountType = "PERCENTAGE" // 百分比折扣(如8折=80%)
	DiscountTypeFixed      DiscountType = "FIXED"      // 固定金额折扣
	DiscountTypeNone       DiscountType = "NONE"       // 无折扣
)

// String 返回字符串表示
func (d DiscountType) String() string {
	return string(d)
}

// IsValid 验证折扣类型是否有效
func (d DiscountType) IsValid() bool {
	switch d {
	case DiscountTypePercentage, DiscountTypeFixed, DiscountTypeNone:
		return true
	default:
		return false
	}
}

// Validate 验证并返回错误
func (d DiscountType) Validate() error {
	if !d.IsValid() {
		return errors.New(errors.CodeInvalidParam,
			fmt.Sprintf("invalid discount type: %s", d))
	}
	return nil
}

// DiscountRule 折扣规则(值对象,不可变)
type DiscountRule struct {
	Type       DiscountType           // 折扣类型
	Value      money.Decimal          // 折扣值(百分比:0-100, 固定金额:具体金额)
	Conditions map[string]interface{} // 折扣条件(如最小消费金额、特定用户等)
}

// NewDiscountRule 创建折扣规则
func NewDiscountRule(discountType DiscountType, value money.Decimal, conditions map[string]interface{}) (*DiscountRule, error) {
	// 验证折扣类型
	if err := discountType.Validate(); err != nil {
		return nil, err
	}

	// 验证折扣值
	if discountType == DiscountTypePercentage {
		if value.LessThan(money.Zero) || value.GreaterThan(money.NewFromInt(100)) {
			return nil, errors.NewInvalidParam("percentage discount must be between 0 and 100")
		}
	} else if discountType == DiscountTypeFixed {
		if value.LessThan(money.Zero) {
			return nil, errors.NewInvalidParam("fixed discount must be non-negative")
		}
	}

	return &DiscountRule{
		Type:       discountType,
		Value:      value,
		Conditions: conditions,
	}, nil
}

// Apply 应用折扣到原价
// originalAmount: 原价
// 返回: 折后价
func (d *DiscountRule) Apply(originalAmount money.Decimal) money.Decimal {
	if d.Type == DiscountTypeNone {
		return originalAmount
	}

	var discountedAmount money.Decimal
	if d.Type == DiscountTypePercentage {
		// 百分比折扣: 原价 * (100 - 折扣值) / 100
		// 例如: 100元 * 80% = 80元
		percentage := money.Sub(money.NewFromInt(100), d.Value)
		discountedAmount = money.Mul(originalAmount, percentage).Div(money.Hundred)
	} else if d.Type == DiscountTypeFixed {
		// 固定金额折扣: 原价 - 折扣值
		discountedAmount = money.Sub(originalAmount, d.Value)
		// 确保不会是负数
		if discountedAmount.LessThan(money.Zero) {
			discountedAmount = money.Zero
		}
	}

	return money.Round(discountedAmount, 4)
}

// IsApplicable 检查折扣是否适用(根据条件)
func (d *DiscountRule) IsApplicable(context map[string]interface{}) bool {
	// 如果没有条件,则总是适用
	if len(d.Conditions) == 0 {
		return true
	}

	// 检查最小消费金额条件
	if minAmount, ok := d.Conditions["min_amount"].(string); ok {
		minAmountDecimal, err := money.NewFromString(minAmount)
		if err == nil {
			if amount, ok := context["amount"].(money.Decimal); ok {
				if amount.LessThan(minAmountDecimal) {
					return false
				}
			}
		}
	}

	// 检查用户条件
	if requiredUserIDs, ok := d.Conditions["user_ids"].([]string); ok {
		if userID, ok := context["user_id"].(string); ok {
			found := false
			for _, id := range requiredUserIDs {
				if id == userID {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// 其他条件可以继续扩展...

	return true
}

// NoDiscount 创建无折扣规则
func NoDiscount() *DiscountRule {
	return &DiscountRule{
		Type:       DiscountTypeNone,
		Value:      money.Zero,
		Conditions: nil,
	}
}
