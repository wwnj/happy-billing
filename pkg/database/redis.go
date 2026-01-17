package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/wwnj/happy-billing/pkg/config"
	"github.com/wwnj/happy-billing/pkg/tracing"
)

var redisClient *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis(cfg *config.RedisConfig) error {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	// 安装追踪 Hook
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	tracing.InstallRedisHook(client, addr)

	redisClient = client
	return nil
}

// GetRedis 获取 Redis 客户端
func GetRedis() *redis.Client {
	return redisClient
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}
