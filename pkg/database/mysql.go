package database

import (
	"fmt"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"github.com/wwnj/happy-billing/pkg/config"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var mysqlDB *gorm.DB

// InitMySQL 初始化 MySQL 连接
func InitMySQL(cfg *config.MySQLConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
	)

	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: getGormLogger(cfg.LogLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		// 禁用外键约束（由应用层保证数据一致性）
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect mysql: %w", err)
	}

	// 获取底层 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping mysql: %w", err)
	}

	// 安装 OpenTelemetry 追踪插件（使用 uptrace/otelgorm）
	if err := db.Use(otelgorm.NewPlugin(
		otelgorm.WithDBName(cfg.Database),
		otelgorm.WithAttributes(attribute.String("db.system", "mysql")),
	)); err != nil {
		return fmt.Errorf("failed to install otelgorm plugin: %w", err)
	}

	mysqlDB = db
	return nil
}

// GetMySQL 获取 MySQL 连接
func GetMySQL() *gorm.DB {
	return mysqlDB
}

// CloseMySQL 关闭 MySQL 连接
func CloseMySQL() error {
	if mysqlDB != nil {
		sqlDB, err := mysqlDB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// getGormLogger 获取 GORM 日志级别
func getGormLogger(level string) logger.Interface {
	var logLevel logger.LogLevel
	switch level {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Warn
	}
	return logger.Default.LogMode(logLevel)
}
