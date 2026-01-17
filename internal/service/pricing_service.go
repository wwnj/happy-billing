package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/repository"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/utils"
	"gorm.io/gorm"
)

// PricingService 定价服务接口
type PricingService interface {
	// 定价规则管理
	CreatePriceRule(ctx context.Context, req *models.CreatePriceRuleRequest) (*models.PriceRule, error)
	GetPriceRule(ctx context.Context, ruleID string) (*models.PriceRule, error)
	ListPriceRules(ctx context.Context, req *models.PriceRuleListQueryRequest) ([]models.PriceRule, int64, error)

	// 折扣规则管理
	CreateDiscountRule(ctx context.Context, req *models.CreateDiscountRuleRequest) (*models.DiscountRule, error)
	GetDiscountRule(ctx context.Context, discountID string) (*models.DiscountRule, error)
	ListDiscountRules(ctx context.Context, req *models.DiscountRuleListQueryRequest) ([]models.DiscountRule, int64, error)

	// 价格查询和计算
	QueryPrice(ctx context.Context, req *models.PriceQueryRequest) (*models.PriceRule, error)
	CalculatePrice(ctx context.Context, req *models.PriceCalculateRequest) (*models.PriceCalculateResponse, error)
}

type pricingService struct {
	db               *gorm.DB
	redis            *redis.Client
	priceRuleRepo    repository.PriceRuleRepository
	discountRuleRepo repository.DiscountRuleRepository
	productSkuRepo   repository.ProductSkuRepository
}

// NewPricingService 创建定价服务实例
func NewPricingService(
	db *gorm.DB,
	redis *redis.Client,
	priceRuleRepo repository.PriceRuleRepository,
	discountRuleRepo repository.DiscountRuleRepository,
	productSkuRepo repository.ProductSkuRepository,
) PricingService {
	return &pricingService{
		db:               db,
		redis:            redis,
		priceRuleRepo:    priceRuleRepo,
		discountRuleRepo: discountRuleRepo,
		productSkuRepo:   productSkuRepo,
	}
}

// ============================================================================
// 定价规则管理
// ============================================================================

func (s *pricingService) CreatePriceRule(ctx context.Context, req *models.CreatePriceRuleRequest) (*models.PriceRule, error) {
	// 生成定价规则ID
	ruleID, err := utils.GeneratePriceRuleID(ctx, s.redis)
	if err != nil {
		return nil, errors.NewInternalError("生成定价规则ID失败: " + err.Error())
	}

	// 创建定价规则
	rule := &models.PriceRule{
		RuleID:         ruleID,
		RuleCode:       req.RuleCode,
		RuleName:       req.RuleName,
		SpuCode:        req.SpuCode,
		SkuCode:        req.SkuCode,
		RuleType:       req.RuleType,
		PricingDetail:  req.PricingDetail,
		Currency:       req.Currency,
		EffectiveStart: req.EffectiveStart,
		EffectiveEnd:   req.EffectiveEnd,
		Region:         req.Region,
		Priority:       req.Priority,
		Status:         models.StatusEnabled,
	}

	if err := s.priceRuleRepo.Create(ctx, rule); err != nil {
		return nil, errors.NewInternalError("创建定价规则失败: " + err.Error())
	}

	return rule, nil
}

func (s *pricingService) GetPriceRule(ctx context.Context, ruleID string) (*models.PriceRule, error) {
	rule, err := s.priceRuleRepo.GetByID(ctx, ruleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewWithCode(errors.ErrPriceRuleNotFound)
		}
		return nil, errors.NewInternalError("查询定价规则失败: " + err.Error())
	}
	return rule, nil
}

func (s *pricingService) ListPriceRules(ctx context.Context, req *models.PriceRuleListQueryRequest) ([]models.PriceRule, int64, error) {
	rules, total, err := s.priceRuleRepo.List(ctx, req)
	if err != nil {
		return nil, 0, errors.NewInternalError("查询定价规则列表失败: " + err.Error())
	}
	return rules, total, nil
}

// ============================================================================
// 折扣规则管理
// ============================================================================

