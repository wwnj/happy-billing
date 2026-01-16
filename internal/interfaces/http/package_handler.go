package http

import (
	"time"

	"github.com/gin-gonic/gin"

	pkgcmd "github.com/wwnj/happy-billing/internal/application/command/package"
	pkgquery "github.com/wwnj/happy-billing/internal/application/query/package"
	pkgdomain "github.com/wwnj/happy-billing/internal/domain/package"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/money"
)

// PackageHandler 套餐包HTTP处理器
type PackageHandler struct {
	purchaseService *pkgcmd.PurchasePackageService
	packageQuery    *pkgquery.PackageQuery
}

// NewPackageHandler 创建套餐包HTTP处理器
func NewPackageHandler(
	purchaseService *pkgcmd.PurchasePackageService,
	packageQuery *pkgquery.PackageQuery,
) *PackageHandler {
	return &PackageHandler{
		purchaseService: purchaseService,
		packageQuery:    packageQuery,
	}
}

// PurchasePackageRequest 购买套餐包请求
type PurchasePackageRequest struct {
	TenantID    string                 `json:"tenant_id" binding:"required"`
	UserID      string                 `json:"user_id" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	TotalQuota  string                 `json:"total_quota" binding:"required"`
	QuotaUnit   string                 `json:"quota_unit" binding:"required"`
	Price       string                 `json:"price" binding:"required"`
	Currency    string                 `json:"currency"`
	ValidFrom   string                 `json:"valid_from" binding:"required"`
	ValidTo     string                 `json:"valid_to" binding:"required"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PurchasePackage 购买套餐包接口
// POST /api/v1/packages/purchase
func (h *PackageHandler) PurchasePackage(c *gin.Context) {
	var req PurchasePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewInvalidParam(err.Error()))
		return
	}

	// 解析金额
	totalQuota, err := money.NewFromString(req.TotalQuota)
	if err != nil {
		Error(c, errors.NewInvalidParam("invalid total_quota format"))
		return
	}

	price, err := money.NewFromString(req.Price)
	if err != nil {
		Error(c, errors.NewInvalidParam("invalid price format"))
		return
	}

	// 解析时间
	validFrom, err := time.Parse(time.RFC3339, req.ValidFrom)
	if err != nil {
		Error(c, errors.NewInvalidParam("invalid valid_from format, use RFC3339"))
		return
	}

	validTo, err := time.Parse(time.RFC3339, req.ValidTo)
	if err != nil {
		Error(c, errors.NewInvalidParam("invalid valid_to format, use RFC3339"))
		return
	}

	// 执行购买
	cmd := pkgcmd.PurchasePackageCommand{
		TenantID:    req.TenantID,
		UserID:      req.UserID,
		Type:        pkgdomain.PackageType(req.Type),
		Name:        req.Name,
		Description: req.Description,
		TotalQuota:  totalQuota,
		QuotaUnit:   req.QuotaUnit,
		Price:       price,
		Currency:    req.Currency,
		ValidFrom:   validFrom,
		ValidTo:     validTo,
		Metadata:    req.Metadata,
	}

	pkg, err := h.purchaseService.Execute(c.Request.Context(), cmd)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"package_id": pkg.ID,
		"package_no": pkg.PackageNo,
		"status":     "purchased",
	})
}

// ListPackagesRequest 查询套餐包列表请求
type ListPackagesRequest struct {
	PaginationRequest
	TenantID string `form:"tenant_id" binding:"required"`
	UserID   string `form:"user_id"`
	Status   string `form:"status"`
	Type     string `form:"type"`
}

// ListPackages 查询套餐包列表接口
// GET /api/v1/packages
func (h *PackageHandler) ListPackages(c *gin.Context) {
	var req ListPackagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, errors.NewInvalidParam(err.Error()))
		return
	}

	// 查询套餐包列表
	query := pkgquery.ListPackagesQuery{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Status:   pkgdomain.PackageStatus(req.Status),
		Type:     pkgdomain.PackageType(req.Type),
		Offset:   req.GetOffset(),
		Limit:    req.GetLimit(),
	}

	result, err := h.packageQuery.ListPackages(c.Request.Context(), query)
	if err != nil {
		Error(c, err)
		return
	}

	// 返回分页响应
	Success(c, PaginationResponse{
		Total:    result.Total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     result.Packages,
	})
}

// GetPackage 查询套餐包详情接口
// GET /api/v1/packages/:id
func (h *PackageHandler) GetPackage(c *gin.Context) {
	packageID := c.Param("id")
	if packageID == "" {
		Error(c, errors.NewInvalidParam("package_id is required"))
		return
	}

	pkg, err := h.packageQuery.GetPackage(c.Request.Context(), packageID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, pkg)
}

// ListAvailablePackagesRequest 查询可用套餐包请求
type ListAvailablePackagesRequest struct {
	TenantID string `form:"tenant_id" binding:"required"`
	UserID   string `form:"user_id" binding:"required"`
	Type     string `form:"type"`
}

// ListAvailablePackages 查询可用套餐包接口
// GET /api/v1/packages/available
func (h *PackageHandler) ListAvailablePackages(c *gin.Context) {
	var req ListAvailablePackagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, errors.NewInvalidParam(err.Error()))
		return
	}

	packages, err := h.packageQuery.ListAvailablePackages(
		c.Request.Context(),
		req.TenantID,
		req.UserID,
		pkgdomain.PackageType(req.Type),
	)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"packages": packages,
		"count":    len(packages),
	})
}

// GetQuotaSummaryRequest 查询配额汇总请求
type GetQuotaSummaryRequest struct {
	TenantID string `form:"tenant_id" binding:"required"`
	UserID   string `form:"user_id" binding:"required"`
	Type     string `form:"type"`
}

// GetQuotaSummary 查询配额汇总接口
// GET /api/v1/packages/quota/summary
func (h *PackageHandler) GetQuotaSummary(c *gin.Context) {
	var req GetQuotaSummaryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, errors.NewInvalidParam(err.Error()))
		return
	}

	summary, err := h.packageQuery.GetQuotaSummary(
		c.Request.Context(),
		req.TenantID,
		req.UserID,
		pkgdomain.PackageType(req.Type),
	)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, summary)
}

// RegisterRoutes 注册路由
func (h *PackageHandler) RegisterRoutes(router *gin.RouterGroup) {
	packages := router.Group("/packages")
	{
		// 购买套餐包
		packages.POST("/purchase", h.PurchasePackage)

		// 查询套餐包
		packages.GET("", h.ListPackages)
		packages.GET("/available", h.ListAvailablePackages)
		packages.GET("/quota/summary", h.GetQuotaSummary)
		packages.GET("/:id", h.GetPackage)
	}
}
