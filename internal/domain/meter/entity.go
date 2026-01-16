package meter

import (
	"time"

	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// MeterRecord 计量记录(事件溯源)
// 代表一次资源使用的计量记录
type MeterRecord struct {
	ID           string                 // 唯一ID
	TenantID     string                 // 租户ID
	UserID       string                 // 用户ID
	ResourceID   string                 // 资源ID(如GPU实例ID)
	ResourceType ResourceType           // 资源类型
	Quantity     money.Decimal          // 计量值
	Unit         string                 // 单位(秒/Token/GB等)
	StartTime    time.Time              // 计量开始时间
	EndTime      time.Time              // 计量结束时间
	Precision    Precision              // 计量精度
	MetaData     map[string]interface{} // 扩展元数据(如GPU型号、存储类型等)
	CreatedAt    time.Time              // 创建时间
}

// NewMeterRecord 创建新的计量记录
func NewMeterRecord(
	id, tenantID, userID, resourceID string,
	resourceType ResourceType,
	quantity money.Decimal,
	unit string,
	startTime, endTime time.Time,
	precision Precision,
	metadata map[string]interface{},
) (*MeterRecord, error) {
	// 验证资源类型
	if err := resourceType.Validate(); err != nil {
		return nil, err
	}

	// 验证精度
	if err := precision.Validate(); err != nil {
		return nil, err
	}

	// 验证时间范围
	if endTime.Before(startTime) {
		return nil, errors.New(errors.CodeInvalidParam, "end_time must be after start_time")
	}

	// 验证计量值
	if quantity.LessThan(money.Zero) {
		return nil, errors.New(errors.CodeInvalidParam, "quantity must be non-negative")
	}

	return &MeterRecord{
		ID:           id,
		TenantID:     tenantID,
		UserID:       userID,
		ResourceID:   resourceID,
		ResourceType: resourceType,
		Quantity:     quantity,
		Unit:         unit,
		StartTime:    startTime,
		EndTime:      endTime,
		Precision:    precision,
		MetaData:     metadata,
		CreatedAt:    time.Now(),
	}, nil
}

// Duration 获取计量时长(秒)
func (m *MeterRecord) Duration() int64 {
	return int64(m.EndTime.Sub(m.StartTime).Seconds())
}

// IsValid 验证记录是否有效
func (m *MeterRecord) IsValid() bool {
	return m.ID != "" &&
		m.TenantID != "" &&
		m.ResourceID != "" &&
		m.ResourceType.IsValid() &&
		m.Precision.IsValid() &&
		m.Quantity.GreaterThanOrEqual(money.Zero) &&
		!m.EndTime.Before(m.StartTime)
}

// MeterConfig 计量配置(聚合根)
// 定义某种资源类型的计量规则
type MeterConfig struct {
	ID           string                 // 唯一ID
	ResourceType ResourceType           // 资源类型
	Precision    Precision              // 计量精度
	MeterPlugin  string                 // 计量器插件名称
	IsActive     bool                   // 是否激活
	Config       map[string]interface{} // 插件特定配置
	CreatedAt    time.Time              // 创建时间
	UpdatedAt    time.Time              // 更新时间
}

// NewMeterConfig 创建新的计量配置
func NewMeterConfig(
	id string,
	resourceType ResourceType,
	precision Precision,
	meterPlugin string,
	config map[string]interface{},
) (*MeterConfig, error) {
	// 验证资源类型
	if err := resourceType.Validate(); err != nil {
		return nil, err
	}

	// 验证精度
	if err := precision.Validate(); err != nil {
		return nil, err
	}

	// 验证插件名称
	if meterPlugin == "" {
		return nil, errors.New(errors.CodeInvalidParam, "meter_plugin cannot be empty")
	}

	now := time.Now()
	return &MeterConfig{
		ID:           id,
		ResourceType: resourceType,
		Precision:    precision,
		MeterPlugin:  meterPlugin,
		IsActive:     true,
		Config:       config,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Activate 激活配置
func (m *MeterConfig) Activate() {
	m.IsActive = true
	m.UpdatedAt = time.Now()
}

// Deactivate 停用配置
func (m *MeterConfig) Deactivate() {
	m.IsActive = false
	m.UpdatedAt = time.Now()
}

// UpdatePlugin 更新插件配置
func (m *MeterConfig) UpdatePlugin(pluginName string, config map[string]interface{}) error {
	if pluginName == "" {
		return errors.New(errors.CodeInvalidParam, "plugin_name cannot be empty")
	}

	m.MeterPlugin = pluginName
	m.Config = config
	m.UpdatedAt = time.Now()
	return nil
}

// UpdatePrecision 更新计量精度
func (m *MeterConfig) UpdatePrecision(precision Precision) error {
	if err := precision.Validate(); err != nil {
		return err
	}

	m.Precision = precision
	m.UpdatedAt = time.Now()
	return nil
}
