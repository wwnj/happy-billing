package service

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/repository"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/utils"
	"gorm.io/gorm"
)

// TenantService 租户服务接口
type TenantService interface {
	// 注册
	RegisterIndividual(ctx context.Context, req *models.RegisterIndividualRequest) (*models.TenantResponse, error)
	RegisterEnterprise(ctx context.Context, req *models.RegisterEnterpriseRequest) (*models.TenantResponse, error)

	// CRUD
	GetTenant(ctx context.Context, tenantID string) (*models.TenantResponse, error)
	ListTenants(ctx context.Context, tenantType *models.TenantType, page, pageSize int) ([]models.Tenant, int64, error)
	UpdateTenant(ctx context.Context, tenantID string, updates map[string]interface{}) error
	DeleteTenant(ctx context.Context, tenantID string) error

	// 组织管理
	CreateOrganization(ctx context.Context, org *models.Organization) error
	GetOrganizationsByTenant(ctx context.Context, tenantID string) ([]models.Organization, error)

	// 项目管理
	CreateProject(ctx context.Context, project *models.Project) error
	GetProjectsByTenant(ctx context.Context, tenantID string) ([]models.Project, error)

	// 用户管理
	CreateUser(ctx context.Context, user *models.User) error
	GetUsersByTenant(ctx context.Context, tenantID string) ([]models.User, error)
}

type tenantService struct {
	db          *gorm.DB
	redis       *redis.Client
	tenantRepo  repository.TenantRepository
	orgRepo     repository.OrganizationRepository
	projectRepo repository.ProjectRepository
	userRepo    repository.UserRepository
}

// NewTenantService 创建租户服务实例
func NewTenantService(
	db *gorm.DB,
	redis *redis.Client,
	tenantRepo repository.TenantRepository,
	orgRepo repository.OrganizationRepository,
	projectRepo repository.ProjectRepository,
	userRepo repository.UserRepository,
) TenantService {
	return &tenantService{
		db:          db,
		redis:       redis,
		tenantRepo:  tenantRepo,
		orgRepo:     orgRepo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
	}
}

// RegisterIndividual 个人用户注册
// 自动创建：租户 → 默认组织 → 默认项目 → 主账号
func (s *tenantService) RegisterIndividual(ctx context.Context, req *models.RegisterIndividualRequest) (*models.TenantResponse, error) {
	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 生成租户 ID
	tenantID, err := utils.GenerateTenantID(ctx, s.redis)
	if err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("生成租户ID失败: " + err.Error())
	}

	// 2. 创建租户
	tenant := &models.Tenant{
		TenantID:          tenantID,
		TenantCode:        fmt.Sprintf("individual_%s", tenantID),
		Name:              req.Name,
		TenantType:        models.TenantTypeIndividual,
		PreferredCurrency: "CNY",
		Verified:          false,
		Status:            models.StatusEnabled,
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("创建租户失败: " + err.Error())
	}

	// 3. 自动创建默认组织 - "个人工作区"
	orgID, err := utils.GenerateOrganizationID(ctx, s.redis)
	if err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("生成组织ID失败: " + err.Error())
	}

	orgType := models.OrgTypePersonal
	organization := &models.Organization{
		OrganizationID:       orgID,
		TenantID:             tenantID,
		ParentOrganizationID: nil,
		OrgCode:              fmt.Sprintf("personal_workspace_%s", tenantID),
		Name:                 "个人工作区",
		OrgType:              &orgType,
		Level:                1,
		Status:               models.StatusEnabled,
	}

	if err := s.orgRepo.Create(ctx, organization); err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("创建组织失败: " + err.Error())
	}

	// 4. 自动创建默认项目
	projectID, err := utils.GenerateProjectID(ctx, s.redis)
	if err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("生成项目ID失败: " + err.Error())
	}

	project := &models.Project{
		ProjectID:      projectID,
		TenantID:       tenantID,
		OrganizationID: orgID,
		ProjectCode:    fmt.Sprintf("default_project_%s", tenantID),
		Name:           "默认项目",
		Description:    nil,
		Status:         models.StatusEnabled,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("创建项目失败: " + err.Error())
	}

	// 5. 创建主账号用户
	userID, err := utils.GenerateUserID(ctx, s.redis)
	if err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("生成用户ID失败: " + err.Error())
	}

	user := &models.User{
		UserID:    userID,
		TenantID:  tenantID,
		IsPrimary: true,
		RealName:  &req.Name,
		Email:     &req.Email,
		Phone:     &req.Phone,
		Status:    models.StatusEnabled,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("创建用户失败: " + err.Error())
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.NewInternalError("提交事务失败: " + err.Error())
	}

	// 6. 返回完整响应
	return &models.TenantResponse{
		Tenant:         tenant,
		Organizations:  []models.Organization{*organization},
		DefaultProject: project,
		PrimaryUser:    user,
	}, nil
}

