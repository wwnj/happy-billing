package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/wwnj/happy-billing/pkg/errors"
)

// RedisLock Redis分布式锁实现
type RedisLock struct {
	client *redis.Client
}

// NewRedisLock 创建Redis分布式锁
func NewRedisLock(client *redis.Client) *RedisLock {
	return &RedisLock{
		client: client,
	}
}

// Acquire 获取锁(阻塞直到获取成功或超时)
func (r *RedisLock) Acquire(ctx context.Context, key string, ttl time.Duration) (Lock, error) {
	// 生成唯一token
	token := uuid.New().String()

	// 使用SET NX命令获取锁
	lockKey := fmt.Sprintf("lock:%s", key)

	success, err := r.client.SetNX(ctx, lockKey, token, ttl).Result()
	if err != nil {
		return nil, errors.NewCacheError("acquire lock", err)
	}

	if !success {
		return nil, errors.New(errors.CodeLockAcquireFailed, "failed to acquire lock: already held by another process")
	}

	return &redisLockInstance{
		client: r.client,
		key:    lockKey,
		token:  token,
	}, nil
}

// Release 释放锁
func (r *RedisLock) Release(ctx context.Context, lock Lock) error {
	if lock == nil {
		return errors.NewInvalidParam("lock cannot be nil")
	}

	rl, ok := lock.(*redisLockInstance)
	if !ok {
		return errors.NewInvalidParam("invalid lock type")
	}

	// 使用Lua脚本确保只删除自己持有的锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := r.client.Eval(ctx, script, []string{rl.key}, rl.token).Result()
	if err != nil {
		return errors.NewCacheError("release lock", err)
	}

	if result.(int64) == 0 {
		return errors.New(errors.CodeLockReleaseFailed, "failed to release lock: not held by this process")
	}

	return nil
}

// TryAcquire 尝试获取锁(非阻塞)
func (r *RedisLock) TryAcquire(ctx context.Context, key string, ttl time.Duration) (Lock, error) {
	return r.Acquire(ctx, key, ttl)
}

// AcquireWithRetry 获取锁并重试
func (r *RedisLock) AcquireWithRetry(ctx context.Context, key string, ttl time.Duration, maxRetries int, retryInterval time.Duration) (Lock, error) {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		lock, err := r.TryAcquire(ctx, key, ttl)
		if err == nil {
			return lock, nil
		}

		lastErr = err

		// 最后一次重试失败后不再等待
		if i < maxRetries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryInterval):
				// 继续重试
			}
		}
	}

	return nil, errors.Wrap(errors.CodeLockAcquireFailed,
		fmt.Sprintf("failed to acquire lock after %d retries", maxRetries),
		lastErr)
}

// redisLockInstance Redis锁实例
type redisLockInstance struct {
	client *redis.Client
	key    string
	token  string
}

// Key 返回锁的key
func (r *redisLockInstance) Key() string {
	return r.key
}

// Token 返回锁的token
func (r *redisLockInstance) Token() string {
	return r.token
}

// Extend 延长锁的过期时间
func (r *redisLockInstance) Extend(ctx context.Context, ttl time.Duration) error {
	// 使用Lua脚本确保只延长自己持有的锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := r.client.Eval(ctx, script, []string{r.key}, r.token, int(ttl.Seconds())).Result()
	if err != nil {
		return errors.NewCacheError("extend lock", err)
	}

	if result.(int64) == 0 {
		return errors.New(errors.CodeLockReleaseFailed, "failed to extend lock: not held by this process")
	}

	return nil
}

// IsHeld 检查锁是否仍被持有
func (r *redisLockInstance) IsHeld(ctx context.Context) bool {
	value, err := r.client.Get(ctx, r.key).Result()
	if err != nil {
		return false
	}

	return value == r.token
}
