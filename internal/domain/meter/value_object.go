package meter

import (
	"fmt"

	"github.com/wwnj/happy-billing/pkg/errors"
)

// ResourceType 资源类型(值对象)
type ResourceType string

// 资源类型常量
const (
	ResourceTypeGPU       ResourceType = "GPU"       // GPU算力
	ResourceTypeStorage   ResourceType = "STORAGE"   // 存储
	ResourceTypeLLMToken  ResourceType = "LLM_TOKEN" // LLM Token
	ResourceTypeBandwidth ResourceType = "BANDWIDTH" // 带宽
	ResourceTypeCPU       ResourceType = "CPU"       // CPU
	ResourceTypeMemory    ResourceType = "MEMORY"    // 内存
)

// String 返回字符串表示
func (r ResourceType) String() string {
	return string(r)
}

// IsValid 验证资源类型是否有效
func (r ResourceType) IsValid() bool {
	switch r {
	case ResourceTypeGPU, ResourceTypeStorage, ResourceTypeLLMToken,
		ResourceTypeBandwidth, ResourceTypeCPU, ResourceTypeMemory:
		return true
	default:
		return false
	}
}

// Validate 验证并返回错误
func (r ResourceType) Validate() error {
	if !r.IsValid() {
		return errors.New(errors.CodeMeterInvalidType,
			fmt.Sprintf("invalid resource type: %s", r))
	}
	return nil
}

// Precision 计量精度(值对象)
type Precision string

// 计量精度常量
const (
	PrecisionSecond Precision = "SECOND" // 秒级精度
	PrecisionMinute Precision = "MINUTE" // 分钟级精度
	PrecisionHour   Precision = "HOUR"   // 小时级精度
	PrecisionDay    Precision = "DAY"    // 天级精度
)

// String 返回字符串表示
func (p Precision) String() string {
	return string(p)
}

// IsValid 验证精度是否有效
func (p Precision) IsValid() bool {
	switch p {
	case PrecisionSecond, PrecisionMinute, PrecisionHour, PrecisionDay:
		return true
	default:
		return false
	}
}

// Validate 验证并返回错误
func (p Precision) Validate() error {
	if !p.IsValid() {
		return errors.New(errors.CodeMeterInvalidPrecision,
			fmt.Sprintf("invalid precision: %s", p))
	}
	return nil
}

// ToSeconds 转换为秒数(用于聚合计算)
func (p Precision) ToSeconds() int {
	switch p {
	case PrecisionSecond:
		return 1
	case PrecisionMinute:
		return 60
	case PrecisionHour:
		return 3600
	case PrecisionDay:
		return 86400
	default:
		return 1
	}
}
