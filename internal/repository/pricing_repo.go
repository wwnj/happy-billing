package repository

import (
	"context"
	"time"

	"github.com/wwnj/happy-billing/internal/models"
	"gorm.io/gorm"
)

// ============================================================================
// PriceRule Repository Interfaces
// ============================================================================

// PriceRuleRepository 定价规则仓储接口
type PriceRuleRepository interface {
	Create(ctx context.Context, rule *models.PriceRule) error
	GetByID(ctx context.Context, ruleID string) (*models.PriceRule, error)
	GetByCode(ctx context.Context, ruleCode string) (*models.PriceRule, error)
	List(ctx context.Context, req *models.PriceRuleListQueryRequest) ([]models.PriceRule, int64, error)
	Update(ctx context.Context, rule *models.PriceRule) error
	Delete(ctx context.Context, ruleID string) error

	// 查询有效的定价规则
	GetEffectiveRule(ctx context.Context, skuCode, spuCode, region *string) (*models.PriceRule, error)
}

// DiscountRuleRepository 折扣规则仓储接口
type DiscountRuleRepository interface {
	Create(ctx context.Context, rule *models.DiscountRule) error
	GetByID(ctx context.Context, discountID string) (*models.DiscountRule, error)
	List(ctx context.Context, req *models.DiscountRuleListQueryRequest) ([]models.DiscountRule, int64, error)
	Update(ctx context.Context, rule *models.DiscountRule) error
	Delete(ctx context.Context, discountID string) error

	// 查询有效的折扣规则
	GetEffectiveDiscounts(ctx context.Context, tenantID string, spuCode, skuCode *string) ([]models.DiscountRule, error)
}

// ============================================================================
// PriceRule Repository Implementation
// ============================================================================

type priceRuleRepository struct {
	db *gorm.DB
}

// NewPriceRuleRepository 创建定价规则仓储实例
func NewPriceRuleRepository(db *gorm.DB) PriceRuleRepository {
	return &priceRuleRepository{db: db}
}

