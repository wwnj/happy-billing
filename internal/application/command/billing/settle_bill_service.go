package command

import (
	"context"

	"github.com/wwnj/happy-billing/internal/domain/billing"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// SettleBillCommand 结算账单命令
type SettleBillCommand struct {
	BillID     string        // 账单ID
	PaidAmount money.Decimal // 支付金额
}

// SettleBillService 结算账单服务
type SettleBillService struct {
	billRepo billing.BillRepository
}

// NewSettleBillService 创建结算账单服务
func NewSettleBillService(billRepo billing.BillRepository) *SettleBillService {
	return &SettleBillService{
		billRepo: billRepo,
	}
}

// Execute 执行结算账单
func (s *SettleBillService) Execute(ctx context.Context, cmd SettleBillCommand) error {
	// 1. 参数验证
	if cmd.BillID == "" {
		return errors.NewInvalidParam("bill_id cannot be empty")
	}
	if cmd.PaidAmount.LessThanOrEqual(money.Zero) {
		return errors.NewInvalidParam("paid amount must be positive")
	}

	// 2. 查询账单
	bill, err := s.billRepo.FindByID(ctx, cmd.BillID)
	if err != nil {
		return err
	}

	// 3. 执行领域逻辑（结算）
	if err := bill.Settle(cmd.PaidAmount); err != nil {
		return err
	}

	// 4. 持久化账单
	if err := s.billRepo.Update(ctx, bill); err != nil {
		return err
	}

	// 5. 发布领域事件（TODO: 事件发布机制）
	// for _, event := range bill.GetEvents() {
	//     s.eventBus.Publish(ctx, "bill.settled", event)
	// }

	return nil
}
