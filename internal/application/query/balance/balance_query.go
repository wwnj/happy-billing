package query

import (
	"context"
	"time"

	"github.com/wwnj/happy-billing/internal/domain/balance"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// BalanceQuery 余额查询服务
type BalanceQuery struct {
	accountRepo balance.AccountRepository
	txRepo      balance.TransactionRepository
}

// NewBalanceQuery 创建余额查询服务
func NewBalanceQuery(
	accountRepo balance.AccountRepository,
	txRepo balance.TransactionRepository,
) *BalanceQuery {
	return &BalanceQuery{
		accountRepo: accountRepo,
		txRepo:      txRepo,
	}
}

// AccountBalanceDTO 账户余额DTO
type AccountBalanceDTO struct {
	AccountID     string        `json:"account_id"`
	TenantID      string        `json:"tenant_id"`
	UserID        string        `json:"user_id"`
	Balance       money.Decimal `json:"balance"`        // 可用余额
	FrozenBalance money.Decimal `json:"frozen_balance"` // 冻结余额
	TotalBalance  money.Decimal `json:"total_balance"`  // 总余额
	Currency      string        `json:"currency"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// TransactionDTO 交易记录DTO
type TransactionDTO struct {
	ID            string                  `json:"id"`
	TransactionID string                  `json:"transaction_id"`
	AccountID     string                  `json:"account_id"`
	Type          balance.TransactionType `json:"type"`
	Amount        money.Decimal           `json:"amount"`
	BalanceBefore money.Decimal           `json:"balance_before"`
	BalanceAfter  money.Decimal           `json:"balance_after"`
	OrderID       *string                 `json:"order_id,omitempty"`
	Description   string                  `json:"description"`
	CreatedAt     time.Time               `json:"created_at"`
}

// GetAccountBalance 查询账户余额
func (q *BalanceQuery) GetAccountBalance(ctx context.Context, accountID string) (*AccountBalanceDTO, error) {
	account, err := q.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return &AccountBalanceDTO{
		AccountID:     account.ID,
		TenantID:      account.TenantID,
		UserID:        account.UserID,
		Balance:       account.Balance,
		FrozenBalance: account.FrozenBalance,
		TotalBalance:  account.TotalBalance(),
		Currency:      account.Currency,
		UpdatedAt:     account.UpdatedAt,
	}, nil
}

// GetAccountBalanceByUserID 根据用户ID查询账户余额
func (q *BalanceQuery) GetAccountBalanceByUserID(ctx context.Context, tenantID, userID string) (*AccountBalanceDTO, error) {
	account, err := q.accountRepo.FindByUserID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	return &AccountBalanceDTO{
		AccountID:     account.ID,
		TenantID:      account.TenantID,
		UserID:        account.UserID,
		Balance:       account.Balance,
		FrozenBalance: account.FrozenBalance,
		TotalBalance:  account.TotalBalance(),
		Currency:      account.Currency,
		UpdatedAt:     account.UpdatedAt,
	}, nil
}

// ListTransactionsQuery 交易记录列表查询参数
type ListTransactionsQuery struct {
	AccountID       string                  // 账户ID
	TransactionType balance.TransactionType // 交易类型（为空则查询所有）
	StartTime       *time.Time              // 开始时间
	EndTime         *time.Time              // 结束时间
	Offset          int                     // 偏移量
	Limit           int                     // 每页数量
}

// ListTransactionsResult 交易记录列表查询结果
type ListTransactionsResult struct {
	Transactions []*TransactionDTO `json:"transactions"`
	Total        int64             `json:"total"`
	Offset       int               `json:"offset"`
	Limit        int               `json:"limit"`
}

// ListTransactions 查询交易记录列表
func (q *BalanceQuery) ListTransactions(ctx context.Context, query ListTransactionsQuery) (*ListTransactionsResult, error) {
	// 参数验证
	if query.AccountID == "" {
		return nil, errors.NewInvalidParam("account_id cannot be empty")
	}
	if query.Limit <= 0 {
		query.Limit = 20 // 默认每页20条
	}
	if query.Limit > 100 {
		query.Limit = 100 // 最大每页100条
	}

	// 查询交易记录
	transactions, total, err := q.txRepo.ListByAccount(
		ctx,
		query.AccountID,
		query.TransactionType,
		query.StartTime,
		query.EndTime,
		query.Offset,
		query.Limit,
	)
	if err != nil {
		return nil, err
	}

	// 转换为DTO
	dtos := make([]*TransactionDTO, 0, len(transactions))
	for _, tx := range transactions {
		dtos = append(dtos, &TransactionDTO{
			ID:            tx.ID,
			TransactionID: tx.TransactionID,
			AccountID:     tx.AccountID,
			Type:          tx.Type,
			Amount:        tx.Amount,
			BalanceBefore: tx.BalanceBefore,
			BalanceAfter:  tx.BalanceAfter,
			OrderID:       tx.OrderID,
			Description:   tx.Description,
			CreatedAt:     tx.CreatedAt,
		})
	}

	return &ListTransactionsResult{
		Transactions: dtos,
		Total:        total,
		Offset:       query.Offset,
		Limit:        query.Limit,
	}, nil
}

// GetTransaction 查询单个交易记录
func (q *BalanceQuery) GetTransaction(ctx context.Context, transactionID string) (*TransactionDTO, error) {
	tx, err := q.txRepo.FindByTransactionID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	return &TransactionDTO{
		ID:            tx.ID,
		TransactionID: tx.TransactionID,
		AccountID:     tx.AccountID,
		Type:          tx.Type,
		Amount:        tx.Amount,
		BalanceBefore: tx.BalanceBefore,
		BalanceAfter:  tx.BalanceAfter,
		OrderID:       tx.OrderID,
		Description:   tx.Description,
		CreatedAt:     tx.CreatedAt,
	}, nil
}

// SumTransactionAmount 统计交易金额
func (q *BalanceQuery) SumTransactionAmount(
	ctx context.Context,
	accountID string,
	transactionType balance.TransactionType,
	startTime, endTime time.Time,
) (money.Decimal, error) {
	return q.txRepo.SumAmountByType(ctx, accountID, transactionType, startTime, endTime)
}
