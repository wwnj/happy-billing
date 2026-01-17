package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/internal/api/response"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/service"
	"github.com/wwnj/happy-billing/pkg/errors"
)

// CurrencyHandler 货币处理器
type CurrencyHandler struct {
	currencyService service.CurrencyService
}

// NewCurrencyHandler 创建货币处理器实例
func NewCurrencyHandler(currencyService service.CurrencyService) *CurrencyHandler {
	return &CurrencyHandler{
		currencyService: currencyService,
	}
}

// CreateExchangeRate 创建汇率
// @Summary 创建汇率
// @Description 创建新的汇率记录
// @Tags 货币管理
// @Accept json
// @Produce json
// @Param body body models.CreateExchangeRateRequest true "汇率信息"
// @Success 200 {object} response.Response{data=models.ExchangeRate} "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /exchange-rates [post]
func (h *CurrencyHandler) CreateExchangeRate(c *gin.Context) {
	var req models.CreateExchangeRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	rate, err := h.currencyService.CreateExchangeRate(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, rate)
}

// GetExchangeRate 获取汇率
// @Summary 获取汇率
// @Description 查询指定日期的汇率（不传日期则使用最新汇率）
// @Tags 货币管理
// @Accept json
// @Produce json
// @Param from_currency query string true "源币种"
// @Param to_currency query string true "目标币种"
// @Param effective_date query string false "生效日期 (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=models.ExchangeRate} "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "汇率不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /exchange-rates/query [get]
func (h *CurrencyHandler) GetExchangeRate(c *gin.Context) {
	var req models.ExchangeRateQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	rate, err := h.currencyService.GetExchangeRate(c.Request.Context(), req.FromCurrency, req.ToCurrency, req.EffectiveDate)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, rate)
}

// ListExchangeRates 查询汇率列表
// @Summary 查询汇率列表
// @Description 查询汇率列表（支持分页）
// @Tags 货币管理
// @Accept json
// @Produce json
// @Param from_currency query string false "源币种"
// @Param to_currency query string false "目标币种"
// @Param effective_date query string false "生效日期"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=models.PageResult} "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /exchange-rates [get]
func (h *CurrencyHandler) ListExchangeRates(c *gin.Context) {
	var req models.ExchangeRateListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	result, err := h.currencyService.ListExchangeRates(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, result)
}

// ConvertCurrency 货币转换
// @Summary 货币转换
// @Description 将一种货币转换为另一种货币
// @Tags 货币管理
// @Accept json
// @Produce json
// @Param body body models.CurrencyConvertRequest true "转换请求"
// @Success 200 {object} response.Response{data=models.CurrencyConvertResponse} "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "汇率不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /currency/convert [post]
func (h *CurrencyHandler) ConvertCurrency(c *gin.Context) {
	var req models.CurrencyConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	result, err := h.currencyService.ConvertCurrency(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, result)
}
