package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/wwnj/happy-billing/internal/domain/balance"
	"github.com/wwnj/happy-billing/internal/infrastructure/persistence/postgres/model"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// TransactionRepository 交易记录仓储PostgreSQL实现
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository 创建交易记录仓储
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

// Save 保存交易记录
func (r *TransactionRepository) Save(ctx context.Context, transaction *balance.Transaction) error {
	do := model.FromDomainTransaction(transaction)

	if err := r.db.WithContext(ctx).Create(do).Error; err != nil {
		return errors.NewDatabaseError("save transaction", err)
	}

	return nil
}

// FindByID 根据ID查询交易记录
func (r *TransactionRepository) FindByID(ctx context.Context, id string) (*balance.Transaction, error) {
	var do model.TransactionDO

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&do).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("transaction")
		}
		return nil, errors.NewDatabaseError("find transaction by id", err)
	}

	return do.ToDomain()
}

// FindByTransactionID 根据业务交易ID查询（用于幂等性检查）
func (r *TransactionRepository) FindByTransactionID(ctx context.Context, transactionID string) (*balance.Transaction, error) {
	var do model.TransactionDO

	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&do).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("transaction")
		}
		return nil, errors.NewDatabaseError("find transaction by transaction_id", err)
	}

	return do.ToDomain()
}

// Exists 检查交易ID是否已存在（幂等性检查）
func (r *TransactionRepository) Exists(ctx context.Context, transactionID string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&model.TransactionDO{}).
		Where("transaction_id = ?", transactionID).
		Count(&count).Error; err != nil {
		return false, errors.NewDatabaseError("check transaction exists", err)
	}

	return count > 0, nil
}

// ListByAccount 查询账户的交易记录
func (r *TransactionRepository) ListByAccount(
	ctx context.Context,
	accountID string,
	transactionType balance.TransactionType,
	startTime, endTime *time.Time,
	offset, limit int,
) ([]*balance.Transaction, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.TransactionDO{}).Where("account_id = ?", accountID)

	// 按类型过滤
	if transactionType != "" {
		query = query.Where("type = ?", string(transactionType))
	}

	// 按时间范围过滤
	if startTime != nil {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", endTime)
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("count transactions", err)
	}

	// 查询分页数据
	var dos []model.TransactionDO
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&dos).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("list transactions", err)
	}

	// 转换为领域对象
	transactions := make([]*balance.Transaction, 0, len(dos))
	for _, do := range dos {
		transaction, err := do.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, total, nil
}

// ListByOrder 查询订单关联的交易记录
func (r *TransactionRepository) ListByOrder(ctx context.Context, orderID string) ([]*balance.Transaction, error) {
	var dos []model.TransactionDO

	if err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&dos).Error; err != nil {
		return nil, errors.NewDatabaseError("list transactions by order", err)
	}

	// 转换为领域对象
	transactions := make([]*balance.Transaction, 0, len(dos))
	for _, do := range dos {
		transaction, err := do.ToDomain()
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// SumAmountByType 统计指定类型的交易金额
func (r *TransactionRepository) SumAmountByType(
	ctx context.Context,
	accountID string,
	transactionType balance.TransactionType,
	startTime, endTime time.Time,
) (money.Decimal, error) {
	var result struct {
		TotalAmount string
	}

	// 使用原始SQL进行聚合查询
	query := r.db.WithContext(ctx).Model(&model.TransactionDO{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL)), 0) as total_amount").
		Where("account_id = ?", accountID)

	if transactionType != "" {
		query = query.Where("type = ?", string(transactionType))
	}

	query = query.Where("created_at >= ? AND created_at <= ?", startTime, endTime)

	if err := query.Scan(&result).Error; err != nil {
		return money.Zero, errors.NewDatabaseError("sum transaction amount", err)
	}

	total, err := money.NewFromString(result.TotalAmount)
	if err != nil {
		return money.Zero, errors.Wrap(errors.CodeInternalError, "parse total amount", err)
	}

	return total, nil
}
