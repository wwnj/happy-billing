package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/internal/api/response"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/service"
	"github.com/wwnj/happy-billing/pkg/errors"
)

// PricingHandler 定价处理器
type PricingHandler struct {
	pricingService service.PricingService
}

// NewPricingHandler 创建定价处理器
func NewPricingHandler(pricingService service.PricingService) *PricingHandler {
	return &PricingHandler{
		pricingService: pricingService,
	}
}

// ============================================================================
// 定价规则管理
// ============================================================================

// CreatePriceRule 创建定价规则
func (h *PricingHandler) CreatePriceRule(c *gin.Context) {
	var req models.CreatePriceRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	rule, err := h.pricingService.CreatePriceRule(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, rule)
}

// GetPriceRule 获取定价规则详情
func (h *PricingHandler) GetPriceRule(c *gin.Context) {
	ruleID := c.Param("rule_id")

	rule, err := h.pricingService.GetPriceRule(c.Request.Context(), ruleID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, rule)
}

// ListPriceRules 查询定价规则列表
func (h *PricingHandler) ListPriceRules(c *gin.Context) {
	var req models.PriceRuleListQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	// 设置默认分页参数
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	rules, total, err := h.pricingService.ListPriceRules(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	result := &models.PageResult{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     rules,
	}
	response.Success(c, result)
}

// ============================================================================
// 折扣规则管理
// ============================================================================

// CreateDiscountRule 创建折扣规则
func (h *PricingHandler) CreateDiscountRule(c *gin.Context) {
	var req models.CreateDiscountRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	rule, err := h.pricingService.CreateDiscountRule(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, rule)
}

// GetDiscountRule 获取折扣规则详情
func (h *PricingHandler) GetDiscountRule(c *gin.Context) {
	discountID := c.Param("discount_id")

	rule, err := h.pricingService.GetDiscountRule(c.Request.Context(), discountID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, rule)
}

// ListDiscountRules 查询折扣规则列表
func (h *PricingHandler) ListDiscountRules(c *gin.Context) {
	var req models.DiscountRuleListQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	// 设置默认分页参数
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	rules, total, err := h.pricingService.ListDiscountRules(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	result := &models.PageResult{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     rules,
	}
	response.Success(c, result)
}

// ============================================================================
// 价格查询和计算
// ============================================================================

// QueryPrice 查询价格
func (h *PricingHandler) QueryPrice(c *gin.Context) {
	var req models.PriceQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	rule, err := h.pricingService.QueryPrice(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, rule)
}

// CalculatePrice 计算价格
func (h *PricingHandler) CalculatePrice(c *gin.Context) {
	var req models.PriceCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	result, err := h.pricingService.CalculatePrice(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}
