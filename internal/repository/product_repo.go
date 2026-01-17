package repository

import (
	"context"

	"github.com/wwnj/happy-billing/internal/models"
	"gorm.io/gorm"
)

// ProductCategoryRepository 产品分类仓储接口
type ProductCategoryRepository interface {
	Create(ctx context.Context, category *models.ProductCategory) error
	GetByID(ctx context.Context, categoryID string) (*models.ProductCategory, error)
	GetByCode(ctx context.Context, categoryCode string) (*models.ProductCategory, error)
	Update(ctx context.Context, category *models.ProductCategory) error
	Delete(ctx context.Context, categoryID string) error
	List(ctx context.Context, parentID *int64, level *int8) ([]models.ProductCategory, error)
	GetChildren(ctx context.Context, parentID int64) ([]models.ProductCategory, error)
	GetTree(ctx context.Context) ([]models.CategoryTreeNode, error)
}

// ProductSpuRepository SPU仓储接口
type ProductSpuRepository interface {
	Create(ctx context.Context, spu *models.ProductSpu) error
	GetByID(ctx context.Context, spuID string) (*models.ProductSpu, error)
	GetByCode(ctx context.Context, spuCode string) (*models.ProductSpu, error)
	Update(ctx context.Context, spu *models.ProductSpu) error
	Delete(ctx context.Context, spuID string) error
	List(ctx context.Context, req *models.SpuListQueryRequest) ([]models.ProductSpu, int64, error)
	GetByCategoryID(ctx context.Context, categoryID int64) ([]models.ProductSpu, error)
}

// ProductSkuRepository SKU仓储接口
type ProductSkuRepository interface {
	Create(ctx context.Context, sku *models.ProductSku) error
	GetByID(ctx context.Context, skuID string) (*models.ProductSku, error)
	GetByCode(ctx context.Context, skuCode string) (*models.ProductSku, error)
	Update(ctx context.Context, sku *models.ProductSku) error
	Delete(ctx context.Context, skuID string) error
	List(ctx context.Context, req *models.SkuListQueryRequest) ([]models.ProductSku, int64, error)
	GetBySpuID(ctx context.Context, spuID int64) ([]models.ProductSku, error)
	GetByRegion(ctx context.Context, region string) ([]models.ProductSku, error)
}

// ============================================================================
// ProductCategory Repository Implementation
// ============================================================================

type productCategoryRepository struct {
	db *gorm.DB
}

// NewProductCategoryRepository 创建产品分类仓储实例
func NewProductCategoryRepository(db *gorm.DB) ProductCategoryRepository {
	return &productCategoryRepository{db: db}
}

func (r *productCategoryRepository) Create(ctx context.Context, category *models.ProductCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *productCategoryRepository) GetByID(ctx context.Context, categoryID string) (*models.ProductCategory, error) {
	var category models.ProductCategory
	err := r.db.WithContext(ctx).Where("category_id = ?", categoryID).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *productCategoryRepository) GetByCode(ctx context.Context, categoryCode string) (*models.ProductCategory, error) {
	var category models.ProductCategory
	err := r.db.WithContext(ctx).Where("category_code = ?", categoryCode).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *productCategoryRepository) Update(ctx context.Context, category *models.ProductCategory) error {
	return r.db.WithContext(ctx).Model(category).Where("category_id = ?", category.CategoryID).Updates(category).Error
}

func (r *productCategoryRepository) Delete(ctx context.Context, categoryID string) error {
	return r.db.WithContext(ctx).Where("category_id = ?", categoryID).Delete(&models.ProductCategory{}).Error
}

func (r *productCategoryRepository) List(ctx context.Context, parentID *int64, level *int8) ([]models.ProductCategory, error) {
	var categories []models.ProductCategory
	query := r.db.WithContext(ctx)

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	}
	if level != nil {
		query = query.Where("level = ?", *level)
	}

	err := query.Order("sort_order ASC, id ASC").Find(&categories).Error
	return categories, err
}

func (r *productCategoryRepository) GetChildren(ctx context.Context, parentID int64) ([]models.ProductCategory, error) {
	var categories []models.ProductCategory
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("sort_order ASC, id ASC").
		Find(&categories).Error
	return categories, err
}

func (r *productCategoryRepository) GetTree(ctx context.Context) ([]models.CategoryTreeNode, error) {
	// 获取所有分类
	var categories []models.ProductCategory
	err := r.db.WithContext(ctx).Order("level ASC, sort_order ASC").Find(&categories).Error
	if err != nil {
		return nil, err
	}

	// 构建树形结构 - 使用map存储指针以便修改
	categoryMap := make(map[int64]*models.CategoryTreeNode)

	// 第一遍：创建所有节点
	for i := range categories {
		cat := &categories[i]
		node := &models.CategoryTreeNode{
			ProductCategory: *cat,
			Children:        []models.CategoryTreeNode{},
		}
		categoryMap[cat.ID] = node
	}

	// 第二遍：建立父子关系（先处理所有关系）
	for i := range categories {
		cat := &categories[i]
		if cat.ParentID != nil {
			// 子节点 - 添加到父节点的Children
			if parent, ok := categoryMap[*cat.ParentID]; ok {
				if child, ok := categoryMap[cat.ID]; ok {
					parent.Children = append(parent.Children, *child)
				}
			}
		}
	}

	// 第三遍：收集根节点
	var roots []models.CategoryTreeNode
	for i := range categories {
		cat := &categories[i]
		if cat.ParentID == nil {
			// 根节点
			if node, ok := categoryMap[cat.ID]; ok {
				roots = append(roots, *node)
			}
		}
	}

	return roots, nil
}

