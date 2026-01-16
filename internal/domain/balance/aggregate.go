package balance

import (
	"time"

	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// Transaction 交易记录(实体，用于事件溯源)
type Transaction struct {
	ID            string                 // 唯一ID
	TransactionID string                 // 业务交易ID(幂等性key)
	AccountID     string                 // 账户ID
	Type          TransactionType        // 交易类型
	Amount        money.Decimal          // 交易金额
	BalanceBefore money.Decimal          // 交易前余额
	BalanceAfter  money.Decimal          // 交易后余额
	OrderID       *string                // 关联订单ID(可选)
	Description   string                 // 交易描述
	MetaData      map[string]interface{} // 扩展元数据
	CreatedAt     time.Time              // 创建时间
}

// NewTransaction 创建交易记录
func NewTransaction(
	id, transactionID, accountID string,
	transactionType TransactionType,
	amount, balanceBefore, balanceAfter money.Decimal,
	orderID *string,
	description string,
	metadata map[string]interface{},
) (*Transaction, error) {
	// 验证交易类型
	if err := transactionType.Validate(); err != nil {
		return nil, err
	}

	// 验证交易ID(幂等性key)
	if transactionID == "" {
		return nil, errors.NewInvalidParam("transaction_id cannot be empty")
	}

	// 验证账户ID
	if accountID == "" {
		return nil, errors.NewInvalidParam("account_id cannot be empty")
	}

	// 验证金额
	if amount.LessThanOrEqual(money.Zero) {
		return nil, errors.NewInvalidParam("transaction amount must be positive")
	}

	return &Transaction{
		ID:            id,
		TransactionID: transactionID,
		AccountID:     accountID,
		Type:          transactionType,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		OrderID:       orderID,
		Description:   description,
		MetaData:      metadata,
		CreatedAt:     time.Now(),
	}, nil
}

// Account 账户聚合根
type Account struct {
	ID            string        // 唯一ID
	TenantID      string        // 租户ID
	UserID        string        // 用户ID
	Balance       money.Decimal // 可用余额
	FrozenBalance money.Decimal // 冻结余额
	Currency      string        // 货币类型
	Version       int64         // 版本号(乐观锁)
	CreatedAt     time.Time     // 创建时间
	UpdatedAt     time.Time     // 更新时间

	// 领域事件
	events []interface{}
}

// NewAccount 创建新账户
func NewAccount(id, tenantID, userID, currency string) (*Account, error) {
	// 验证必填字段
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
	account := &Account{
		ID:            id,
		TenantID:      tenantID,
		UserID:        userID,
		Balance:       money.Zero,
		FrozenBalance: money.Zero,
		Currency:      currency,
		Version:       0,
		CreatedAt:     now,
		UpdatedAt:     now,
		events:        make([]interface{}, 0),
	}

	// 发布账户创建事件
	account.AddEvent(&AccountCreatedEvent{
		AccountID: id,
		TenantID:  tenantID,
		UserID:    userID,
		CreatedAt: now,
	})

	return account, nil
}

// Charge 充值(增加可用余额)
func (a *Account) Charge(amount money.Decimal, transactionID string, description string) error {
	// 验证金额
	if amount.LessThanOrEqual(money.Zero) {
		return errors.NewInvalidParam("charge amount must be positive")
	}

	// 验证交易ID
	if transactionID == "" {
		return errors.NewInvalidParam("transaction_id cannot be empty")
	}

	// 记录交易前余额
	balanceBefore := a.Balance

	// 更新余额
	a.Balance = money.Add(a.Balance, amount)
	a.UpdatedAt = time.Now()
	a.Version++

	// 发布充值事件
	a.AddEvent(&AccountChargedEvent{
		AccountID:     a.ID,
		TransactionID: transactionID,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  a.Balance,
		Description:   description,
		ChargedAt:     time.Now(),
	})

	return nil
}

// Deduct 扣费(减少可用余额)
func (a *Account) Deduct(amount money.Decimal, transactionID string, orderID *string, description string) error {
	// 验证金额
	if amount.LessThanOrEqual(money.Zero) {
		return errors.NewInvalidParam("deduct amount must be positive")
	}

	// 验证交易ID
	if transactionID == "" {
		return errors.NewInvalidParam("transaction_id cannot be empty")
	}

	// 检查余额是否足够
	if a.Balance.LessThan(amount) {
		return errors.NewInsufficientBalance()
	}

	// 记录交易前余额
	balanceBefore := a.Balance

	// 更新余额
	a.Balance = money.Sub(a.Balance, amount)
	a.UpdatedAt = time.Now()
	a.Version++

	// 发布扣费事件
	a.AddEvent(&AccountDeductedEvent{
		AccountID:     a.ID,
		TransactionID: transactionID,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  a.Balance,
		OrderID:       orderID,
		Description:   description,
		DeductedAt:    time.Now(),
	})

	return nil
}

// Freeze 冻结金额(从可用余额转到冻结余额)
func (a *Account) Freeze(amount money.Decimal, transactionID string, description string) error {
	// 验证金额
	if amount.LessThanOrEqual(money.Zero) {
		return errors.NewInvalidParam("freeze amount must be positive")
	}

	// 检查可用余额是否足够
	if a.Balance.LessThan(amount) {
		return errors.NewInsufficientBalance()
	}

	// 记录交易前余额
	balanceBefore := a.Balance
	frozenBefore := a.FrozenBalance

	// 更新余额
	a.Balance = money.Sub(a.Balance, amount)
	a.FrozenBalance = money.Add(a.FrozenBalance, amount)
	a.UpdatedAt = time.Now()
	a.Version++

	// 发布冻结事件
	a.AddEvent(&AccountFrozenEvent{
		AccountID:           a.ID,
		TransactionID:       transactionID,
		Amount:              amount,
		BalanceBefore:       balanceBefore,
		BalanceAfter:        a.Balance,
		FrozenBalanceBefore: frozenBefore,
		FrozenBalanceAfter:  a.FrozenBalance,
		Description:         description,
		FrozenAt:            time.Now(),
	})

	return nil
}

// Unfreeze 解冻金额(从冻结余额转回可用余额)
func (a *Account) Unfreeze(amount money.Decimal, transactionID string, description string) error {
	// 验证金额
	if amount.LessThanOrEqual(money.Zero) {
		return errors.NewInvalidParam("unfreeze amount must be positive")
	}

	// 检查冻结余额是否足够
	if a.FrozenBalance.LessThan(amount) {
		return errors.New(errors.CodeInvalidParam, "insufficient frozen balance")
	}

	// 记录交易前余额
	balanceBefore := a.Balance
	frozenBefore := a.FrozenBalance

	// 更新余额
	a.FrozenBalance = money.Sub(a.FrozenBalance, amount)
	a.Balance = money.Add(a.Balance, amount)
	a.UpdatedAt = time.Now()
	a.Version++

	// 发布解冻事件
	a.AddEvent(&AccountUnfrozenEvent{
		AccountID:           a.ID,
		TransactionID:       transactionID,
		Amount:              amount,
		BalanceBefore:       balanceBefore,
		BalanceAfter:        a.Balance,
		FrozenBalanceBefore: frozenBefore,
		FrozenBalanceAfter:  a.FrozenBalance,
		Description:         description,
		UnfrozenAt:          time.Now(),
	})

	return nil
}

// Refund 退款(增加可用余额)
func (a *Account) Refund(amount money.Decimal, transactionID string, orderID *string, description string) error {
	// 验证金额
	if amount.LessThanOrEqual(money.Zero) {
		return errors.NewInvalidParam("refund amount must be positive")
	}

	// 记录交易前余额
	balanceBefore := a.Balance

	// 更新余额
	a.Balance = money.Add(a.Balance, amount)
	a.UpdatedAt = time.Now()
	a.Version++

	// 发布退款事件
	a.AddEvent(&AccountRefundedEvent{
		AccountID:     a.ID,
		TransactionID: transactionID,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  a.Balance,
		OrderID:       orderID,
		Description:   description,
		RefundedAt:    time.Now(),
	})

	return nil
}

// Adjust 调整余额(人工调账，可正可负)
func (a *Account) Adjust(amount money.Decimal, transactionID string, reason string) error {
	// 验证交易ID
	if transactionID == "" {
		return errors.NewInvalidParam("transaction_id cannot be empty")
	}

	// 验证原因
	if reason == "" {
		return errors.NewInvalidParam("adjust reason cannot be empty")
	}

	// 记录交易前余额
	balanceBefore := a.Balance

	// 更新余额
	a.Balance = money.Add(a.Balance, amount)

	// 确保余额不为负
	if a.Balance.LessThan(money.Zero) {
		return errors.New(errors.CodeInvalidParam, "adjustment would result in negative balance")
	}

	a.UpdatedAt = time.Now()
	a.Version++

	// 发布调整事件
	a.AddEvent(&AccountAdjustedEvent{
		AccountID:     a.ID,
		TransactionID: transactionID,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  a.Balance,
		Reason:        reason,
		AdjustedAt:    time.Now(),
	})

	return nil
}

// TotalBalance 总余额(可用余额 + 冻结余额)
func (a *Account) TotalBalance() money.Decimal {
	return money.Add(a.Balance, a.FrozenBalance)
}

// HasSufficientBalance 检查是否有足够的可用余额
func (a *Account) HasSufficientBalance(amount money.Decimal) bool {
	return a.Balance.GreaterThanOrEqual(amount)
}

// 领域事件管理

// AddEvent 添加领域事件
func (a *Account) AddEvent(event interface{}) {
	a.events = append(a.events, event)
}

// GetEvents 获取所有领域事件
func (a *Account) GetEvents() []interface{} {
	return a.events
}

// ClearEvents 清空领域事件(发布后调用)
func (a *Account) ClearEvents() {
	a.events = make([]interface{}, 0)
}

// HasEvents 是否有未发布的领域事件
func (a *Account) HasEvents() bool {
	return len(a.events) > 0
}
