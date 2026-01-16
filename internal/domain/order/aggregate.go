package order

import (
	"time"

	"github.com/wwnj/happy-billing/internal/domain/meter"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// OrderItem 订单项(实体)
type OrderItem struct {
	ID           string                 // 唯一ID
	OrderID      string                 // 订单ID
	ResourceType meter.ResourceType     // 资源类型
	ResourceID   string                 // 资源ID
	Quantity     money.Decimal          // 数量
	UnitPrice    money.Decimal          // 单价
	Amount       money.Decimal          // 金额
	StartTime    *time.Time             // 计量开始时间(后付费)
	EndTime      *time.Time             // 计量结束时间(后付费)
	MetaData     map[string]interface{} // 扩展元数据
}

// NewOrderItem 创建订单项
func NewOrderItem(
	id, orderID string,
	resourceType meter.ResourceType,
	resourceID string,
	quantity, unitPrice money.Decimal,
	startTime, endTime *time.Time,
	metadata map[string]interface{},
) (*OrderItem, error) {
	// 验证资源类型
	if err := resourceType.Validate(); err != nil {
		return nil, err
	}

	// 验证数量和单价
	if quantity.LessThan(money.Zero) {
		return nil, errors.NewInvalidParam("quantity must be non-negative")
	}

	if unitPrice.LessThan(money.Zero) {
		return nil, errors.NewInvalidParam("unit_price must be non-negative")
	}

	// 计算金额
	amount := money.Mul(quantity, unitPrice)

	return &OrderItem{
		ID:           id,
		OrderID:      orderID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Quantity:     quantity,
		UnitPrice:    unitPrice,
		Amount:       money.Round(amount, 4),
		StartTime:    startTime,
		EndTime:      endTime,
		MetaData:     metadata,
	}, nil
}

// Order 订单聚合根
type Order struct {
	ID          string        // 唯一ID
	OrderNo     string        // 订单号(业务主键)
	TenantID    string        // 租户ID
	UserID      string        // 用户ID
	OrderType   OrderType     // 订单类型
	PaymentMode PaymentMode   // 支付方式
	Status      OrderStatus   // 订单状态
	Items       []*OrderItem  // 订单项
	TotalAmount money.Decimal // 总金额
	PaidAmount  money.Decimal // 已支付金额
	Currency    string        // 货币类型
	Remark      string        // 备注
	CreatedAt   time.Time     // 创建时间
	UpdatedAt   time.Time     // 更新时间
	PaidAt      *time.Time    // 支付时间
	CompletedAt *time.Time    // 完成时间
	CancelledAt *time.Time    // 取消时间

	// 领域事件(事件溯源)
	events []interface{}
}

// NewOrder 创建新订单
func NewOrder(
	id, orderNo, tenantID, userID string,
	orderType OrderType,
	paymentMode PaymentMode,
	currency string,
) (*Order, error) {
	// 验证订单类型
	if err := orderType.Validate(); err != nil {
		return nil, err
	}

	// 验证支付方式
	if err := paymentMode.Validate(); err != nil {
		return nil, err
	}

	// 验证必填字段
	if orderNo == "" {
		return nil, errors.NewInvalidParam("order_no cannot be empty")
	}
	if tenantID == "" {
		return nil, errors.NewInvalidParam("tenant_id cannot be empty")
	}
	if userID == "" {
		return nil, errors.NewInvalidParam("user_id cannot be empty")
	}

	// 默认货币
	if currency == "" {
		currency = "CNY"
	}

	now := time.Now()
	order := &Order{
		ID:          id,
		OrderNo:     orderNo,
		TenantID:    tenantID,
		UserID:      userID,
		OrderType:   orderType,
		PaymentMode: paymentMode,
		Status:      OrderStatusPending,
		Items:       make([]*OrderItem, 0),
		TotalAmount: money.Zero,
		PaidAmount:  money.Zero,
		Currency:    currency,
		CreatedAt:   now,
		UpdatedAt:   now,
		events:      make([]interface{}, 0),
	}

	// 发布订单创建事件
	order.AddEvent(&OrderCreatedEvent{
		OrderID:   id,
		OrderNo:   orderNo,
		TenantID:  tenantID,
		UserID:    userID,
		OrderType: orderType,
		CreatedAt: now,
	})

	return order, nil
}

// AddItem 添加订单项
func (o *Order) AddItem(item *OrderItem) error {
	if o.Status.IsFinal() {
		return errors.New(errors.CodeOrderInvalidStatus, "cannot add item to finalized order")
	}

	o.Items = append(o.Items, item)

	// 重新计算总金额
	o.RecalculateTotalAmount()
	o.UpdatedAt = time.Now()

	return nil
}

// RecalculateTotalAmount 重新计算总金额
func (o *Order) RecalculateTotalAmount() {
	total := money.Zero
	for _, item := range o.Items {
		total = money.Add(total, item.Amount)
	}
	o.TotalAmount = money.Round(total, 4)
}

// Pay 支付订单
func (o *Order) Pay(amount money.Decimal) error {
	// 验证状态
	if o.Status != OrderStatusPending {
		return errors.New(errors.CodeOrderInvalidStatus,
			"only pending orders can be paid")
	}

	// 验证金额
	if amount.LessThanOrEqual(money.Zero) {
		return errors.NewInvalidParam("payment amount must be positive")
	}

	// 更新已支付金额
	o.PaidAmount = money.Add(o.PaidAmount, amount)

	// 检查是否已全额支付
	if o.PaidAmount.GreaterThanOrEqual(o.TotalAmount) {
		o.Status = OrderStatusPaid
		now := time.Now()
		o.PaidAt = &now

		// 发布订单支付事件
		o.AddEvent(&OrderPaidEvent{
			OrderID:    o.ID,
			PaidAmount: o.PaidAmount,
			PaidAt:     now,
		})
	}

	o.UpdatedAt = time.Now()
	return nil
}

// Complete 完成订单
func (o *Order) Complete() error {
	// 验证状态(已支付或待处理状态可以完成)
	if o.Status != OrderStatusPaid && o.Status != OrderStatusPending {
		return errors.New(errors.CodeOrderInvalidStatus,
			"only paid or pending orders can be completed")
	}

	o.Status = OrderStatusCompleted
	now := time.Now()
	o.CompletedAt = &now
	o.UpdatedAt = now

	// 发布订单完成事件
	o.AddEvent(&OrderCompletedEvent{
		OrderID:     o.ID,
		CompletedAt: now,
	})

	return nil
}

// Cancel 取消订单
func (o *Order) Cancel(reason string) error {
	// 验证状态(已完成或已退款的订单不能取消)
	if o.Status == OrderStatusCompleted || o.Status == OrderStatusRefunded {
		return errors.New(errors.CodeOrderCannotCancel,
			"completed or refunded orders cannot be cancelled")
	}

	if o.Status == OrderStatusCancelled {
		return errors.New(errors.CodeOrderInvalidStatus, "order already cancelled")
	}

	o.Status = OrderStatusCancelled
	o.Remark = reason
	now := time.Now()
	o.CancelledAt = &now
	o.UpdatedAt = now

	// 发布订单取消事件
	o.AddEvent(&OrderCancelledEvent{
		OrderID:     o.ID,
		Reason:      reason,
		CancelledAt: now,
	})

	return nil
}

// Fail 订单失败
func (o *Order) Fail(reason string) error {
	if o.Status.IsFinal() {
		return errors.New(errors.CodeOrderInvalidStatus, "order already in final status")
	}

	o.Status = OrderStatusFailed
	o.Remark = reason
	o.UpdatedAt = time.Now()

	// 发布订单失败事件
	o.AddEvent(&OrderFailedEvent{
		OrderID:  o.ID,
		Reason:   reason,
		FailedAt: time.Now(),
	})

	return nil
}

// Refund 退款
func (o *Order) Refund(amount money.Decimal, reason string) error {
	// 验证状态(只有已完成的订单可以退款)
	if o.Status != OrderStatusCompleted {
		return errors.New(errors.CodeOrderInvalidStatus, "only completed orders can be refunded")
	}

	// 验证退款金额
	if amount.LessThanOrEqual(money.Zero) || amount.GreaterThan(o.PaidAmount) {
		return errors.NewInvalidParam("invalid refund amount")
	}

	o.Status = OrderStatusRefunded
	o.Remark = reason
	o.UpdatedAt = time.Now()

	// 发布退款事件
	o.AddEvent(&OrderRefundedEvent{
		OrderID:      o.ID,
		RefundAmount: amount,
		Reason:       reason,
		RefundedAt:   time.Now(),
	})

	return nil
}

// 领域事件管理

// AddEvent 添加领域事件
func (o *Order) AddEvent(event interface{}) {
	o.events = append(o.events, event)
}

// GetEvents 获取所有领域事件
func (o *Order) GetEvents() []interface{} {
	return o.events
}

// ClearEvents 清空领域事件(发布后调用)
func (o *Order) ClearEvents() {
	o.events = make([]interface{}, 0)
}

// HasEvents 是否有未发布的领域事件
func (o *Order) HasEvents() bool {
	return len(o.events) > 0
}

// 辅助方法

// IsPending 是否待处理
func (o *Order) IsPending() bool {
	return o.Status == OrderStatusPending
}

// IsPaid 是否已支付
func (o *Order) IsPaid() bool {
	return o.Status == OrderStatusPaid
}

// IsCompleted 是否已完成
func (o *Order) IsCompleted() bool {
	return o.Status == OrderStatusCompleted
}

// IsCancelled 是否已取消
func (o *Order) IsCancelled() bool {
	return o.Status == OrderStatusCancelled
}

// IsFailed 是否失败
func (o *Order) IsFailed() bool {
	return o.Status == OrderStatusFailed
}

// OutstandingAmount 未付金额
func (o *Order) OutstandingAmount() money.Decimal {
	return money.Sub(o.TotalAmount, o.PaidAmount)
}