// RegisterEnterprise 企业用户注册
func (s *tenantService) RegisterEnterprise(ctx context.Context, req *models.RegisterEnterpriseRequest) (*models.TenantResponse, error) {
	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 生成租户 ID
	tenantID, err := utils.GenerateTenantID(ctx, s.redis)
	if err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("生成租户ID失败: " + err.Error())
	}

	// 2. 创建租户
	tenant := &models.Tenant{
		TenantID:          tenantID,
		TenantCode:        fmt.Sprintf("enterprise_%s", tenantID),
		Name:              req.CompanyName,
		TenantType:        models.TenantTypeEnterprise,
		PreferredCurrency: "CNY",
		Verified:          false,
		Status:            models.StatusEnabled,
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("创建租户失败: " + err.Error())
	}

	// 3. 自动创建总部组织
	orgID, err := utils.GenerateOrganizationID(ctx, s.redis)
	if err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("生成组织ID失败: " + err.Error())
	}

	orgType := models.OrgTypeHeadquarters
	organization := &models.Organization{
		OrganizationID:       orgID,
		TenantID:             tenantID,
		ParentOrganizationID: nil,
		OrgCode:              fmt.Sprintf("headquarters_%s", tenantID),
		Name:                 "总部",
		OrgType:              &orgType,
		Level:                1,
		Status:               models.StatusEnabled,
	}

	if err := s.orgRepo.Create(ctx, organization); err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("创建组织失败: " + err.Error())
	}

	// 4. 自动创建默认项目
	projectID, err := utils.GenerateProjectID(ctx, s.redis)
	if err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("生成项目ID失败: " + err.Error())
	}

	project := &models.Project{
		ProjectID:      projectID,
		TenantID:       tenantID,
		OrganizationID: orgID,
		ProjectCode:    fmt.Sprintf("default_project_%s", tenantID),
		Name:           "默认项目",
		Description:    nil,
		Status:         models.StatusEnabled,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("创建项目失败: " + err.Error())
	}

	// 5. 创建管理员用户
	userID, err := utils.GenerateUserID(ctx, s.redis)
	if err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("生成用户ID失败: " + err.Error())
	}

	user := &models.User{
		UserID:    userID,
		TenantID:  tenantID,
		IsPrimary: true,
		Email:     &req.Email,
		Phone:     &req.Phone,
		Status:    models.StatusEnabled,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("创建用户失败: " + err.Error())
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.NewInternalError("提交事务失败: " + err.Error())
	}

	// 6. 返回完整响应
	return &models.TenantResponse{
		Tenant:         tenant,
		Organizations:  []models.Organization{*organization},
		DefaultProject: project,
		PrimaryUser:    user,
	}, nil
}

// GetTenant 获取租户详细信息
func (s *tenantService) GetTenant(ctx context.Context, tenantID string) (*models.TenantResponse, error) {
	// 1. 获取租户
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewWithCode(errors.ErrTenantNotFound)
		}
		return nil, errors.NewInternalError("查询租户失败: " + err.Error())
	}

	// 2. 获取组织列表
	organizations, err := s.orgRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("查询组织失败: " + err.Error())
	}

	// 3. 获取项目列表
	projects, err := s.projectRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("查询项目失败: " + err.Error())
	}

	// 4. 获取主账号用户
	primaryUser, err := s.userRepo.GetPrimaryUser(ctx, tenantID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.NewInternalError("查询用户失败: " + err.Error())
	}

	// 5. 构造响应
	response := &models.TenantResponse{
		Tenant:        tenant,
		Organizations: organizations,
		PrimaryUser:   primaryUser,
	}

	if len(projects) > 0 {
		response.DefaultProject = &projects[0]
	}

	return response, nil
}

