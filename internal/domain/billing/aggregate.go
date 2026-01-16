package billing

import (
	"time"

	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// BillItem 账单明细（实体）
type BillItem struct {
	ID          string                 // 唯一ID
	BillID      string                 // 账单ID
	Type        BillItemType           // 明细类型
	OrderID     *string                // 关联订单ID
	Description string                 // 描述
	Amount      money.Decimal          // 金额
	Quantity    money.Decimal          // 数量
	UnitPrice   money.Decimal          // 单价
	Discount    money.Decimal          // 折扣金额
	TaxAmount   money.Decimal          // 税额
	TotalAmount money.Decimal          // 总金额（含税）
	MetaData    map[string]interface{} // 扩展元数据
	CreatedAt   time.Time              // 创建时间
}

// NewBillItem 创建账单明细
func NewBillItem(
	id, billID string,
	itemType BillItemType,
	orderID *string,
	description string,
	amount money.Decimal,
	metadata map[string]interface{},
) (*BillItem, error) {
	// 验证必填字段
	if id == "" {
		return nil, errors.NewInvalidParam("bill item id cannot be empty")
	}
	if billID == "" {
		return nil, errors.NewInvalidParam("bill_id cannot be empty")
	}

	// 验证类型
	if err := itemType.Validate(); err != nil {
		return nil, err
	}

	// 验证金额
	if amount.IsZero() {
		return nil, errors.NewInvalidParam("bill item amount cannot be zero")
	}

	return &BillItem{
		ID:          id,
		BillID:      billID,
		Type:        itemType,
		OrderID:     orderID,
		Description: description,
		Amount:      amount,
		Quantity:    money.Zero,
		UnitPrice:   money.Zero,
		Discount:    money.Zero,
		TaxAmount:   money.Zero,
		TotalAmount: amount,
		MetaData:    metadata,
		CreatedAt:   time.Now(),
	}, nil
}

// SetQuantityAndPrice 设置数量和单价
func (item *BillItem) SetQuantityAndPrice(quantity, unitPrice money.Decimal) {
	item.Quantity = quantity
	item.UnitPrice = unitPrice
	item.Amount = money.Mul(quantity, unitPrice)
	item.TotalAmount = money.Sub(money.Add(item.Amount, item.TaxAmount), item.Discount)
}

// SetDiscount 设置折扣
func (item *BillItem) SetDiscount(discount money.Decimal) error {
	if discount.LessThan(money.Zero) {
		return errors.NewInvalidParam("discount cannot be negative")
	}
	if discount.GreaterThan(item.Amount) {
		return errors.NewInvalidParam("discount cannot exceed amount")
	}

	item.Discount = discount
	item.TotalAmount = money.Sub(money.Add(item.Amount, item.TaxAmount), item.Discount)
	return nil
}

// SetTax 设置税额
func (item *BillItem) SetTax(taxAmount money.Decimal) error {
	if taxAmount.LessThan(money.Zero) {
		return errors.NewInvalidParam("tax amount cannot be negative")
	}

	item.TaxAmount = taxAmount
	item.TotalAmount = money.Sub(money.Add(item.Amount, item.TaxAmount), item.Discount)
	return nil
}

// Bill 账单聚合根
type Bill struct {
	ID                string        // 唯一ID
	BillNo            string        // 账单号
	TenantID          string        // 租户ID
	UserID            string        // 用户ID
	Cycle             BillCycle     // 账单周期
	Status            BillStatus    // 账单状态
	PeriodStart       time.Time     // 账期开始时间
	PeriodEnd         time.Time     // 账期结束时间
	TotalAmount       money.Decimal // 总金额
	DiscountAmount    money.Decimal // 折扣金额
	TaxAmount         money.Decimal // 税额
	ActualAmount      money.Decimal // 实际应付金额
	PaidAmount        money.Decimal // 已支付金额
	OutstandingAmount money.Decimal // 未付金额
	Currency          string        // 货币类型
	Items             []*BillItem   // 账单明细
	DueDate           *time.Time    // 到期日期
	SettledAt         *time.Time    // 结算时间
	CreatedAt         time.Time     // 创建时间
	UpdatedAt         time.Time     // 更新时间

	// 领域事件
	events []interface{}
}

// NewBill 创建新账单
func NewBill(
	id, billNo, tenantID, userID string,
	cycle BillCycle,
	periodStart, periodEnd time.Time,
	currency string,
) (*Bill, error) {
	// 验证必填字段
	if tenantID == "" {
		return nil, errors.NewInvalidParam("tenant_id cannot be empty")
	}
	if userID == "" {
		return nil, errors.NewInvalidParam("user_id cannot be empty")
	}
	if billNo == "" {
		return nil, errors.NewInvalidParam("bill_no cannot be empty")
	}

	// 验证账单周期
	if err := cycle.Validate(); err != nil {
		return nil, err
	}

	// 验证账期时间
	if periodStart.After(periodEnd) {
		return nil, errors.NewInvalidParam("period_start must be before period_end")
	}

	// 默认货币
	if currency == "" {
		currency = "CNY"
	}

	now := time.Now()
	bill := &Bill{
		ID:                id,
		BillNo:            billNo,
		TenantID:          tenantID,
		UserID:            userID,
		Cycle:             cycle,
		Status:            BillStatusPending,
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		TotalAmount:       money.Zero,
		DiscountAmount:    money.Zero,
		TaxAmount:         money.Zero,
		ActualAmount:      money.Zero,
		PaidAmount:        money.Zero,
		OutstandingAmount: money.Zero,
		Currency:          currency,
		Items:             make([]*BillItem, 0),
		CreatedAt:         now,
		UpdatedAt:         now,
		events:            make([]interface{}, 0),
	}

	// 发布账单创建事件
	bill.AddEvent(&BillGeneratedEvent{
		BillID:      id,
		BillNo:      billNo,
		TenantID:    tenantID,
		UserID:      userID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		GeneratedAt: now,
	})

	return bill, nil
}

// AddItem 添加账单明细
func (b *Bill) AddItem(item *BillItem) error {
	if item == nil {
		return errors.NewInvalidParam("bill item cannot be nil")
	}

	if item.BillID != b.ID {
		return errors.NewInvalidParam("bill item does not belong to this bill")
	}

	b.Items = append(b.Items, item)
	b.RecalculateAmount()

	return nil
}

// RecalculateAmount 重新计算账单金额
func (b *Bill) RecalculateAmount() {
	var totalAmount money.Decimal = money.Zero
	var discountAmount money.Decimal = money.Zero
	var taxAmount money.Decimal = money.Zero

	for _, item := range b.Items {
		totalAmount = money.Add(totalAmount, item.Amount)
		discountAmount = money.Add(discountAmount, item.Discount)
		taxAmount = money.Add(taxAmount, item.TaxAmount)
	}

	b.TotalAmount = totalAmount
	b.DiscountAmount = discountAmount
	b.TaxAmount = taxAmount
	b.ActualAmount = money.Sub(money.Add(totalAmount, taxAmount), discountAmount)
	b.OutstandingAmount = money.Sub(b.ActualAmount, b.PaidAmount)
	b.UpdatedAt = time.Now()
}

// Settle 结算账单
func (b *Bill) Settle(paidAmount money.Decimal) error {
	// 验证状态
	if b.Status != BillStatusPending && b.Status != BillStatusOverdue {
		return errors.New(errors.CodeBillAlreadyPaid, "bill already settled or cancelled")
	}

	// 验证金额
	if paidAmount.LessThan(money.Zero) {
		return errors.NewInvalidParam("paid amount cannot be negative")
	}

	// 记录支付金额
	b.PaidAmount = money.Add(b.PaidAmount, paidAmount)
	b.OutstandingAmount = money.Sub(b.ActualAmount, b.PaidAmount)

	// 检查是否全部支付
	if b.OutstandingAmount.LessThanOrEqual(money.Zero) {
		b.Status = BillStatusSettled
		now := time.Now()
		b.SettledAt = &now

		// 发布账单结算事件
		b.AddEvent(&BillSettledEvent{
			BillID:     b.ID,
			BillNo:     b.BillNo,
			PaidAmount: paidAmount,
			SettledAt:  now,
		})
	}

	b.UpdatedAt = time.Now()
	return nil
}

// Cancel 取消账单
func (b *Bill) Cancel(reason string) error {
	// 验证状态
	if b.Status.IsFinal() {
		return errors.New(errors.CodeInvalidParam, "cannot cancel finalized bill")
	}

	b.Status = BillStatusCancelled
	b.UpdatedAt = time.Now()

	// 发布账单取消事件
	b.AddEvent(&BillCancelledEvent{
		BillID:      b.ID,
		BillNo:      b.BillNo,
		Reason:      reason,
		CancelledAt: time.Now(),
	})

	return nil
}

// MarkOverdue 标记为逾期
func (b *Bill) MarkOverdue() error {
	// 只能将待结算状态标记为逾期
	if b.Status != BillStatusPending {
		return errors.New(errors.CodeInvalidParam, "only pending bills can be marked as overdue")
	}

	b.Status = BillStatusOverdue
	b.UpdatedAt = time.Now()

	// 发布账单逾期事件
	b.AddEvent(&BillOverdueEvent{
		BillID:            b.ID,
		BillNo:            b.BillNo,
		ActualAmount:      b.ActualAmount,
		OutstandingAmount: b.OutstandingAmount,
		OverdueAt:         time.Now(),
	})

	return nil
}

// SetDueDate 设置到期日期
func (b *Bill) SetDueDate(dueDate time.Time) error {
	if dueDate.Before(b.PeriodEnd) {
		return errors.NewInvalidParam("due date must be after period end")
	}

	b.DueDate = &dueDate
	b.UpdatedAt = time.Now()
	return nil
}

// IsOverdue 是否逾期
func (b *Bill) IsOverdue() bool {
	if b.DueDate == nil {
		return false
	}

	return time.Now().After(*b.DueDate) && b.Status == BillStatusPending
}

// 领域事件管理

// AddEvent 添加领域事件
func (b *Bill) AddEvent(event interface{}) {
	b.events = append(b.events, event)
}

// GetEvents 获取所有领域事件
func (b *Bill) GetEvents() []interface{} {
	return b.events
}

// ClearEvents 清空领域事件（发布后调用）
func (b *Bill) ClearEvents() {
	b.events = make([]interface{}, 0)
}

// HasEvents 是否有未发布的领域事件
func (b *Bill) HasEvents() bool {
	return len(b.events) > 0
}
