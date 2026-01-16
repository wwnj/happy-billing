package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/wwnj/happy-billing/internal/domain/billing"
	"github.com/wwnj/happy-billing/internal/infrastructure/persistence/postgres/model"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// BillRepository 账单仓储PostgreSQL实现
type BillRepository struct {
	db *gorm.DB
}

// NewBillRepository 创建账单仓储
func NewBillRepository(db *gorm.DB) *BillRepository {
	return &BillRepository{
		db: db,
	}
}

// Save 保存账单（新建）
func (r *BillRepository) Save(ctx context.Context, bill *billing.Bill) error {
	do := model.FromDomainBill(bill)

	if err := r.db.WithContext(ctx).Create(do).Error; err != nil {
		return errors.NewDatabaseError("save bill", err)
	}

	return nil
}

// Update 更新账单
func (r *BillRepository) Update(ctx context.Context, bill *billing.Bill) error {
	do := model.FromDomainBill(bill)

	if err := r.db.WithContext(ctx).Save(do).Error; err != nil {
		return errors.NewDatabaseError("update bill", err)
	}

	return nil
}

// FindByID 根据ID查询账单
func (r *BillRepository) FindByID(ctx context.Context, id string) (*billing.Bill, error) {
	var do model.BillDO

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&do).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("bill")
		}
		return nil, errors.NewDatabaseError("find bill by id", err)
	}

	return do.ToDomain()
}

// FindByBillNo 根据账单号查询
func (r *BillRepository) FindByBillNo(ctx context.Context, billNo string) (*billing.Bill, error) {
	var do model.BillDO

	if err := r.db.WithContext(ctx).Where("bill_no = ?", billNo).First(&do).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("bill")
		}
		return nil, errors.NewDatabaseError("find bill by bill_no", err)
	}

	return do.ToDomain()
}

// ListByUser 查询用户的账单列表
func (r *BillRepository) ListByUser(
	ctx context.Context,
	tenantID, userID string,
	status billing.BillStatus,
	startTime, endTime *time.Time,
	offset, limit int,
) ([]*billing.Bill, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.BillDO{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID)

	// 按状态过滤
	if status != "" {
		query = query.Where("status = ?", string(status))
	}

	// 按账期时间范围过滤
	if startTime != nil {
		query = query.Where("period_start >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("period_end <= ?", endTime)
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("count bills by user", err)
	}

	// 查询分页数据
	var dos []model.BillDO
	if err := query.Offset(offset).Limit(limit).Order("period_start DESC").Find(&dos).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("list bills by user", err)
	}

	// 转换为领域对象
	bills := make([]*billing.Bill, 0, len(dos))
	for _, do := range dos {
		bill, err := do.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		bills = append(bills, bill)
	}

	return bills, total, nil
}

// ListByTenant 查询租户的账单列表
func (r *BillRepository) ListByTenant(
	ctx context.Context,
	tenantID string,
	status billing.BillStatus,
	offset, limit int,
) ([]*billing.Bill, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.BillDO{}).Where("tenant_id = ?", tenantID)

	// 按状态过滤
	if status != "" {
		query = query.Where("status = ?", string(status))
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("count bills by tenant", err)
	}

	// 查询分页数据
	var dos []model.BillDO
	if err := query.Offset(offset).Limit(limit).Order("period_start DESC").Find(&dos).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("list bills by tenant", err)
	}

	// 转换为领域对象
	bills := make([]*billing.Bill, 0, len(dos))
	for _, do := range dos {
		bill, err := do.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		bills = append(bills, bill)
	}

	return bills, total, nil
}

// ListOverdue 查询逾期账单
func (r *BillRepository) ListOverdue(
	ctx context.Context,
	asOfDate time.Time,
	offset, limit int,
) ([]*billing.Bill, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.BillDO{}).
		Where("status = ? AND due_date < ?", string(billing.BillStatusPending), asOfDate)

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("count overdue bills", err)
	}

	// 查询分页数据
	var dos []model.BillDO
	if err := query.Offset(offset).Limit(limit).Order("due_date ASC").Find(&dos).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("list overdue bills", err)
	}

	// 转换为领域对象
	bills := make([]*billing.Bill, 0, len(dos))
	for _, do := range dos {
		bill, err := do.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		bills = append(bills, bill)
	}

	return bills, total, nil
}

// SumAmountByUser 统计用户在指定时间范围内的账单总额
func (r *BillRepository) SumAmountByUser(
	ctx context.Context,
	tenantID, userID string,
	startTime, endTime time.Time,
) (money.Decimal, error) {
	var result struct {
		TotalAmount string
	}

	// 使用原始SQL进行聚合查询
	if err := r.db.WithContext(ctx).
		Model(&model.BillDO{}).
		Select("COALESCE(SUM(CAST(actual_amount AS DECIMAL)), 0) as total_amount").
		Where("tenant_id = ? AND user_id = ? AND period_start >= ? AND period_end <= ? AND status != ?",
			tenantID, userID, startTime, endTime, string(billing.BillStatusCancelled)).
		Scan(&result).Error; err != nil {
		return money.Zero, errors.NewDatabaseError("sum amount by user", err)
	}

	total, err := money.NewFromString(result.TotalAmount)
	if err != nil {
		return money.Zero, errors.Wrap(errors.CodeInternalError, "parse total amount", err)
	}

	return total, nil
}

// SumAmountByTenant 统计租户在指定时间范围内的账单总额
func (r *BillRepository) SumAmountByTenant(
	ctx context.Context,
	tenantID string,
	startTime, endTime time.Time,
) (money.Decimal, error) {
	var result struct {
		TotalAmount string
	}

	// 使用原始SQL进行聚合查询
	if err := r.db.WithContext(ctx).
		Model(&model.BillDO{}).
		Select("COALESCE(SUM(CAST(actual_amount AS DECIMAL)), 0) as total_amount").
		Where("tenant_id = ? AND period_start >= ? AND period_end <= ? AND status != ?",
			tenantID, startTime, endTime, string(billing.BillStatusCancelled)).
		Scan(&result).Error; err != nil {
		return money.Zero, errors.NewDatabaseError("sum amount by tenant", err)
	}

	total, err := money.NewFromString(result.TotalAmount)
	if err != nil {
		return money.Zero, errors.Wrap(errors.CodeInternalError, "parse total amount", err)
	}

	return total, nil
}
