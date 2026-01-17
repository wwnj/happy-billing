package repository

import (
	"context"
	"time"

	"github.com/wwnj/happy-billing/internal/models"
	"gorm.io/gorm"
)

// ExchangeRateRepository 汇率仓储接口
type ExchangeRateRepository interface {
	Create(ctx context.Context, rate *models.ExchangeRate) error
	GetRate(ctx context.Context, fromCurrency, toCurrency models.Currency, effectiveDate time.Time) (*models.ExchangeRate, error)
	GetLatestRate(ctx context.Context, fromCurrency, toCurrency models.Currency) (*models.ExchangeRate, error)
	ListRates(ctx context.Context, req *models.ExchangeRateListQuery) ([]models.ExchangeRate, int64, error)
}

type exchangeRateRepository struct {
	db *gorm.DB
}

// NewExchangeRateRepository 创建汇率仓储实例
func NewExchangeRateRepository(db *gorm.DB) ExchangeRateRepository {
	return &exchangeRateRepository{db: db}
}

func (r *exchangeRateRepository) Create(ctx context.Context, rate *models.ExchangeRate) error {
	return r.db.WithContext(ctx).Create(rate).Error
}

func (r *exchangeRateRepository) GetRate(ctx context.Context, fromCurrency, toCurrency models.Currency, effectiveDate time.Time) (*models.ExchangeRate, error) {
	var rate models.ExchangeRate
	err := r.db.WithContext(ctx).
		Where("from_currency = ? AND to_currency = ? AND effective_date = ?", fromCurrency, toCurrency, effectiveDate.Format("2006-01-02")).
		First(&rate).Error
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *exchangeRateRepository) GetLatestRate(ctx context.Context, fromCurrency, toCurrency models.Currency) (*models.ExchangeRate, error) {
	var rate models.ExchangeRate
	err := r.db.WithContext(ctx).
		Where("from_currency = ? AND to_currency = ?", fromCurrency, toCurrency).
		Order("effective_date DESC").
		First(&rate).Error
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *exchangeRateRepository) ListRates(ctx context.Context, req *models.ExchangeRateListQuery) ([]models.ExchangeRate, int64, error) {
	var rates []models.ExchangeRate
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ExchangeRate{})

	if req.FromCurrency != nil {
		query = query.Where("from_currency = ?", *req.FromCurrency)
	}
	if req.ToCurrency != nil {
		query = query.Where("to_currency = ?", *req.ToCurrency)
	}
	if req.EffectiveDate != nil {
		query = query.Where("effective_date = ?", req.EffectiveDate.Format("2006-01-02"))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.PageSize
	err := query.
		Order("effective_date DESC, created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&rates).Error

	return rates, total, err
}
