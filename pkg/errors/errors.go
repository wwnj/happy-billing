package errors

import (
	"errors"
	"fmt"
)

// ErrorCode 错误码类型
type ErrorCode string

// 错误码常量定义
const (
	// 通用错误码 (1000-1999)
	CodeSuccess         ErrorCode = "SUCCESS"
	CodeInternalError   ErrorCode = "INTERNAL_ERROR"
	CodeInvalidParam    ErrorCode = "INVALID_PARAM"
	CodeNotFound        ErrorCode = "NOT_FOUND"
	CodeAlreadyExists   ErrorCode = "ALREADY_EXISTS"
	CodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	CodeForbidden       ErrorCode = "FORBIDDEN"
	CodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// 计量相关错误 (2000-2999)
	CodeMeterInvalidType      ErrorCode = "METER_INVALID_TYPE"
	CodeMeterInvalidPrecision ErrorCode = "METER_INVALID_PRECISION"
	CodeMeterCollectFailed    ErrorCode = "METER_COLLECT_FAILED"
	CodeMeterAggregateFailed  ErrorCode = "METER_AGGREGATE_FAILED"

	// 定价相关错误 (3000-3999)
	CodePriceRuleNotFound    ErrorCode = "PRICE_RULE_NOT_FOUND"
	CodePriceCalculateFailed ErrorCode = "PRICE_CALCULATE_FAILED"
	CodeDiscountInvalid      ErrorCode = "DISCOUNT_INVALID"

	// 订单相关错误 (4000-4999)
	CodeOrderNotFound      ErrorCode = "ORDER_NOT_FOUND"
	CodeOrderInvalidStatus ErrorCode = "ORDER_INVALID_STATUS"
	CodeOrderCannotCancel  ErrorCode = "ORDER_CANNOT_CANCEL"
	CodeOrderCreateFailed  ErrorCode = "ORDER_CREATE_FAILED"

	// 余额相关错误 (5000-5999)
	CodeAccountNotFound      ErrorCode = "ACCOUNT_NOT_FOUND"
	CodeInsufficientBalance  ErrorCode = "INSUFFICIENT_BALANCE"
	CodeBalanceDeductFailed  ErrorCode = "BALANCE_DEDUCT_FAILED"
	CodeTransactionDuplicate ErrorCode = "TRANSACTION_DUPLICATE"
	CodeTransactionNotFound  ErrorCode = "TRANSACTION_NOT_FOUND"

	// 账单相关错误 (6000-6999)
	CodeBillNotFound       ErrorCode = "BILL_NOT_FOUND"
	CodeBillAlreadyPaid    ErrorCode = "BILL_ALREADY_PAID"
	CodeBillGenerateFailed ErrorCode = "BILL_GENERATE_FAILED"

	// 套餐包相关错误 (7000-7999)
	CodePackageNotFound          ErrorCode = "PACKAGE_NOT_FOUND"
	CodePackageExpired           ErrorCode = "PACKAGE_EXPIRED"
	CodePackageQuotaInsufficient ErrorCode = "PACKAGE_QUOTA_INSUFFICIENT"
	CodePackageNotAvailable      ErrorCode = "PACKAGE_NOT_AVAILABLE"
	CodeInvalidStatus            ErrorCode = "INVALID_STATUS"

	// 基础设施相关错误 (9000-9999)
	CodeDatabaseError     ErrorCode = "DATABASE_ERROR"
	CodeCacheError        ErrorCode = "CACHE_ERROR"
	CodeMessageQueueError ErrorCode = "MESSAGE_QUEUE_ERROR"
	CodeLockAcquireFailed ErrorCode = "LOCK_ACQUIRE_FAILED"
	CodeLockReleaseFailed ErrorCode = "LOCK_RELEASE_FAILED"
)

// BizError 业务错误结构体
type BizError struct {
	Code    ErrorCode // 错误码
	Message string    // 错误消息
	Cause   error     // 原始错误(用于错误链)
}

// Error 实现error接口
func (e *BizError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现Unwrap接口,支持errors.Is和errors.As
func (e *BizError) Unwrap() error {
	return e.Cause
}

// New 创建新的业务错误
func New(code ErrorCode, message string) *BizError {
	return &BizError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装错误,添加错误链
func Wrap(code ErrorCode, message string, err error) *BizError {
	return &BizError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// Is 判断错误码是否匹配
func Is(err error, code ErrorCode) bool {
	var bizErr *BizError
	if errors.As(err, &bizErr) {
		return bizErr.Code == code
	}
	return false
}

// IsCode 判断错误码是否匹配（Is的别名，为了更清晰的语义）
func IsCode(err error, code ErrorCode) bool {
	return Is(err, code)
}

// GetCode 获取错误码
func GetCode(err error) ErrorCode {
	var bizErr *BizError
	if errors.As(err, &bizErr) {
		return bizErr.Code
	}
	return CodeInternalError
}

// GetMessage 获取错误消息
func GetMessage(err error) string {
	var bizErr *BizError
	if errors.As(err, &bizErr) {
		return bizErr.Message
	}
	return err.Error()
}

// 便捷函数

// NewInvalidParam 创建无效参数错误
func NewInvalidParam(message string) *BizError {
	return New(CodeInvalidParam, message)
}

// NewNotFound 创建资源不存在错误
func NewNotFound(resource string) *BizError {
	return New(CodeNotFound, fmt.Sprintf("%s not found", resource))
}

// NewAlreadyExists 创建资源已存在错误
func NewAlreadyExists(resource string) *BizError {
	return New(CodeAlreadyExists, fmt.Sprintf("%s already exists", resource))
}

// NewInternalError 创建内部错误
func NewInternalError(message string, cause error) *BizError {
	return Wrap(CodeInternalError, message, cause)
}

// NewInsufficientBalance 创建余额不足错误
func NewInsufficientBalance() *BizError {
	return New(CodeInsufficientBalance, "insufficient balance")
}

// NewTransactionDuplicate 创建交易重复错误
func NewTransactionDuplicate() *BizError {
	return New(CodeTransactionDuplicate, "transaction already processed")
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(operation string, cause error) *BizError {
	return Wrap(CodeDatabaseError, fmt.Sprintf("database %s failed", operation), cause)
}

// NewCacheError 创建缓存错误
func NewCacheError(operation string, cause error) *BizError {
	return Wrap(CodeCacheError, fmt.Sprintf("cache %s failed", operation), cause)
}
