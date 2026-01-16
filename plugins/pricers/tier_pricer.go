package pricers

import (
	"context"
	"fmt"

	"github.com/wwnj/happy-billing/internal/domain/pricing"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// TierPricer 阶梯定价器插件
// 支持阶梯定价策略：用量越多单价越低
type TierPricer struct{}

// NewTierPricer 创建阶梯定价器
func NewTierPricer() *TierPricer {
	return &TierPricer{}
}

// Name 返回插件名称
func (t *TierPricer) Name() string {
	return "tier_pricer"
}

// Calculate 计算阶梯价格
// 算法：
// 1. 遍历所有阶梯价格（已按StartRange升序排列）
// 2. 计算每个阶梯内的用量和费用
// 3. 累加所有阶梯费用得到总费用
func (t *TierPricer) Calculate(ctx context.Context, quantity money.Decimal, rule *pricing.PriceRule, context map[string]interface{}) (money.Decimal, error) {
	// 验证定价类型
	if rule.PriceType != pricing.PriceTypeTier {
		return money.Zero, errors.New(
			errors.CodePriceCalculateFailed,
			fmt.Sprintf("tier pricer only supports TIER price type, got %s", rule.PriceType),
		)
	}

	// 验证阶梯价格是否存在
	if len(rule.TierPrices) == 0 {
		return money.Zero, errors.New(
			errors.CodePriceCalculateFailed,
			"no tier prices configured",
		)
	}

	// 计算阶梯价格
	totalAmount := money.Zero
	remainingQuantity := quantity

	for _, tier := range rule.TierPrices {
		if remainingQuantity.LessThanOrEqual(money.Zero) {
			break
		}

		// 计算当前阶梯能消耗的量
		var tierQuantity money.Decimal

		// 检查是否在当前阶梯范围内
		if remainingQuantity.LessThan(tier.StartRange) {
			// 用量未达到当前阶梯起点，跳过
			continue
		}

		// 计算当前阶梯的实际用量
		if tier.EndRange.IsZero() {
			// 无上限，全部消耗
			tierQuantity = money.Sub(remainingQuantity, tier.StartRange)
		} else {
			// 有上限
			if remainingQuantity.LessThan(tier.EndRange) {
				// 用量在当前阶梯范围内
				tierQuantity = money.Sub(remainingQuantity, tier.StartRange)
			} else {
				// 用量超过当前阶梯上限
				tierQuantity = money.Sub(tier.EndRange, tier.StartRange)
			}
		}

		// 累加当前阶梯费用
		tierAmount := money.Mul(tierQuantity, tier.UnitPrice)
		totalAmount = money.Add(totalAmount, tierAmount)

		// 更新剩余量（用于下一个阶梯计算）
		if !tier.EndRange.IsZero() && remainingQuantity.GreaterThan(tier.EndRange) {
			remainingQuantity = money.Sub(remainingQuantity, tier.EndRange)
		} else {
			remainingQuantity = money.Zero
		}
	}

	// 四舍五入到4位小数
	return money.Round(totalAmount, 4), nil
}

// ApplyDiscount 应用折扣
func (t *TierPricer) ApplyDiscount(ctx context.Context, amount money.Decimal, discount *pricing.DiscountRule) (money.Decimal, error) {
	if discount == nil {
		return amount, nil
	}

	discountedAmount := discount.Apply(amount)
	return money.Round(discountedAmount, 4), nil
}

// Validate 验证阶梯定价配置
func (t *TierPricer) Validate(rule *pricing.PriceRule) error {
	// 检查定价类型
	if rule.PriceType != pricing.PriceTypeTier {
		return fmt.Errorf("price type must be TIER")
	}

	// 检查是否有阶梯价格
	if len(rule.TierPrices) == 0 {
		return fmt.Errorf("at least one tier price required")
	}

	// 检查阶梯价格是否连续
	for i := 0; i < len(rule.TierPrices); i++ {
		tier := rule.TierPrices[i]

		// 检查单价非负
		if tier.UnitPrice.LessThan(money.Zero) {
			return fmt.Errorf("tier %d: unit price must be non-negative", i)
		}

		// 检查范围有效性
		if tier.StartRange.LessThan(money.Zero) {
			return fmt.Errorf("tier %d: start range must be non-negative", i)
		}

		if !tier.EndRange.IsZero() && tier.EndRange.LessThanOrEqual(tier.StartRange) {
			return fmt.Errorf("tier %d: end range must be greater than start range", i)
		}

		// 检查阶梯连续性（当前阶梯的起点应该等于上一阶梯的终点）
		if i > 0 {
			prevTier := rule.TierPrices[i-1]
			if !prevTier.EndRange.IsZero() && !tier.StartRange.Equal(prevTier.EndRange) {
				return fmt.Errorf("tier %d: start range must equal previous tier end range for continuity", i)
			}
		} else {
			// 第一个阶梯应该从0开始
			if !tier.StartRange.IsZero() {
				return fmt.Errorf("first tier must start from 0")
			}
		}
	}

	// 检查最后一个阶梯是否有上限（建议无上限）
	lastTier := rule.TierPrices[len(rule.TierPrices)-1]
	if !lastTier.EndRange.IsZero() {
		// 警告：最后一个阶梯有上限，可能导致超出范围的用量无法计费
		// 这里只记录，不返回错误
	}

	return nil
}

// init 在包初始化时自动注册阶梯定价器
func init() {
	tierPricer := NewTierPricer()
	if err := Register(tierPricer); err != nil {
		panic(fmt.Sprintf("failed to register tier pricer: %v", err))
	}
}
