package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/internal/api/response"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/service"
	"github.com/wwnj/happy-billing/pkg/errors"
)

// ProductHandler 产品处理器
type ProductHandler struct {
	productService service.ProductService
}

// NewProductHandler 创建产品处理器
func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// ============================================================================
// 产品分类 API
// ============================================================================

// CreateCategory 创建产品分类
//
// @Summary 创建产品分类
// @Description 创建新的产品分类，支持树形结构
// @Tags 产品分类
// @Accept json
// @Produce json
// @Param request body models.CreateCategoryRequest true "分类信息"
// @Success 200 {object} response.Response{data=models.ProductCategory}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/products/categories [post]
func (h *ProductHandler) CreateCategory(c *gin.Context) {
	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}

	result, err := h.productService.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}

// GetCategory 获取产品分类详情
//
// @Summary 获取产品分类详情
// @Description 根据分类ID获取分类详细信息
// @Tags 产品分类
// @Accept json
// @Produce json
// @Param category_id path string true "分类ID"
// @Success 200 {object} response.Response{data=models.ProductCategory}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/products/categories/{category_id} [get]
func (h *ProductHandler) GetCategory(c *gin.Context) {
	categoryID := c.Param("category_id")
	if categoryID == "" {
		response.BadRequest(c, "分类ID不能为空")
		return
	}

	result, err := h.productService.GetCategory(c.Request.Context(), categoryID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}

// GetCategoryTree 获取分类树
//
// @Summary 获取分类树
// @Description 获取完整的产品分类树形结构
// @Tags 产品分类
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]models.CategoryTreeNode}
// @Failure 500 {object} response.Response
// @Router /api/v1/products/categories/tree [get]
func (h *ProductHandler) GetCategoryTree(c *gin.Context) {
	result, err := h.productService.GetCategoryTree(c.Request.Context())
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}

// ListCategories 列出产品分类
//
// @Summary 列出产品分类
// @Description 根据条件查询产品分类列表
// @Tags 产品分类
// @Accept json
// @Produce json
// @Param parent_id query int false "父分类ID"
// @Param level query int false "分类层级"
// @Success 200 {object} response.Response{data=[]models.ProductCategory}
// @Failure 500 {object} response.Response
// @Router /api/v1/products/categories [get]
func (h *ProductHandler) ListCategories(c *gin.Context) {
	var parentID *int64
	var level *int8

	if parentIDStr := c.Query("parent_id"); parentIDStr != "" {
		if pid, err := strconv.ParseInt(parentIDStr, 10, 64); err == nil {
			parentID = &pid
		}
	}

	if levelStr := c.Query("level"); levelStr != "" {
		if l, err := strconv.ParseInt(levelStr, 10, 8); err == nil {
			l8 := int8(l)
			level = &l8
		}
	}

	result, err := h.productService.ListCategories(c.Request.Context(), parentID, level)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}

// ============================================================================
// SPU API
// ============================================================================

// CreateSpu 创建SPU
//
// @Summary 创建SPU
// @Description 创建新的标准产品单元
// @Tags SPU
// @Accept json
// @Produce json
// @Param request body models.CreateSpuRequest true "SPU信息"
// @Success 200 {object} response.Response{data=models.ProductSpu}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/products/spu [post]
func (h *ProductHandler) CreateSpu(c *gin.Context) {
	var req models.CreateSpuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}

	result, err := h.productService.CreateSpu(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}

