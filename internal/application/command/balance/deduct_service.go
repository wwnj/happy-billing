package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/wwnj/happy-billing/internal/domain/balance"
	"github.com/wwnj/happy-billing/internal/infrastructure/lock"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/idempotent"
	"github.com/wwnj/happy-billing/pkg/money"
)

// DeductCommand 扣费命令
type DeductCommand struct {
	AccountID     string        // 账户ID
	Amount        money.Decimal // 扣费金额
	TransactionID string        // 交易ID（幂等性key）
	OrderID       *string       // 关联订单ID（可选）
	Description   string        // 描述
}

// DeductService 扣费服务
type DeductService struct {
	accountRepo balance.AccountRepository
	txRepo      balance.TransactionRepository
	lockClient  lock.DistributedLock
	executor    *idempotent.Executor
	keyGen      *idempotent.IdempotencyKey
}

// NewDeductService 创建扣费服务
func NewDeductService(
	accountRepo balance.AccountRepository,
	txRepo balance.TransactionRepository,
	lockClient lock.DistributedLock,
	idempotencyChecker idempotent.IdempotencyChecker,
) *DeductService {
	executor := idempotent.NewExecutor(idempotencyChecker)
	keyGen := idempotent.NewIdempotencyKey("deduct")

	return &DeductService{
		accountRepo: accountRepo,
		txRepo:      txRepo,
		lockClient:  lockClient,
		executor:    executor,
		keyGen:      keyGen,
	}
}

// Execute 执行扣费
// 高并发扣费的三重保证：
// 1. Redis分布式锁 - 防止并发竞争
// 2. 数据库乐观锁(version) - 防止并发更新冲突
// 3. 幂等性(transaction_id) - 防止重复处理
func (s *DeductService) Execute(ctx context.Context, cmd DeductCommand) error {
	// 参数验证
	if cmd.AccountID == "" {
		return errors.NewInvalidParam("account_id cannot be empty")
	}
	if cmd.TransactionID == "" {
		return errors.NewInvalidParam("transaction_id cannot be empty")
	}
	if cmd.Amount.LessThanOrEqual(money.Zero) {
		return errors.NewInvalidParam("deduct amount must be positive")
	}

	// 生成幂等性key
	idempotencyKey := s.keyGen.Generate(cmd.TransactionID)

	// 幂等性执行
	return s.executor.Execute(
		ctx,
		idempotencyKey,
		24*time.Hour, // Redis key有效期24小时
		func(ctx context.Context) error {
			// 1. 获取分布式锁
			lockKey := "account:" + cmd.AccountID
			lockInstance, err := s.lockClient.Acquire(ctx, lockKey, 10*time.Second)
			if err != nil {
				return errors.Wrap(errors.CodeLockAcquireFailed, "acquire account lock failed", err)
			}
			defer s.lockClient.Release(ctx, lockInstance)

			// 2. 数据库幂等性检查（双重保证）
			exists, err := s.txRepo.Exists(ctx, cmd.TransactionID)
			if err != nil {
				return err
			}
			if exists {
				return nil // 已处理
			}

			// 3. 加载聚合根
			account, err := s.accountRepo.FindByID(ctx, cmd.AccountID)
			if err != nil {
				return err
			}

			// 4. 执行领域逻辑（包含余额充足性检查）
			balanceBefore := account.Balance
			if err := account.Deduct(cmd.Amount, cmd.TransactionID, cmd.OrderID, cmd.Description); err != nil {
				return err
			}

			// 5. 创建交易记录
			transaction, err := balance.NewTransaction(
				uuid.New().String(),
				cmd.TransactionID,
				cmd.AccountID,
				balance.TransactionTypeDeduct,
				cmd.Amount,
				balanceBefore,   // 扣费前余额
				account.Balance, // 扣费后余额
				cmd.OrderID,
				cmd.Description,
				nil, // 无扩展元数据
			)
			if err != nil {
				return err
			}

			// 6. 持久化（乐观锁）
			if err := s.accountRepo.Update(ctx, account); err != nil {
				return err
			}

			// 7. 保存交易记录
			if err := s.txRepo.Save(ctx, transaction); err != nil {
				return err
			}

			// 8. 发布领域事件（TODO: 事件发布机制）
			// for _, event := range account.GetEvents() {
			//     s.eventBus.Publish(ctx, "balance.deducted", event)
			// }

			return nil
		},
	)
}
