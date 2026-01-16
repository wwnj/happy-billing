package pkgdomain

import (
	"time"

	"github.com/wwnj/happy-billing/pkg/money"
)

// PackagePurchasedEvent 套餐包购买事件
type PackagePurchasedEvent struct {
	PackageID  string        // 套餐包ID
	PackageNo  string        // 套餐包号
	TenantID   string        // 租户ID
	UserID     string        // 用户ID
	Type       PackageType   // 套餐包类型
	TotalQuota money.Decimal // 总配额
	Price      money.Decimal // 价格
	ValidFrom  time.Time     // 生效时间
	ValidTo    time.Time     // 过期时间
	OccurredAt time.Time     // 发生时间
}

// EventType 返回事件类型
func (e *PackagePurchasedEvent) EventType() string {
	return "package.purchased"
}

// PackageConsumedEvent 配额消费事件
type PackageConsumedEvent struct {
	PackageID      string        // 套餐包ID
	PackageNo      string        // 套餐包号
	TenantID       string        // 租户ID
	UserID         string        // 用户ID
	ConsumedQuota  money.Decimal // 消费配额
	RemainingQuota money.Decimal // 剩余配额
	OrderID        string        // 关联订单ID
	Description    string        // 描述
	OccurredAt     time.Time     // 发生时间
}

// EventType 返回事件类型
func (e *PackageConsumedEvent) EventType() string {
	return "package.consumed"
}

// PackageExpiredEvent 套餐包过期事件
type PackageExpiredEvent struct {
	PackageID      string        // 套餐包ID
	PackageNo      string        // 套餐包号
	TenantID       string        // 租户ID
	UserID         string        // 用户ID
	RemainingQuota money.Decimal // 剩余配额
	ValidTo        time.Time     // 过期时间
	OccurredAt     time.Time     // 发生时间
}

// EventType 返回事件类型
func (e *PackageExpiredEvent) EventType() string {
	return "package.expired"
}

// PackageExhaustedEvent 配额耗尽事件
type PackageExhaustedEvent struct {
	PackageID   string        // 套餐包ID
	PackageNo   string        // 套餐包号
	TenantID    string        // 租户ID
	UserID      string        // 用户ID
	UsedQuota   money.Decimal // 已使用配额
	ExhaustedAt time.Time     // 耗尽时间
	OccurredAt  time.Time     // 发生时间
}

// EventType 返回事件类型
func (e *PackageExhaustedEvent) EventType() string {
	return "package.exhausted"
}

// PackageCancelledEvent 套餐包取消事件
type PackageCancelledEvent struct {
	PackageID      string        // 套餐包ID
	PackageNo      string        // 套餐包号
	TenantID       string        // 租户ID
	UserID         string        // 用户ID
	OldStatus      PackageStatus // 原状态
	RemainingQuota money.Decimal // 剩余配额
	Reason         string        // 取消原因
	CancelledAt    time.Time     // 取消时间
	OccurredAt     time.Time     // 发生时间
}

// EventType 返回事件类型
func (e *PackageCancelledEvent) EventType() string {
	return "package.cancelled"
}
