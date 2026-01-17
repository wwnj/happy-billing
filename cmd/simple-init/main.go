package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 直接连接MySQL
	dsn := "billing_user:billing_pass_2024@tcp(127.0.0.1:3306)/happy_billing?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping: %v", err)
	}

	fmt.Println("Connected to MySQL successfully")

	// 创建表
	createTable := `
CREATE TABLE IF NOT EXISTS exchange_rates (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  from_currency VARCHAR(8) NOT NULL,
  to_currency VARCHAR(8) NOT NULL,
  rate DECIMAL(18,8) NOT NULL,
  effective_date DATE NOT NULL,
  source VARCHAR(64),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_currency_date (from_currency, to_currency, effective_date),
  INDEX idx_date (effective_date),
  INDEX idx_from_to (from_currency, to_currency)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	fmt.Println("✓ Table created")

	// 插入数据
	insertData := `
INSERT IGNORE INTO exchange_rates (from_currency, to_currency, rate, effective_date, source) VALUES
('CNY', 'USD', 0.13850000, '2026-01-17', 'BANK'),
('CNY', 'EUR', 0.12750000, '2026-01-17', 'BANK'),
('CNY', 'JPY', 20.35000000, '2026-01-17', 'BANK'),
('CNY', 'GBP', 0.10980000, '2026-01-17', 'BANK'),
('CNY', 'HKD', 1.08500000, '2026-01-17', 'BANK');
`

	result, err := db.Exec(insertData)
	if err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}

	rows, _ := result.RowsAffected()
	fmt.Printf("✓ Inserted %d rates\n", rows)

	fmt.Println("\nAll done!")
}
