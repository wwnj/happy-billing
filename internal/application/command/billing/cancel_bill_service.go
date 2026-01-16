package command

import (
	"context"

	"github.com/wwnj/happy-billing/internal/domain/billing"
	"github.com/wwnj/happy-billing/pkg/errors"
)

// CancelBillCommand 取消账单命令
type CancelBillCommand struct {
	BillID string // 账单ID
	Reason string // 取消原因
}

// CancelBillService 取消账单服务
type CancelBillService struct {
	billRepo billing.BillRepository
}

// NewCancelBillService 创建取消账单服务
func NewCancelBillService(billRepo billing.BillRepository) *CancelBillService {
	return &CancelBillService{
		billRepo: billRepo,
	}
}

// Execute 执行取消账单
func (s *CancelBillService) Execute(ctx context.Context, cmd CancelBillCommand) error {
	// 1. 参数验证
	if cmd.BillID == "" {
		return errors.NewInvalidParam("bill_id cannot be empty")
	}
	if cmd.Reason == "" {
		return errors.NewInvalidParam("reason cannot be empty")
	}

	// 2. 查询账单
	bill, err := s.billRepo.FindByID(ctx, cmd.BillID)
	if err != nil {
		return err
	}

	// 3. 执行领域逻辑（取消）
	if err := bill.Cancel(cmd.Reason); err != nil {
		return err
	}

	// 4. 持久化账单
	if err := s.billRepo.Update(ctx, bill); err != nil {
		return err
	}

	// 5. 发布领域事件（TODO: 事件发布机制）
	// for _, event := range bill.GetEvents() {
	//     s.eventBus.Publish(ctx, "bill.cancelled", event)
	// }

	return nil
}
