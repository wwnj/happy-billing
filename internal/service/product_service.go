package service

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/repository"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/utils"
	"gorm.io/gorm"
)

// ProductService 产品服务接口
type ProductService interface {
	// 产品分类
	CreateCategory(ctx context.Context, req *models.CreateCategoryRequest) (*models.ProductCategory, error)
	GetCategory(ctx context.Context, categoryID string) (*models.ProductCategory, error)
	GetCategoryTree(ctx context.Context) ([]models.CategoryTreeNode, error)
	ListCategories(ctx context.Context, parentID *int64, level *int8) ([]models.ProductCategory, error)
	UpdateCategory(ctx context.Context, categoryID string, updates map[string]interface{}) error
	DeleteCategory(ctx context.Context, categoryID string) error

	// SPU管理
	CreateSpu(ctx context.Context, req *models.CreateSpuRequest) (*models.ProductSpu, error)
	GetSpu(ctx context.Context, spuID string) (*models.ProductSpuResponse, error)
	GetSpuByCode(ctx context.Context, spuCode string) (*models.ProductSpu, error)
	ListSpu(ctx context.Context, req *models.SpuListQueryRequest) (*models.PageResult, error)
	UpdateSpu(ctx context.Context, spuID string, updates map[string]interface{}) error
	DeleteSpu(ctx context.Context, spuID string) error

	// SKU管理
	CreateSku(ctx context.Context, req *models.CreateSkuRequest) (*models.ProductSku, error)
	GetSku(ctx context.Context, skuID string) (*models.ProductSkuResponse, error)
	GetSkuByCode(ctx context.Context, skuCode string) (*models.ProductSku, error)
	ListSku(ctx context.Context, req *models.SkuListQueryRequest) (*models.PageResult, error)
	GetSkusBySpu(ctx context.Context, spuID string) ([]models.ProductSku, error)
	UpdateSku(ctx context.Context, skuID string, updates map[string]interface{}) error
	DeleteSku(ctx context.Context, skuID string) error
}

type productService struct {
	db           *gorm.DB
	redis        *redis.Client
	categoryRepo repository.ProductCategoryRepository
	spuRepo      repository.ProductSpuRepository
	skuRepo      repository.ProductSkuRepository
}

// NewProductService 创建产品服务实例
func NewProductService(
	db *gorm.DB,
	redis *redis.Client,
	categoryRepo repository.ProductCategoryRepository,
	spuRepo repository.ProductSpuRepository,
	skuRepo repository.ProductSkuRepository,
) ProductService {
	return &productService{
		db:           db,
		redis:        redis,
		categoryRepo: categoryRepo,
		spuRepo:      spuRepo,
		skuRepo:      skuRepo,
	}
}

// ============================================================================
// 产品分类服务
// ============================================================================

// CreateCategory 创建产品分类
func (s *productService) CreateCategory(ctx context.Context, req *models.CreateCategoryRequest) (*models.ProductCategory, error) {
	// 生成分类 ID
	categoryID, err := utils.GenerateCategoryID(ctx, s.redis)
	if err != nil {
		return nil, errors.NewInternalError("生成分类ID失败: " + err.Error())
	}

	// 检查分类编码是否已存在
	existing, _ := s.categoryRepo.GetByCode(ctx, req.CategoryCode)
	if existing != nil {
		return nil, errors.New(errors.ErrInvalidParams, "分类编码已存在")
	}

	category := &models.ProductCategory{
		CategoryID:   categoryID,
		CategoryCode: req.CategoryCode,
		CategoryName: req.CategoryName,
		ParentID:     req.ParentID,
		Level:        req.Level,
		SortOrder:    req.SortOrder,
		Icon:         req.Icon,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, errors.NewInternalError("创建分类失败: " + err.Error())
	}

	return category, nil
}

