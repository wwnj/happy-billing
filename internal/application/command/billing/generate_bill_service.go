package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/wwnj/happy-billing/internal/domain/billing"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// OrderData 订单数据（简化版，用于账单生成）
type OrderData struct {
	ID          string
	Amount      money.Decimal
	Description string
	CreatedAt   time.Time
}

// OrderQueryService 订单查询服务接口（用于账单生成）
type OrderQueryService interface {
	// ListOrdersByUserInPeriod 查询用户在指定时间范围内的订单
	ListOrdersByUserInPeriod(ctx context.Context, tenantID, userID string, startTime, endTime time.Time) ([]*OrderData, error)
}

// GenerateBillCommand 生成账单命令
type GenerateBillCommand struct {
	TenantID    string            // 租户ID
	UserID      string            // 用户ID
	Cycle       billing.BillCycle // 账单周期
	PeriodStart time.Time         // 账期开始时间
	PeriodEnd   time.Time         // 账期结束时间
	DueDate     *time.Time        // 到期日期（可选）
	Currency    string            // 货币类型
}

// GenerateBillService 生成账单服务
type GenerateBillService struct {
	billRepo     billing.BillRepository
	billItemRepo billing.BillItemRepository
	orderQuery   OrderQueryService
}

// NewGenerateBillService 创建生成账单服务
func NewGenerateBillService(
	billRepo billing.BillRepository,
	billItemRepo billing.BillItemRepository,
	orderQuery OrderQueryService,
) *GenerateBillService {
	return &GenerateBillService{
		billRepo:     billRepo,
		billItemRepo: billItemRepo,
		orderQuery:   orderQuery,
	}
}

// Execute 执行生成账单
func (s *GenerateBillService) Execute(ctx context.Context, cmd GenerateBillCommand) (*billing.Bill, error) {
	// 1. 参数验证
	if err := s.validateCommand(cmd); err != nil {
		return nil, err
	}

	// 2. 查询账期内的订单数据
	orders, err := s.orderQuery.ListOrdersByUserInPeriod(
		ctx,
		cmd.TenantID,
		cmd.UserID,
		cmd.PeriodStart,
		cmd.PeriodEnd,
	)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternalError, "query orders failed", err)
	}

	// 如果没有订单，不生成账单
	if len(orders) == 0 {
		return nil, errors.New(errors.CodeInvalidParam, "no orders found in the period")
	}

	// 3. 生成账单号
	billNo := s.generateBillNo(cmd.TenantID, cmd.UserID, cmd.PeriodStart)

	// 4. 创建账单聚合根
	bill, err := billing.NewBill(
		uuid.New().String(),
		billNo,
		cmd.TenantID,
		cmd.UserID,
		cmd.Cycle,
		cmd.PeriodStart,
		cmd.PeriodEnd,
		cmd.Currency,
	)
	if err != nil {
		return nil, err
	}

	// 5. 设置到期日期
	if cmd.DueDate != nil {
		if err := bill.SetDueDate(*cmd.DueDate); err != nil {
			return nil, err
		}
	}

	// 6. 添加账单明细
	items := make([]*billing.BillItem, 0, len(orders))
	for _, order := range orders {
		item, err := billing.NewBillItem(
			uuid.New().String(),
			bill.ID,
			billing.BillItemTypeOrder,
			&order.ID,
			order.Description,
			order.Amount,
			nil, // 暂不使用元数据
		)
		if err != nil {
			return nil, errors.Wrap(errors.CodeInternalError, "create bill item failed", err)
		}

		if err := bill.AddItem(item); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	// 7. 持久化账单
	if err := s.billRepo.Save(ctx, bill); err != nil {
		return nil, err
	}

	// 8. 批量保存账单明细
	if err := s.billItemRepo.BatchSave(ctx, items); err != nil {
		// 如果明细保存失败，需要回滚账单（简化处理，实际应使用事务）
		return nil, errors.Wrap(errors.CodeInternalError, "save bill items failed", err)
	}

	// 9. 发布领域事件（TODO: 事件发布机制）
	// for _, event := range bill.GetEvents() {
	//     s.eventBus.Publish(ctx, "bill.generated", event)
	// }

	return bill, nil
}

// validateCommand 验证命令参数
func (s *GenerateBillService) validateCommand(cmd GenerateBillCommand) error {
	if cmd.TenantID == "" {
		return errors.NewInvalidParam("tenant_id cannot be empty")
	}
	if cmd.UserID == "" {
		return errors.NewInvalidParam("user_id cannot be empty")
	}
	if err := cmd.Cycle.Validate(); err != nil {
		return err
	}
	if cmd.PeriodStart.After(cmd.PeriodEnd) {
		return errors.NewInvalidParam("period_start must be before period_end")
	}
	if cmd.Currency == "" {
		cmd.Currency = "CNY"
	}

	return nil
}

// generateBillNo 生成账单号
// 格式: BILL-{YYYYMM}-{TenantID}-{UserID}
func (s *GenerateBillService) generateBillNo(tenantID, userID string, periodStart time.Time) string {
	return fmt.Sprintf("BILL-%s-%s-%s",
		periodStart.Format("200601"),
		tenantID[:8],
		userID[:8],
	)
}

// GenerateBillsForTenant 批量生成租户下所有用户的账单
func (s *GenerateBillService) GenerateBillsForTenant(
	ctx context.Context,
	tenantID string,
	cycle billing.BillCycle,
	periodStart, periodEnd time.Time,
) ([]*billing.Bill, error) {
	// TODO: 查询租户下所有有订单的用户
	// TODO: 为每个用户生成账单
	// 这是一个批量操作，实际应该使用消息队列异步处理

	return nil, errors.New(errors.CodeInternalError, "not implemented yet")
}
