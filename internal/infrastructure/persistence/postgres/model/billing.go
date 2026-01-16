package model

import (
	"database/sql"
	"time"

	"github.com/wwnj/happy-billing/internal/domain/billing"
	"github.com/wwnj/happy-billing/pkg/money"
)

// BillDO 账单数据对象
type BillDO struct {
	ID                string       `gorm:"column:id;primaryKey;type:varchar(64)"`
	BillNo            string       `gorm:"column:bill_no;uniqueIndex:uk_bill_no;type:varchar(64);not null"`
	TenantID          string       `gorm:"column:tenant_id;type:varchar(64);not null;index:idx_tenant_user"`
	UserID            string       `gorm:"column:user_id;type:varchar(64);not null;index:idx_tenant_user"`
	Cycle             string       `gorm:"column:cycle;type:varchar(20);not null"`
	Status            string       `gorm:"column:status;type:varchar(20);not null;index:idx_status"`
	PeriodStart       time.Time    `gorm:"column:period_start;not null;index:idx_period"`
	PeriodEnd         time.Time    `gorm:"column:period_end;not null;index:idx_period"`
	TotalAmount       string       `gorm:"column:total_amount;type:decimal(20,4);not null;default:0"`
	DiscountAmount    string       `gorm:"column:discount_amount;type:decimal(20,4);not null;default:0"`
	TaxAmount         string       `gorm:"column:tax_amount;type:decimal(20,4);not null;default:0"`
	ActualAmount      string       `gorm:"column:actual_amount;type:decimal(20,4);not null;default:0"`
	PaidAmount        string       `gorm:"column:paid_amount;type:decimal(20,4);not null;default:0"`
	OutstandingAmount string       `gorm:"column:outstanding_amount;type:decimal(20,4);not null;default:0"`
	Currency          string       `gorm:"column:currency;type:varchar(10);not null;default:'CNY'"`
	DueDate           sql.NullTime `gorm:"column:due_date;index:idx_due_date"`
	SettledAt         sql.NullTime `gorm:"column:settled_at"`
	CreatedAt         time.Time    `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt         time.Time    `gorm:"column:updated_at;not null;autoUpdateTime"`
}

// TableName 指定表名
func (BillDO) TableName() string {
	return "bills"
}

// ToDomain 转换为领域对象
func (do *BillDO) ToDomain() (*billing.Bill, error) {
	totalAmount, err := money.NewFromString(do.TotalAmount)
	if err != nil {
		return nil, err
	}

	discountAmount, err := money.NewFromString(do.DiscountAmount)
	if err != nil {
		return nil, err
	}

	taxAmount, err := money.NewFromString(do.TaxAmount)
	if err != nil {
		return nil, err
	}

	actualAmount, err := money.NewFromString(do.ActualAmount)
	if err != nil {
		return nil, err
	}

	paidAmount, err := money.NewFromString(do.PaidAmount)
	if err != nil {
		return nil, err
	}

	outstandingAmount, err := money.NewFromString(do.OutstandingAmount)
	if err != nil {
		return nil, err
	}

	bill := &billing.Bill{
		ID:                do.ID,
		BillNo:            do.BillNo,
		TenantID:          do.TenantID,
		UserID:            do.UserID,
		Cycle:             billing.BillCycle(do.Cycle),
		Status:            billing.BillStatus(do.Status),
		PeriodStart:       do.PeriodStart,
		PeriodEnd:         do.PeriodEnd,
		TotalAmount:       totalAmount,
		DiscountAmount:    discountAmount,
		TaxAmount:         taxAmount,
		ActualAmount:      actualAmount,
		PaidAmount:        paidAmount,
		OutstandingAmount: outstandingAmount,
		Currency:          do.Currency,
		CreatedAt:         do.CreatedAt,
		UpdatedAt:         do.UpdatedAt,
	}

	if do.DueDate.Valid {
		bill.DueDate = &do.DueDate.Time
	}

	if do.SettledAt.Valid {
		bill.SettledAt = &do.SettledAt.Time
	}

	return bill, nil
}

// FromDomainBill 从领域对象转换
func FromDomainBill(bill *billing.Bill) *BillDO {
	do := &BillDO{
		ID:                bill.ID,
		BillNo:            bill.BillNo,
		TenantID:          bill.TenantID,
		UserID:            bill.UserID,
		Cycle:             string(bill.Cycle),
		Status:            string(bill.Status),
		PeriodStart:       bill.PeriodStart,
		PeriodEnd:         bill.PeriodEnd,
		TotalAmount:       bill.TotalAmount.String(),
		DiscountAmount:    bill.DiscountAmount.String(),
		TaxAmount:         bill.TaxAmount.String(),
		ActualAmount:      bill.ActualAmount.String(),
		PaidAmount:        bill.PaidAmount.String(),
		OutstandingAmount: bill.OutstandingAmount.String(),
		Currency:          bill.Currency,
		CreatedAt:         bill.CreatedAt,
		UpdatedAt:         bill.UpdatedAt,
	}

	if bill.DueDate != nil {
		do.DueDate = sql.NullTime{Time: *bill.DueDate, Valid: true}
	}

	if bill.SettledAt != nil {
		do.SettledAt = sql.NullTime{Time: *bill.SettledAt, Valid: true}
	}

	return do
}

// BillItemDO 账单明细数据对象
type BillItemDO struct {
	ID          string    `gorm:"column:id;primaryKey;type:varchar(64)"`
	BillID      string    `gorm:"column:bill_id;type:varchar(64);not null;index:idx_bill_id"`
	Type        string    `gorm:"column:type;type:varchar(20);not null"`
	OrderID     *string   `gorm:"column:order_id;type:varchar(64);index:idx_order_id"`
	Description string    `gorm:"column:description;type:text"`
	Amount      string    `gorm:"column:amount;type:decimal(20,4);not null"`
	Quantity    string    `gorm:"column:quantity;type:decimal(20,4);not null;default:0"`
	UnitPrice   string    `gorm:"column:unit_price;type:decimal(20,4);not null;default:0"`
	Discount    string    `gorm:"column:discount;type:decimal(20,4);not null;default:0"`
	TaxAmount   string    `gorm:"column:tax_amount;type:decimal(20,4);not null;default:0"`
	TotalAmount string    `gorm:"column:total_amount;type:decimal(20,4);not null"`
	MetaData    string    `gorm:"column:metadata;type:jsonb"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;autoCreateTime"`
}