func (s *pricingService) CreateDiscountRule(ctx context.Context, req *models.CreateDiscountRuleRequest) (*models.DiscountRule, error) {
	// 生成折扣规则ID
	discountID, err := utils.GenerateDiscountID(ctx, s.redis)
	if err != nil {
		return nil, errors.NewInternalError("生成折扣规则ID失败: " + err.Error())
	}

	// 创建折扣规则
	rule := &models.DiscountRule{
		DiscountID:     discountID,
		DiscountName:   req.DiscountName,
		DiscountType:   req.DiscountType,
		DiscountValue:  req.DiscountValue,
		TargetType:     req.TargetType,
		TargetID:       req.TargetID,
		EffectiveStart: req.EffectiveStart,
		EffectiveEnd:   req.EffectiveEnd,
		MaxDiscount:    req.MaxDiscount,
		MinAmount:      req.MinAmount,
		Status:         models.StatusEnabled,
	}

	if err := s.discountRuleRepo.Create(ctx, rule); err != nil {
		return nil, errors.NewInternalError("创建折扣规则失败: " + err.Error())
	}

	return rule, nil
}

func (s *pricingService) GetDiscountRule(ctx context.Context, discountID string) (*models.DiscountRule, error) {
	rule, err := s.discountRuleRepo.GetByID(ctx, discountID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewWithCode(errors.ErrDiscountRuleNotFound)
		}
		return nil, errors.NewInternalError("查询折扣规则失败: " + err.Error())
	}
	return rule, nil
}

func (s *pricingService) ListDiscountRules(ctx context.Context, req *models.DiscountRuleListQueryRequest) ([]models.DiscountRule, int64, error) {
	rules, total, err := s.discountRuleRepo.List(ctx, req)
	if err != nil {
		return nil, 0, errors.NewInternalError("查询折扣规则列表失败: " + err.Error())
	}
	return rules, total, nil
}

// ============================================================================
// 价格查询和计算
// ============================================================================

func (s *pricingService) QueryPrice(ctx context.Context, req *models.PriceQueryRequest) (*models.PriceRule, error) {
	// 查询SKU信息
	sku, err := s.productSkuRepo.GetByCode(ctx, req.SkuCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrInvalidParams, "SKU不存在")
		}
		return nil, errors.NewInternalError("查询SKU失败: " + err.Error())
	}

	// 查询有效的定价规则
	rule, err := s.priceRuleRepo.GetEffectiveRule(ctx, &req.SkuCode, &sku.SpuCode, req.Region)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrInvalidParams, "未找到有效的定价规则")
		}
		return nil, errors.NewInternalError("查询定价规则失败: " + err.Error())
	}

	return rule, nil
}

func (s *pricingService) CalculatePrice(ctx context.Context, req *models.PriceCalculateRequest) (*models.PriceCalculateResponse, error) {
	// 1. 查询定价规则
	priceReq := &models.PriceQueryRequest{
		SkuCode:  req.SkuCode,
		Region:   req.Region,
		TenantID: &req.TenantID,
	}
	rule, err := s.QueryPrice(ctx, priceReq)
	if err != nil {
		return nil, err
	}

	// 2. 计算原价
	originalPrice, err := s.calculateOriginalPrice(rule, req.Quantity, req.StartTime, req.EndTime)
	if err != nil {
		return nil, errors.NewInternalError("计算原价失败: " + err.Error())
	}

	// 3. 查询折扣规则
	sku, _ := s.productSkuRepo.GetByCode(ctx, req.SkuCode)
	discounts, _ := s.discountRuleRepo.GetEffectiveDiscounts(ctx, req.TenantID, &sku.SpuCode, &req.SkuCode)

	// 4. 计算折扣金额
	discountAmount := s.calculateDiscountAmount(originalPrice, discounts)

	// 5. 计算最终价格
	finalPrice := originalPrice - discountAmount
	if finalPrice < 0 {
		finalPrice = 0
	}

	return &models.PriceCalculateResponse{
		OriginalPrice:  originalPrice,
		DiscountAmount: discountAmount,
		FinalPrice:     finalPrice,
		Currency:       rule.Currency,
		PriceRule:      rule,
		Discounts:      discounts,
	}, nil
}

