package http

import (
	"time"

	"github.com/gin-gonic/gin"

	balanceCmd "github.com/wwnj/happy-billing/internal/application/command/balance"
	balanceQuery "github.com/wwnj/happy-billing/internal/application/query/balance"
	"github.com/wwnj/happy-billing/internal/domain/balance"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// BalanceHandler 余额HTTP处理器
type BalanceHandler struct {
	chargeService *balanceCmd.ChargeService
	deductService *balanceCmd.DeductService
	balanceQuery  *balanceQuery.BalanceQuery
}

// NewBalanceHandler 创建余额HTTP处理器
func NewBalanceHandler(
	chargeService *balanceCmd.ChargeService,
	deductService *balanceCmd.DeductService,
	balanceQuery *balanceQuery.BalanceQuery,
) *BalanceHandler {
	return &BalanceHandler{
		chargeService: chargeService,
		deductService: deductService,
		balanceQuery:  balanceQuery,
	}
}

// ChargeRequest 充值请求
type ChargeRequest struct {
	AccountID     string `json:"account_id" binding:"required"`
	Amount        string `json:"amount" binding:"required"`
	TransactionID string `json:"transaction_id" binding:"required"`
	Description   string `json:"description"`
}

// Charge 充值接口
// POST /api/v1/balance/charge
func (h *BalanceHandler) Charge(c *gin.Context) {
	var req ChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewInvalidParam(err.Error()))
		return
	}

	// 解析金额
	amount, err := money.NewFromString(req.Amount)
	if err != nil {
		Error(c, errors.NewInvalidParam("invalid amount format"))
		return
	}

	// 执行充值
	cmd := balanceCmd.ChargeCommand{
		AccountID:     req.AccountID,
		Amount:        amount,
		TransactionID: req.TransactionID,
		Description:   req.Description,
	}

	if err := h.chargeService.Execute(c.Request.Context(), cmd); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"transaction_id": req.TransactionID,
		"status":         "success",
	})
}

// DeductRequest 扣费请求
type DeductRequest struct {
	AccountID     string  `json:"account_id" binding:"required"`
	Amount        string  `json:"amount" binding:"required"`
	TransactionID string  `json:"transaction_id" binding:"required"`
	OrderID       *string `json:"order_id"`
	Description   string  `json:"description"`
}

// Deduct 扣费接口（内部使用）
// POST /api/v1/balance/deduct
func (h *BalanceHandler) Deduct(c *gin.Context) {
	var req DeductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewInvalidParam(err.Error()))
		return
	}

	// 解析金额
	amount, err := money.NewFromString(req.Amount)
	if err != nil {
		Error(c, errors.NewInvalidParam("invalid amount format"))
		return
	}

	// 执行扣费
	cmd := balanceCmd.DeductCommand{
		AccountID:     req.AccountID,
		Amount:        amount,
		TransactionID: req.TransactionID,
		OrderID:       req.OrderID,
		Description:   req.Description,
	}

	if err := h.deductService.Execute(c.Request.Context(), cmd); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"transaction_id": req.TransactionID,
		"status":         "success",
	})
}

// GetBalance 查询余额接口
// GET /api/v1/balance/:account_id
func (h *BalanceHandler) GetBalance(c *gin.Context) {
	accountID := c.Param("account_id")
	if accountID == "" {
		Error(c, errors.NewInvalidParam("account_id is required"))
		return
	}

	balance, err := h.balanceQuery.GetAccountBalance(c.Request.Context(), accountID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, balance)
}

// GetBalanceByUserID 根据用户ID查询余额接口
// GET /api/v1/balance/user/:tenant_id/:user_id
func (h *BalanceHandler) GetBalanceByUserID(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	userID := c.Param("user_id")

	if tenantID == "" || userID == "" {
		Error(c, errors.NewInvalidParam("tenant_id and user_id are required"))
		return
	}

	balance, err := h.balanceQuery.GetAccountBalanceByUserID(c.Request.Context(), tenantID, userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, balance)
}

// ListTransactionsRequest 交易记录列表请求
type ListTransactionsRequest struct {
	PaginationRequest
	AccountID       string `form:"account_id" binding:"required"`
	TransactionType string `form:"transaction_type"`
	StartTime       string `form:"start_time"`
	EndTime         string `form:"end_time"`
}

// ListTransactions 查询交易记录列表接口
// GET /api/v1/balance/transactions
func (h *BalanceHandler) ListTransactions(c *gin.Context) {
	var req ListTransactionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, errors.NewInvalidParam(err.Error()))
		return
	}

	// 解析时间范围
	var startTime, endTime *time.Time
	if req.StartTime != "" {
		t, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			Error(c, errors.NewInvalidParam("invalid start_time format, use RFC3339"))
			return
		}
		startTime = &t
	}
	if req.EndTime != "" {
		t, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			Error(c, errors.NewInvalidParam("invalid end_time format, use RFC3339"))
			return
		}
		endTime = &t
	}

	// 查询交易记录
	query := balanceQuery.ListTransactionsQuery{
		AccountID:       req.AccountID,
		TransactionType: balance.TransactionType(req.TransactionType),
		StartTime:       startTime,
		EndTime:         endTime,
		Offset:          req.GetOffset(),
		Limit:           req.GetLimit(),
	}

	result, err := h.balanceQuery.ListTransactions(c.Request.Context(), query)
	if err != nil {
		Error(c, err)
		return
	}

	// 返回分页响应
	Success(c, PaginationResponse{
		Total:    result.Total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     result.Transactions,
	})
}

// GetTransaction 查询单个交易记录接口
// GET /api/v1/balance/transactions/:transaction_id
func (h *BalanceHandler) GetTransaction(c *gin.Context) {
	transactionID := c.Param("transaction_id")
	if transactionID == "" {
		Error(c, errors.NewInvalidParam("transaction_id is required"))
		return
	}

	transaction, err := h.balanceQuery.GetTransaction(c.Request.Context(), transactionID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, transaction)
}

// RegisterRoutes 注册路由
func (h *BalanceHandler) RegisterRoutes(router *gin.RouterGroup) {
	balance := router.Group("/balance")
	{
		// 充值
		balance.POST("/charge", h.Charge)

		// 扣费（内部接口）
		balance.POST("/deduct", h.Deduct)

		// 查询余额
		balance.GET("/:account_id", h.GetBalance)
		balance.GET("/user/:tenant_id/:user_id", h.GetBalanceByUserID)

		// 查询交易记录
		balance.GET("/transactions", h.ListTransactions)
		balance.GET("/transactions/:transaction_id", h.GetTransaction)
	}
}
