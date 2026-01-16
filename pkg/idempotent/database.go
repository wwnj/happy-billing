package idempotent

import (
	"context"
	"time"
)

// DatabaseIdempotencyChecker 数据库幂等性检查器
// 依赖数据库唯一索引保证幂等性（如transactions表的transaction_id唯一索引）
type DatabaseIdempotencyChecker struct {
	storage IdempotencyStorage
}

// IdempotencyStorage 幂等性存储接口
// 由具体的仓储实现（如PostgreSQL TransactionRepository）
type IdempotencyStorage interface {
	// Exists 检查记录是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Save 保存记录（依赖唯一索引约束）
	Save(ctx context.Context, key string, ttl time.Duration) error

	// Delete 删除记录
	Delete(ctx context.Context, key string) error
}

// NewDatabaseIdempotencyChecker 创建数据库幂等性检查器
func NewDatabaseIdempotencyChecker(storage IdempotencyStorage) *DatabaseIdempotencyChecker {
	return &DatabaseIdempotencyChecker{
		storage: storage,
	}
}

// Check 检查key是否已存在
func (d *DatabaseIdempotencyChecker) Check(ctx context.Context, key string) (bool, error) {
	return d.storage.Exists(ctx, key)
}

// Mark 标记key已处理
func (d *DatabaseIdempotencyChecker) Mark(ctx context.Context, key string, ttl time.Duration) error {
	return d.storage.Save(ctx, key, ttl)
}

// Delete 删除key
func (d *DatabaseIdempotencyChecker) Delete(ctx context.Context, key string) error {
	return d.storage.Delete(ctx, key)
}

// CheckAndMark 原子性检查并标记
// 数据库版本通过唯一索引约束保证原子性
// 如果插入失败（违反唯一约束），表示已存在
func (d *DatabaseIdempotencyChecker) CheckAndMark(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	err := d.storage.Save(ctx, key, ttl)
	if err != nil {
		// 如果是唯一约束违反错误，返回false（已存在）
		// 具体判断逻辑由storage实现层处理
		exists, checkErr := d.storage.Exists(ctx, key)
		if checkErr != nil {
			return false, checkErr
		}
		if exists {
			// 已存在，返回false表示非首次请求
			return false, nil
		}
		// 其他错误，返回错误
		return false, err
	}

	// 保存成功，返回true表示首次请求
	return true, nil
}
