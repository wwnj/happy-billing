package idempotent

/*
幂等性工具使用示例

## 1. 基本使用（Redis版本）

```go
import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wwnj/happy-billing/pkg/idempotent"
)

// 初始化Redis客户端
redisClient := redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

// 创建Redis幂等性检查器
checker := idempotent.NewRedisIdempotencyChecker(redisClient)

// 创建执行器
executor := idempotent.NewExecutor(checker)

// 创建key生成器
keyGen := idempotent.NewIdempotencyKey("transaction")

// 执行幂等性操作
transactionID := "tx_20250116_001"
key := keyGen.Generate(transactionID)

err := executor.Execute(
	context.Background(),
	key,
	24*time.Hour, // key有效期24小时
	func(ctx context.Context) error {
		// 实际业务逻辑：扣费操作
		return deductBalance(ctx, accountID, amount, transactionID)
	},
)
```

## 2. 余额扣费场景（完整示例）

```go
// internal/application/command/balance/deduct_service.go

type DeductService struct {
	accountRepo balance.AccountRepository
	txRepo      balance.TransactionRepository
	lockClient  lock.DistributedLock
	executor    *idempotent.Executor
	keyGen      *idempotent.IdempotencyKey
}

func NewDeductService(
	accountRepo balance.AccountRepository,
	txRepo balance.TransactionRepository,
	lockClient lock.DistributedLock,
	redisClient *redis.Client,
) *DeductService {
	// 创建Redis幂等性检查器
	checker := idempotent.NewRedisIdempotencyChecker(redisClient)
	executor := idempotent.NewExecutor(checker)
	keyGen := idempotent.NewIdempotencyKey("deduct")

	return &DeductService{
		accountRepo: accountRepo,
		txRepo:      txRepo,
		lockClient:  lockClient,
		executor:    executor,
		keyGen:      keyGen,
	}
}

func (s *DeductService) Execute(ctx context.Context, cmd DeductCommand) error {
	// 生成幂等性key
	idempotencyKey := s.keyGen.Generate(cmd.TransactionID)

	// 幂等性执行
	return s.executor.Execute(
		ctx,
		idempotencyKey,
		24*time.Hour, // Redis key有效期24小时
		func(ctx context.Context) error {
			// 1. 获取分布式锁
			lock, err := s.lockClient.Acquire(ctx, "account:"+cmd.AccountID, 10*time.Second)
			if err != nil {
				return err
			}
			defer s.lockClient.Release(ctx, lock)

			// 2. 数据库幂等性检查（双重保证）
			if exists, err := s.txRepo.Exists(ctx, cmd.TransactionID); err != nil {
				return err
			} else if exists {
				return nil // 已处理
			}

			// 3. 加载聚合根
			account, err := s.accountRepo.FindByID(ctx, cmd.AccountID)
			if err != nil {
				return err
			}

			// 4. 执行领域逻辑
			if err := account.Deduct(cmd.Amount, cmd.TransactionID, cmd.OrderID, cmd.Description); err != nil {
				return err
			}

			// 5. 持久化（乐观锁）
			if err := s.accountRepo.Update(ctx, account); err != nil {
				return err
			}

			// 6. 发布领域事件
			// ...

			return nil
		},
	)
}
```

## 3. 带回滚的幂等性执行

```go
// 如果业务逻辑执行失败，自动删除Redis标记，允许重试
err := executor.ExecuteWithRollback(
	context.Background(),
	key,
	24*time.Hour,
	func(ctx context.Context) error {
		// 业务逻辑
		if err := doSomething(); err != nil {
			// 执行失败，会自动删除Redis标记
			return err
		}
		return nil
	},
)
```

## 4. 数据库版本（持久化保证）

```go
// 基于TransactionRepository实现IdempotencyStorage接口
type TransactionIdempotencyStorage struct {
	repo balance.TransactionRepository
}

func (s *TransactionIdempotencyStorage) Exists(ctx context.Context, key string) (bool, error) {
	return s.repo.Exists(ctx, key)
}

func (s *TransactionIdempotencyStorage) Save(ctx context.Context, key string, ttl time.Duration) error {
	// 创建一个临时交易记录（仅用于幂等性标记）
	// 实际业务中，交易记录由业务逻辑创建
	return s.repo.Save(ctx, &Transaction{
		TransactionID: key,
		// ... 其他字段
	})
}

// 使用数据库版本
storage := &TransactionIdempotencyStorage{repo: transactionRepo}
checker := idempotent.NewDatabaseIdempotencyChecker(storage)
executor := idempotent.NewExecutor(checker)
```

## 5. Redis vs 数据库对比

| 特性 | Redis版本 | 数据库版本 |
|------|----------|-----------|
| **性能** | 极快（内存操作） | 较慢（磁盘IO） |
| **持久化** | 可能丢失（取决于持久化配置） | 完全持久化 |
| **适用场景** | 高并发场景，短期幂等性保证 | 长期幂等性保证，严格一致性要求 |
| **TTL支持** | 原生支持 | 需要额外清理机制 |
| **原子性** | SET NX原生支持 | 依赖唯一索引约束 |

## 6. 推荐实践

### 双重幂等性保证（Redis + 数据库）

```go
// 1. 先用Redis快速过滤（性能优先）
idempotencyKey := keyGen.Generate(transactionID)
err := executor.Execute(ctx, idempotencyKey, 24*time.Hour, func(ctx context.Context) error {
	// 2. 数据库唯一索引保证最终一致性（可靠性优先）
	if exists, err := txRepo.Exists(ctx, transactionID); err != nil {
		return err
	} else if exists {
		return nil
	}

	// 3. 执行业务逻辑...
	return nil
})
```

### Key命名规范

```go
// 业务前缀
keyGen := idempotent.NewIdempotencyKey("billing")

// 不同类型操作使用不同key
deductKey := keyGen.GenerateWithType("deduct", transactionID)    // billing:deduct:tx_001
chargeKey := keyGen.GenerateWithType("charge", transactionID)    // billing:charge:tx_002
refundKey := keyGen.GenerateWithType("refund", transactionID)    // billing:refund:tx_003
```

### TTL设置建议

- **短周期操作**（秒级/分钟级）：1小时
- **日常业务**（订单/支付）：24小时
- **月度账单**：30天
- **年度报表**：365天

### 错误处理

```go
err := executor.Execute(ctx, key, ttl, fn)
if err != nil {
	// 区分幂等性错误和业务错误
	if errors.IsCode(err, errors.CodeInternal) {
		// 幂等性基础设施错误（Redis故障等）
		// 可以降级到仅依赖数据库唯一索引
		log.Error("idempotency infrastructure error", zap.Error(err))
	} else {
		// 业务逻辑错误
		return err
	}
}
```

## 7. 注意事项

1. **Redis故障降级**：Redis故障时，依赖数据库唯一索引保证幂等性
2. **Key冲突**：使用业务前缀和类型避免不同业务的key冲突
3. **TTL合理性**：根据业务特点设置合理的TTL，避免无限期存储
4. **监控告警**：监控幂等性检查失败率，及时发现问题
5. **压力测试**：验证高并发场景下的幂等性保证

*/