func (r *priceRuleRepository) Create(ctx context.Context, rule *models.PriceRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *priceRuleRepository) GetByID(ctx context.Context, ruleID string) (*models.PriceRule, error) {
	var rule models.PriceRule
	err := r.db.WithContext(ctx).
		Where("rule_id = ?", ruleID).
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *priceRuleRepository) GetByCode(ctx context.Context, ruleCode string) (*models.PriceRule, error) {
	var rule models.PriceRule
	err := r.db.WithContext(ctx).
		Where("rule_code = ?", ruleCode).
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *priceRuleRepository) List(ctx context.Context, req *models.PriceRuleListQueryRequest) ([]models.PriceRule, int64, error) {
	var rules []models.PriceRule
	var total int64

	query := r.db.WithContext(ctx).Model(&models.PriceRule{})

	// 过滤条件
	if req.SpuCode != nil {
		query = query.Where("spu_code = ?", *req.SpuCode)
	}
	if req.SkuCode != nil {
		query = query.Where("sku_code = ?", *req.SkuCode)
	}
	if req.RuleType != nil {
		query = query.Where("rule_type = ?", *req.RuleType)
	}
	if req.Region != nil {
		query = query.Where("region = ?", *req.Region)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	err := query.
		Order("priority DESC, created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&rules).Error

	return rules, total, err
}

func (r *priceRuleRepository) Update(ctx context.Context, rule *models.PriceRule) error {
	return r.db.WithContext(ctx).
		Model(&models.PriceRule{}).
		Where("rule_id = ?", rule.RuleID).
		Updates(rule).Error
}

func (r *priceRuleRepository) Delete(ctx context.Context, ruleID string) error {
	return r.db.WithContext(ctx).
		Where("rule_id = ?", ruleID).
		Delete(&models.PriceRule{}).Error
}

func (r *priceRuleRepository) GetEffectiveRule(ctx context.Context, skuCode, spuCode, region *string) (*models.PriceRule, error) {
	var rule models.PriceRule
	now := time.Now()

	query := r.db.WithContext(ctx).Model(&models.PriceRule{}).
		Where("status = ?", models.StatusEnabled).
		Where("effective_start <= ?", now).
		Where("effective_end IS NULL OR effective_end >= ?", now)

	// 优先级：SKU级别 > SPU+地域 > SPU级别
	if skuCode != nil {
		query = query.Where("sku_code = ?", *skuCode)
	} else if spuCode != nil && region != nil {
		query = query.Where("spu_code = ? AND region = ?", *spuCode, *region)
	} else if spuCode != nil {
		query = query.Where("spu_code = ?", *spuCode)
	}

	err := query.
		Order("priority DESC, created_at DESC").
		First(&rule).Error

	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// ============================================================================
// DiscountRule Repository Implementation
// ============================================================================

type discountRuleRepository struct {
	db *gorm.DB
}

// NewDiscountRuleRepository 创建折扣规则仓储实例
func NewDiscountRuleRepository(db *gorm.DB) DiscountRuleRepository {
	return &discountRuleRepository{db: db}
}

func (r *discountRuleRepository) Create(ctx context.Context, rule *models.DiscountRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *discountRuleRepository) GetByID(ctx context.Context, discountID string) (*models.DiscountRule, error) {
	var rule models.DiscountRule
	err := r.db.WithContext(ctx).
		Where("discount_id = ?", discountID).
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *discountRuleRepository) List(ctx context.Context, req *models.DiscountRuleListQueryRequest) ([]models.DiscountRule, int64, error) {
	var rules []models.DiscountRule
	var total int64

	query := r.db.WithContext(ctx).Model(&models.DiscountRule{})

	// 过滤条件
	if req.TargetType != nil {
		query = query.Where("target_type = ?", *req.TargetType)
	}
	if req.TargetID != nil {
		query = query.Where("target_id = ?", *req.TargetID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&rules).Error

	return rules, total, err
}

func (r *discountRuleRepository) Update(ctx context.Context, rule *models.DiscountRule) error {
	return r.db.WithContext(ctx).
		Model(&models.DiscountRule{}).
		Where("discount_id = ?", rule.DiscountID).
		Updates(rule).Error
}

func (r *discountRuleRepository) Delete(ctx context.Context, discountID string) error {
	return r.db.WithContext(ctx).
		Where("discount_id = ?", discountID).
		Delete(&models.DiscountRule{}).Error
}

func (r *discountRuleRepository) GetEffectiveDiscounts(ctx context.Context, tenantID string, spuCode, skuCode *string) ([]models.DiscountRule, error) {
	var rules []models.DiscountRule
	now := time.Now()

	query := r.db.WithContext(ctx).Model(&models.DiscountRule{}).
		Where("status = ?", models.StatusEnabled).
		Where("effective_start <= ?", now).
		Where("effective_end IS NULL OR effective_end >= ?", now)

	// 查询租户级别折扣
	tenantQuery := query.Session(&gorm.Session{}).
		Where("target_type = ? AND target_id = ?", models.DiscountTargetTypeTenant, tenantID)

	// 查询SPU级别折扣
	var spuQuery *gorm.DB
	if spuCode != nil {
		spuQuery = query.Session(&gorm.Session{}).
			Where("target_type = ? AND target_id = ?", models.DiscountTargetTypeSPU, *spuCode)
	}

	// 查询SKU级别折扣
	var skuQuery *gorm.DB
	if skuCode != nil {
		skuQuery = query.Session(&gorm.Session{}).
			Where("target_type = ? AND target_id = ?", models.DiscountTargetTypeSKU, *skuCode)
	}

	// 合并查询结果
	if err := tenantQuery.Find(&rules).Error; err != nil {
		return nil, err
	}

	if spuQuery != nil {
		var spuRules []models.DiscountRule
		if err := spuQuery.Find(&spuRules).Error; err == nil {
			rules = append(rules, spuRules...)
		}
	}

	if skuQuery != nil {
		var skuRules []models.DiscountRule
		if err := skuQuery.Find(&skuRules).Error; err == nil {
			rules = append(rules, skuRules...)
		}
	}

	return rules, nil
}
