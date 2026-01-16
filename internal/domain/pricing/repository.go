package pricing

import (
	"context"

	"github.com/wwnj/happy-billing/internal/domain/meter"
)

// PriceRuleRepository 定价规则仓储接口
type PriceRuleRepository interface {
	// Save 保存定价规则
	Save(ctx context.Context, rule *PriceRule) error

	// FindByID 根据ID查询定价规则
	FindByID(ctx context.Context, id string) (*PriceRule, error)

	// FindByResourceType 根据资源类型查询定价规则
	// 返回激活的、有效的定价规则(按优先级排序)
	FindByResourceType(ctx context.Context, resourceType meter.ResourceType) ([]*PriceRule, error)

	// FindEffective 查询当前有效的定价规则
	// resourceType: 资源类型
	// 返回: 按优先级排序的有效规则(优先级数字越小越靠前)
	FindEffective(ctx context.Context, resourceType meter.ResourceType) (*PriceRule, error)

	// ListActive 查询所有激活的定价规则
	ListActive(ctx context.Context) ([]*PriceRule, error)

	// Update 更新定价规则
	Update(ctx context.Context, rule *PriceRule) error

	// Delete 删除定价规则
	Delete(ctx context.Context, id string) error
}

// TierPriceRepository 阶梯价格仓储接口
type TierPriceRepository interface {
	// Save 保存阶梯价格
	Save(ctx context.Context, tierPrice *TierPrice) error

	// BatchSave 批量保存阶梯价格
	BatchSave(ctx context.Context, tierPrices []*TierPrice) error

	// FindByRuleID 查询指定规则的所有阶梯价格
	// 返回按StartRange升序排列的阶梯价格列表
	FindByRuleID(ctx context.Context, ruleID string) ([]*TierPrice, error)

	// Update 更新阶梯价格
	Update(ctx context.Context, tierPrice *TierPrice) error

	// Delete 删除阶梯价格
	Delete(ctx context.Context, id string) error

	// DeleteByRuleID 删除指定规则的所有阶梯价格
	DeleteByRuleID(ctx context.Context, ruleID string) error
}
