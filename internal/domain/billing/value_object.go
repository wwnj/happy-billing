package billing

import "github.com/wwnj/happy-billing/pkg/errors"

// BillStatus 账单状态
type BillStatus string

const (
	BillStatusPending   BillStatus = "PENDING"   // 待结算
	BillStatusSettled   BillStatus = "SETTLED"   // 已结算
	BillStatusCancelled BillStatus = "CANCELLED" // 已取消
	BillStatusOverdue   BillStatus = "OVERDUE"   // 已逾期
)

// Validate 验证账单状态
func (s BillStatus) Validate() error {
	switch s {
	case BillStatusPending, BillStatusSettled, BillStatusCancelled, BillStatusOverdue:
		return nil
	default:
		return errors.New(errors.CodeInvalidParam, "invalid bill status: "+string(s))
	}
}

// IsFinal 是否为最终状态
func (s BillStatus) IsFinal() bool {
	return s == BillStatusSettled || s == BillStatusCancelled
}

// BillCycle 账单周期
type BillCycle string

const (
	BillCycleMonthly   BillCycle = "MONTHLY"   // 月度账单
	BillCycleQuarterly BillCycle = "QUARTERLY" // 季度账单
	BillCycleYearly    BillCycle = "YEARLY"    // 年度账单
	BillCycleCustom    BillCycle = "CUSTOM"    // 自定义周期
)

// Validate 验证账单周期
func (c BillCycle) Validate() error {
	switch c {
	case BillCycleMonthly, BillCycleQuarterly, BillCycleYearly, BillCycleCustom:
		return nil
	default:
		return errors.New(errors.CodeInvalidParam, "invalid bill cycle: "+string(c))
	}
}

// BillItemType 账单明细类型
type BillItemType string

const (
	BillItemTypeOrder    BillItemType = "ORDER"    // 订单费用
	BillItemTypeCharge   BillItemType = "CHARGE"   // 充值
	BillItemTypeRefund   BillItemType = "REFUND"   // 退款
	BillItemTypeAdjust   BillItemType = "ADJUST"   // 调整
	BillItemTypeDiscount BillItemType = "DISCOUNT" // 折扣
)

// Validate 验证账单明细类型
func (t BillItemType) Validate() error {
	switch t {
	case BillItemTypeOrder, BillItemTypeCharge, BillItemTypeRefund, BillItemTypeAdjust, BillItemTypeDiscount:
		return nil
	default:
		return errors.New(errors.CodeInvalidParam, "invalid bill item type: "+string(t))
	}
}

// IsIncome 是否为收入类型
func (t BillItemType) IsIncome() bool {
	return t == BillItemTypeCharge
}

// IsExpense 是否为支出类型
func (t BillItemType) IsExpense() bool {
	return t == BillItemTypeOrder
}

// IsAdjustment 是否为调整类型
func (t BillItemType) IsAdjustment() bool {
	return t == BillItemTypeRefund || t == BillItemTypeAdjust || t == BillItemTypeDiscount
}
