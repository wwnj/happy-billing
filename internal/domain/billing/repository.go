package billing

import (
	"context"
	"time"

	"github.com/wwnj/happy-billing/pkg/money"
)

// BillRepository 账单仓储接口
type BillRepository interface {
	// Save 保存账单（新建）
	Save(ctx context.Context, bill *Bill) error

	// Update 更新账单
	Update(ctx context.Context, bill *Bill) error

	// FindByID 根据ID查询账单
	FindByID(ctx context.Context, id string) (*Bill, error)

	// FindByBillNo 根据账单号查询
	FindByBillNo(ctx context.Context, billNo string) (*Bill, error)

	// ListByUser 查询用户的账单列表
	// status: 账单状态（为空则查询所有状态）
	// startTime, endTime: 账期时间范围（为空则不限制）
	// offset, limit: 分页参数
	ListByUser(ctx context.Context, tenantID, userID string, status BillStatus, startTime, endTime *time.Time, offset, limit int) ([]*Bill, int64, error)

	// ListByTenant 查询租户的账单列表
	ListByTenant(ctx context.Context, tenantID string, status BillStatus, offset, limit int) ([]*Bill, int64, error)

	// ListOverdue 查询逾期账单
	ListOverdue(ctx context.Context, asOfDate time.Time, offset, limit int) ([]*Bill, int64, error)

	// SumAmountByUser 统计用户在指定时间范围内的账单总额
	SumAmountByUser(ctx context.Context, tenantID, userID string, startTime, endTime time.Time) (money.Decimal, error)

	// SumAmountByTenant 统计租户在指定时间范围内的账单总额
	SumAmountByTenant(ctx context.Context, tenantID string, startTime, endTime time.Time) (money.Decimal, error)
}

// BillItemRepository 账单明细仓储接口
type BillItemRepository interface {
	// Save 保存账单明细
	Save(ctx context.Context, item *BillItem) error

	// BatchSave 批量保存账单明细
	BatchSave(ctx context.Context, items []*BillItem) error

	// FindByID 根据ID查询账单明细
	FindByID(ctx context.Context, id string) (*BillItem, error)

	// ListByBill 查询账单的明细列表
	ListByBill(ctx context.Context, billID string) ([]*BillItem, error)

	// ListByOrder 查询订单关联的账单明细
	ListByOrder(ctx context.Context, orderID string) ([]*BillItem, error)

	// SumAmountByBill 统计账单的明细总额
	SumAmountByBill(ctx context.Context, billID string) (money.Decimal, error)
}
