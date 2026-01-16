package query

import (
	"context"
	"time"

	"github.com/wwnj/happy-billing/internal/domain/billing"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// BillQuery 账单查询服务
type BillQuery struct {
	billRepo     billing.BillRepository
	billItemRepo billing.BillItemRepository
}

// NewBillQuery 创建账单查询服务
func NewBillQuery(
	billRepo billing.BillRepository,
	billItemRepo billing.BillItemRepository,
) *BillQuery {
	return &BillQuery{
		billRepo:     billRepo,
		billItemRepo: billItemRepo,
	}
}

// BillDTO 账单DTO
type BillDTO struct {
	ID                string             `json:"id"`
	BillNo            string             `json:"bill_no"`
	TenantID          string             `json:"tenant_id"`
	UserID            string             `json:"user_id"`
	Cycle             billing.BillCycle  `json:"cycle"`
	Status            billing.BillStatus `json:"status"`
	PeriodStart       time.Time          `json:"period_start"`
	PeriodEnd         time.Time          `json:"period_end"`
	TotalAmount       money.Decimal      `json:"total_amount"`
	DiscountAmount    money.Decimal      `json:"discount_amount"`
	TaxAmount         money.Decimal      `json:"tax_amount"`
	ActualAmount      money.Decimal      `json:"actual_amount"`
	PaidAmount        money.Decimal      `json:"paid_amount"`
	OutstandingAmount money.Decimal      `json:"outstanding_amount"`
	Currency          string             `json:"currency"`
	DueDate           *time.Time         `json:"due_date,omitempty"`
	SettledAt         *time.Time         `json:"settled_at,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// BillItemDTO 账单明细DTO
type BillItemDTO struct {
	ID          string               `json:"id"`
	BillID      string               `json:"bill_id"`
	Type        billing.BillItemType `json:"type"`
	OrderID     *string              `json:"order_id,omitempty"`
	Description string               `json:"description"`
	Amount      money.Decimal        `json:"amount"`
	Quantity    money.Decimal        `json:"quantity"`
	UnitPrice   money.Decimal        `json:"unit_price"`
	Discount    money.Decimal        `json:"discount"`
	TaxAmount   money.Decimal        `json:"tax_amount"`
	TotalAmount money.Decimal        `json:"total_amount"`
	CreatedAt   time.Time            `json:"created_at"`
}

// BillDetailDTO 账单详情DTO
type BillDetailDTO struct {
	Bill  *BillDTO       `json:"bill"`
	Items []*BillItemDTO `json:"items"`
}

// GetBill 查询账单
func (q *BillQuery) GetBill(ctx context.Context, billID string) (*BillDTO, error) {
	bill, err := q.billRepo.FindByID(ctx, billID)
	if err != nil {
		return nil, err
	}

	return toBillDTO(bill), nil
}

// GetBillDetail 查询账单详情（包含明细）
func (q *BillQuery) GetBillDetail(ctx context.Context, billID string) (*BillDetailDTO, error) {
	// 查询账单
	bill, err := q.billRepo.FindByID(ctx, billID)
	if err != nil {
		return nil, err
	}

	// 查询账单明细
	items, err := q.billItemRepo.ListByBill(ctx, billID)
	if err != nil {
		return nil, err
	}

	// 转换为DTO
	itemDTOs := make([]*BillItemDTO, 0, len(items))
	for _, item := range items {
		itemDTOs = append(itemDTOs, toBillItemDTO(item))
	}

	return &BillDetailDTO{
		Bill:  toBillDTO(bill),
		Items: itemDTOs,
	}, nil
}

// ListBillsQuery 账单列表查询参数
type ListBillsQuery struct {
	TenantID  string             // 租户ID
	UserID    string             // 用户ID（为空则查询租户所有账单）
	Status    billing.BillStatus // 账单状态
	StartTime *time.Time         // 账期开始时间
	EndTime   *time.Time         // 账期结束时间
	Offset    int                // 偏移量
	Limit     int                // 每页数量
}

// ListBillsResult 账单列表查询结果
type ListBillsResult struct {
	Bills  []*BillDTO `json:"bills"`
	Total  int64      `json:"total"`
	Offset int        `json:"offset"`
	Limit  int        `json:"limit"`
}

// ListBills 查询账单列表
func (q *BillQuery) ListBills(ctx context.Context, query ListBillsQuery) (*ListBillsResult, error) {
	// 参数验证
	if query.TenantID == "" {
		return nil, errors.NewInvalidParam("tenant_id cannot be empty")
	}
	if query.Limit <= 0 {
		query.Limit = 20 // 默认每页20条
	}
	if query.Limit > 100 {
		query.Limit = 100 // 最大每页100条
	}

	var bills []*billing.Bill
	var total int64
	var err error

	// 根据是否有用户ID选择查询方式
	if query.UserID != "" {
		bills, total, err = q.billRepo.ListByUser(
			ctx,
			query.TenantID,
			query.UserID,
			query.Status,
			query.StartTime,
			query.EndTime,
			query.Offset,
			query.Limit,
		)
	} else {
		bills, total, err = q.billRepo.ListByTenant(
			ctx,
			query.TenantID,
			query.Status,
			query.Offset,
			query.Limit,
		)
	}

	if err != nil {
		return nil, err
	}

	// 转换为DTO
	dtos := make([]*BillDTO, 0, len(bills))
	for _, bill := range bills {
		dtos = append(dtos, toBillDTO(bill))
	}

	return &ListBillsResult{
		Bills:  dtos,
		Total:  total,
		Offset: query.Offset,
		Limit:  query.Limit,
	}, nil
}

// ListOverdueBills 查询逾期账单
func (q *BillQuery) ListOverdueBills(ctx context.Context, asOfDate time.Time, offset, limit int) (*ListBillsResult, error) {
	bills, total, err := q.billRepo.ListOverdue(ctx, asOfDate, offset, limit)
	if err != nil {
		return nil, err
	}

	// 转换为DTO
	dtos := make([]*BillDTO, 0, len(bills))
	for _, bill := range bills {
		dtos = append(dtos, toBillDTO(bill))
	}

	return &ListBillsResult{
		Bills:  dtos,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}, nil
}

// SumBillAmount 统计账单总额
func (q *BillQuery) SumBillAmount(
	ctx context.Context,
	tenantID, userID string,
	startTime, endTime time.Time,
) (money.Decimal, error) {
	if userID != "" {
		return q.billRepo.SumAmountByUser(ctx, tenantID, userID, startTime, endTime)
	}
	return q.billRepo.SumAmountByTenant(ctx, tenantID, startTime, endTime)
}

// 转换函数

func toBillDTO(bill *billing.Bill) *BillDTO {
	return &BillDTO{
		ID:                bill.ID,
		BillNo:            bill.BillNo,
		TenantID:          bill.TenantID,
		UserID:            bill.UserID,
		Cycle:             bill.Cycle,
		Status:            bill.Status,
		PeriodStart:       bill.PeriodStart,
		PeriodEnd:         bill.PeriodEnd,
		TotalAmount:       bill.TotalAmount,
		DiscountAmount:    bill.DiscountAmount,
		TaxAmount:         bill.TaxAmount,
		ActualAmount:      bill.ActualAmount,
		PaidAmount:        bill.PaidAmount,
		OutstandingAmount: bill.OutstandingAmount,
		Currency:          bill.Currency,
		DueDate:           bill.DueDate,
		SettledAt:         bill.SettledAt,
		CreatedAt:         bill.CreatedAt,
		UpdatedAt:         bill.UpdatedAt,
	}
}

func toBillItemDTO(item *billing.BillItem) *BillItemDTO {
	return &BillItemDTO{
		ID:          item.ID,
		BillID:      item.BillID,
		Type:        item.Type,
		OrderID:     item.OrderID,
		Description: item.Description,
		Amount:      item.Amount,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		Discount:    item.Discount,
		TaxAmount:   item.TaxAmount,
		TotalAmount: item.TotalAmount,
		CreatedAt:   item.CreatedAt,
	}
}
