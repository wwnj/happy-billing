package order

import (
	"context"
	"time"
)

// OrderRepository 订单仓储接口
type OrderRepository interface {
	// Save 保存订单(新建)
	Save(ctx context.Context, order *Order) error

	// Update 更新订单
	Update(ctx context.Context, order *Order) error

	// FindByID 根据ID查询订单
	FindByID(ctx context.Context, id string) (*Order, error)

	// FindByOrderNo 根据订单号查询订单
	FindByOrderNo(ctx context.Context, orderNo string) (*Order, error)

	// ListByUser 查询用户的订单列表
	// userID: 用户ID
	// status: 订单状态(为空则查询所有状态)
	// offset, limit: 分页参数
	ListByUser(ctx context.Context, userID string, status OrderStatus, offset, limit int) ([]*Order, int64, error)

	// ListByTenant 查询租户的订单列表
	// tenantID: 租户ID
	// status: 订单状态(为空则查询所有状态)
	// startTime, endTime: 时间范围(为空则不限制)
	// offset, limit: 分页参数
	ListByTenant(ctx context.Context, tenantID string, status OrderStatus, startTime, endTime *time.Time, offset, limit int) ([]*Order, int64, error)

	// CountByStatus 统计指定状态的订单数量
	CountByStatus(ctx context.Context, tenantID string, status OrderStatus) (int64, error)

	// SumAmountByPeriod 统计指定时间段的订单总金额
	// tenantID: 租户ID
	// startTime, endTime: 时间范围
	// status: 订单状态(为空则统计所有状态)
	SumAmountByPeriod(ctx context.Context, tenantID string, startTime, endTime time.Time, status OrderStatus) (map[string]interface{}, error)
}

// OrderItemRepository 订单项仓储接口
type OrderItemRepository interface {
	// Save 保存订单项
	Save(ctx context.Context, item *OrderItem) error

	// BatchSave 批量保存订单项
	BatchSave(ctx context.Context, items []*OrderItem) error

	// FindByOrderID 查询指定订单的所有订单项
	FindByOrderID(ctx context.Context, orderID string) ([]*OrderItem, error)

	// Update 更新订单项
	Update(ctx context.Context, item *OrderItem) error

	// Delete 删除订单项
	Delete(ctx context.Context, id string) error

	// DeleteByOrderID 删除指定订单的所有订单项
	DeleteByOrderID(ctx context.Context, orderID string) error
}
