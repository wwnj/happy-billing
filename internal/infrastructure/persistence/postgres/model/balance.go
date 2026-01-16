package model

import (
	"time"

	"github.com/wwnj/happy-billing/internal/domain/balance"
	"github.com/wwnj/happy-billing/pkg/money"
)

// AccountDO 账户数据对象（DO - Data Object）
type AccountDO struct {
	ID            string    `gorm:"column:id;primaryKey;type:varchar(64)"`
	TenantID      string    `gorm:"column:tenant_id;type:varchar(64);not null;index:idx_tenant_user"`
	UserID        string    `gorm:"column:user_id;type:varchar(64);not null;index:idx_tenant_user"`
	Balance       string    `gorm:"column:balance;type:decimal(20,4);not null;default:0"`
	FrozenBalance string    `gorm:"column:frozen_balance;type:decimal(20,4);not null;default:0"`
	Currency      string    `gorm:"column:currency;type:varchar(10);not null;default:'CNY'"`
	Version       int64     `gorm:"column:version;not null;default:0"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

// TableName 指定表名
func (AccountDO) TableName() string {
	return "accounts"
}

// ToDomain 转换为领域对象
func (do *AccountDO) ToDomain() (*balance.Account, error) {
	balanceDecimal, err := money.NewFromString(do.Balance)
	if err != nil {
		return nil, err
	}

	frozenBalanceDecimal, err := money.NewFromString(do.FrozenBalance)
	if err != nil {
		return nil, err
	}

	account := &balance.Account{
		ID:            do.ID,
		TenantID:      do.TenantID,
		UserID:        do.UserID,
		Balance:       balanceDecimal,
		FrozenBalance: frozenBalanceDecimal,
		Currency:      do.Currency,
		Version:       do.Version,
		CreatedAt:     do.CreatedAt,
		UpdatedAt:     do.UpdatedAt,
	}

	return account, nil
}

// FromDomain 从领域对象转换
func FromDomainAccount(account *balance.Account) *AccountDO {
	return &AccountDO{
		ID:            account.ID,
		TenantID:      account.TenantID,
		UserID:        account.UserID,
		Balance:       account.Balance.String(),
		FrozenBalance: account.FrozenBalance.String(),
		Currency:      account.Currency,
		Version:       account.Version,
		CreatedAt:     account.CreatedAt,
		UpdatedAt:     account.UpdatedAt,
	}
}

// TransactionDO 交易记录数据对象
type TransactionDO struct {
	ID            string    `gorm:"column:id;primaryKey;type:varchar(64)"`
	TransactionID string    `gorm:"column:transaction_id;uniqueIndex:uk_transaction_id;type:varchar(64);not null"`
	AccountID     string    `gorm:"column:account_id;index:idx_account_id;type:varchar(64);not null"`
	Type          string    `gorm:"column:type;type:varchar(20);not null"`
	Amount        string    `gorm:"column:amount;type:decimal(20,4);not null"`
	BalanceBefore string    `gorm:"column:balance_before;type:decimal(20,4);not null"`
	BalanceAfter  string    `gorm:"column:balance_after;type:decimal(20,4);not null"`
	OrderID       *string   `gorm:"column:order_id;type:varchar(64);index:idx_order_id"`
	Description   string    `gorm:"column:description;type:text"`
	MetaData      string    `gorm:"column:metadata;type:jsonb"` // PostgreSQL JSONB类型
	CreatedAt     time.Time `gorm:"column:created_at;not null;autoCreateTime;index:idx_created_at"`
}

// TableName 指定表名
func (TransactionDO) TableName() string {
	return "transactions"
}

// ToDomain 转换为领域对象
func (do *TransactionDO) ToDomain() (*balance.Transaction, error) {
	amount, err := money.NewFromString(do.Amount)
	if err != nil {
		return nil, err
	}

	balanceBefore, err := money.NewFromString(do.BalanceBefore)
	if err != nil {
		return nil, err
	}

	balanceAfter, err := money.NewFromString(do.BalanceAfter)
	if err != nil {
		return nil, err
	}

	// 解析metadata（简化处理，实际项目中可能需要更复杂的JSON解析）
	var metadata map[string]interface{}
	// TODO: 解析JSONB字段

	transaction := &balance.Transaction{
		ID:            do.ID,
		TransactionID: do.TransactionID,
		AccountID:     do.AccountID,
		Type:          balance.TransactionType(do.Type),
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		OrderID:       do.OrderID,
		Description:   do.Description,
		MetaData:      metadata,
		CreatedAt:     do.CreatedAt,
	}

	return transaction, nil
}

// FromDomainTransaction 从领域对象转换
func FromDomainTransaction(transaction *balance.Transaction) *TransactionDO {
	// TODO: 序列化metadata为JSONB
	metadataJSON := "{}"

	return &TransactionDO{
		ID:            transaction.ID,
		TransactionID: transaction.TransactionID,
		AccountID:     transaction.AccountID,
		Type:          string(transaction.Type),
		Amount:        transaction.Amount.String(),
		BalanceBefore: transaction.BalanceBefore.String(),
		BalanceAfter:  transaction.BalanceAfter.String(),
		OrderID:       transaction.OrderID,
		Description:   transaction.Description,
		MetaData:      metadataJSON,
		CreatedAt:     transaction.CreatedAt,
	}
}
