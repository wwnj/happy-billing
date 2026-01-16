package meters

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wwnj/happy-billing/internal/domain/meter"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// GPUMeter GPU计量器插件
// 按秒级精度计量GPU使用时长
type GPUMeter struct {
	// 可以添加配置字段,如API客户端等
}

// NewGPUMeter 创建GPU计量器
func NewGPUMeter() *GPUMeter {
	return &GPUMeter{}
}

// Name 返回插件名称
func (g *GPUMeter) Name() string {
	return "gpu_meter"
}

// Collect 采集GPU计量数据
func (g *GPUMeter) Collect(ctx context.Context, resourceID string, metadata map[string]interface{}) (*meter.MeterRecord, error) {
	// 从metadata中提取必要信息
	tenantID, ok := metadata["tenant_id"].(string)
	if !ok || tenantID == "" {
		return nil, errors.NewInvalidParam("tenant_id is required")
	}

	userID, ok := metadata["user_id"].(string)
	if !ok || userID == "" {
		return nil, errors.NewInvalidParam("user_id is required")
	}

	startTime, ok := metadata["start_time"].(time.Time)
	if !ok {
		// 尝试从字符串解析
		if startTimeStr, ok := metadata["start_time"].(string); ok {
			var err error
			startTime, err = time.Parse(time.RFC3339, startTimeStr)
			if err != nil {
				return nil, errors.NewInvalidParam("invalid start_time format")
			}
		} else {
			return nil, errors.NewInvalidParam("start_time is required")
		}
	}

	endTime, ok := metadata["end_time"].(time.Time)
	if !ok {
		// 尝试从字符串解析
		if endTimeStr, ok := metadata["end_time"].(string); ok {
			var err error
			endTime, err = time.Parse(time.RFC3339, endTimeStr)
			if err != nil {
				return nil, errors.NewInvalidParam("invalid end_time format")
			}
		} else {
			// 如果没有提供end_time,使用当前时间
			endTime = time.Now()
		}
	}

	// 计算使用时长(秒)
	duration := endTime.Sub(startTime).Seconds()
	if duration < 0 {
		return nil, errors.NewInvalidParam("end_time must be after start_time")
	}

	quantity := money.NewFromFloat(duration)

	// 创建计量记录
	record, err := meter.NewMeterRecord(
		uuid.New().String(),
		tenantID,
		userID,
		resourceID,
		meter.ResourceTypeGPU,
		quantity,
		"seconds", // 单位:秒
		startTime,
		endTime,
		meter.PrecisionSecond, // 秒级精度
		metadata,
	)
	if err != nil {
		return nil, errors.Wrap(errors.CodeMeterCollectFailed, "failed to create meter record", err)
	}

	return record, nil
}

// Aggregate 聚合GPU计量记录
// 将多条记录的时长相加
func (g *GPUMeter) Aggregate(ctx context.Context, records []*meter.MeterRecord) (money.Decimal, error) {
	if len(records) == 0 {
		return money.Zero, nil
	}

	total := money.Zero
	for _, record := range records {
		// 验证记录类型
		if record.ResourceType != meter.ResourceTypeGPU {
			return money.Zero, errors.New(
				errors.CodeMeterAggregateFailed,
				fmt.Sprintf("invalid resource type: expected GPU, got %s", record.ResourceType),
			)
		}

		// 累加计量值
		total = money.Add(total, record.Quantity)
	}

	return total, nil
}

// Validate 验证GPU计量器配置
func (g *GPUMeter) Validate(config map[string]interface{}) error {
	// GPU计量器目前不需要特殊配置
	// 这里可以验证配置参数,如GPU型号、地域等

	// 示例:检查是否配置了GPU型号
	if gpuModel, ok := config["gpu_model"].(string); ok {
		validModels := []string{"A100", "V100", "T4", "A10", "H100"}
		isValid := false
		for _, model := range validModels {
			if gpuModel == model {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("unsupported GPU model: %s", gpuModel)
		}
	}

	return nil
}

// init 在包初始化时自动注册GPU计量器
func init() {
	gpuMeter := NewGPUMeter()
	if err := Register(gpuMeter); err != nil {
		panic(fmt.Sprintf("failed to register GPU meter: %v", err))
	}
}
