package meter

import (
	"context"

	"github.com/wwnj/happy-billing/pkg/money"
)

// MeterPlugin 计量器插件接口
// 这是可插拔设计的核心,不同资源类型可以实现不同的计量逻辑
type MeterPlugin interface {
	// Name 返回插件名称(唯一标识)
	Name() string

	// Collect 采集计量数据
	// resourceID: 资源ID
	// metadata: 扩展元数据(可以传递资源特定的信息)
	// 返回: 计量记录
	Collect(ctx context.Context, resourceID string, metadata map[string]interface{}) (*MeterRecord, error)

	// Aggregate 聚合多条计量记录
	// records: 待聚合的计量记录
	// 返回: 聚合后的计量值
	Aggregate(ctx context.Context, records []*MeterRecord) (money.Decimal, error)

	// Validate 验证插件配置是否有效
	// config: 插件配置(从MeterConfig.Config传入)
	// 返回: 如果配置无效则返回错误
	Validate(config map[string]interface{}) error
}
