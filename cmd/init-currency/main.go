package main

import (
	"fmt"
	"log"
	"time"

	"github.com/wwnj/happy-billing/internal/models"
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

	// 初始化数据库
	if err := database.InitAll(cfg); err != nil {
		log.Fatalf("Failed to init databases: %v", err)
	}
	defer database.CloseAll()

	db := database.GetMySQL()

	// AutoMigrate 汇率表
	fmt.Println("Creating exchange_rates table...")
	if err := db.AutoMigrate(&models.ExchangeRate{}); err != nil {
		log.Fatalf("Failed to migrate exchange_rates table: %v", err)
	}

	fmt.Println("✓ exchange_rates table created")

	// 插入初始汇率数据
	fmt.Println("Inserting initial exchange rates...")

	rates := []models.ExchangeRate{
		{FromCurrency: "CNY", ToCurrency: "USD", Rate: 0.1385, EffectiveDate: mustParseDate("2026-01-17"), Source: "BANK"},
		{FromCurrency: "CNY", ToCurrency: "EUR", Rate: 0.1275, EffectiveDate: mustParseDate("2026-01-17"), Source: "BANK"},
		{FromCurrency: "CNY", ToCurrency: "JPY", Rate: 20.35, EffectiveDate: mustParseDate("2026-01-17"), Source: "BANK"},
		{FromCurrency: "CNY", ToCurrency: "GBP", Rate: 0.1098, EffectiveDate: mustParseDate("2026-01-17"), Source: "BANK"},
		{FromCurrency: "CNY", ToCurrency: "HKD", Rate: 1.085, EffectiveDate: mustParseDate("2026-01-17"), Source: "BANK"},
	}

	for _, rate := range rates {
		if err := db.Create(&rate).Error; err != nil {
			fmt.Printf("Warning: Failed to insert rate %s->%s: %v\n", rate.FromCurrency, rate.ToCurrency, err)
		} else {
			fmt.Printf("✓ Inserted rate: %s -> %s = %f\n", rate.FromCurrency, rate.ToCurrency, rate.Rate)
		}
	}

	fmt.Println("\nMigration completed!")
}

func mustParseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}
