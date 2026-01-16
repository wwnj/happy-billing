package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wwnj/happy-billing/pkg/errors"
)

// Response 统一响应结构
type Response struct {
	Code    string      `json:"code"`               // 业务错误码
	Message string      `json:"message"`            // 错误消息
	Data    interface{} `json:"data,omitempty"`     // 响应数据
	TraceID string      `json:"trace_id,omitempty"` // 追踪ID
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    string(errors.CodeSuccess),
		Message: "success",
		Data:    data,
		TraceID: getTraceID(c),
	})
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	// 解析业务错误
	code := errors.GetCode(err)
	message := errors.GetMessage(err)

	// 映射HTTP状态码
	httpStatus := mapErrorCodeToHTTPStatus(code)

	c.JSON(httpStatus, Response{
		Code:    string(code),
		Message: message,
		TraceID: getTraceID(c),
	})
}

// mapErrorCodeToHTTPStatus 映射错误码到HTTP状态码
func mapErrorCodeToHTTPStatus(code errors.ErrorCode) int {
	switch code {
	case errors.CodeSuccess:
		return http.StatusOK
	case errors.CodeInvalidParam:
		return http.StatusBadRequest
	case errors.CodeNotFound:
		return http.StatusNotFound
	case errors.CodeAlreadyExists:
		return http.StatusConflict
	case errors.CodeUnauthorized:
		return http.StatusUnauthorized
	case errors.CodeForbidden:
		return http.StatusForbidden
	case errors.CodeTooManyRequests:
		return http.StatusTooManyRequests
	case errors.CodeInsufficientBalance:
		return http.StatusPaymentRequired
	default:
		return http.StatusInternalServerError
	}
}

// getTraceID 获取追踪ID
func getTraceID(c *gin.Context) string {
	traceID, exists := c.Get("trace_id")
	if !exists {
		return ""
	}
	return traceID.(string)
}

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// GetOffset 获取偏移量
func (p *PaginationRequest) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.GetLimit()
}

// GetLimit 获取每页数量
func (p *PaginationRequest) GetLimit() int {
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return p.PageSize
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Data     interface{} `json:"data"`
}
