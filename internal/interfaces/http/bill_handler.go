package http

import (
	"time"

	"github.com/gin-gonic/gin"

	billingCmd "github.com/wwnj/happy-billing/internal/application/command/billing"
	billingQuery "github.com/wwnj/happy-billing/internal/application/query/billing"
	"github.com/wwnj/happy-billing/internal/domain/billing"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// BillHandler 账单HTTP处理器
type BillHandler struct {
	generateService *billingCmd.GenerateBillService
	settleService   *billingCmd.SettleBillService
	cancelService   *billingCmd.CancelBillService
	billQuery       *billingQuery.BillQuery
}

// NewBillHandler 创建账单HTTP处理器
func NewBillHandler(
	generateService *billingCmd.GenerateBillService,
	settleService *billingCmd.SettleBillService,
	cancelService *billingCmd.CancelBillService,
	billQuery *billingQuery.BillQuery,
) *BillHandler {
	return &BillHandler{
		generateService: generateService,
		settleService:   settleService,
		cancelService:   cancelService,
		billQuery:       billQuery,
	}
}

// GenerateBillRequest 生成账单请求
type GenerateBillRequest struct {
	TenantID    string `json:"tenant_id" binding:"required"`
	UserID      string `json:"user_id" binding:"required"`
	Cycle       string `json:"cycle" binding:"required"`
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
	DueDate     string `json:"due_date"`
	Currency    string `json:"currency"`
}

// GenerateBill 生成账单接口（内部使用）
// POST /api/v1/bills/generate
func (h *BillHandler) GenerateBill(c *gin.Context) {
	var req GenerateBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewInvalidParam(err.Error()))
		return
	}

	// 解析时间
	periodStart, err := time.Parse(time.RFC3339, req.PeriodStart)
	if err != nil {
		Error(c, errors.NewInvalidParam("invalid period_start format, use RFC3339"))
		return
	}

	periodEnd, err := time.Parse(time.RFC3339, req.PeriodEnd)
	if err != nil {
		Error(c, errors.NewInvalidParam("invalid period_end format, use RFC3339"))
		return
	}

	var dueDate *time.Time
	if req.DueDate != "" {
		t, err := time.Parse(time.RFC3339, req.DueDate)
		if err != nil {
			Error(c, errors.NewInvalidParam("invalid due_date format, use RFC3339"))
			return
		}
		dueDate = &t
	}

	// 执行生成账单
	cmd := billingCmd.GenerateBillCommand{
		TenantID:    req.TenantID,
		UserID:      req.UserID,
		Cycle:       billing.BillCycle(req.Cycle),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		DueDate:     dueDate,
		Currency:    req.Currency,
	}

	bill, err := h.generateService.Execute(c.Request.Context(), cmd)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"bill_id": bill.ID,
		"bill_no": bill.BillNo,
		"status":  "generated",
	})
}

// SettleBillRequest 结算账单请求
type SettleBillRequest struct {
	PaidAmount string `json:"paid_amount" binding:"required"`
}

// SettleBill 结算账单接口
// POST /api/v1/bills/:id/settle
func (h *BillHandler) SettleBill(c *gin.Context) {
	billID := c.Param("id")
	if billID == "" {
		Error(c, errors.NewInvalidParam("bill_id is required"))
		return
	}

	var req SettleBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewInvalidParam(err.Error()))
		return
	}

	// 解析金额
	paidAmount, err := money.NewFromString(req.PaidAmount)
	if err != nil {
		Error(c, errors.NewInvalidParam("invalid paid_amount format"))
		return
	}

	// 执行结算
	cmd := billingCmd.SettleBillCommand{
		BillID:     billID,
		PaidAmount: paidAmount,
	}

	if err := h.settleService.Execute(c.Request.Context(), cmd); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"bill_id": billID,
		"status":  "settled",
	})
}

