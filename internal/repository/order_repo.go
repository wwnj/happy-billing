package repository

import (
	"context"

	"github.com/wwnj/happy-billing/internal/models"
	"gorm.io/gorm"
)

// ============================================================================
// Order Repository Interfaces
// ============================================================================

// OrderRepository 订单仓储接口
type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, orderID string) (*models.Order, error)
	List(ctx context.Context, req *models.OrderListQueryRequest) ([]models.Order, int64, error)
	Update(ctx context.Context, order *models.Order) error
	UpdateStatus(ctx context.Context, orderID string, status models.OrderStatus) error
}

// OrderItemRepository 订单明细仓储接口
type OrderItemRepository interface {
	Create(ctx context.Context, item *models.OrderItem) error
	CreateBatch(ctx context.Context, items []models.OrderItem) error
	GetByOrderID(ctx context.Context, orderID string) ([]models.OrderItem, error)
}

// ResourceInstanceRepository 资源实例仓储接口
type ResourceInstanceRepository interface {
	Create(ctx context.Context, instance *models.ResourceInstance) error
	GetByID(ctx context.Context, instanceID string) (*models.ResourceInstance, error)
	GetByOrderID(ctx context.Context, orderID string) ([]models.ResourceInstance, error)
	ListByTenant(ctx context.Context, tenantID string, status *models.ResourceStatus) ([]models.ResourceInstance, error)
	UpdateStatus(ctx context.Context, instanceID string, status models.ResourceStatus) error
}

// ============================================================================
// Order Repository Implementation
// ============================================================================

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓储实例
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) GetByID(ctx context.Context, orderID string) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) List(ctx context.Context, req *models.OrderListQueryRequest) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Order{})

	// 过滤条件
	if req.TenantID != nil {
		query = query.Where("tenant_id = ?", *req.TenantID)
	}
	if req.ProjectID != nil {
		query = query.Where("project_id = ?", *req.ProjectID)
	}
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.OrderType != nil {
		query = query.Where("order_type = ?", *req.OrderType)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Currency != nil {
		query = query.Where("currency = ?", *req.Currency)
	}
	if req.Keyword != nil {
		query = query.Where("order_no LIKE ? OR order_id LIKE ?", "%"+*req.Keyword+"%", "%"+*req.Keyword+"%")
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
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepository) Update(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("order_id = ?", order.OrderID).
		Updates(order).Error
}

func (r *orderRepository) UpdateStatus(ctx context.Context, orderID string, status models.OrderStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("order_id = ?", orderID).
		Update("status", status).Error
}

// ============================================================================
// OrderItem Repository Implementation
// ============================================================================

type orderItemRepository struct {
	db *gorm.DB
}

// NewOrderItemRepository 创建订单明细仓储实例
func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &orderItemRepository{db: db}
}

func (r *orderItemRepository) Create(ctx context.Context, item *models.OrderItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *orderItemRepository) CreateBatch(ctx context.Context, items []models.OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&items).Error
}

func (r *orderItemRepository) GetByOrderID(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	var items []models.OrderItem
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Find(&items).Error
	return items, err
}

// ============================================================================
// ResourceInstance Repository Implementation
// ============================================================================

type resourceInstanceRepository struct {
	db *gorm.DB
}

// NewResourceInstanceRepository 创建资源实例仓储实例
func NewResourceInstanceRepository(db *gorm.DB) ResourceInstanceRepository {
	return &resourceInstanceRepository{db: db}
}

func (r *resourceInstanceRepository) Create(ctx context.Context, instance *models.ResourceInstance) error {
	return r.db.WithContext(ctx).Create(instance).Error
}

func (r *resourceInstanceRepository) GetByID(ctx context.Context, instanceID string) (*models.ResourceInstance, error) {
	var instance models.ResourceInstance
	err := r.db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *resourceInstanceRepository) GetByOrderID(ctx context.Context, orderID string) ([]models.ResourceInstance, error) {
	var instances []models.ResourceInstance
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Find(&instances).Error
	return instances, err
}

func (r *resourceInstanceRepository) ListByTenant(ctx context.Context, tenantID string, status *models.ResourceStatus) ([]models.ResourceInstance, error) {
	var instances []models.ResourceInstance
	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Order("created_at DESC").Find(&instances).Error
	return instances, err
}

func (r *resourceInstanceRepository) UpdateStatus(ctx context.Context, instanceID string, status models.ResourceStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.ResourceInstance{}).
		Where("instance_id = ?", instanceID).
		Update("status", status).Error
}
