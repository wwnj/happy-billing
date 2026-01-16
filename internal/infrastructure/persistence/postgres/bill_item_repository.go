package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/wwnj/happy-billing/internal/domain/billing"
	"github.com/wwnj/happy-billing/internal/infrastructure/persistence/postgres/model"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// BillItemRepository 账单明细仓储PostgreSQL实现
type BillItemRepository struct {
	db *gorm.DB
}

// NewBillItemRepository 创建账单明细仓储
func NewBillItemRepository(db *gorm.DB) *BillItemRepository {
	return &BillItemRepository{
		db: db,
	}
}

// Save 保存账单明细
func (r *BillItemRepository) Save(ctx context.Context, item *billing.BillItem) error {
	do := model.FromDomainBillItem(item)

	if err := r.db.WithContext(ctx).Create(do).Error; err != nil {
		return errors.NewDatabaseError("save bill item", err)
	}

	return nil
}

// BatchSave 批量保存账单明细
func (r *BillItemRepository) BatchSave(ctx context.Context, items []*billing.BillItem) error {
	if len(items) == 0 {
		return nil
	}

	// 转换为DO
	dos := make([]model.BillItemDO, 0, len(items))
	for _, item := range items {
		dos = append(dos, *model.FromDomainBillItem(item))
	}

	// 批量插入
	if err := r.db.WithContext(ctx).CreateInBatches(dos, 100).Error; err != nil {
		return errors.NewDatabaseError("batch save bill items", err)
	}

	return nil
}

// FindByID 根据ID查询账单明细
func (r *BillItemRepository) FindByID(ctx context.Context, id string) (*billing.BillItem, error) {
	var do model.BillItemDO

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&do).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("bill item")
		}
		return nil, errors.NewDatabaseError("find bill item by id", err)
	}

	return do.ToDomain()
}

// ListByBill 查询账单的明细列表
func (r *BillItemRepository) ListByBill(ctx context.Context, billID string) ([]*billing.BillItem, error) {
	var dos []model.BillItemDO

	if err := r.db.WithContext(ctx).
		Where("bill_id = ?", billID).
		Order("created_at ASC").
		Find(&dos).Error; err != nil {
		return nil, errors.NewDatabaseError("list bill items by bill", err)
	}

	// 转换为领域对象
	items := make([]*billing.BillItem, 0, len(dos))
	for _, do := range dos {
		item, err := do.ToDomain()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// ListByOrder 查询订单关联的账单明细
func (r *BillItemRepository) ListByOrder(ctx context.Context, orderID string) ([]*billing.BillItem, error) {
	var dos []model.BillItemDO

	if err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at ASC").
		Find(&dos).Error; err != nil {
		return nil, errors.NewDatabaseError("list bill items by order", err)
	}

	// 转换为领域对象
	items := make([]*billing.BillItem, 0, len(dos))
	for _, do := range dos {
		item, err := do.ToDomain()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// SumAmountByBill 统计账单的明细总额
func (r *BillItemRepository) SumAmountByBill(ctx context.Context, billID string) (money.Decimal, error) {
	var result struct {
		TotalAmount string
	}

	// 使用原始SQL进行聚合查询
	if err := r.db.WithContext(ctx).
		Model(&model.BillItemDO{}).
		Select("COALESCE(SUM(CAST(total_amount AS DECIMAL)), 0) as total_amount").
		Where("bill_id = ?", billID).
		Scan(&result).Error; err != nil {
		return money.Zero, errors.NewDatabaseError("sum amount by bill", err)
	}

	total, err := money.NewFromString(result.TotalAmount)
	if err != nil {
		return money.Zero, errors.Wrap(errors.CodeInternalError, "parse total amount", err)
	}

	return total, nil
}
