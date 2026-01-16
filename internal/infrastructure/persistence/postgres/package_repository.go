package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	pkgdomain "github.com/wwnj/happy-billing/internal/domain/package"
	"github.com/wwnj/happy-billing/internal/infrastructure/persistence/postgres/model"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// PackageRepository 套餐包仓储PostgreSQL实现
type PackageRepository struct {
	db *gorm.DB
}

// NewPackageRepository 创建套餐包仓储
func NewPackageRepository(db *gorm.DB) *PackageRepository {
	return &PackageRepository{
		db: db,
	}
}

// Save 保存套餐包（新增）
func (r *PackageRepository) Save(ctx context.Context, pkg *pkgdomain.Package) error {
	var do model.PackageDO
	do.FromDomain(pkg)

	if err := r.db.WithContext(ctx).Create(&do).Error; err != nil {
		return errors.NewDatabaseError("save package", err)
	}

	return nil
}

// Update 更新套餐包
func (r *PackageRepository) Update(ctx context.Context, pkg *pkgdomain.Package) error {
	var do model.PackageDO
	do.FromDomain(pkg)

	result := r.db.WithContext(ctx).
		Model(&model.PackageDO{}).
		Where("id = ?", pkg.ID).
		Updates(map[string]interface{}{
			"status":          do.Status,
			"used_quota":      do.UsedQuota,
			"remaining_quota": do.RemainingQuota,
			"exhausted_at":    do.ExhaustedAt,
			"cancelled_at":    do.CancelledAt,
			"metadata":        do.Metadata,
			"updated_at":      time.Now(),
		})

	if result.Error != nil {
		return errors.NewDatabaseError("update package", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New(errors.CodePackageNotFound, "package not found")
	}

	return nil
}

// FindByID 根据ID查询套餐包
func (r *PackageRepository) FindByID(ctx context.Context, id string) (*pkgdomain.Package, error) {
	var do model.PackageDO

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&do).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodePackageNotFound, "package not found")
		}
		return nil, errors.NewDatabaseError("find package by id", err)
	}

	return do.ToDomain()
}

// FindByPackageNo 根据套餐包号查询
func (r *PackageRepository) FindByPackageNo(ctx context.Context, packageNo string) (*pkgdomain.Package, error) {
	var do model.PackageDO

	if err := r.db.WithContext(ctx).Where("package_no = ?", packageNo).First(&do).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodePackageNotFound, "package not found")
		}
		return nil, errors.NewDatabaseError("find package by package_no", err)
	}

	return do.ToDomain()
}

// ListByUser 查询用户的套餐包列表
func (r *PackageRepository) ListByUser(
	ctx context.Context,
	tenantID, userID string,
	status pkgdomain.PackageStatus,
	packageType pkgdomain.PackageType,
	offset, limit int,
) ([]*pkgdomain.Package, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.PackageDO{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID)

	// 可选过滤条件
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if packageType != "" {
		query = query.Where("type = ?", packageType)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("count packages", err)
	}

	// 查询数据
	var dos []model.PackageDO
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&dos).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("list packages", err)
	}

	// 转换为领域模型
	packages := make([]*pkgdomain.Package, 0, len(dos))
	for _, do := range dos {
		pkg, err := do.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		packages = append(packages, pkg)
	}

	return packages, total, nil
}

// ListAvailableByUser 查询用户的可用套餐包列表（状态=ACTIVE，未过期，有余额）
func (r *PackageRepository) ListAvailableByUser(
	ctx context.Context,
	tenantID, userID string,
	packageType pkgdomain.PackageType,
) ([]*pkgdomain.Package, error) {
	now := time.Now()

	query := r.db.WithContext(ctx).Model(&model.PackageDO{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Where("status = ?", pkgdomain.PackageStatusActive).
		Where("valid_from <= ? AND valid_to > ?", now, now).
		Where("remaining_quota > ?", "0")

	// 可选过滤类型
	if packageType != "" {
		query = query.Where("type = ?", packageType)
	}

	// 按到期时间排序，先到期的优先使用
	var dos []model.PackageDO
	if err := query.Order("valid_to ASC").Find(&dos).Error; err != nil {
		return nil, errors.NewDatabaseError("list available packages", err)
	}

	// 转换为领域模型
	packages := make([]*pkgdomain.Package, 0, len(dos))
	for _, do := range dos {
		pkg, err := do.ToDomain()
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}

// ListExpired 查询过期的套餐包（valid_to < now AND status = ACTIVE）
func (r *PackageRepository) ListExpired(ctx context.Context, now time.Time, limit int) ([]*pkgdomain.Package, error) {
	var dos []model.PackageDO

	if err := r.db.WithContext(ctx).
		Where("status = ?", pkgdomain.PackageStatusActive).
		Where("valid_to < ?", now).
		Limit(limit).
		Find(&dos).Error; err != nil {
		return nil, errors.NewDatabaseError("list expired packages", err)
	}

	// 转换为领域模型
	packages := make([]*pkgdomain.Package, 0, len(dos))
	for _, do := range dos {
		pkg, err := do.ToDomain()
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}

// SumQuotaByUser 统计用户某类型套餐包的配额汇总
func (r *PackageRepository) SumQuotaByUser(
	ctx context.Context,
	tenantID, userID string,
	packageType pkgdomain.PackageType,
) (totalQuota, usedQuota, remainingQuota money.Decimal, err error) {
	type QuotaSum struct {
		TotalQuota     string
		UsedQuota      string
		RemainingQuota string
	}

	var result QuotaSum

	query := r.db.WithContext(ctx).Model(&model.PackageDO{}).
		Select("COALESCE(SUM(total_quota), 0) as total_quota, COALESCE(SUM(used_quota), 0) as used_quota, COALESCE(SUM(remaining_quota), 0) as remaining_quota").
		Where("tenant_id = ? AND user_id = ?", tenantID, userID)

	// 可选过滤类型
	if packageType != "" {
		query = query.Where("type = ?", packageType)
	}

	// 只统计激活状态的套餐包
	query = query.Where("status = ?", pkgdomain.PackageStatusActive)

	if err := query.Scan(&result).Error; err != nil {
		return money.Zero, money.Zero, money.Zero, errors.NewDatabaseError("sum package quota", err)
	}

	totalQuota, err = money.NewFromString(result.TotalQuota)
	if err != nil {
		return money.Zero, money.Zero, money.Zero, err
	}

	usedQuota, err = money.NewFromString(result.UsedQuota)
	if err != nil {
		return money.Zero, money.Zero, money.Zero, err
	}

	remainingQuota, err = money.NewFromString(result.RemainingQuota)
	if err != nil {
		return money.Zero, money.Zero, money.Zero, err
	}

	return totalQuota, usedQuota, remainingQuota, nil
}