// GetSpu 获取SPU详情
//
// @Summary 获取SPU详情
// @Description 根据SPU ID获取详细信息，包含分类信息
// @Tags SPU
// @Accept json
// @Produce json
// @Param spu_id path string true "SPU ID"
// @Success 200 {object} response.Response{data=models.ProductSpuResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/products/spu/{spu_id} [get]
func (h *ProductHandler) GetSpu(c *gin.Context) {
	spuID := c.Param("spu_id")
	if spuID == "" {
		response.BadRequest(c, "SPU ID不能为空")
		return
	}

	result, err := h.productService.GetSpu(c.Request.Context(), spuID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}

// ListSpu 列出SPU
//
// @Summary 列出SPU
// @Description 分页查询SPU列表，支持多条件筛选
// @Tags SPU
// @Accept json
// @Produce json
// @Param category_id query int false "分类ID"
// @Param product_type query string false "产品类型"
// @Param brand query string false "品牌"
// @Param status query int false "状态"
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=models.PageResult}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/products/spu [get]
func (h *ProductHandler) ListSpu(c *gin.Context) {
	var req models.SpuListQueryRequest

	// 解析分页参数
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 解析筛选条件
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if cid, err := strconv.ParseInt(categoryIDStr, 10, 64); err == nil {
			req.CategoryID = &cid
		}
	}

	if productTypeStr := c.Query("product_type"); productTypeStr != "" {
		pt := models.ProductType(productTypeStr)
		req.ProductType = &pt
	}

	if brand := c.Query("brand"); brand != "" {
		req.Brand = &brand
	}

	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status := models.Status(s)
			req.Status = &status
		}
	}

	if keyword := c.Query("keyword"); keyword != "" {
		req.Keyword = &keyword
	}

	result, err := h.productService.ListSpu(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}

// ============================================================================
// SKU API
// ============================================================================

// CreateSku 创建SKU
//
// @Summary 创建SKU
// @Description 创建新的库存保管单元
// @Tags SKU
// @Accept json
// @Produce json
// @Param request body models.CreateSkuRequest true "SKU信息"
// @Success 200 {object} response.Response{data=models.ProductSku}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/products/sku [post]
func (h *ProductHandler) CreateSku(c *gin.Context) {
	var req models.CreateSkuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}

	result, err := h.productService.CreateSku(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}

// GetSku 获取SKU详情
//
// @Summary 获取SKU详情
// @Description 根据SKU ID获取详细信息，包含SPU和分类信息
// @Tags SKU
// @Accept json
// @Produce json
// @Param sku_id path string true "SKU ID"
// @Success 200 {object} response.Response{data=models.ProductSkuResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/products/sku/{sku_id} [get]
func (h *ProductHandler) GetSku(c *gin.Context) {
	skuID := c.Param("sku_id")
	if skuID == "" {
		response.BadRequest(c, "SKU ID不能为空")
		return
	}

	result, err := h.productService.GetSku(c.Request.Context(), skuID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}

// ListSku 列出SKU
//
// @Summary 列出SKU
// @Description 分页查询SKU列表，支持多条件筛选
// @Tags SKU
// @Accept json
// @Produce json
// @Param spu_id query int false "SPU ID"
// @Param region query string false "地域"
// @Param stock_type query string false "库存类型"
// @Param status query int false "状态"
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=models.PageResult}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/products/sku [get]
func (h *ProductHandler) ListSku(c *gin.Context) {
	var req models.SkuListQueryRequest

	// 解析分页参数
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 解析筛选条件
	if spuIDStr := c.Query("spu_id"); spuIDStr != "" {
		if sid, err := strconv.ParseInt(spuIDStr, 10, 64); err == nil {
			req.SpuID = &sid
		}
	}

	if region := c.Query("region"); region != "" {
		req.Region = &region
	}

	if stockTypeStr := c.Query("stock_type"); stockTypeStr != "" {
		st := models.StockType(stockTypeStr)
		req.StockType = &st
	}

	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status := models.Status(s)
			req.Status = &status
		}
	}

	if keyword := c.Query("keyword"); keyword != "" {
		req.Keyword = &keyword
	}

	result, err := h.productService.ListSku(c.Request.Context(), &req)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}

// GetSkusBySpu 获取指定SPU的所有SKU
//
// @Summary 获取SPU的所有SKU
// @Description 获取指定SPU下的所有SKU列表
// @Tags SKU
// @Accept json
// @Produce json
// @Param spu_id path string true "SPU ID"
// @Success 200 {object} response.Response{data=[]models.ProductSku}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/products/spu/{spu_id}/skus [get]
func (h *ProductHandler) GetSkusBySpu(c *gin.Context) {
	spuID := c.Param("spu_id")
	if spuID == "" {
		response.BadRequest(c, "SPU ID不能为空")
		return
	}

	result, err := h.productService.GetSkusBySpu(c.Request.Context(), spuID)
	if err != nil {
		if bizErr, ok := err.(*errors.BizError); ok {
			response.Error(c, bizErr)
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}
