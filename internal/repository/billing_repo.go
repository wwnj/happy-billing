package repository

import (
	"context"

	"github.com/wwnj/happy-billing/internal/models"
	"gorm.io/gorm"
)

// ============================================================================
// Bill Repository Interface
// ============================================================================

// BillRepository 账单仓储接口
type BillRepository interface {
	Create(ctx context.Context, bill *models.Bill) error
	GetByID(ctx context.Context, billID string) (*models.Bill, error)
	GetByOrderID(ctx context.Context, orderID string) ([]models.Bill, error)
	List(ctx context.Context, req *models.BillListQueryRequest) ([]models.Bill, int64, error)
	UpdateStatus(ctx context.Context, billID string, status models.BillStatus) error
}

// PaymentRepository 支付记录仓储接口
type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, paymentID string) (*models.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) ([]models.Payment, error)
	GetByBillID(ctx context.Context, billID string) ([]models.Payment, error)
	UpdateStatus(ctx context.Context, paymentID string, status models.PaymentStatus) error
}

// AccountBalanceRepository 账户余额仓储接口
type AccountBalanceRepository interface {
	GetByTenantID(ctx context.Context, tenantID string) (*models.AccountBalance, error)
	Create(ctx context.Context, balance *models.AccountBalance) error
	UpdateBalance(ctx context.Context, tenantID string, balanceDelta, frozenDelta float64) error
}

// BalanceTransactionRepository 余额变动记录仓储接口
type BalanceTransactionRepository interface {
	Create(ctx context.Context, transaction *models.BalanceTransaction) error
	ListByTenant(ctx context.Context, tenantID string, limit int) ([]models.BalanceTransaction, error)
	ListByTenantPaginated(ctx context.Context, tenantID string, page, pageSize int) ([]models.BalanceTransaction, int64, error)
}

// ============================================================================
// Bill Repository Implementation
// ============================================================================

type billRepository struct {
	db *gorm.DB
}

// NewBillRepository 创建账单仓储实例
func NewBillRepository(db *gorm.DB) BillRepository {
	return &billRepository{db: db}
}

func (r *billRepository) Create(ctx context.Context, bill *models.Bill) error {
	return r.db.WithContext(ctx).Create(bill).Error
}

func (r *billRepository) GetByID(ctx context.Context, billID string) (*models.Bill, error) {
	var bill models.Bill
	err := r.db.WithContext(ctx).
		Where("bill_id = ?", billID).
		First(&bill).Error
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

func (r *billRepository) GetByOrderID(ctx context.Context, orderID string) ([]models.Bill, error) {
	var bills []models.Bill
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&bills).Error
	return bills, err
}

func (r *billRepository) List(ctx context.Context, req *models.BillListQueryRequest) ([]models.Bill, int64, error) {
	var bills []models.Bill
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Bill{})

	// 过滤条件
	if req.TenantID != nil {
		query = query.Where("tenant_id = ?", *req.TenantID)
	}
	if req.ProjectID != nil {
		query = query.Where("project_id = ?", *req.ProjectID)
	}
	if req.OrderID != nil {
		query = query.Where("order_id = ?", *req.OrderID)
	}
	if req.BillType != nil {
		query = query.Where("bill_type = ?", *req.BillType)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&bills).Error

	return bills, total, err
}

func (r *billRepository) UpdateStatus(ctx context.Context, billID string, status models.BillStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.Bill{}).
		Where("bill_id = ?", billID).
		Update("status", status).Error
}

// ============================================================================
// Payment Repository Implementation
// ============================================================================

type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository 创建支付记录仓储实例
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepository) GetByID(ctx context.Context, paymentID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Where("payment_id = ?", paymentID).
		First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByOrderID(ctx context.Context, orderID string) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) GetByBillID(ctx context.Context, billID string) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.WithContext(ctx).
		Where("bill_id = ?", billID).
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, paymentID string, status models.PaymentStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("payment_id = ?", paymentID).
		Update("status", status).Error
}

// ============================================================================
// AccountBalance Repository Implementation
// ============================================================================

type accountBalanceRepository struct {
	db *gorm.DB
}

// NewAccountBalanceRepository 创建账户余额仓储实例
func NewAccountBalanceRepository(db *gorm.DB) AccountBalanceRepository {
	return &accountBalanceRepository{db: db}
}

func (r *accountBalanceRepository) GetByTenantID(ctx context.Context, tenantID string) (*models.AccountBalance, error) {
	var balance models.AccountBalance
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func (r *accountBalanceRepository) Create(ctx context.Context, balance *models.AccountBalance) error {
	return r.db.WithContext(ctx).Create(balance).Error
}

func (r *accountBalanceRepository) UpdateBalance(ctx context.Context, tenantID string, balanceDelta, frozenDelta float64) error {
	return r.db.WithContext(ctx).
		Model(&models.AccountBalance{}).
		Where("tenant_id = ?", tenantID).
		Updates(map[string]interface{}{
			"balance":        gorm.Expr("balance + ?", balanceDelta),
			"frozen_balance": gorm.Expr("frozen_balance + ?", frozenDelta),
		}).Error
}

// ============================================================================
// BalanceTransaction Repository Implementation
// ============================================================================

type balanceTransactionRepository struct {
	db *gorm.DB
}

// NewBalanceTransactionRepository 创建余额变动记录仓储实例
func NewBalanceTransactionRepository(db *gorm.DB) BalanceTransactionRepository {
	return &balanceTransactionRepository{db: db}
}

func (r *balanceTransactionRepository) Create(ctx context.Context, transaction *models.BalanceTransaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

func (r *balanceTransactionRepository) ListByTenant(ctx context.Context, tenantID string, limit int) ([]models.BalanceTransaction, error) {
	var transactions []models.BalanceTransaction
	query := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&transactions).Error
	return transactions, err
}

func (r *balanceTransactionRepository) ListByTenantPaginated(ctx context.Context, tenantID string, page, pageSize int) ([]models.BalanceTransaction, int64, error) {
	var transactions []models.BalanceTransaction
	var total int64

	query := r.db.WithContext(ctx).Model(&models.BalanceTransaction{}).
		Where("tenant_id = ?", tenantID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&transactions).Error

	return transactions, total, err
}
