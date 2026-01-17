package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/repository"
	"github.com/wwnj/happy-billing/pkg/errors"
	"gorm.io/gorm"
)

// CurrencyService 货币服务接口
type CurrencyService interface {
	// 汇率管理
	CreateExchangeRate(ctx context.Context, req *models.CreateExchangeRateRequest) (*models.ExchangeRate, error)
	GetExchangeRate(ctx context.Context, fromCurrency, toCurrency models.Currency, date *time.Time) (*models.ExchangeRate, error)
	ListExchangeRates(ctx context.Context, req *models.ExchangeRateListQuery) (*models.PageResult, error)

	// 货币转换
	ConvertCurrency(ctx context.Context, req *models.CurrencyConvertRequest) (*models.CurrencyConvertResponse, error)

	// 计算本位币金额（总是转换为CNY）
	CalculateBaseCurrencyAmount(ctx context.Context, amount float64, currency models.Currency, date *time.Time) (float64, float64, error)
}

type currencyService struct {
	redis        *redis.Client
	exchangeRepo repository.ExchangeRateRepository
}

// NewCurrencyService 创建货币服务实例
func NewCurrencyService(
	redis *redis.Client,
	exchangeRepo repository.ExchangeRateRepository,
) CurrencyService {
	return &currencyService{
		redis:        redis,
		exchangeRepo: exchangeRepo,
	}
}

func (s *currencyService) CreateExchangeRate(ctx context.Context, req *models.CreateExchangeRateRequest) (*models.ExchangeRate, error) {
	rate := &models.ExchangeRate{
		FromCurrency:  req.FromCurrency,
		ToCurrency:    req.ToCurrency,
		Rate:          req.Rate,
		EffectiveDate: req.EffectiveDate,
		Source:        req.Source,
	}

	if err := s.exchangeRepo.Create(ctx, rate); err != nil {
		return nil, errors.NewInternalError("创建汇率失败: " + err.Error())
	}

	// 更新Redis缓存
	cacheKey := fmt.Sprintf("rate:%s:%s:%s", req.FromCurrency, req.ToCurrency, req.EffectiveDate.Format("2006-01-02"))
	s.redis.Set(ctx, cacheKey, req.Rate, 24*time.Hour)

	return rate, nil
}

func (s *currencyService) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency models.Currency, date *time.Time) (*models.ExchangeRate, error) {
	// 如果是同一币种，汇率为1
	if fromCurrency == toCurrency {
		now := time.Now()
		if date == nil {
			date = &now
		}
		return &models.ExchangeRate{
			FromCurrency:  fromCurrency,
			ToCurrency:    toCurrency,
			Rate:          1.0,
			EffectiveDate: *date,
			Source:        "SYSTEM",
		}, nil
	}

	var effectiveDate time.Time
	if date == nil {
		effectiveDate = time.Now()
	} else {
		effectiveDate = *date
	}

	// 尝试从Redis缓存获取
	cacheKey := fmt.Sprintf("rate:%s:%s:%s", fromCurrency, toCurrency, effectiveDate.Format("2006-01-02"))
	val, err := s.redis.Get(ctx, cacheKey).Float64()
	if err == nil {
		return &models.ExchangeRate{
			FromCurrency:  fromCurrency,
			ToCurrency:    toCurrency,
			Rate:          val,
			EffectiveDate: effectiveDate,
			Source:        "CACHE",
		}, nil
	}

	// 从数据库查询
	rate, err := s.exchangeRepo.GetRate(ctx, fromCurrency, toCurrency, effectiveDate)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 尝试获取最新汇率
			rate, err = s.exchangeRepo.GetLatestRate(ctx, fromCurrency, toCurrency)
			if err != nil {
				return nil, errors.Newf(errors.ErrNotFound, "未找到 %s -> %s 的汇率", fromCurrency, toCurrency)
			}
		} else {
			return nil, errors.NewInternalError("查询汇率失败: " + err.Error())
		}
	}

	// 缓存到Redis
	s.redis.Set(ctx, cacheKey, rate.Rate, 24*time.Hour)

	return rate, nil
}

func (s *currencyService) ListExchangeRates(ctx context.Context, req *models.ExchangeRateListQuery) (*models.PageResult, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	rates, total, err := s.exchangeRepo.ListRates(ctx, req)
	if err != nil {
		return nil, errors.NewInternalError("查询汇率列表失败: " + err.Error())
	}

	return &models.PageResult{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     rates,
	}, nil
}

func (s *currencyService) ConvertCurrency(ctx context.Context, req *models.CurrencyConvertRequest) (*models.CurrencyConvertResponse, error) {
	// 获取汇率
	rate, err := s.GetExchangeRate(ctx, req.FromCurrency, req.ToCurrency, req.Date)
	if err != nil {
		return nil, err
	}

	// 转换金额
	convertedAmount := req.Amount * rate.Rate

	return &models.CurrencyConvertResponse{
		OriginalAmount:    req.Amount,
		OriginalCurrency:  req.FromCurrency,
		ConvertedAmount:   convertedAmount,
		ConvertedCurrency: req.ToCurrency,
		ExchangeRate:      rate.Rate,
		EffectiveDate:     rate.EffectiveDate,
	}, nil
}

func (s *currencyService) CalculateBaseCurrencyAmount(ctx context.Context, amount float64, currency models.Currency, date *time.Time) (float64, float64, error) {
	// 如果已经是本位币（CNY），直接返回
	if currency == models.CurrencyCNY {
		return amount, 1.0, nil
	}

	// 获取汇率：currency -> CNY
	// 注意：数据库中存储的是 CNY -> currency 的汇率，需要取倒数
	rate, err := s.GetExchangeRate(ctx, models.CurrencyCNY, currency, date)
	if err != nil {
		return 0, 0, err
	}

	// 转换为本位币
	// 如果rate是 1 CNY = 0.14 USD，则 1 USD = 1/0.14 CNY
	baseCurrencyAmount := amount / rate.Rate

	return baseCurrencyAmount, rate.Rate, nil
}
