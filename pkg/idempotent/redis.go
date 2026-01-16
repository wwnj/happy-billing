package idempotent

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wwnj/happy-billing/pkg/errors"
)

// RedisIdempotencyChecker Redis幂等性检查器
type RedisIdempotencyChecker struct {
	client *redis.Client
}

// NewRedisIdempotencyChecker 创建Redis幂等性检查器
func NewRedisIdempotencyChecker(client *redis.Client) *RedisIdempotencyChecker {
	return &RedisIdempotencyChecker{
		client: client,
	}
}

// Check 检查key是否已存在
func (r *RedisIdempotencyChecker) Check(ctx context.Context, key string) (bool, error) {
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, errors.NewCacheError("check idempotency key", err)
	}

	return exists > 0, nil
}

// Mark 标记key已处理
func (r *RedisIdempotencyChecker) Mark(ctx context.Context, key string, ttl time.Duration) error {
	if ttl == 0 {
		// 永不过期（不推荐，但支持）
		if err := r.client.Set(ctx, key, "1", 0).Err(); err != nil {
			return errors.NewCacheError("mark idempotency key", err)
		}
	} else {
		// 带过期时间
		if err := r.client.SetEx(ctx, key, "1", ttl).Err(); err != nil {
			return errors.NewCacheError("mark idempotency key", err)
		}
	}

	return nil
}

// Delete 删除key
func (r *RedisIdempotencyChecker) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return errors.NewCacheError("delete idempotency key", err)
	}

	return nil
}

// CheckAndMark 原子性检查并标记
// 使用SET NX命令实现原子操作
// 返回true表示成功标记（首次处理），false表示已存在（重复请求）
func (r *RedisIdempotencyChecker) CheckAndMark(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	success, err := r.client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return false, errors.NewCacheError("check and mark idempotency key", err)
	}

	return success, nil
}
