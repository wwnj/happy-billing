package pkgcmd

import (
	"context"
	"time"

	pkgdomain "github.com/wwnj/happy-billing/internal/domain/package"
	"github.com/wwnj/happy-billing/pkg/errors"
)

// ExpirePackageService 过期套餐包处理服务
// 用于定时任务，批量标记过期的套餐包
type ExpirePackageService struct {
	packageRepo pkgdomain.PackageRepository
}

// NewExpirePackageService 创建过期套餐包处理服务
func NewExpirePackageService(packageRepo pkgdomain.PackageRepository) *ExpirePackageService {
	return &ExpirePackageService{
		packageRepo: packageRepo,
	}
}

// Execute 执行过期处理
// 逻辑：
// 1. 查询过期的套餐包（status=ACTIVE && valid_to < now）
// 2. 批量标记为EXPIRED状态
// 3. 发布领域事件
func (s *ExpirePackageService) Execute(ctx context.Context, batchSize int) (int, error) {
	if batchSize <= 0 {
		batchSize = 100 // 默认每批处理100个
	}

	now := time.Now()

	// 1. 查询过期的套餐包
	packages, err := s.packageRepo.ListExpired(ctx, now, batchSize)
	if err != nil {
		return 0, errors.Wrap(errors.CodeInternalError, "failed to list expired packages", err)
	}

	if len(packages) == 0 {
		return 0, nil
	}

	// 2. 批量标记为过期
	expiredCount := 0
	for _, pkg := range packages {
		// 执行领域逻辑（标记过期）
		if err := pkg.MarkExpired(); err != nil {
			// 记录错误但继续处理其他套餐包
			continue
		}

		// 持久化
		if err := s.packageRepo.Update(ctx, pkg); err != nil {
			// 记录错误但继续处理其他套餐包
			continue
		}

		expiredCount++

		// TODO: 发布领域事件到消息队列（Kafka）
	}

	return expiredCount, nil
}
