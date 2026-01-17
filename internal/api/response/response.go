package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/pkg/errors"
)

// Response 统一响应结构
type Response struct {
	Code    errors.ErrorCode `json:"code"`    // 业务错误码
	Message string           `json:"message"` // 消息
	Data    interface{}      `json:"data"`    // 数据
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errors.Success,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errors.Success,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, err *errors.BizError) {
	c.JSON(err.HTTPStatus(), Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    err.Detail,
	})
}

// ErrorWithCode 错误响应（使用错误码）
func ErrorWithCode(c *gin.Context, code errors.ErrorCode) {
	err := errors.NewWithCode(code)
	c.JSON(err.HTTPStatus(), Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    nil,
	})
}

// ErrorWithMessage 错误响应（自定义消息）
func ErrorWithMessage(c *gin.Context, code errors.ErrorCode, message string) {
	err := errors.New(code, message)
	c.JSON(err.HTTPStatus(), Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    nil,
	})
}

// InternalError 服务器内部错误
func InternalError(c *gin.Context) {
	ErrorWithCode(c, errors.ErrInternalServer)
}

// BadRequest 请求参数错误
func BadRequest(c *gin.Context, message string) {
	ErrorWithMessage(c, errors.ErrInvalidParams, message)
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	ErrorWithMessage(c, errors.ErrNotFound, message)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context) {
	ErrorWithCode(c, errors.ErrUnauthorized)
}

// Forbidden 禁止访问
func Forbidden(c *gin.Context) {
	ErrorWithCode(c, errors.ErrForbidden)
}