// GetCategory 获取产品分类
func (s *productService) GetCategory(ctx context.Context, categoryID string) (*models.ProductCategory, error) {
	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewWithCode(errors.ErrCategoryNotFound)
		}
		return nil, errors.NewInternalError("查询分类失败: " + err.Error())
	}
	return category, nil
}

// GetCategoryTree 获取分类树
func (s *productService) GetCategoryTree(ctx context.Context) ([]models.CategoryTreeNode, error) {
	tree, err := s.categoryRepo.GetTree(ctx)
	if err != nil {
		return nil, errors.NewInternalError("查询分类树失败: " + err.Error())
	}
	return tree, nil
}

// ListCategories 列出分类
func (s *productService) ListCategories(ctx context.Context, parentID *int64, level *int8) ([]models.ProductCategory, error) {
	categories, err := s.categoryRepo.List(ctx, parentID, level)
	if err != nil {
		return nil, errors.NewInternalError("查询分类列表失败: " + err.Error())
	}
	return categories, nil
}

// UpdateCategory 更新产品分类
func (s *productService) UpdateCategory(ctx context.Context, categoryID string, updates map[string]interface{}) error {
	// 检查分类是否存在
	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewWithCode(errors.ErrCategoryNotFound)
		}
		return errors.NewInternalError("查询分类失败: " + err.Error())
	}

	// 应用更新
	if name, ok := updates["category_name"]; ok {
		category.CategoryName = name.(string)
	}
	if sortOrder, ok := updates["sort_order"]; ok {
		category.SortOrder = sortOrder.(int)
	}
	if icon, ok := updates["icon"]; ok {
		if icon != nil {
			iconStr := icon.(string)
			category.Icon = &iconStr
		}
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return errors.NewInternalError("更新分类失败: " + err.Error())
	}

	return nil
}

// DeleteCategory 删除产品分类
func (s *productService) DeleteCategory(ctx context.Context, categoryID string) error {
	// 检查是否有子分类
	children, err := s.categoryRepo.GetChildren(ctx, 0) // 需要先获取ID
	if err != nil {
		return errors.NewInternalError("查询子分类失败: " + err.Error())
	}
	if len(children) > 0 {
		return errors.New(errors.ErrInvalidParams, "该分类下存在子分类，无法删除")
	}

	if err := s.categoryRepo.Delete(ctx, categoryID); err != nil {
		return errors.NewInternalError("删除分类失败: " + err.Error())
	}

	return nil
}

// ============================================================================
// SPU服务
// ============================================================================

// CreateSpu 创建SPU
func (s *productService) CreateSpu(ctx context.Context, req *models.CreateSpuRequest) (*models.ProductSpu, error) {
	// 生成SPU ID
	spuID, err := utils.GenerateSPUID(ctx, s.redis)
	if err != nil {
		return nil, errors.NewInternalError("生成SPU ID失败: " + err.Error())
	}

	// 检查SPU编码是否已存在
	existing, _ := s.spuRepo.GetByCode(ctx, req.SpuCode)
	if existing != nil {
		return nil, errors.New(errors.ErrInvalidParams, "SPU编码已存在")
	}

	// 转换规格模板
	var specTemplate *models.SpecTemplate
	if req.SpecTemplate != nil {
		st := models.SpecTemplate(req.SpecTemplate)
		specTemplate = &st
	}

	spu := &models.ProductSpu{
		SpuID:        spuID,
		SpuCode:      req.SpuCode,
		SpuName:      req.SpuName,
		CategoryID:   req.CategoryID,
		ProductType:  req.ProductType,
		Brand:        req.Brand,
		Description:  req.Description,
		BillingUnit:  req.BillingUnit,
		SpecTemplate: specTemplate,
		Status:       models.StatusEnabled,
	}

	if err := s.spuRepo.Create(ctx, spu); err != nil {
		return nil, errors.NewInternalError("创建SPU失败: " + err.Error())
	}

	return spu, nil
}