// CancelBillRequest 取消账单请求
type CancelBillRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// CancelBill 取消账单接口
// POST /api/v1/bills/:id/cancel
func (h *BillHandler) CancelBill(c *gin.Context) {
	billID := c.Param("id")
	if billID == "" {
		Error(c, errors.NewInvalidParam("bill_id is required"))
		return
	}

	var req CancelBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewInvalidParam(err.Error()))
		return
	}

	// 执行取消
	cmd := billingCmd.CancelBillCommand{
		BillID: billID,
		Reason: req.Reason,
	}

	if err := h.cancelService.Execute(c.Request.Context(), cmd); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"bill_id": billID,
		"status":  "cancelled",
	})
}

// GetBill 查询账单详情接口
// GET /api/v1/bills/:id
func (h *BillHandler) GetBill(c *gin.Context) {
	billID := c.Param("id")
	if billID == "" {
		Error(c, errors.NewInvalidParam("bill_id is required"))
		return
	}

	// 查询账单详情（包含明细）
	billDetail, err := h.billQuery.GetBillDetail(c.Request.Context(), billID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, billDetail)
}

// ListBillsRequest 账单列表请求
type ListBillsRequest struct {
	PaginationRequest
	TenantID  string `form:"tenant_id" binding:"required"`
	UserID    string `form:"user_id"`
	Status    string `form:"status"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

// ListBills 查询账单列表接口
// GET /api/v1/bills
func (h *BillHandler) ListBills(c *gin.Context) {
	var req ListBillsRequest
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

	// 查询账单列表
	query := billingQuery.ListBillsQuery{
		TenantID:  req.TenantID,
		UserID:    req.UserID,
		Status:    billing.BillStatus(req.Status),
		StartTime: startTime,
		EndTime:   endTime,
		Offset:    req.GetOffset(),
		Limit:     req.GetLimit(),
	}

	result, err := h.billQuery.ListBills(c.Request.Context(), query)
	if err != nil {
		Error(c, err)
		return
	}

	// 返回分页响应
	Success(c, PaginationResponse{
		Total:    result.Total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     result.Bills,
	})
}

// ListOverdueBills 查询逾期账单接口
// GET /api/v1/bills/overdue
func (h *BillHandler) ListOverdueBills(c *gin.Context) {
	var req PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, errors.NewInvalidParam(err.Error()))
		return
	}

	// 查询逾期账单
	result, err := h.billQuery.ListOverdueBills(
		c.Request.Context(),
		time.Now(),
		req.GetOffset(),
		req.GetLimit(),
	)
	if err != nil {
		Error(c, err)
		return
	}

	// 返回分页响应
	Success(c, PaginationResponse{
		Total:    result.Total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     result.Bills,
	})
}

// GetBillStatistics 账单统计接口
// GET /api/v1/bills/statistics
func (h *BillHandler) GetBillStatistics(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	userID := c.Query("user_id")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if tenantID == "" {
		Error(c, errors.NewInvalidParam("tenant_id is required"))
		return
	}

	// 解析时间范围
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		Error(c, errors.NewInvalidParam("invalid start_time format, use RFC3339"))
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		Error(c, errors.NewInvalidParam("invalid end_time format, use RFC3339"))
		return
	}

	// 统计账单总额
	totalAmount, err := h.billQuery.SumBillAmount(
		c.Request.Context(),
		tenantID,
		userID,
		startTime,
		endTime,
	)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"tenant_id":    tenantID,
		"user_id":      userID,
		"start_time":   startTime.Format(time.RFC3339),
		"end_time":     endTime.Format(time.RFC3339),
		"total_amount": totalAmount.String(),
	})
}

// RegisterRoutes 注册路由
func (h *BillHandler) RegisterRoutes(router *gin.RouterGroup) {
	bills := router.Group("/bills")
	{
		// 生成账单（内部接口）
		bills.POST("/generate", h.GenerateBill)

		// 查询账单
		bills.GET("", h.ListBills)
		bills.GET("/:id", h.GetBill)
		bills.GET("/overdue", h.ListOverdueBills)
		bills.GET("/statistics", h.GetBillStatistics)

		// 操作账单
		bills.POST("/:id/settle", h.SettleBill)
		bills.POST("/:id/cancel", h.CancelBill)
	}
}
