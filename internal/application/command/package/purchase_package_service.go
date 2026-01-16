package pkgcmd

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	pkgdomain "github.com/wwnj/happy-billing/internal/domain/package"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// PurchasePackageCommand 购买套餐包命令
type PurchasePackageCommand struct {
	TenantID    string                 // 租户ID
	UserID      string                 // 用户ID
	Type        pkgdomain.PackageType  // 套餐包类型
	Name        string                 // 套餐包名称
	Description string                 // 描述
	TotalQuota  money.Decimal          // 总配额
	QuotaUnit   string                 // 配额单位
	Price       money.Decimal          // 价格
	Currency    string                 // 货币类型
	ValidFrom   time.Time              // 生效时间
	ValidTo     time.Time              // 过期时间
	Metadata    map[string]interface{} // 扩展元数据
}

// PurchasePackageService 购买套餐包服务
type PurchasePackageService struct {
	packageRepo pkgdomain.PackageRepository
}

// NewPurchasePackageService 创建购买套餐包服务
func NewPurchasePackageService(packageRepo pkgdomain.PackageRepository) *PurchasePackageService {
	return &PurchasePackageService{
		packageRepo: packageRepo,
	}
}

// Execute 执行购买套餐包
func (s *PurchasePackageService) Execute(ctx context.Context, cmd PurchasePackageCommand) (*pkgdomain.Package, error) {
	// 1. 参数验证
	if err := s.validate(cmd); err != nil {
		return nil, err
	}

	// 2. 生成套餐包号（PKG-{YYYYMM}-{TenantID前6位}-{UserID前6位}-{随机4位}）
	packageNo := s.generatePackageNo(cmd.TenantID, cmd.UserID)

	// 3. 创建套餐包聚合根
	pkg, err := pkgdomain.NewPackage(
		uuid.New().String(),
		packageNo,
		cmd.TenantID,
		cmd.UserID,
		cmd.Type,
		cmd.Name,
		cmd.Description,
		cmd.TotalQuota,
		cmd.QuotaUnit,
		cmd.Price,
		cmd.Currency,
		cmd.ValidFrom,
		cmd.ValidTo,
	)
	if err != nil {
		return nil, err
	}

	// 设置扩展元数据
	if cmd.Metadata != nil {
		pkg.Metadata = cmd.Metadata
	}

	// 4. 持久化套餐包
	if err := s.packageRepo.Save(ctx, pkg); err != nil {
		return nil, errors.Wrap(errors.CodeInternalError, "failed to save package", err)
	}

	// 5. 发布领域事件（事件已在NewPackage中添加）
	// TODO: 发布到消息队列（Kafka）

	return pkg, nil
}

// validate 验证命令参数
func (s *PurchasePackageService) validate(cmd PurchasePackageCommand) error {
	if cmd.TenantID == "" || cmd.UserID == "" {
		return errors.NewInvalidParam("tenant_id and user_id are required")
	}

	if err := cmd.Type.Validate(); err != nil {
		return err
	}

	if cmd.Name == "" {
		return errors.NewInvalidParam("name is required")
	}

	if cmd.TotalQuota.LessThanOrEqual(money.Zero) {
		return errors.NewInvalidParam("total_quota must be greater than 0")
	}

	if cmd.QuotaUnit == "" {
		return errors.NewInvalidParam("quota_unit is required")
	}

	if cmd.Price.LessThan(money.Zero) {
		return errors.NewInvalidParam("price must be greater than or equal to 0")
	}

	if cmd.Currency == "" {
		cmd.Currency = "CNY"
	}

	if cmd.ValidFrom.IsZero() || cmd.ValidTo.IsZero() {
		return errors.NewInvalidParam("valid_from and valid_to are required")
	}

	if cmd.ValidFrom.After(cmd.ValidTo) {
		return errors.NewInvalidParam("valid_from must be before valid_to")
	}

	return nil
}

// generatePackageNo 生成套餐包号
func (s *PurchasePackageService) generatePackageNo(tenantID, userID string) string {
	// 格式：PKG-{YYYYMM}-{TenantID前6位}-{UserID前6位}-{随机4位}
	now := time.Now()

	tenantPrefix := tenantID
	if len(tenantID) > 6 {
		tenantPrefix = tenantID[:6]
	}

	userPrefix := userID
	if len(userID) > 6 {
		userPrefix = userID[:6]
	}

	randomSuffix := uuid.New().String()[:4]

	return fmt.Sprintf("PKG-%s-%s-%s-%s",
		now.Format("200601"),
		tenantPrefix,
		userPrefix,
		randomSuffix,
	)
}
