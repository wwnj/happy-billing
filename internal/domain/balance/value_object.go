package balance

import (
	"fmt"

	"github.com/wwnj/happy-billing/pkg/errors"
)

// TransactionType 交易类型(值对象)
type TransactionType string

// 交易类型常量
const (
	TransactionTypeCharge   TransactionType = "CHARGE"   // 充值
	TransactionTypeDeduct   TransactionType = "DEDUCT"   // 扣费
	TransactionTypeFreeze   TransactionType = "FREEZE"   // 冻结
	TransactionTypeUnfreeze TransactionType = "UNFREEZE" // 解冻
	TransactionTypeRefund   TransactionType = "REFUND"   // 退款
	TransactionTypeAdjust   TransactionType = "ADJUST"   // 调整(人工调账)
)

// String 返回字符串表示
func (t TransactionType) String() string {
	return string(t)
}

// IsValid 验证交易类型是否有效
func (t TransactionType) IsValid() bool {
	switch t {
	case TransactionTypeCharge, TransactionTypeDeduct, TransactionTypeFreeze,
		TransactionTypeUnfreeze, TransactionTypeRefund, TransactionTypeAdjust:
		return true
	default:
		return false
	}
}

// Validate 验证并返回错误
func (t TransactionType) Validate() error {
	if !t.IsValid() {
		return errors.New(errors.CodeInvalidParam,
			fmt.Sprintf("invalid transaction type: %s", t))
	}
	return nil
}

// IsIncome 是否为收入类型(增加余额)
func (t TransactionType) IsIncome() bool {
	return t == TransactionTypeCharge ||
		t == TransactionTypeUnfreeze ||
		t == TransactionTypeRefund
}

// IsExpense 是否为支出类型(减少余额)
func (t TransactionType) IsExpense() bool {
	return t == TransactionTypeDeduct ||
		t == TransactionTypeFreeze
}
