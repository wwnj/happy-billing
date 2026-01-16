package pricing

import (
	"time"

	"github.com/wwnj/happy-billing/internal/domain/meter"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// TierPrice 阶梯价格(实体)
type TierPrice struct {
	ID         string        // 唯一ID
	RuleID     string        // 所属定价规则ID
	StartRange money.Decimal // 起始范围(包含)
	EndRange   money.Decimal // 结束范围(不包含,0表示无上限)
	UnitPrice  money.Decimal // 单价
	Currency   string        // 货币类型(CNY/USD等)
}

// NewTierPrice 创建阶梯价格
func NewTierPrice(id, ruleID string, startRange, endRange, unitPrice money.Decimal, currency string) (*TierPrice, error) {
	// 验证范围
	if startRange.LessThan(money.Zero) {
		return nil, errors.NewInvalidParam("start_range must be non-negative")
	}

	if !endRange.IsZero() && endRange.LessThanOrEqual(startRange) {
		return nil, errors.NewInvalidParam("end_range must be greater than start_range")
	}

	// 验证单价
	if unitPrice.LessThan(money.Zero) {
		return nil, errors.NewInvalidParam("unit_price must be non-negative")
	}

	// 验证货币
	if currency == "" {
		currency = "CNY" // 默认人民币
	}

	return &TierPrice{
		ID:         id,
		RuleID:     ruleID,
		StartRange: startRange,
		EndRange:   endRange,
		UnitPrice:  unitPrice,
		Currency:   currency,
	}, nil
}

// IsInRange 判断数量是否在当前阶梯范围内
func (t *TierPrice) IsInRange(quantity money.Decimal) bool {
	// 检查下限
	if quantity.LessThan(t.StartRange) {
		return false
	}

	// 检查上限(0表示无上限)
	if !t.EndRange.IsZero() && quantity.GreaterThanOrEqual(t.EndRange) {
		return false
	}

	return true
}

// Calculate 计算当前阶梯的费用
// quantity: 使用量
// 返回: 费用金额
func (t *TierPrice) Calculate(quantity money.Decimal) money.Decimal {
	if !t.IsInRange(quantity) {
		return money.Zero
	}

	// 计算当前阶梯的实际用量
	actualQuantity := quantity
	if !t.EndRange.IsZero() && actualQuantity.GreaterThan(t.EndRange) {
		actualQuantity = t.EndRange
	}
	actualQuantity = money.Sub(actualQuantity, t.StartRange)

	// 费用 = 用量 * 单价
	return money.Mul(actualQuantity, t.UnitPrice)
}

// PriceRule 定价规则(聚合根)
type PriceRule struct {
	ID           string             // 唯一ID
	Name         string             // 规则名称
	ResourceType meter.ResourceType // 资源类型
	PriceType    PriceType          // 定价类型
	TierPrices   []*TierPrice       // 阶梯价格列表(按StartRange升序排列)
	DiscountRule *DiscountRule      // 折扣规则
	ValidFrom    time.Time          // 生效时间
	ValidTo      time.Time          // 失效时间(零值表示永久有效)
	IsActive     bool               // 是否激活
	Priority     int                // 优先级(用于处理规则冲突,数字越小优先级越高)
	CreatedAt    time.Time          // 创建时间
	UpdatedAt    time.Time          // 更新时间
}

// NewPriceRule 创建定价规则
func NewPriceRule(
	id, name string,
	resourceType meter.ResourceType,
	priceType PriceType,
	validFrom, validTo time.Time,
	priority int,
) (*PriceRule, error) {
	// 验证资源类型
	if err := resourceType.Validate(); err != nil {
		return nil, err
	}

	// 验证定价类型
	if err := priceType.Validate(); err != nil {
		return nil, err
	}

	// 验证名称
	if name == "" {
		return nil, errors.NewInvalidParam("name cannot be empty")
	}

	// 验证时间范围
	if !validTo.IsZero() && validTo.Before(validFrom) {
		return nil, errors.NewInvalidParam("valid_to must be after valid_from")
	}

	now := time.Now()
	return &PriceRule{
		ID:           id,
		Name:         name,
		ResourceType: resourceType,
		PriceType:    priceType,
		TierPrices:   make([]*TierPrice, 0),
		DiscountRule: NoDiscount(),
		ValidFrom:    validFrom,
		ValidTo:      validTo,
		IsActive:     true,
		Priority:     priority,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// AddTierPrice 添加阶梯价格
func (p *PriceRule) AddTierPrice(tierPrice *TierPrice) error {
	// 验证定价类型是否支持阶梯
	if p.PriceType != PriceTypeTier {
		return errors.New(errors.CodeInvalidParam, "only TIER price type supports tier prices")
	}

	// 检查范围是否重叠
	for _, existing := range p.TierPrices {
		if tierPrice.IsInRange(existing.StartRange) || existing.IsInRange(tierPrice.StartRange) {
			return errors.New(errors.CodeInvalidParam, "tier price range overlaps with existing tier")
		}
	}

	p.TierPrices = append(p.TierPrices, tierPrice)
	p.UpdatedAt = time.Now()
	return nil
}

// SetDiscountRule 设置折扣规则
func (p *PriceRule) SetDiscountRule(discount *DiscountRule) {
	p.DiscountRule = discount
	p.UpdatedAt = time.Now()
}

// Calculate 计算费用
// quantity: 使用量
// context: 上下文信息(用于折扣条件判断)
// 返回: 费用金额
func (p *PriceRule) Calculate(quantity money.Decimal, context map[string]interface{}) (money.Decimal, error) {
	if !p.IsActive {
		return money.Zero, errors.New(errors.CodePriceRuleNotFound, "price rule is not active")
	}

	// 检查是否在有效期内
	now := time.Now()
	if now.Before(p.ValidFrom) {
		return money.Zero, errors.New(errors.CodeInvalidParam, "price rule not yet effective")
	}
	if !p.ValidTo.IsZero() && now.After(p.ValidTo) {
		return money.Zero, errors.New(errors.CodeInvalidParam, "price rule has expired")
	}

	var amount money.Decimal

	switch p.PriceType {
	case PriceTypeTier:
		// 阶梯定价
		amount = p.calculateTierPrice(quantity)
	case PriceTypeFlat:
		// 固定定价(使用第一个阶梯的单价)
		if len(p.TierPrices) > 0 {
			amount = money.Mul(quantity, p.TierPrices[0].UnitPrice)
		}
	default:
		return money.Zero, errors.New(errors.CodePriceCalculateFailed, "unsupported price type")
	}

	// 应用折扣
	if p.DiscountRule != nil && p.DiscountRule.IsApplicable(context) {
		amount = p.DiscountRule.Apply(amount)
	}

	return money.Round(amount, 4), nil
}

// calculateTierPrice 计算阶梯价格
func (p *PriceRule) calculateTierPrice(quantity money.Decimal) money.Decimal {
	totalAmount := money.Zero
	remainingQuantity := quantity

	for _, tier := range p.TierPrices {
		if remainingQuantity.LessThanOrEqual(money.Zero) {
			break
		}

		// 计算当前阶梯能消耗的量
		var tierQuantity money.Decimal
		if tier.EndRange.IsZero() {
			// 无上限,全部消耗
			tierQuantity = remainingQuantity
		} else {
			// 有上限
			maxTierQuantity := money.Sub(tier.EndRange, tier.StartRange)
			tierQuantity = money.Min(remainingQuantity, maxTierQuantity)
		}

		// 累加当前阶梯费用
		tierAmount := money.Mul(tierQuantity, tier.UnitPrice)
		totalAmount = money.Add(totalAmount, tierAmount)

		// 扣减已消耗的量
		remainingQuantity = money.Sub(remainingQuantity, tierQuantity)
	}

	return totalAmount
}

// Activate 激活规则
func (p *PriceRule) Activate() {
	p.IsActive = true
	p.UpdatedAt = time.Now()
}

// Deactivate 停用规则
func (p *PriceRule) Deactivate() {
	p.IsActive = false
	p.UpdatedAt = time.Now()
}

// IsEffective 检查规则是否有效
func (p *PriceRule) IsEffective(at time.Time) bool {
	if !p.IsActive {
		return false
	}

	if at.Before(p.ValidFrom) {
		return false
	}

	if !p.ValidTo.IsZero() && at.After(p.ValidTo) {
		return false
	}

	return true
}