// GetSpu 获取SPU详情（包含分类信息）
func (s *productService) GetSpu(ctx context.Context, spuID string) (*models.ProductSpuResponse, error) {
	spu, err := s.spuRepo.GetByID(ctx, spuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewWithCode(errors.ErrProductNotFound)
		}
		return nil, errors.NewInternalError("查询SPU失败: " + err.Error())
	}

	// 获取分类信息
	var category *models.ProductCategory
	if spu.CategoryID > 0 {
		cat, _ := s.categoryRepo.GetByID(ctx, "")
		category = cat
	}

	return &models.ProductSpuResponse{
		ProductSpu: *spu,
		Category:   category,
	}, nil
}

// GetSpuByCode 根据编码获取SPU
func (s *productService) GetSpuByCode(ctx context.Context, spuCode string) (*models.ProductSpu, error) {
	spu, err := s.spuRepo.GetByCode(ctx, spuCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewWithCode(errors.ErrProductNotFound)
		}
		return nil, errors.NewInternalError("查询SPU失败: " + err.Error())
	}
	return spu, nil
}

// ListSpu 列出SPU
func (s *productService) ListSpu(ctx context.Context, req *models.SpuListQueryRequest) (*models.PageResult, error) {
	spus, total, err := s.spuRepo.List(ctx, req)
	if err != nil {
		return nil, errors.NewInternalError("查询SPU列表失败: " + err.Error())
	}

	return &models.PageResult{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     spus,
	}, nil
}

// UpdateSpu 更新SPU
func (s *productService) UpdateSpu(ctx context.Context, spuID string, updates map[string]interface{}) error {
	// 检查SPU是否存在
	spu, err := s.spuRepo.GetByID(ctx, spuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewWithCode(errors.ErrProductNotFound)
		}
		return errors.NewInternalError("查询SPU失败: " + err.Error())
	}

	// 应用更新（这里简化处理，实际应该更细致）
	if name, ok := updates["spu_name"]; ok {
		spu.SpuName = name.(string)
	}
	if desc, ok := updates["description"]; ok {
		if desc != nil {
			descStr := desc.(string)
			spu.Description = &descStr
		}
	}
	if status, ok := updates["status"]; ok {
		spu.Status = status.(models.Status)
	}

	if err := s.spuRepo.Update(ctx, spu); err != nil {
		return errors.NewInternalError("更新SPU失败: " + err.Error())
	}

	return nil
}

// DeleteSpu 删除SPU
func (s *productService) DeleteSpu(ctx context.Context, spuID string) error {
	// 检查是否有关联的SKU
	spu, err := s.spuRepo.GetByID(ctx, spuID)
	if err != nil {
		return errors.NewWithCode(errors.ErrProductNotFound)
	}

	skus, _ := s.skuRepo.GetBySpuID(ctx, spu.ID)
	if len(skus) > 0 {
		return errors.New(errors.ErrInvalidParams, "该SPU下存在SKU，无法删除")
	}

	if err := s.spuRepo.Delete(ctx, spuID); err != nil {
		return errors.NewInternalError("删除SPU失败: " + err.Error())
	}

	return nil
}

// ============================================================================
// SKU服务
// ============================================================================

// CreateSku 创建SKU
func (s *productService) CreateSku(ctx context.Context, req *models.CreateSkuRequest) (*models.ProductSku, error) {
	// 生成SKU ID
	skuID, err := utils.GenerateSKUID(ctx, s.redis)
	if err != nil {
		return nil, errors.NewInternalError("生成SKU ID失败: " + err.Error())
	}

	// 检查SKU编码是否已存在
	existing, _ := s.skuRepo.GetByCode(ctx, req.SkuCode)
	if existing != nil {
		return nil, errors.New(errors.ErrInvalidParams, "SKU编码已存在")
	}

	// 获取SPU信息
	spu, err := s.spuRepo.GetByID(ctx, "")
	if err != nil {
		return nil, errors.NewWithCode(errors.ErrProductNotFound)
	}

	sku := &models.ProductSku{
		SkuID:      skuID,
		SkuCode:    req.SkuCode,
		SpuID:      req.SpuID,
		SpuCode:    spu.SpuCode,
		SkuName:    req.SkuName,
		SpecValues: models.SpecValues(req.SpecValues),
		Region:     req.Region,
		StockType:  req.StockType,
		Status:     models.StatusEnabled,
	}

	if err := s.skuRepo.Create(ctx, sku); err != nil {
		return nil, errors.NewInternalError("创建SKU失败: " + err.Error())
	}

	return sku, nil
}