// ListTenants 获取租户列表
func (s *tenantService) ListTenants(ctx context.Context, tenantType *models.TenantType, page, pageSize int) ([]models.Tenant, int64, error) {
	offset := (page - 1) * pageSize
	return s.tenantRepo.List(ctx, tenantType, offset, pageSize)
}

// UpdateTenant 更新租户
func (s *tenantService) UpdateTenant(ctx context.Context, tenantID string, updates map[string]interface{}) error {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewWithCode(errors.ErrTenantNotFound)
		}
		return errors.NewInternalError("查询租户失败: " + err.Error())
	}

	// 这里可以添加字段验证逻辑

	// 更新字段
	if name, ok := updates["name"].(string); ok {
		tenant.Name = name
	}
	if currency, ok := updates["preferred_currency"].(string); ok {
		tenant.PreferredCurrency = currency
	}

	return s.tenantRepo.Update(ctx, tenant)
}

// DeleteTenant 删除租户
func (s *tenantService) DeleteTenant(ctx context.Context, tenantID string) error {
	// TODO: 添加级联删除逻辑（删除关联的组织、项目、用户等）
	return s.tenantRepo.Delete(ctx, tenantID)
}

// CreateOrganization 创建组织
func (s *tenantService) CreateOrganization(ctx context.Context, org *models.Organization) error {
	// 验证租户是否存在
	_, err := s.tenantRepo.GetByID(ctx, org.TenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewWithCode(errors.ErrTenantNotFound)
		}
		return errors.NewInternalError("查询租户失败: " + err.Error())
	}

	// 生成组织 ID
	orgID, err := utils.GenerateOrganizationID(ctx, s.redis)
	if err != nil {
		return errors.NewInternalError("生成组织ID失败: " + err.Error())
	}

	org.OrganizationID = orgID
	org.Status = models.StatusEnabled

	return s.orgRepo.Create(ctx, org)
}

// GetOrganizationsByTenant 获取租户的组织列表
func (s *tenantService) GetOrganizationsByTenant(ctx context.Context, tenantID string) ([]models.Organization, error) {
	return s.orgRepo.GetByTenantID(ctx, tenantID)
}

// CreateProject 创建项目
func (s *tenantService) CreateProject(ctx context.Context, project *models.Project) error {
	// 验证租户是否存在
	_, err := s.tenantRepo.GetByID(ctx, project.TenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewWithCode(errors.ErrTenantNotFound)
		}
		return errors.NewInternalError("查询租户失败: " + err.Error())
	}

	// 验证组织是否存在
	_, err = s.orgRepo.GetByID(ctx, project.OrganizationID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewWithCode(errors.ErrOrgNotFound)
		}
		return errors.NewInternalError("查询组织失败: " + err.Error())
	}

	// 生成项目 ID
	projectID, err := utils.GenerateProjectID(ctx, s.redis)
	if err != nil {
		return errors.NewInternalError("生成项目ID失败: " + err.Error())
	}

	project.ProjectID = projectID
	project.Status = models.StatusEnabled

	return s.projectRepo.Create(ctx, project)
}

// GetProjectsByTenant 获取租户的项目列表
func (s *tenantService) GetProjectsByTenant(ctx context.Context, tenantID string) ([]models.Project, error) {
	return s.projectRepo.GetByTenantID(ctx, tenantID)
}

// CreateUser 创建用户
func (s *tenantService) CreateUser(ctx context.Context, user *models.User) error {
	// 验证租户是否存在
	_, err := s.tenantRepo.GetByID(ctx, user.TenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewWithCode(errors.ErrTenantNotFound)
		}
		return errors.NewInternalError("查询租户失败: " + err.Error())
	}

	// 生成用户 ID
	userID, err := utils.GenerateUserID(ctx, s.redis)
	if err != nil {
		return errors.NewInternalError("生成用户ID失败: " + err.Error())
	}

	user.UserID = userID
	user.Status = models.StatusEnabled

	return s.userRepo.Create(ctx, user)
}

// GetUsersByTenant 获取租户的用户列表
func (s *tenantService) GetUsersByTenant(ctx context.Context, tenantID string) ([]models.User, error) {
	return s.userRepo.GetByTenantID(ctx, tenantID)
}
