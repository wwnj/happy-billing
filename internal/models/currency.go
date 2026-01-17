package models

import (
	"time"
)

// ExchangeRate 汇率模型
type ExchangeRate struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	FromCurrency  Currency  `gorm:"column:from_currency;size:8;not null" json:"from_currency"`
	ToCurrency    Currency  `gorm:"column:to_currency;size:8;not null" json:"to_currency"`
	Rate          float64   `gorm:"column:rate;type:decimal(18,8);not null" json:"rate"`
	EffectiveDate time.Time `gorm:"column:effective_date;type:date;not null" json:"effective_date"`
	Source        string    `gorm:"column:source;size:64" json:"source"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 表名
func (ExchangeRate) TableName() string {
	return "exchange_rates"
}

// ExchangeRateQuery 汇率查询请求
type ExchangeRateQuery struct {
	FromCurrency  Currency   `json:"from_currency" form:"from_currency" binding:"required"`
	ToCurrency    Currency   `json:"to_currency" form:"to_currency" binding:"required"`
	EffectiveDate *time.Time `json:"effective_date,omitempty" form:"effective_date"`
}

// ExchangeRateListQuery 汇率列表查询
type ExchangeRateListQuery struct {
	Pagination
	FromCurrency  *Currency  `json:"from_currency,omitempty" form:"from_currency"`
	ToCurrency    *Currency  `json:"to_currency,omitempty" form:"to_currency"`
	EffectiveDate *time.Time `json:"effective_date,omitempty" form:"effective_date"`
}

// CreateExchangeRateRequest 创建汇率请求
type CreateExchangeRateRequest struct {
	FromCurrency  Currency  `json:"from_currency" binding:"required"`
	ToCurrency    Currency  `json:"to_currency" binding:"required"`
	Rate          float64   `json:"rate" binding:"required,gt=0"`
	EffectiveDate time.Time `json:"effective_date" binding:"required"`
	Source        string    `json:"source"`
}

// CurrencyConvertRequest 货币转换请求
type CurrencyConvertRequest struct {
	Amount       float64    `json:"amount" binding:"required,gt=0"`
	FromCurrency Currency   `json:"from_currency" binding:"required"`
	ToCurrency   Currency   `json:"to_currency" binding:"required"`
	Date         *time.Time `json:"date,omitempty"` // 不传则使用今日汇率
}

// CurrencyConvertResponse 货币转换响应
type CurrencyConvertResponse struct {
	OriginalAmount    float64   `json:"original_amount"`
	OriginalCurrency  Currency  `json:"original_currency"`
	ConvertedAmount   float64   `json:"converted_amount"`
	ConvertedCurrency Currency  `json:"converted_currency"`
	ExchangeRate      float64   `json:"exchange_rate"`
	EffectiveDate     time.Time `json:"effective_date"`
}
