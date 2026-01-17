package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/internal/api/response"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/service"
	"github.com/wwnj/happy-billing/pkg/errors"
)

// BillHandler 账单处理器
type BillHandler struct {
	billService service.BillService
}

// NewBillHandler 创建账单处理器实例
func NewBillHandler(billService service.BillService) *BillHandler {
	return &BillHandler{
		billService: billService,
	}
}

// ============================================================================
// 账单查询
// ============================================================================

// GetBill 获取账单详情
// @Summary 获取账单详情
// @Description 根据账单ID获取账单详细信息
// @Tags 账单管理
// @Accept json
// @Produce json
// @Param bill_id path string true "账单ID"
// @Success 200 {object} response.Response{data=models.Bill} "成功"
// @Failure 404 {object} response.Response "账单不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /bills/{bill_id} [get]
func (h *BillHandler) GetBill(c *gin.Context) {
	billID := c.Param("bill_id")
	if billID == "" {
		response.Error(c, errors.New(errors.ErrInvalidParams, "账单ID不能为空"))
		return
	}

	bill, err := h.billService.GetBill(c.Request.Context(), billID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, bill)
}

// ListBills 查询账单列表
// @Summary 查询账单列表
// @Description 根据条件查询账单列表（支持分页）
// @Tags 账单管理
// @Accept json
// @Produce json
// @Param tenant_id query string false "租户ID"
// @Param project_id query string false "项目ID"
// @Param order_id query string false "订单ID"
// @Param bill_type query string false "账单类型"
// @Param status query string false "账单状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=models.PageResult} "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /bills [get]
func (h *BillHandler) ListBills(c *gin.Context) {
	var req models.BillListQueryRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	result, err := h.billService.ListBills(c.Request.Context(), &req)
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

// GetOrderBills 获取订单的所有账单
// @Summary 获取订单的所有账单
// @Description 根据订单ID获取该订单的所有账单
// @Tags 账单管理
// @Accept json
// @Produce json
// @Param order_id path string true "订单ID"
// @Success 200 {object} response.Response{data=[]models.Bill} "成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /orders/{order_id}/bills [get]
func (h *BillHandler) GetOrderBills(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		response.Error(c, errors.New(errors.ErrInvalidParams, "订单ID不能为空"))
		return
	}

	bills, err := h.billService.GetBillsByOrderID(c.Request.Context(), orderID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, bills)
}
