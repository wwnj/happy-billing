package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/wwnj/happy-billing/pkg/config"
)

var clickhouseDB *sql.DB

// InitClickHouse 初始化 ClickHouse 连接
func InitClickHouse(cfg *config.ClickHouseConfig) error {
	// 构建连接选项
	options := &clickhouse.Options{
		Addr: cfg.Addresses,
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout:     cfg.DialTimeout,
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	}

	// 连接数据库
	db := clickhouse.OpenDB(options)

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	clickhouseDB = db
	return nil
}

// GetClickHouse 获取 ClickHouse 连接
func GetClickHouse() *sql.DB {
	return clickhouseDB
}

// CloseClickHouse 关闭 ClickHouse 连接
func CloseClickHouse() error {
	if clickhouseDB != nil {
		return clickhouseDB.Close()
	}
	return nil
}

// BatchInsert ClickHouse 批量插入助手函数
func BatchInsert(table string, columns []string, rows [][]interface{}) error {
	if clickhouseDB == nil {
		return fmt.Errorf("clickhouse not initialized")
	}

	if len(rows) == 0 {
		return nil
	}

	// 构建批量插入语句
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ","),
		strings.Join(placeholders, ","),
	)

	// 开始事务
	tx, err := clickhouseDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 准备语句
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// 批量插入
	for _, row := range rows {
		if _, err := stmt.Exec(row...); err != nil {
			return fmt.Errorf("failed to exec statement: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
