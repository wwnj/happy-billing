package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode int

const (
	// 成功
	Success ErrorCode = 0

	// 通用错误 (1000-1999)
	ErrInternalServer ErrorCode = 1000
	ErrInvalidParams  ErrorCode = 1001
	ErrNotFound       ErrorCode = 1002
	ErrUnauthorized   ErrorCode = 1003
	ErrForbidden      ErrorCode = 1004
	ErrTooManyRequest ErrorCode = 1005

	// 租户相关错误 (2000-2099)
	ErrTenantNotFound      ErrorCode = 2000
	ErrTenantAlreadyExists ErrorCode = 2001
	ErrTenantDisabled      ErrorCode = 2002
	ErrOrgNotFound         ErrorCode = 2010
	ErrProjectNotFound     ErrorCode = 2020
	ErrUserNotFound        ErrorCode = 2030

	// 产品相关错误 (3000-3099)
	ErrProductNotFound  ErrorCode = 3000
	ErrProductDisabled  ErrorCode = 3001
	ErrSKUNotFound      ErrorCode = 3010
	ErrSKUOutOfStock    ErrorCode = 3011
	ErrCategoryNotFound ErrorCode = 3020

	// 定价相关错误 (4000-4099)
	ErrPriceRuleNotFound  ErrorCode = 4000
	ErrPriceCalculateFail ErrorCode = 4001
	ErrDiscountInvalid    ErrorCode = 4010

	// 订单相关错误 (5000-5099)
	ErrOrderNotFound      ErrorCode = 5000
	ErrOrderStatusInvalid ErrorCode = 5001
	ErrOrderCannotCancel  ErrorCode = 5002
	ErrResourceNotFound   ErrorCode = 5010

	// 账单相关错误 (6000-6099)
	ErrBillNotFound      ErrorCode = 6000
	ErrBillAlreadyPaid   ErrorCode = 6001
	ErrBillAmountInvalid ErrorCode = 6002

	// 支付相关错误 (7000-7099)
	ErrPaymentFailed       ErrorCode = 7000
	ErrBalanceInsufficient ErrorCode = 7001
	ErrPaymentDuplicate    ErrorCode = 7002
	ErrRefundFailed        ErrorCode = 7010

	// 计量相关错误 (8000-8099)
	ErrMeteringDataInvalid ErrorCode = 8000
	ErrMeteringSubmitFail  ErrorCode = 8001
)

// BizError 业务错误
type BizError struct {
	Code    ErrorCode
	Message string
	Detail  interface{}
}

// Error 实现 error 接口
func (e *BizError) Error() string {
	if e.Detail != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New 创建业务错误
func New(code ErrorCode, message string) *BizError {
	return &BizError{
		Code:    code,
		Message: message,
	}
}

// Newf 创建带格式化消息的业务错误
func Newf(code ErrorCode, format string, args ...interface{}) *BizError {
	return &BizError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// WithDetail 添加错误详情
func (e *BizError) WithDetail(detail interface{}) *BizError {
	e.Detail = detail
	return e
}

// HTTPStatus 获取对应的 HTTP 状态码
func (e *BizError) HTTPStatus() int {
	switch e.Code {
	case Success:
		return http.StatusOK
	case ErrInvalidParams:
		return http.StatusBadRequest
	case ErrNotFound, ErrTenantNotFound, ErrProductNotFound, ErrOrderNotFound, ErrBillNotFound:
		return http.StatusNotFound
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrTooManyRequest:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// 预定义错误消息
var errorMessages = map[ErrorCode]string{
	Success:                "成功",
	ErrInternalServer:      "服务器内部错误",
	ErrInvalidParams:       "请求参数无效",
	ErrNotFound:            "资源不存在",
	ErrUnauthorized:        "未授权",
	ErrForbidden:           "禁止访问",
	ErrTooManyRequest:      "请求过于频繁",
	ErrTenantNotFound:      "租户不存在",
	ErrTenantAlreadyExists: "租户已存在",
	ErrTenantDisabled:      "租户已被禁用",
	ErrOrgNotFound:         "组织不存在",
	ErrProjectNotFound:     "项目不存在",
	ErrUserNotFound:        "用户不存在",
	ErrProductNotFound:     "产品不存在",
	ErrProductDisabled:     "产品已下架",
	ErrSKUNotFound:         "SKU不存在",
	ErrSKUOutOfStock:       "SKU库存不足",
	ErrCategoryNotFound:    "产品分类不存在",
	ErrPriceRuleNotFound:   "定价规则不存在",
	ErrPriceCalculateFail:  "价格计算失败",
	ErrDiscountInvalid:     "折扣规则无效",
	ErrOrderNotFound:       "订单不存在",
	ErrOrderStatusInvalid:  "订单状态无效",
	ErrOrderCannotCancel:   "订单无法取消",
	ErrResourceNotFound:    "资源实例不存在",
	ErrBillNotFound:        "账单不存在",
	ErrBillAlreadyPaid:     "账单已支付",
	ErrBillAmountInvalid:   "账单金额无效",
	ErrPaymentFailed:       "支付失败",
	ErrBalanceInsufficient: "余额不足",
	ErrPaymentDuplicate:    "重复支付",
	ErrRefundFailed:        "退款失败",
	ErrMeteringDataInvalid: "计量数据无效",
	ErrMeteringSubmitFail:  "计量数据提交失败",
}

// GetMessage 获取错误消息
func GetMessage(code ErrorCode) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// NewWithCode 使用错误码创建错误
func NewWithCode(code ErrorCode) *BizError {
	return &BizError{
		Code:    code,
		Message: GetMessage(code),
	}
}

// NewInternalError 创建内部错误
func NewInternalError(message string) *BizError {
	return &BizError{
		Code:    ErrInternalServer,
		Message: message,
	}
}
