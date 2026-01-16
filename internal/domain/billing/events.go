package billing

import (
	"time"

	"github.com/wwnj/happy-billing/pkg/money"
)

// BillGeneratedEvent 账单生成事件
type BillGeneratedEvent struct {
	BillID      string    // 账单ID
	BillNo      string    // 账单号
	TenantID    string    // 租户ID
	UserID      string    // 用户ID
	PeriodStart time.Time // 账期开始时间
	PeriodEnd   time.Time // 账期结束时间
	GeneratedAt time.Time // 生成时间
}

// EventType 返回事件类型
func (e *BillGeneratedEvent) EventType() string {
	return "bill.generated"
}

// BillSettledEvent 账单结算事件
type BillSettledEvent struct {
	BillID     string        // 账单ID
	BillNo     string        // 账单号
	PaidAmount money.Decimal // 支付金额
	SettledAt  time.Time     // 结算时间
}

// EventType 返回事件类型
func (e *BillSettledEvent) EventType() string {
	return "bill.settled"
}

// BillCancelledEvent 账单取消事件
type BillCancelledEvent struct {
	BillID      string    // 账单ID
	BillNo      string    // 账单号
	Reason      string    // 取消原因
	CancelledAt time.Time // 取消时间
}

// EventType 返回事件类型
func (e *BillCancelledEvent) EventType() string {
	return "bill.cancelled"
}

// BillOverdueEvent 账单逾期事件
type BillOverdueEvent struct {
	BillID            string        // 账单ID
	BillNo            string        // 账单号
	ActualAmount      money.Decimal // 应付金额
	OutstandingAmount money.Decimal // 未付金额
	OverdueAt         time.Time     // 逾期时间
}

// EventType 返回事件类型
func (e *BillOverdueEvent) EventType() string {
	return "bill.overdue"
}
