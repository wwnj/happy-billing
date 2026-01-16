package balance

import (
	"time"

	"github.com/wwnj/happy-billing/pkg/money"
)

// AccountCreatedEvent 账户创建事件
type AccountCreatedEvent struct {
	AccountID string    // 账户ID
	TenantID  string    // 租户ID
	UserID    string    // 用户ID
	CreatedAt time.Time // 创建时间
}

// EventType 返回事件类型
func (e *AccountCreatedEvent) EventType() string {
	return "account.created"
}

// AccountChargedEvent 账户充值事件
type AccountChargedEvent struct {
	AccountID     string        // 账户ID
	TransactionID string        // 交易ID
	Amount        money.Decimal // 充值金额
	BalanceBefore money.Decimal // 充值前余额
	BalanceAfter  money.Decimal // 充值后余额
	Description   string        // 描述
	ChargedAt     time.Time     // 充值时间
}

// EventType 返回事件类型
func (e *AccountChargedEvent) EventType() string {
	return "account.charged"
}

// AccountDeductedEvent 账户扣费事件
type AccountDeductedEvent struct {
	AccountID     string        // 账户ID
	TransactionID string        // 交易ID
	Amount        money.Decimal // 扣费金额
	BalanceBefore money.Decimal // 扣费前余额
	BalanceAfter  money.Decimal // 扣费后余额
	OrderID       *string       // 关联订单ID
	Description   string        // 描述
	DeductedAt    time.Time     // 扣费时间
}

// EventType 返回事件类型
func (e *AccountDeductedEvent) EventType() string {
	return "account.deducted"
}

// AccountFrozenEvent 账户冻结事件
type AccountFrozenEvent struct {
	AccountID           string        // 账户ID
	TransactionID       string        // 交易ID
	Amount              money.Decimal // 冻结金额
	BalanceBefore       money.Decimal // 冻结前可用余额
	BalanceAfter        money.Decimal // 冻结后可用余额
	FrozenBalanceBefore money.Decimal // 冻结前冻结余额
	FrozenBalanceAfter  money.Decimal // 冻结后冻结余额
	Description         string        // 描述
	FrozenAt            time.Time     // 冻结时间
}

// EventType 返回事件类型
func (e *AccountFrozenEvent) EventType() string {
	return "account.frozen"
}

// AccountUnfrozenEvent 账户解冻事件
type AccountUnfrozenEvent struct {
	AccountID           string        // 账户ID
	TransactionID       string        // 交易ID
	Amount              money.Decimal // 解冻金额
	BalanceBefore       money.Decimal // 解冻前可用余额
	BalanceAfter        money.Decimal // 解冻后可用余额
	FrozenBalanceBefore money.Decimal // 解冻前冻结余额
	FrozenBalanceAfter  money.Decimal // 解冻后冻结余额
	Description         string        // 描述
	UnfrozenAt          time.Time     // 解冻时间
}

// EventType 返回事件类型
func (e *AccountUnfrozenEvent) EventType() string {
	return "account.unfrozen"
}

// AccountRefundedEvent 账户退款事件
type AccountRefundedEvent struct {
	AccountID     string        // 账户ID
	TransactionID string        // 交易ID
	Amount        money.Decimal // 退款金额
	BalanceBefore money.Decimal // 退款前余额
	BalanceAfter  money.Decimal // 退款后余额
	OrderID       *string       // 关联订单ID
	Description   string        // 描述
	RefundedAt    time.Time     // 退款时间
}

// EventType 返回事件类型
func (e *AccountRefundedEvent) EventType() string {
	return "account.refunded"
}

// AccountAdjustedEvent 账户调整事件
type AccountAdjustedEvent struct {
	AccountID     string        // 账户ID
	TransactionID string        // 交易ID
	Amount        money.Decimal // 调整金额(正数增加，负数减少)
	BalanceBefore money.Decimal // 调整前余额
	BalanceAfter  money.Decimal // 调整后余额
	Reason        string        // 调整原因
	AdjustedAt    time.Time     // 调整时间
}

// EventType 返回事件类型
func (e *AccountAdjustedEvent) EventType() string {
	return "account.adjusted"
}