// ============================================================================
// 私有方法
// ============================================================================

// calculateOriginalPrice 计算原价
func (s *pricingService) calculateOriginalPrice(rule *models.PriceRule, quantity float64, startTime, endTime time.Time) (float64, error) {
	switch rule.RuleType {
	case models.PriceRuleTypeFixed:
		return s.calculateFixed(rule, quantity)
	case models.PriceRuleTypeTiered:
		return s.calculateTiered(rule, quantity)
	case models.PriceRuleTypeTimeBased:
		return s.calculateTimeBased(rule, quantity, startTime, endTime)
	case models.PriceRuleTypePackage:
		return s.calculatePackage(rule, quantity)
	default:
		return 0, errors.New(errors.ErrInvalidParams, "不支持的定价类型")
	}
}

// calculateFixed 固定价格
func (s *pricingService) calculateFixed(rule *models.PriceRule, quantity float64) (float64, error) {
	var detail models.FixedPricingDetail
	bytes, _ := json.Marshal(rule.PricingDetail)
	if err := json.Unmarshal(bytes, &detail); err != nil {
		return 0, err
	}
	return detail.UnitPrice * quantity, nil
}

// calculateTiered 阶梯价格
func (s *pricingService) calculateTiered(rule *models.PriceRule, quantity float64) (float64, error) {
	var detail models.TieredPricingDetail
	bytes, _ := json.Marshal(rule.PricingDetail)
	if err := json.Unmarshal(bytes, &detail); err != nil {
		return 0, err
	}

	totalPrice := 0.0
	remainingQty := quantity

	for _, tier := range detail.Tiers {
		if remainingQty <= 0 {
			break
		}

		tierSize := 0.0
		if tier.To == nil {
			// 最后一档，无上限
			tierSize = remainingQty
		} else {
			tierSize = *tier.To - tier.From
			if tierSize > remainingQty {
				tierSize = remainingQty
			}
		}

		totalPrice += tierSize * tier.UnitPrice
		remainingQty -= tierSize
	}

	return totalPrice, nil
}

// calculateTimeBased 时段价格
func (s *pricingService) calculateTimeBased(rule *models.PriceRule, quantity float64, startTime, endTime time.Time) (float64, error) {
	var detail models.TimeBasedPricingDetail
	bytes, _ := json.Marshal(rule.PricingDetail)
	if err := json.Unmarshal(bytes, &detail); err != nil {
		return 0, err
	}

	// 简化实现：使用平均价格
	totalUnitPrice := 0.0
	for _, period := range detail.Periods {
		totalUnitPrice += period.UnitPrice
	}
	avgUnitPrice := totalUnitPrice / float64(len(detail.Periods))

	return avgUnitPrice * quantity, nil
}

// calculatePackage 资源包价格
func (s *pricingService) calculatePackage(rule *models.PriceRule, quantity float64) (float64, error) {
	var detail models.PackagePricingDetail
	bytes, _ := json.Marshal(rule.PricingDetail)
	if err := json.Unmarshal(bytes, &detail); err != nil {
		return 0, err
	}

	// 计算需要购买的资源包数量
	packageCount := (quantity + detail.PackageSize - 1) / detail.PackageSize
	return detail.PackagePrice * packageCount, nil
}

// calculateDiscountAmount 计算折扣金额
func (s *pricingService) calculateDiscountAmount(originalPrice float64, discounts []models.DiscountRule) float64 {
	totalDiscount := 0.0

	for _, discount := range discounts {
		// 检查最小消费限制
		if discount.MinAmount != nil && originalPrice < *discount.MinAmount {
			continue
		}

		discountAmount := 0.0
		switch discount.DiscountType {
		case models.DiscountTypePercentage:
			// 百分比折扣
			discountAmount = originalPrice * discount.DiscountValue
		case models.DiscountTypeAmount:
			// 金额折扣
			discountAmount = discount.DiscountValue
		}

		// 检查最大折扣限制
		if discount.MaxDiscount != nil && discountAmount > *discount.MaxDiscount {
			discountAmount = *discount.MaxDiscount
		}

		totalDiscount += discountAmount
	}

	return totalDiscount
}
