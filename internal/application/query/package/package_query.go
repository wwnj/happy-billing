package pkgquery

import (
	"context"
	"time"

	pkgdomain "github.com/wwnj/happy-billing/internal/domain/package"
	"github.com/wwnj/happy-billing/pkg/money"
)

// PackageDTO 套餐包视图
type PackageDTO struct {
	ID             string                 `json:"id"`
	PackageNo      string                 `json:"package_no"`
	TenantID       string                 `json:"tenant_id"`
	UserID         string                 `json:"user_id"`
	Type           string                 `json:"type"`
	Status         string                 `json:"status"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	TotalQuota     string                 `json:"total_quota"`
	UsedQuota      string                 `json:"used_quota"`
	RemainingQuota string                 `json:"remaining_quota"`
	QuotaUnit      string                 `json:"quota_unit"`
	UsageRate      string                 `json:"usage_rate"` // 使用率百分比
	Price          string                 `json:"price"`
	Currency       string                 `json:"currency"`
	PurchasedAt    time.Time              `json:"purchased_at"`
	ValidFrom      time.Time              `json:"valid_from"`
	ValidTo        time.Time              `json:"valid_to"`
	ExhaustedAt    *time.Time             `json:"exhausted_at,omitempty"`
	CancelledAt    *time.Time             `json:"cancelled_at,omitempty"`
	IsAvailable    bool                   `json:"is_available"`
	IsExpired      bool                   `json:"is_expired"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ListPackagesQuery 查询套餐包列表请求
type ListPackagesQuery struct {
	TenantID string
	UserID   string
	Status   pkgdomain.PackageStatus
	Type     pkgdomain.PackageType
	Offset   int
	Limit    int
}

// ListPackagesResult 查询套餐包列表结果
type ListPackagesResult struct {
	Packages []*PackageDTO `json:"packages"`
	Total    int64         `json:"total"`
}

// PackageQuotaSummaryDTO 套餐包配额汇总视图
type PackageQuotaSummaryDTO struct {
	TenantID       string `json:"tenant_id"`
	UserID         string `json:"user_id"`
	Type           string `json:"type"`
	TotalQuota     string `json:"total_quota"`
	UsedQuota      string `json:"used_quota"`
	RemainingQuota string `json:"remaining_quota"`
	UsageRate      string `json:"usage_rate"` // 使用率百分比
}

// PackageQuery 套餐包查询服务
type PackageQuery struct {
	packageRepo pkgdomain.PackageRepository
}

// NewPackageQuery 创建套餐包查询服务
func NewPackageQuery(packageRepo pkgdomain.PackageRepository) *PackageQuery {
	return &PackageQuery{
		packageRepo: packageRepo,
	}
}

// GetPackage 查询套餐包详情
func (q *PackageQuery) GetPackage(ctx context.Context, packageID string) (*PackageDTO, error) {
	pkg, err := q.packageRepo.FindByID(ctx, packageID)
	if err != nil {
		return nil, err
	}

	return toPackageDTO(pkg), nil
}

// ListPackages 查询套餐包列表
func (q *PackageQuery) ListPackages(ctx context.Context, query ListPackagesQuery) (*ListPackagesResult, error) {
	packages, total, err := q.packageRepo.ListByUser(
		ctx,
		query.TenantID,
		query.UserID,
		query.Status,
		query.Type,
		query.Offset,
		query.Limit,
	)
	if err != nil {
		return nil, err
	}

	dtos := make([]*PackageDTO, 0, len(packages))
	for _, pkg := range packages {
		dtos = append(dtos, toPackageDTO(pkg))
	}

	return &ListPackagesResult{
		Packages: dtos,
		Total:    total,
	}, nil
}

// ListAvailablePackages 查询可用套餐包列表
func (q *PackageQuery) ListAvailablePackages(
	ctx context.Context,
	tenantID, userID string,
	packageType pkgdomain.PackageType,
) ([]*PackageDTO, error) {
	packages, err := q.packageRepo.ListAvailableByUser(ctx, tenantID, userID, packageType)
	if err != nil {
		return nil, err
	}

	dtos := make([]*PackageDTO, 0, len(packages))
	for _, pkg := range packages {
		dtos = append(dtos, toPackageDTO(pkg))
	}

	return dtos, nil
}

// GetQuotaSummary 查询配额汇总
func (q *PackageQuery) GetQuotaSummary(
	ctx context.Context,
	tenantID, userID string,
	packageType pkgdomain.PackageType,
) (*PackageQuotaSummaryDTO, error) {
	totalQuota, usedQuota, remainingQuota, err := q.packageRepo.SumQuotaByUser(
		ctx,
		tenantID,
		userID,
		packageType,
	)
	if err != nil {
		return nil, err
	}

	// 计算使用率
	usageRate := money.Zero
	if totalQuota.GreaterThan(money.Zero) {
		usageRate = money.Div(usedQuota, totalQuota)
		usageRate = money.Mul(usageRate, money.NewFromInt(100))
	}

	return &PackageQuotaSummaryDTO{
		TenantID:       tenantID,
		UserID:         userID,
		Type:           string(packageType),
		TotalQuota:     totalQuota.String(),
		UsedQuota:      usedQuota.String(),
		RemainingQuota: remainingQuota.String(),
		UsageRate:      usageRate.String(),
	}, nil
}

// toPackageDTO 转换为DTO
func toPackageDTO(pkg *pkgdomain.Package) *PackageDTO {
	return &PackageDTO{
		ID:             pkg.ID,
		PackageNo:      pkg.PackageNo,
		TenantID:       pkg.TenantID,
		UserID:         pkg.UserID,
		Type:           string(pkg.Type),
		Status:         string(pkg.Status),
		Name:           pkg.Name,
		Description:    pkg.Description,
		TotalQuota:     pkg.TotalQuota.String(),
		UsedQuota:      pkg.UsedQuota.String(),
		RemainingQuota: pkg.RemainingQuota.String(),
		QuotaUnit:      pkg.QuotaUnit,
		UsageRate:      pkg.GetUsageRate().String(),
		Price:          pkg.Price.String(),
		Currency:       pkg.Currency,
		PurchasedAt:    pkg.PurchasedAt,
		ValidFrom:      pkg.ValidFrom,
		ValidTo:        pkg.ValidTo,
		ExhaustedAt:    pkg.ExhaustedAt,
		CancelledAt:    pkg.CancelledAt,
		IsAvailable:    pkg.IsAvailable(),
		IsExpired:      pkg.IsExpired(),
		Metadata:       pkg.Metadata,
	}
}
