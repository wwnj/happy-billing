package pkgdomain

import (
	"time"

	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// Package 套餐包聚合根
type Package struct {
	// 基本信息
	ID          string        // 套餐包ID
	PackageNo   string        // 套餐包号
	TenantID    string        // 租户ID
	UserID      string        // 用户ID
	Type        PackageType   // 套餐包类型
	Status      PackageStatus // 套餐包状态
	Name        string        // 套餐包名称
	Description string        // 描述

	// 配额信息
	TotalQuota     money.Decimal // 总配额
	UsedQuota      money.Decimal // 已使用配额
	RemainingQuota money.Decimal // 剩余配额
	QuotaUnit      string        // 配额单位（如：秒、GB、次）

	// 金额信息
	Price    money.Decimal // 购买价格
	Currency string        // 货币类型

	// 有效期
	PurchasedAt time.Time  // 购买时间
	ValidFrom   time.Time  // 生效时间
	ValidTo     time.Time  // 过期时间
	ExhaustedAt *time.Time // 耗尽时间
	CancelledAt *time.Time // 取消时间

	// 元数据
	Metadata map[string]interface{} // 扩展元数据

	// 领域事件
	events []interface{}
}

// NewPackage 创建套餐包
func NewPackage(
	id, packageNo, tenantID, userID string,
	packageType PackageType,
	name, description string,
	totalQuota money.Decimal,
	quotaUnit string,
	price money.Decimal,
	currency string,
	validFrom, validTo time.Time,
) (*Package, error) {
	// 参数验证
	if id == "" || packageNo == "" || tenantID == "" || userID == "" {
		return nil, errors.New(errors.CodeInvalidParam, "id, package_no, tenant_id, user_id are required")
	}

	if err := packageType.Validate(); err != nil {
		return nil, err
	}

	if totalQuota.LessThanOrEqual(money.Zero) {
		return nil, errors.New(errors.CodeInvalidParam, "total_quota must be greater than 0")
	}

	if validFrom.After(validTo) {
		return nil, errors.New(errors.CodeInvalidParam, "valid_from must be before valid_to")
	}

	now := time.Now()
	pkg := &Package{
		ID:             id,
		PackageNo:      packageNo,
		TenantID:       tenantID,
		UserID:         userID,
		Type:           packageType,
		Status:         PackageStatusActive,
		Name:           name,
		Description:    description,
		TotalQuota:     totalQuota,
		UsedQuota:      money.Zero,
		RemainingQuota: totalQuota,
		QuotaUnit:      quotaUnit,
		Price:          price,
		Currency:       currency,
		PurchasedAt:    now,
		ValidFrom:      validFrom,
		ValidTo:        validTo,
		Metadata:       make(map[string]interface{}),
		events:         []interface{}{},
	}

	// 发布套餐包创建事件
	pkg.AddEvent(&PackagePurchasedEvent{
		PackageID:  pkg.ID,
		PackageNo:  pkg.PackageNo,
		TenantID:   pkg.TenantID,
		UserID:     pkg.UserID,
		Type:       pkg.Type,
		TotalQuota: pkg.TotalQuota,
		Price:      pkg.Price,
		ValidFrom:  pkg.ValidFrom,
		ValidTo:    pkg.ValidTo,
		OccurredAt: now,
	})

	return pkg, nil
}

// Consume 消费配额
func (p *Package) Consume(quantity money.Decimal, orderID, description string) error {
	// 状态检查
	if p.Status != PackageStatusActive {
		return errors.New(errors.CodePackageNotAvailable, "package is not active")
	}

	// 过期检查
	if time.Now().After(p.ValidTo) {
		return errors.New(errors.CodePackageExpired, "package has expired")
	}

	// 配额检查
	if quantity.LessThanOrEqual(money.Zero) {
		return errors.New(errors.CodeInvalidParam, "consume quantity must be greater than 0")
	}

	if p.RemainingQuota.LessThan(quantity) {
		return errors.New(errors.CodePackageQuotaInsufficient, "insufficient package quota")
	}

	// 扣减配额
	p.UsedQuota = money.Add(p.UsedQuota, quantity)
	p.RemainingQuota = money.Sub(p.RemainingQuota, quantity)

	// 发布配额消费事件
	p.AddEvent(&PackageConsumedEvent{
		PackageID:      p.ID,
		PackageNo:      p.PackageNo,
		TenantID:       p.TenantID,
		UserID:         p.UserID,
		ConsumedQuota:  quantity,
		RemainingQuota: p.RemainingQuota,
		OrderID:        orderID,
		Description:    description,
		OccurredAt:     time.Now(),
	})

	// 检查是否耗尽
	if p.RemainingQuota.LessThanOrEqual(money.Zero) {
		p.MarkExhausted()
	}

	return nil
}

// MarkExpired 标记为过期
func (p *Package) MarkExpired() error {
	if p.Status == PackageStatusCancelled {
		return errors.New(errors.CodeInvalidStatus, "cannot mark cancelled package as expired")
	}

	if p.Status == PackageStatusExpired {
		return nil // 已经是过期状态，幂等处理
	}

	p.Status = PackageStatusExpired

	// 发布过期事件
	p.AddEvent(&PackageExpiredEvent{
		PackageID:      p.ID,
		PackageNo:      p.PackageNo,
		TenantID:       p.TenantID,
		UserID:         p.UserID,
		RemainingQuota: p.RemainingQuota,
		ValidTo:        p.ValidTo,
		OccurredAt:     time.Now(),
	})

	return nil
}

// MarkExhausted 标记为耗尽
func (p *Package) MarkExhausted() error {
	if p.Status == PackageStatusCancelled {
		return errors.New(errors.CodeInvalidStatus, "cannot mark cancelled package as exhausted")
	}

	if p.Status == PackageStatusExhausted {
		return nil // 已经是耗尽状态，幂等处理
	}

	p.Status = PackageStatusExhausted
	now := time.Now()
	p.ExhaustedAt = &now

	// 发布耗尽事件
	p.AddEvent(&PackageExhaustedEvent{
		PackageID:   p.ID,
		PackageNo:   p.PackageNo,
		TenantID:    p.TenantID,
		UserID:      p.UserID,
		UsedQuota:   p.UsedQuota,
		ExhaustedAt: now,
		OccurredAt:  now,
	})

	return nil
}

// Cancel 取消套餐包
func (p *Package) Cancel(reason string) error {
	if p.Status == PackageStatusCancelled {
		return nil // 已经取消，幂等处理
	}

	if p.Status == PackageStatusExhausted {
		return errors.New(errors.CodeInvalidStatus, "cannot cancel exhausted package")
	}

	oldStatus := p.Status
	p.Status = PackageStatusCancelled
	now := time.Now()
	p.CancelledAt = &now

	// 发布取消事件
	p.AddEvent(&PackageCancelledEvent{
		PackageID:      p.ID,
		PackageNo:      p.PackageNo,
		TenantID:       p.TenantID,
		UserID:         p.UserID,
		OldStatus:      oldStatus,
		RemainingQuota: p.RemainingQuota,
		Reason:         reason,
		CancelledAt:    now,
		OccurredAt:     now,
	})

	return nil
}

// IsExpired 是否已过期
func (p *Package) IsExpired() bool {
	return time.Now().After(p.ValidTo)
}

// IsAvailable 是否可用
func (p *Package) IsAvailable() bool {
	return p.Status == PackageStatusActive &&
		!p.IsExpired() &&
		p.RemainingQuota.GreaterThan(money.Zero)
}

// GetUsageRate 获取使用率（百分比）
func (p *Package) GetUsageRate() money.Decimal {
	if p.TotalQuota.Equal(money.Zero) {
		return money.Zero
	}
	// 使用率 = (已使用配额 / 总配额) * 100
	rate := money.Div(p.UsedQuota, p.TotalQuota)
	return money.Mul(rate, money.NewFromInt(100))
}

// AddEvent 添加领域事件
func (p *Package) AddEvent(event interface{}) {
	p.events = append(p.events, event)
}

// GetEvents 获取所有领域事件
func (p *Package) GetEvents() []interface{} {
	return p.events
}

// ClearEvents 清空领域事件
func (p *Package) ClearEvents() {
	p.events = []interface{}{}
}
