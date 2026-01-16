package pkgdomain

import (
	"context"
	"time"

	"github.com/wwnj/happy-billing/pkg/money"
)

// PackageRepository 套餐包仓储接口
type PackageRepository interface {
	// Save 保存套餐包（新增）
	Save(ctx context.Context, pkg *Package) error

	// Update 更新套餐包
	Update(ctx context.Context, pkg *Package) error

	// FindByID 根据ID查询套餐包
	FindByID(ctx context.Context, id string) (*Package, error)

	// FindByPackageNo 根据套餐包号查询
	FindByPackageNo(ctx context.Context, packageNo string) (*Package, error)

	// ListByUser 查询用户的套餐包列表
	// status: 可选，过滤状态
	// packageType: 可选，过滤类型
	// offset, limit: 分页参数
	ListByUser(
		ctx context.Context,
		tenantID, userID string,
		status PackageStatus,
		packageType PackageType,
		offset, limit int,
	) ([]*Package, int64, error)

	// ListAvailableByUser 查询用户的可用套餐包列表（状态=ACTIVE，未过期，有余额）
	// 按优先级排序：先到期的优先
	ListAvailableByUser(
		ctx context.Context,
		tenantID, userID string,
		packageType PackageType,
	) ([]*Package, error)

	// ListExpired 查询过期的套餐包（valid_to < now AND status = ACTIVE）
	// 用于定时任务批量标记过期
	ListExpired(ctx context.Context, now time.Time, limit int) ([]*Package, error)

	// SumQuotaByUser 统计用户某类型套餐包的配额汇总
	SumQuotaByUser(
		ctx context.Context,
		tenantID, userID string,
		packageType PackageType,
	) (totalQuota, usedQuota, remainingQuota money.Decimal, err error)
}
