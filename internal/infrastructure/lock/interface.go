package lock

import (
	"context"
	"time"
)

// DistributedLock 分布式锁接口
type DistributedLock interface {
	// Acquire 获取锁
	// key: 锁的唯一标识
	// ttl: 锁的过期时间
	// 返回: 锁对象和错误
	Acquire(ctx context.Context, key string, ttl time.Duration) (Lock, error)

	// Release 释放锁
	// lock: 要释放的锁对象
	// 返回: 错误
	Release(ctx context.Context, lock Lock) error

	// TryAcquire 尝试获取锁(非阻塞)
	// key: 锁的唯一标识
	// ttl: 锁的过期时间
	// 返回: 锁对象(如果获取失败返回nil)和错误
	TryAcquire(ctx context.Context, key string, ttl time.Duration) (Lock, error)

	// AcquireWithRetry 获取锁并重试
	// key: 锁的唯一标识
	// ttl: 锁的过期时间
	// maxRetries: 最大重试次数
	// retryInterval: 重试间隔
	// 返回: 锁对象和错误
	AcquireWithRetry(ctx context.Context, key string, ttl time.Duration, maxRetries int, retryInterval time.Duration) (Lock, error)
}

// Lock 锁对象接口
type Lock interface {
	// Key 返回锁的key
	Key() string

	// Token 返回锁的token(用于验证锁的持有者)
	Token() string

	// Extend 延长锁的过期时间
	// ttl: 新的过期时间
	// 返回: 错误
	Extend(ctx context.Context, ttl time.Duration) error

	// IsHeld 检查锁是否仍被持有
	// 返回: true表示锁仍有效
	IsHeld(ctx context.Context) bool
}
