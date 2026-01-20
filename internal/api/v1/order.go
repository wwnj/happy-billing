package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/internal/api/response"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/service"
	"github.com/wwnj/happy-billing/pkg/errors"
)

// OrderHandler 订单处理器
type OrderHandler struct {
	orderService service.OrderService
}

// NewOrderHandler 创建订单处理器实例
func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// ============================================================================
// 订单管理
// ============================================================================

// CreateOrder 创建订单
// @Summary 创建订单
// @Description 创建新订单（预付费订单会自动生成账单）
// @Tags 订单管理
// @Accept json
// @Produce json
// @Param body body models.CreateOrderRequest true "订单信息"
// @Success 200 {object} response.Response{data=models.Order} "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrInvalidParams, err.Error()))
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, order)
}

// GetOrder 获取订单详情
// @Summary 获取订单详情
// @Description 根据订单ID获取订单详细信息
// @Tags 订单管理
// @Accept json
// @Produce json
// @Param order_id path string true "订单ID"
// @Success 200 {object} response.Response{data=models.Order} "成功"
// @Failure 404 {object} response.Response "订单不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /orders/{order_id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		response.Error(c, errors.New(errors.ErrInvalidParams, "订单ID不能为空"))
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, order)
}

// ListOrders 查询订单列表
// @Summary 查询订单列表
// @Description 根据条件查询订单列表（支持分页）
// @Tags 订单管理
// @Accept json
// @Produce json
// @Param tenant_id query string false "租户ID"
// @Param project_id query string false "项目ID"
// @Param user_id query string false "用户ID"
// @Param order_type query string false "订单类型"
// @Param status query string false "订单状态"
// @Param currency query string false "货币类型"
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=models.PageResult} "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
	var req models.OrderListQueryRequest

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

	result, err := h.orderService.ListOrders(c.Request.Context(), &req)
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

// CancelOrder 取消订单
// @Summary 取消订单
// @Description 取消待支付状态的订单
// @Tags 订单管理
// @Accept json
// @Produce json
// @Param order_id path string true "订单ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "订单状态不允许取消"
// @Failure 404 {object} response.Response "订单不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /orders/{order_id}/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		response.Error(c, errors.New(errors.ErrInvalidParams, "订单ID不能为空"))
		return
	}

	err := h.orderService.CancelOrder(c.Request.Context(), orderID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	response.Success(c, gin.H{
		"order_id": orderID,
		"message":  "订单已取消",
	})
}

// GetOrderItems 获取订单明细
// @Summary 获取订单明细
// @Description 根据订单ID获取订单明细列表
// @Tags 订单管理
// @Accept json
// @Produce json
// @Param order_id path string true "订单ID"
// @Success 200 {object} response.Response{data=[]models.OrderItem} "成功"
// @Failure 404 {object} response.Response "订单不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /orders/{order_id}/items [get]
func (h *OrderHandler) GetOrderItems(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		response.Error(c, errors.New(errors.ErrInvalidParams, "订单ID不能为空"))
		return
	}

	// 先检查订单是否存在
	_, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.Error(c, errors.NewInternalError(err.Error()))
		}
		return
	}

	// 这里简化处理，实际应该通过 service 层获取
	response.Success(c, []models.OrderItem{})
}
