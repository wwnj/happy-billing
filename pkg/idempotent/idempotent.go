package idempotent

import (
	"context"
	"time"

	"github.com/wwnj/happy-billing/pkg/errors"
)

// IdempotencyChecker 幂等性检查器接口
type IdempotencyChecker interface {
	// Check 检查key是否已存在（已处理过）
	// 返回true表示已处理，false表示未处理
	Check(ctx context.Context, key string) (bool, error)

	// Mark 标记key已处理
	// ttl: 过期时间（0表示永不过期，但不推荐）
	Mark(ctx context.Context, key string, ttl time.Duration) error

	// Delete 删除key（用于异常回滚）
	Delete(ctx context.Context, key string) error

	// CheckAndMark 原子性检查并标记（如果后端支持）
	// 返回true表示成功标记（首次处理），false表示已存在（重复请求）
	CheckAndMark(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

// ExecuteFunc 幂等性执行函数类型
type ExecuteFunc func(ctx context.Context) error

// Executor 幂等性执行器
// 封装了检查-执行-标记的完整流程
type Executor struct {
	checker IdempotencyChecker
}

// NewExecutor 创建幂等性执行器
func NewExecutor(checker IdempotencyChecker) *Executor {
	return &Executor{
		checker: checker,
	}
}

// Execute 幂等性执行
// 1. 检查key是否已处理
// 2. 如果已处理，直接返回nil（幂等）
// 3. 如果未处理，执行fn
// 4. 执行成功后标记key已处理
// 5. 执行失败时不标记（下次可重试）
func (e *Executor) Execute(ctx context.Context, key string, ttl time.Duration, fn ExecuteFunc) error {
	// 1. 检查是否已处理
	processed, err := e.checker.Check(ctx, key)
	if err != nil {
		return errors.Wrap(errors.CodeInternalError, "idempotency check failed", err)
	}

	// 2. 已处理，直接返回（幂等）
	if processed {
		return nil
	}

	// 3. 执行业务逻辑
	if err := fn(ctx); err != nil {
		return err
	}

	// 4. 标记已处理
	if err := e.checker.Mark(ctx, key, ttl); err != nil {
		// 标记失败只记录日志，不影响业务结果
		// 下次请求会重新执行，但如果是数据库操作，依赖唯一索引保证幂等性
		return errors.Wrap(errors.CodeInternalError, "idempotency mark failed", err)
	}

	return nil
}

// ExecuteWithRollback 幂等性执行（带回滚）
// 如果执行失败，自动删除标记（允许重试）
func (e *Executor) ExecuteWithRollback(ctx context.Context, key string, ttl time.Duration, fn ExecuteFunc) error {
	// 1. 原子性检查并标记
	isFirst, err := e.checker.CheckAndMark(ctx, key, ttl)
	if err != nil {
		return errors.Wrap(errors.CodeInternalError, "idempotency check and mark failed", err)
	}

	// 2. 非首次请求，直接返回（幂等）
	if !isFirst {
		return nil
	}

	// 3. 执行业务逻辑
	if err := fn(ctx); err != nil {
		// 执行失败，删除标记（允许重试）
		_ = e.checker.Delete(ctx, key)
		return err
	}

	return nil
}

// IdempotencyKey 幂等性key生成器
type IdempotencyKey struct {
	prefix string
}

// NewIdempotencyKey 创建key生成器
func NewIdempotencyKey(prefix string) *IdempotencyKey {
	return &IdempotencyKey{
		prefix: prefix,
	}
}

// Generate 生成幂等性key
// 格式: {prefix}:{id}
func (k *IdempotencyKey) Generate(id string) string {
	if k.prefix == "" {
		return id
	}
	return k.prefix + ":" + id
}

// GenerateWithType 生成带类型的幂等性key
// 格式: {prefix}:{type}:{id}
func (k *IdempotencyKey) GenerateWithType(typ string, id string) string {
	if k.prefix == "" {
		return typ + ":" + id
	}
	return k.prefix + ":" + typ + ":" + id
}