// GetSku 获取SKU详情（包含SPU和分类信息）
func (s *productService) GetSku(ctx context.Context, skuID string) (*models.ProductSkuResponse, error) {
	sku, err := s.skuRepo.GetByID(ctx, skuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewWithCode(errors.ErrSKUNotFound)
		}
		return nil, errors.NewInternalError("查询SKU失败: " + err.Error())
	}

	// 获取SPU信息
	var spu *models.ProductSpu
	var category *models.ProductCategory
	if sku.SpuID > 0 {
		spuData, _ := s.spuRepo.GetByID(ctx, "")
		spu = spuData
		if spu != nil && spu.CategoryID > 0 {
			cat, _ := s.categoryRepo.GetByID(ctx, "")
			category = cat
		}
	}

	return &models.ProductSkuResponse{
		ProductSku: *sku,
		Spu:        spu,
		Category:   category,
	}, nil
}

// GetSkuByCode 根据编码获取SKU
func (s *productService) GetSkuByCode(ctx context.Context, skuCode string) (*models.ProductSku, error) {
	sku, err := s.skuRepo.GetByCode(ctx, skuCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewWithCode(errors.ErrSKUNotFound)
		}
		return nil, errors.NewInternalError("查询SKU失败: " + err.Error())
	}
	return sku, nil
}

// ListSku 列出SKU
func (s *productService) ListSku(ctx context.Context, req *models.SkuListQueryRequest) (*models.PageResult, error) {
	skus, total, err := s.skuRepo.List(ctx, req)
	if err != nil {
		return nil, errors.NewInternalError("查询SKU列表失败: " + err.Error())
	}

	return &models.PageResult{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     skus,
	}, nil
}

// GetSkusBySpu 获取指定SPU的所有SKU
func (s *productService) GetSkusBySpu(ctx context.Context, spuID string) ([]models.ProductSku, error) {
	spu, err := s.spuRepo.GetByID(ctx, spuID)
	if err != nil {
		return nil, errors.NewWithCode(errors.ErrProductNotFound)
	}

	skus, err := s.skuRepo.GetBySpuID(ctx, spu.ID)
	if err != nil {
		return nil, errors.NewInternalError("查询SKU列表失败: " + err.Error())
	}

	return skus, nil
}

// UpdateSku 更新SKU
func (s *productService) UpdateSku(ctx context.Context, skuID string, updates map[string]interface{}) error {
	// 检查SKU是否存在
	sku, err := s.skuRepo.GetByID(ctx, skuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewWithCode(errors.ErrSKUNotFound)
		}
		return errors.NewInternalError("查询SKU失败: " + err.Error())
	}

	// 应用更新
	if name, ok := updates["sku_name"]; ok {
		sku.SkuName = name.(string)
	}
	if stockType, ok := updates["stock_type"]; ok {
		if stockType != nil {
			st := stockType.(models.StockType)
			sku.StockType = &st
		}
	}
	if status, ok := updates["status"]; ok {
		sku.Status = status.(models.Status)
	}

	if err := s.skuRepo.Update(ctx, sku); err != nil {
		return errors.NewInternalError("更新SKU失败: " + err.Error())
	}

	return nil
}

// DeleteSku 删除SKU
func (s *productService) DeleteSku(ctx context.Context, skuID string) error {
	if err := s.skuRepo.Delete(ctx, skuID); err != nil {
		return errors.NewInternalError("删除SKU失败: " + err.Error())
	}
	return nil
}
