package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/internal/api/response"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/service"
	"github.com/wwnj/happy-billing/pkg/errors"
)

// TenantHandler 租户处理器
type TenantHandler struct {
	tenantService service.TenantService
}

// NewTenantHandler 创建租户处理器
func NewTenantHandler(tenantService service.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// RegisterIndividual 个人用户注册
//
// @Summary 个人用户注册
// @Description 注册个人开发者账号，自动创建租户、组织、项目和用户
// @Tags 租户
// @Accept json
// @Produce json
// @Param request body models.RegisterIndividualRequest true "注册信息)"
// @Success 200 {object} response.Response{data=models.TenantResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/tenants/register/individual [post]
func (h *TenantHandler) RegisterIndividual(c *gin.Context) {
	var req models.RegisterIndividualRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}

	result, err := h.tenantService.RegisterIndividual(c.Request.Context(), &req)
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

// RegisterEnterprise 企业用户注册
//
// @Summary 企业用户注册
// @Description 注册企业账号，自动创建租户、总部组织、默认项目和管理员用户
// @Tags 租户
// @Accept json
// @Produce json
// @Param request body models.RegisterEnterpriseRequest true "注册信息)"
// @Success 200 {object} response.Response{data=models.TenantResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/tenants/register/enterprise [post]
func (h *TenantHandler) RegisterEnterprise(c *gin.Context) {
	var req models.RegisterEnterpriseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}

	result, err := h.tenantService.RegisterEnterprise(c.Request.Context(), &req)
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

// GetTenant 获取租户详情
//
// @Summary 获取租户详情
// @Description 获取租户详细信息，包括组织、项目和主账号用户
// @Tags 租户
// @Accept json
// @Produce json
// @Param tenant_id path string true "租户ID)"
// @Success 200 {object} response.Response{data=models.TenantResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/tenants/{tenant_id} [get]
func (h *TenantHandler) GetTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.BadRequest(c, "租户ID不能为空")
		return
	}

	result, err := h.tenantService.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		if err.Error() == errors.NewWithCode(errors.ErrTenantNotFound).Error() {
			response.NotFound(c, "租户不存在")
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, result)
}

// ListTenants 获取租户列表
//
// @Summary 获取租户列表
// @Description 分页查询租户列表，支持按租户类型筛选
// @Tags 租户
// @Accept json
// @Produce json
// @Param tenant_type query string false "租户类型: ENTERPRISE/INDIVIDUAL)"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=object{list=[]models.Tenant,total=int64}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/tenants [get]
func (h *TenantHandler) ListTenants(c *gin.Context) {
	// 解析查询参数
	tenantTypeStr := c.Query("tenant_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 参数验证
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var tenantType *models.TenantType
	if tenantTypeStr != "" {
		t := models.TenantType(tenantTypeStr)
		tenantType = &t
	}

	// 调用服务层
	tenants, total, err := h.tenantService.ListTenants(c.Request.Context(), tenantType, page, pageSize)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Success(c, gin.H{
		"list":      tenants,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateOrganization 创建组织
//
// @Summary 创建组织
// @Description 为租户创建新的组织/部门
// @Tags 租户-组织
// @Accept json
// @Produce json
// @Param request body models.Organization true "组织信息)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/organizations [post]
func (h *TenantHandler) CreateOrganization(c *gin.Context) {
	var org models.Organization
	if err := c.ShouldBindJSON(&org); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}

	if err := h.tenantService.CreateOrganization(c.Request.Context(), &org); err != nil {
		if err.Error() == errors.NewWithCode(errors.ErrTenantNotFound).Error() {
			response.NotFound(c, "租户不存在")
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, org)
}

// GetOrganizationsByTenant 获取租户的组织列表
//
// @Summary 获取租户的组织列表
// @Description 查询租户下的所有组织
// @Tags 租户-组织
// @Accept json
// @Produce json
// @Param tenant_id path string true "租户ID)"
// @Success 200 {object} response.Response{data=[]models.Organization}
// @Failure 500 {object} response.Response
// @Router /api/v1/tenants/{tenant_id}/organizations [get]
func (h *TenantHandler) GetOrganizationsByTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.BadRequest(c, "租户ID不能为空")
		return
	}

	organizations, err := h.tenantService.GetOrganizationsByTenant(c.Request.Context(), tenantID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Success(c, organizations)
}

// CreateProject 创建项目
//
// @Summary 创建项目
// @Description 为组织创建新项目
// @Tags 租户-项目
// @Accept json
// @Produce json
// @Param request body models.Project true "项目信息)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/projects [post]
func (h *TenantHandler) CreateProject(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}

	if err := h.tenantService.CreateProject(c.Request.Context(), &project); err != nil {
		if err.Error() == errors.NewWithCode(errors.ErrTenantNotFound).Error() {
			response.NotFound(c, "租户不存在")
		} else if err.Error() == errors.NewWithCode(errors.ErrOrgNotFound).Error() {
			response.NotFound(c, "组织不存在")
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, project)
}

// GetProjectsByTenant 获取租户的项目列表
//
// @Summary 获取租户的项目列表
// @Description 查询租户下的所有项目
// @Tags 租户-项目
// @Accept json
// @Produce json
// @Param tenant_id path string true "租户ID)"
// @Success 200 {object} response.Response{data=[]models.Project}
// @Failure 500 {object} response.Response
// @Router /api/v1/tenants/{tenant_id}/projects [get]
func (h *TenantHandler) GetProjectsByTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.BadRequest(c, "租户ID不能为空")
		return
	}

	projects, err := h.tenantService.GetProjectsByTenant(c.Request.Context(), tenantID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Success(c, projects)
}

// CreateUser 创建用户
//
// @Summary 创建用户
// @Description 为租户创建新用户
// @Tags 租户-用户
// @Accept json
// @Produce json
// @Param request body models.User true "用户信息)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users [post]
func (h *TenantHandler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}

	if err := h.tenantService.CreateUser(c.Request.Context(), &user); err != nil {
		if err.Error() == errors.NewWithCode(errors.ErrTenantNotFound).Error() {
			response.NotFound(c, "租户不存在")
		} else {
			response.InternalError(c)
		}
		return
	}

	response.Success(c, user)
}

// GetUsersByTenant 获取租户的用户列表
//
// @Summary 获取租户的用户列表
// @Description 查询租户下的所有用户
// @Tags 租户-用户
// @Accept json
// @Produce json
// @Param tenant_id path string true "租户ID)"
// @Success 200 {object} response.Response{data=[]models.User}
// @Failure 500 {object} response.Response
// @Router /api/v1/tenants/{tenant_id}/users [get]
func (h *TenantHandler) GetUsersByTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		response.BadRequest(c, "租户ID不能为空")
		return
	}

	users, err := h.tenantService.GetUsersByTenant(c.Request.Context(), tenantID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Success(c, users)
}
