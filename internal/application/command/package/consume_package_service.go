package pkgcmd

import (
	"context"

	pkgdomain "github.com/wwnj/happy-billing/internal/domain/package"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// ConsumePackageCommand 消费套餐包命令
type ConsumePackageCommand struct {
	TenantID    string                // 租户ID
	UserID      string                // 用户ID
	Type        pkgdomain.PackageType // 套餐包类型
	Quantity    money.Decimal         // 消费数量
	OrderID     string                // 关联订单ID
	Description string                // 描述
}

// ConsumePackageResult 消费套餐包结果
type ConsumePackageResult struct {
	ConsumedFromPackage money.Decimal // 从套餐包消费的数量
	RemainingQuantity   money.Decimal // 剩余需要消费的数量（需要从余额扣费）
	PackagesUsed        []string      // 使用的套餐包ID列表
}

// ConsumePackageService 消费套餐包服务
type ConsumePackageService struct {
	packageRepo pkgdomain.PackageRepository
}

// NewConsumePackageService 创建消费套餐包服务
func NewConsumePackageService(packageRepo pkgdomain.PackageRepository) *ConsumePackageService {
	return &ConsumePackageService{
		packageRepo: packageRepo,
	}
}

// Execute 执行消费套餐包
// 逻辑：
// 1. 查询用户可用的套餐包（按到期时间排序）
// 2. 依次从套餐包中扣除配额，直到满足需求或套餐包耗尽
// 3. 返回从套餐包消费的数量和剩余需要消费的数量
func (s *ConsumePackageService) Execute(ctx context.Context, cmd ConsumePackageCommand) (*ConsumePackageResult, error) {
	// 1. 参数验证
	if err := s.validate(cmd); err != nil {
		return nil, err
	}

	// 2. 查询用户可用的套餐包（按到期时间排序，先到期的优先）
	packages, err := s.packageRepo.ListAvailableByUser(ctx, cmd.TenantID, cmd.UserID, cmd.Type)
	if err != nil {
		return nil, err
	}

	// 如果没有可用套餐包，直接返回（全部从余额扣费）
	if len(packages) == 0 {
		return &ConsumePackageResult{
			ConsumedFromPackage: money.Zero,
			RemainingQuantity:   cmd.Quantity,
			PackagesUsed:        []string{},
		}, nil
	}

	// 3. 依次从套餐包中扣除配额
	remainingQuantity := cmd.Quantity
	consumedTotal := money.Zero
	packagesUsed := []string{}

	for _, pkg := range packages {
		// 如果已经满足需求，跳出循环
		if remainingQuantity.LessThanOrEqual(money.Zero) {
			break
		}

		// 计算本次从该套餐包消费的数量
		var consumeAmount money.Decimal
		if pkg.RemainingQuota.GreaterThanOrEqual(remainingQuantity) {
			// 套餐包余额充足，全部从套餐包扣除
			consumeAmount = remainingQuantity
		} else {
			// 套餐包余额不足，只扣除剩余的
			consumeAmount = pkg.RemainingQuota
		}

		// 执行消费（领域逻辑）
		if err := pkg.Consume(consumeAmount, cmd.OrderID, cmd.Description); err != nil {
			// 如果消费失败（如套餐包状态变化），记录日志但继续尝试下一个套餐包
			continue
		}

		// 持久化套餐包
		if err := s.packageRepo.Update(ctx, pkg); err != nil {
			return nil, errors.Wrap(errors.CodeInternalError, "failed to update package", err)
		}

		// 统计
		consumedTotal = money.Add(consumedTotal, consumeAmount)
		remainingQuantity = money.Sub(remainingQuantity, consumeAmount)
		packagesUsed = append(packagesUsed, pkg.ID)

		// TODO: 发布领域事件到消息队列（Kafka）
	}

	// 4. 返回结果
	return &ConsumePackageResult{
		ConsumedFromPackage: consumedTotal,
		RemainingQuantity:   remainingQuantity,
		PackagesUsed:        packagesUsed,
	}, nil
}

// validate 验证命令参数
func (s *ConsumePackageService) validate(cmd ConsumePackageCommand) error {
	if cmd.TenantID == "" || cmd.UserID == "" {
		return errors.NewInvalidParam("tenant_id and user_id are required")
	}

	if err := cmd.Type.Validate(); err != nil {
		return err
	}

	if cmd.Quantity.LessThanOrEqual(money.Zero) {
		return errors.NewInvalidParam("quantity must be greater than 0")
	}

	if cmd.OrderID == "" {
		return errors.NewInvalidParam("order_id is required")
	}

	return nil
}
