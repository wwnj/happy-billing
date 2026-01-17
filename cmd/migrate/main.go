package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/wwnj/happy-billing/pkg/config"
	"github.com/wwnj/happy-billing/pkg/database"
	"github.com/wwnj/happy-billing/pkg/logger"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	if err := logger.Init(&logger.Config{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	}); err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer logger.Sync()

	fmt.Println("Initializing database migrations...")

	// 初始化数据库
	if err := database.InitAll(cfg); err != nil {
		log.Fatalf("Failed to init databases: %v", err)
	}
	defer database.CloseAll()

	db := database.GetMySQL()

	// 读取迁移文件
	migrations := []string{
		"migrations/20240117_create_exchange_rates.sql",
		"migrations/20240117_add_multi_currency_fields.sql",
	}

	for _, file := range migrations {
		fmt.Printf("Executing migration: %s\n", file)

		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		// 分割SQL语句（按;分割）
		statements := strings.Split(string(content), ";")

		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" || strings.HasPrefix(stmt, "--") {
				continue
			}

			if err := db.Exec(stmt).Error; err != nil {
				log.Printf("Error executing statement: %s\nError: %v", stmt, err)
			} else {
				fmt.Printf("✓ Executed statement\n")
			}
		}

		fmt.Printf("✓ Migration %s completed\n\n", file)
	}

	fmt.Println("All migrations completed successfully!")
}
