package order

import (
	"time"

	"github.com/wwnj/happy-billing/pkg/money"
)

// OrderCreatedEvent 订单创建事件
type OrderCreatedEvent struct {
	OrderID   string    // 订单ID
	OrderNo   string    // 订单号
	TenantID  string    // 租户ID
	UserID    string    // 用户ID
	OrderType OrderType // 订单类型
	CreatedAt time.Time // 创建时间
}

// EventType 返回事件类型
func (e *OrderCreatedEvent) EventType() string {
	return "order.created"
}

// OrderPaidEvent 订单支付事件
type OrderPaidEvent struct {
	OrderID    string        // 订单ID
	PaidAmount money.Decimal // 支付金额
	PaidAt     time.Time     // 支付时间
}

// EventType 返回事件类型
func (e *OrderPaidEvent) EventType() string {
	return "order.paid"
}

// OrderCompletedEvent 订单完成事件
type OrderCompletedEvent struct {
	OrderID     string    // 订单ID
	CompletedAt time.Time // 完成时间
}

// EventType 返回事件类型
func (e *OrderCompletedEvent) EventType() string {
	return "order.completed"
}

// OrderCancelledEvent 订单取消事件
type OrderCancelledEvent struct {
	OrderID     string    // 订单ID
	Reason      string    // 取消原因
	CancelledAt time.Time // 取消时间
}

// EventType 返回事件类型
func (e *OrderCancelledEvent) EventType() string {
	return "order.cancelled"
}

// OrderFailedEvent 订单失败事件
type OrderFailedEvent struct {
	OrderID  string    // 订单ID
	Reason   string    // 失败原因
	FailedAt time.Time // 失败时间
}

// EventType 返回事件类型
func (e *OrderFailedEvent) EventType() string {
	return "order.failed"
}

// OrderRefundedEvent 订单退款事件
type OrderRefundedEvent struct {
	OrderID      string        // 订单ID
	RefundAmount money.Decimal // 退款金额
	Reason       string        // 退款原因
	RefundedAt   time.Time     // 退款时间
}

// EventType 返回事件类型
func (e *OrderRefundedEvent) EventType() string {
	return "order.refunded"
}
