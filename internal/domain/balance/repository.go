package balance

import (
	"context"
	"time"

	"github.com/wwnj/happy-billing/pkg/money"
)

// AccountRepository 账户仓储接口
type AccountRepository interface {
	// Save 保存账户(新建)
	Save(ctx context.Context, account *Account) error

	// Update 更新账户(带乐观锁)
	// 返回错误如果version不匹配(并发冲突)
	Update(ctx context.Context, account *Account) error

	// FindByID 根据ID查询账户
	FindByID(ctx context.Context, id string) (*Account, error)

	// FindByUserID 根据用户ID查询账户
	FindByUserID(ctx context.Context, tenantID, userID string) (*Account, error)

	// FindByTenant 查询租户下的所有账户
	FindByTenant(ctx context.Context, tenantID string, offset, limit int) ([]*Account, int64, error)

	// Lock 锁定账户(用于分布式事务)
	// 返回带锁的账户，使用完毕后需要调用Unlock
	Lock(ctx context.Context, accountID string) (*Account, error)

	// Unlock 解锁账户
	Unlock(ctx context.Context, accountID string) error

	// SumBalanceByTenant 统计租户的总余额
	SumBalanceByTenant(ctx context.Context, tenantID string) (money.Decimal, error)
}

// TransactionRepository 交易记录仓储接口
type TransactionRepository interface {
	// Save 保存交易记录
	Save(ctx context.Context, transaction *Transaction) error

	// FindByID 根据ID查询交易记录
	FindByID(ctx context.Context, id string) (*Transaction, error)

	// FindByTransactionID 根据业务交易ID查询(用于幂等性检查)
	FindByTransactionID(ctx context.Context, transactionID string) (*Transaction, error)

	// Exists 检查交易ID是否已存在(幂等性检查)
	Exists(ctx context.Context, transactionID string) (bool, error)

	// ListByAccount 查询账户的交易记录
	// accountID: 账户ID
	// transactionType: 交易类型(为空则查询所有类型)
	// startTime, endTime: 时间范围(为空则不限制)
	// offset, limit: 分页参数
	ListByAccount(ctx context.Context, accountID string, transactionType TransactionType, startTime, endTime *time.Time, offset, limit int) ([]*Transaction, int64, error)

	// ListByOrder 查询订单关联的交易记录
	ListByOrder(ctx context.Context, orderID string) ([]*Transaction, error)

	// SumAmountByType 统计指定类型的交易金额
	// accountID: 账户ID
	// transactionType: 交易类型
	// startTime, endTime: 时间范围
	SumAmountByType(ctx context.Context, accountID string, transactionType TransactionType, startTime, endTime time.Time) (money.Decimal, error)
}
