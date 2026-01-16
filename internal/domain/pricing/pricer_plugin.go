package pricing

import (
	"context"

	"github.com/wwnj/happy-billing/pkg/money"
)

// PricerPlugin 定价器插件接口
// 支持不同的定价策略实现
type PricerPlugin interface {
	// Name 返回插件名称(唯一标识)
	Name() string

	// Calculate 计算价格
	// quantity: 使用量
	// rule: 定价规则
	// context: 上下文信息(用于复杂定价逻辑)
	// 返回: 计算后的金额
	Calculate(ctx context.Context, quantity money.Decimal, rule *PriceRule, context map[string]interface{}) (money.Decimal, error)

	// ApplyDiscount 应用折扣
	// amount: 原价
	// discount: 折扣规则
	// 返回: 折后价
	ApplyDiscount(ctx context.Context, amount money.Decimal, discount *DiscountRule) (money.Decimal, error)

	// Validate 验证定价配置
	// rule: 定价规则
	// 返回: 如果配置无效则返回错误
	Validate(rule *PriceRule) error
}