// ============================================================================
// ProductSpu Repository Implementation
// ============================================================================

type productSpuRepository struct {
	db *gorm.DB
}

// NewProductSpuRepository 创建SPU仓储实例
func NewProductSpuRepository(db *gorm.DB) ProductSpuRepository {
	return &productSpuRepository{db: db}
}

func (r *productSpuRepository) Create(ctx context.Context, spu *models.ProductSpu) error {
	return r.db.WithContext(ctx).Create(spu).Error
}

func (r *productSpuRepository) GetByID(ctx context.Context, spuID string) (*models.ProductSpu, error) {
	var spu models.ProductSpu
	err := r.db.WithContext(ctx).Where("spu_id = ?", spuID).First(&spu).Error
	if err != nil {
		return nil, err
	}
	return &spu, nil
}

func (r *productSpuRepository) GetByCode(ctx context.Context, spuCode string) (*models.ProductSpu, error) {
	var spu models.ProductSpu
	err := r.db.WithContext(ctx).Where("spu_code = ?", spuCode).First(&spu).Error
	if err != nil {
		return nil, err
	}
	return &spu, nil
}

func (r *productSpuRepository) Update(ctx context.Context, spu *models.ProductSpu) error {
	return r.db.WithContext(ctx).Model(spu).Where("spu_id = ?", spu.SpuID).Updates(spu).Error
}

func (r *productSpuRepository) Delete(ctx context.Context, spuID string) error {
	return r.db.WithContext(ctx).Where("spu_id = ?", spuID).Delete(&models.ProductSpu{}).Error
}

func (r *productSpuRepository) List(ctx context.Context, req *models.SpuListQueryRequest) ([]models.ProductSpu, int64, error) {
	var spus []models.ProductSpu
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ProductSpu{})

	// 条件过滤
	if req.CategoryID != nil {
		query = query.Where("category_id = ?", *req.CategoryID)
	}
	if req.ProductType != nil {
		query = query.Where("product_type = ?", *req.ProductType)
	}
	if req.Brand != nil {
		query = query.Where("brand = ?", *req.Brand)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Keyword != nil && *req.Keyword != "" {
		keyword := "%" + *req.Keyword + "%"
		query = query.Where("spu_name LIKE ? OR spu_code LIKE ?", keyword, keyword)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := req.GetOffset()
	limit := req.GetLimit()
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&spus).Error

	return spus, total, err
}

func (r *productSpuRepository) GetByCategoryID(ctx context.Context, categoryID int64) ([]models.ProductSpu, error) {
	var spus []models.ProductSpu
	err := r.db.WithContext(ctx).
		Where("category_id = ? AND status = ?", categoryID, models.StatusEnabled).
		Order("created_at DESC").
		Find(&spus).Error
	return spus, err
}

// ============================================================================
// ProductSku Repository Implementation
// ============================================================================

type productSkuRepository struct {
	db *gorm.DB
}

// NewProductSkuRepository 创建SKU仓储实例
func NewProductSkuRepository(db *gorm.DB) ProductSkuRepository {
	return &productSkuRepository{db: db}
}

func (r *productSkuRepository) Create(ctx context.Context, sku *models.ProductSku) error {
	return r.db.WithContext(ctx).Create(sku).Error
}

func (r *productSkuRepository) GetByID(ctx context.Context, skuID string) (*models.ProductSku, error) {
	var sku models.ProductSku
	err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&sku).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *productSkuRepository) GetByCode(ctx context.Context, skuCode string) (*models.ProductSku, error) {
	var sku models.ProductSku
	err := r.db.WithContext(ctx).Where("sku_code = ?", skuCode).First(&sku).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *productSkuRepository) Update(ctx context.Context, sku *models.ProductSku) error {
	return r.db.WithContext(ctx).Model(sku).Where("sku_id = ?", sku.SkuID).Updates(sku).Error
}

func (r *productSkuRepository) Delete(ctx context.Context, skuID string) error {
	return r.db.WithContext(ctx).Where("sku_id = ?", skuID).Delete(&models.ProductSku{}).Error
}

func (r *productSkuRepository) List(ctx context.Context, req *models.SkuListQueryRequest) ([]models.ProductSku, int64, error) {
	var skus []models.ProductSku
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ProductSku{})

	// 条件过滤
	if req.SpuID != nil {
		query = query.Where("spu_id = ?", *req.SpuID)
	}
	if req.Region != nil {
		query = query.Where("region = ?", *req.Region)
	}
	if req.StockType != nil {
		query = query.Where("stock_type = ?", *req.StockType)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Keyword != nil && *req.Keyword != "" {
		keyword := "%" + *req.Keyword + "%"
		query = query.Where("sku_name LIKE ? OR sku_code LIKE ?", keyword, keyword)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := req.GetOffset()
	limit := req.GetLimit()
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&skus).Error

	return skus, total, err
}

func (r *productSkuRepository) GetBySpuID(ctx context.Context, spuID int64) ([]models.ProductSku, error) {
	var skus []models.ProductSku
	err := r.db.WithContext(ctx).
		Where("spu_id = ? AND status = ?", spuID, models.StatusEnabled).
		Order("created_at DESC").
		Find(&skus).Error
	return skus, err
}

func (r *productSkuRepository) GetByRegion(ctx context.Context, region string) ([]models.ProductSku, error) {
	var skus []models.ProductSku
	err := r.db.WithContext(ctx).
		Where("region = ? AND status = ?", region, models.StatusEnabled).
		Order("created_at DESC").
		Find(&skus).Error
	return skus, err
}
