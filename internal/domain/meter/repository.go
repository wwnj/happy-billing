package meter

import (
	"context"
	"time"

	"github.com/wwnj/happy-billing/pkg/money"
)

// MeterRecordRepository 计量记录仓储接口
type MeterRecordRepository interface {
	// Save 保存单条计量记录
	Save(ctx context.Context, record *MeterRecord) error

	// BatchSave 批量保存计量记录(用于高性能写入)
	BatchSave(ctx context.Context, records []*MeterRecord) error

	// FindByID 根据ID查询计量记录
	FindByID(ctx context.Context, id string) (*MeterRecord, error)

	// QueryByRange 按时间范围查询计量记录
	// tenantID: 租户ID
	// resourceType: 资源类型(可为空,表示查询所有类型)
	// start, end: 时间范围
	QueryByRange(ctx context.Context, tenantID string, resourceType ResourceType, start, end time.Time) ([]*MeterRecord, error)

	// QueryByResourceID 查询指定资源的计量记录
	// resourceID: 资源ID
	// start, end: 时间范围
	QueryByResourceID(ctx context.Context, resourceID string, start, end time.Time) ([]*MeterRecord, error)

	// Aggregate 聚合计量数据
	// tenantID: 租户ID
	// resourceType: 资源类型
	// start, end: 时间范围
	// 返回: 聚合后的总计量值
	Aggregate(ctx context.Context, tenantID string, resourceType ResourceType, start, end time.Time) (money.Decimal, error)
}

// MeterConfigRepository 计量配置仓储接口
type MeterConfigRepository interface {
	// Save 保存计量配置
	Save(ctx context.Context, config *MeterConfig) error

	// FindByID 根据ID查询配置
	FindByID(ctx context.Context, id string) (*MeterConfig, error)

	// FindByResourceType 根据资源类型查询配置
	FindByResourceType(ctx context.Context, resourceType ResourceType) (*MeterConfig, error)

	// ListActive 查询所有激活的配置
	ListActive(ctx context.Context) ([]*MeterConfig, error)

	// Update 更新配置
	Update(ctx context.Context, config *MeterConfig) error

	// Delete 删除配置
	Delete(ctx context.Context, id string) error
}
