package database

import (
	"fmt"

	"github.com/wwnj/happy-billing/pkg/config"
	"github.com/wwnj/happy-billing/pkg/logger"
	"go.uber.org/zap"
)

// InitAll 初始化所有数据库连接
func InitAll(cfg *config.Config) error {
	// 初始化 MySQL
	if err := InitMySQL(&cfg.MySQL); err != nil {
		return fmt.Errorf("failed to init mysql: %w", err)
	}
	logger.Info("MySQL connected successfully")

	// 初始化 ClickHouse (可选)
	if len(cfg.ClickHouse.Addresses) > 0 {
		if err := InitClickHouse(&cfg.ClickHouse); err != nil {
			return fmt.Errorf("failed to init clickhouse: %w", err)
		}
		logger.Info("ClickHouse connected successfully")
	}

	// 初始化 Redis
	if err := InitRedis(&cfg.Redis); err != nil {
		return fmt.Errorf("failed to init redis: %w", err)
	}
	logger.Info("Redis connected successfully")

	return nil
}

// CloseAll 关闭所有数据库连接
func CloseAll() {
	if err := CloseMySQL(); err != nil {
		logger.Error("Failed to close MySQL", zap.Error(err))
	}

	if err := CloseClickHouse(); err != nil {
		logger.Error("Failed to close ClickHouse", zap.Error(err))
	}

	if err := CloseRedis(); err != nil {
		logger.Error("Failed to close Redis", zap.Error(err))
	}

	logger.Info("All database connections closed")
}
