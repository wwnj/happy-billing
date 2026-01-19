package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/internal/api/response"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/service"
	"github.com/wwnj/happy-billing/pkg/errors"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	paymentService service.PaymentService
}

// NewPaymentHandler 创建支付处理器实例
func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// ============================================================================
// 支付操作
// ============================================================================

// CreatePayment 创建支付
// @Summary 创建支付
// @Description 创建支付记录并处理支付（支持余额支付和第三方支付）
// @Tags 支付管理
// @Accept json
// @Produce json
// @Param body body models.CreatePaymentRequest true "支付信息"
// @Success 200 {object} response.Response{data=models.Payment} "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /payments [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	payment, err := h.paymentService.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, payment)
}

// GetPayment 获取支付详情
// @Summary 获取支付详情
// @Description 根据支付ID获取支付详细信息
// @Tags 支付管理
// @Accept json
// @Produce json
// @Param payment_id path string true "支付ID"
// @Success 200 {object} response.Response{data=models.Payment} "成功"
// @Failure 404 {object} response.Response "支付记录不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /payments/{payment_id} [get]
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID := c.Param("payment_id")
	if paymentID == "" {
		response.Error(c, errors.New(errors.ErrInvalidParams, "支付ID不能为空"))
		return
	}

	payment, err := h.paymentService.GetPayment(c.Request.Context(), paymentID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, payment)
}

// ============================================================================
// 余额操作
// ============================================================================

// GetBalance 获取账户余额
// @Summary 获取账户余额
// @Description 根据租户ID获取账户余额信息
// @Tags 账户余额
// @Accept json
// @Produce json
// @Param tenant_id path string true "租户ID"
// @Success 200 {object} response.Response{data=models.AccountBalance} "成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /tenants/{tenant_id}/balance [get]
func (h *PaymentHandler) GetBalance(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.Error(c, errors.New(errors.ErrInvalidParams, "租户ID不能为空"))
		return
	}

	balance, err := h.paymentService.GetBalance(c.Request.Context(), tenantID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, balance)
}

// Recharge 账户充值
// @Summary 账户充值
// @Description 为租户账户充值
// @Tags 账户余额
// @Accept json
// @Produce json
// @Param tenant_id path string true "租户ID"
// @Param body body object{amount=float64} true "充值金额"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /tenants/{tenant_id}/balance/recharge [post]
func (h *PaymentHandler) Recharge(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.Error(c, errors.New(errors.ErrInvalidParams, "租户ID不能为空"))
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	err := h.paymentService.Recharge(c.Request.Context(), tenantID, req.Amount)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, gin.H{
		"tenant_id": tenantID,
		"amount":    req.Amount,
		"message":   "充值成功",
	})
}

// GetBalanceTransactions 获取余额变动记录
// @Summary 获取余额变动记录
// @Description 获取租户的余额变动记录列表（支持分页）
// @Tags 账户余额
// @Accept json
// @Produce json
// @Param tenant_id path string true "租户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=object} "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /tenants/{tenant_id}/balance/transactions [get]
func (h *PaymentHandler) GetBalanceTransactions(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.Error(c, errors.New(errors.ErrInvalidParams, "租户ID不能为空"))
		return
	}

	// 获取分页参数
	page := 1
	pageSize := 20
	if p, ok := c.GetQuery("page"); ok && p != "" {
		if pInt, err := strconv.Atoi(p); err == nil && pInt > 0 {
			page = pInt
		}
	}
	if ps, ok := c.GetQuery("page_size"); ok && ps != "" {
		if psInt, err := strconv.Atoi(ps); err == nil && psInt > 0 {
			pageSize = psInt
		}
	}

	transactions, total, err := h.paymentService.GetBalanceTransactions(c.Request.Context(), tenantID, page, pageSize)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, gin.H{
		"data":      transactions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
