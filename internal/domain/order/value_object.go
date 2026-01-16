package order

import (
	"fmt"

	"github.com/wwnj/happy-billing/pkg/errors"
)

// OrderStatus 订单状态(值对象)
type OrderStatus string

// 订单状态常量
const (
	OrderStatusPending   OrderStatus = "PENDING"   // 待处理
	OrderStatusPaid      OrderStatus = "PAID"      // 已支付
	OrderStatusCompleted OrderStatus = "COMPLETED" // 已完成
	OrderStatusCancelled OrderStatus = "CANCELLED" // 已取消
	OrderStatusFailed    OrderStatus = "FAILED"    // 失败
	OrderStatusRefunded  OrderStatus = "REFUNDED"  // 已退款
)

// String 返回字符串表示
func (o OrderStatus) String() string {
	return string(o)
}

// IsValid 验证订单状态是否有效
func (o OrderStatus) IsValid() bool {
	switch o {
	case OrderStatusPending, OrderStatusPaid, OrderStatusCompleted,
		OrderStatusCancelled, OrderStatusFailed, OrderStatusRefunded:
		return true
	default:
		return false
	}
}

// Validate 验证并返回错误
func (o OrderStatus) Validate() error {
	if !o.IsValid() {
		return errors.New(errors.CodeInvalidParam,
			fmt.Sprintf("invalid order status: %s", o))
	}
	return nil
}

// IsFinal 是否为最终状态(不可再变更)
func (o OrderStatus) IsFinal() bool {
	return o == OrderStatusCompleted ||
		o == OrderStatusCancelled ||
		o == OrderStatusFailed ||
		o == OrderStatusRefunded
}

// OrderType 订单类型(值对象)
type OrderType string

// 订单类型常量
const (
	OrderTypePostpaid OrderType = "POSTPAID" // 后付费(按量计费)
	OrderTypePrepaid  OrderType = "PREPAID"  // 预付费(预充值)
	OrderTypePackage  OrderType = "PACKAGE"  // 套餐包
	OrderTypeRecharge OrderType = "RECHARGE" // 充值订单
)

// String 返回字符串表示
func (o OrderType) String() string {
	return string(o)
}

// IsValid 验证订单类型是否有效
func (o OrderType) IsValid() bool {
	switch o {
	case OrderTypePostpaid, OrderTypePrepaid, OrderTypePackage, OrderTypeRecharge:
		return true
	default:
		return false
	}
}

// Validate 验证并返回错误
func (o OrderType) Validate() error {
	if !o.IsValid() {
		return errors.New(errors.CodeInvalidParam,
			fmt.Sprintf("invalid order type: %s", o))
	}
	return nil
}

// PaymentMode 支付方式(值对象)
type PaymentMode string

// 支付方式常量
const (
	PaymentModeBalance PaymentMode = "BALANCE" // 余额支付
	PaymentModeAlipay  PaymentMode = "ALIPAY"  // 支付宝
	PaymentModeWechat  PaymentMode = "WECHAT"  // 微信支付
	PaymentModeBank    PaymentMode = "BANK"    // 银行转账
	PaymentModeOffline PaymentMode = "OFFLINE" // 线下支付
)

// String 返回字符串表示
func (p PaymentMode) String() string {
	return string(p)
}

// IsValid 验证支付方式是否有效
func (p PaymentMode) IsValid() bool {
	switch p {
	case PaymentModeBalance, PaymentModeAlipay, PaymentModeWechat,
		PaymentModeBank, PaymentModeOffline:
		return true
	default:
		return false
	}
}

// Validate 验证并返回错误
func (p PaymentMode) Validate() error {
	if !p.IsValid() {
		return errors.New(errors.CodeInvalidParam,
			fmt.Sprintf("invalid payment mode: %s", p))
	}
	return nil
}