// TableName 指定表名
func (BillItemDO) TableName() string {
	return "bill_items"
}

// ToDomain 转换为领域对象
func (do *BillItemDO) ToDomain() (*billing.BillItem, error) {
	amount, err := money.NewFromString(do.Amount)
	if err != nil {
		return nil, err
	}

	quantity, err := money.NewFromString(do.Quantity)
	if err != nil {
		return nil, err
	}

	unitPrice, err := money.NewFromString(do.UnitPrice)
	if err != nil {
		return nil, err
	}

	discount, err := money.NewFromString(do.Discount)
	if err != nil {
		return nil, err
	}

	taxAmount, err := money.NewFromString(do.TaxAmount)
	if err != nil {
		return nil, err
	}

	totalAmount, err := money.NewFromString(do.TotalAmount)
	if err != nil {
		return nil, err
	}

	// TODO: 解析JSONB字段
	var metadata map[string]interface{}

	item := &billing.BillItem{
		ID:          do.ID,
		BillID:      do.BillID,
		Type:        billing.BillItemType(do.Type),
		OrderID:     do.OrderID,
		Description: do.Description,
		Amount:      amount,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		Discount:    discount,
		TaxAmount:   taxAmount,
		TotalAmount: totalAmount,
		MetaData:    metadata,
		CreatedAt:   do.CreatedAt,
	}

	return item, nil
}

// FromDomainBillItem 从领域对象转换
func FromDomainBillItem(item *billing.BillItem) *BillItemDO {
	// TODO: 序列化metadata为JSONB
	metadataJSON := "{}"

	return &BillItemDO{
		ID:          item.ID,
		BillID:      item.BillID,
		Type:        string(item.Type),
		OrderID:     item.OrderID,
		Description: item.Description,
		Amount:      item.Amount.String(),
		Quantity:    item.Quantity.String(),
		UnitPrice:   item.UnitPrice.String(),
		Discount:    item.Discount.String(),
		TaxAmount:   item.TaxAmount.String(),
		TotalAmount: item.TotalAmount.String(),
		MetaData:    metadataJSON,
		CreatedAt:   item.CreatedAt,
	}
}
