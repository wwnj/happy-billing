package repository

import (
	"context"

	"github.com/wwnj/happy-billing/internal/models"
	"gorm.io/gorm"
)

// TenantRepository 租户仓储接口
type TenantRepository interface {
	// Tenant CRUD
	Create(ctx context.Context, tenant *models.Tenant) error
	GetByID(ctx context.Context, tenantID string) (*models.Tenant, error)
	GetByCode(ctx context.Context, tenantCode string) (*models.Tenant, error)
	Update(ctx context.Context, tenant *models.Tenant) error
	Delete(ctx context.Context, tenantID string) error
	List(ctx context.Context, tenantType *models.TenantType, offset, limit int) ([]models.Tenant, int64, error)

	// Verification
	UpdateVerified(ctx context.Context, tenantID string, verified bool, verifyType models.VerifyType, info *models.VerifiedInfo) error
}

// OrganizationRepository 组织仓储接口
type OrganizationRepository interface {
	Create(ctx context.Context, org *models.Organization) error
	GetByID(ctx context.Context, organizationID string) (*models.Organization, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]models.Organization, error)
	Update(ctx context.Context, org *models.Organization) error
	Delete(ctx context.Context, organizationID string) error
	GetChildren(ctx context.Context, parentOrgID string) ([]models.Organization, error)
}

// ProjectRepository 项目仓储接口
type ProjectRepository interface {
	Create(ctx context.Context, project *models.Project) error
	GetByID(ctx context.Context, projectID string) (*models.Project, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]models.Project, error)
	GetByOrganizationID(ctx context.Context, organizationID string) ([]models.Project, error)
	Update(ctx context.Context, project *models.Project) error
	Delete(ctx context.Context, projectID string) error
	List(ctx context.Context, tenantID string, offset, limit int) ([]models.Project, int64, error)
}

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, userID string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByPhone(ctx context.Context, phone string) (*models.User, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]models.User, error)
	GetPrimaryUser(ctx context.Context, tenantID string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, userID string) error
}

// VerificationRepository 认证仓储接口
type VerificationRepository interface {
	Create(ctx context.Context, verification *models.Verification) error
	GetByID(ctx context.Context, verificationID string) (*models.Verification, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]models.Verification, error)
	Update(ctx context.Context, verification *models.Verification) error
	UpdateStatus(ctx context.Context, verificationID string, status models.VerificationStatus, rejectReason *string) error
	List(ctx context.Context, status *models.VerificationStatus, offset, limit int) ([]models.Verification, int64, error)
}

// ============================================================================
// Tenant Repository Implementation
// ============================================================================

type tenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository 创建租户仓储实例
func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *tenantRepository) GetByID(ctx context.Context, tenantID string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) GetByCode(ctx context.Context, tenantCode string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.WithContext(ctx).Where("tenant_code = ?", tenantCode).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *models.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

func (r *tenantRepository) Delete(ctx context.Context, tenantID string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&models.Tenant{}).Error
}

func (r *tenantRepository) List(ctx context.Context, tenantType *models.TenantType, offset, limit int) ([]models.Tenant, int64, error) {
	var tenants []models.Tenant
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Tenant{})
	if tenantType != nil {
		query = query.Where("tenant_type = ?", *tenantType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset(offset).Limit(limit).Find(&tenants).Error
	return tenants, total, err
}

func (r *tenantRepository) UpdateVerified(ctx context.Context, tenantID string, verified bool, verifyType models.VerifyType, info *models.VerifiedInfo) error {
	updates := map[string]interface{}{
		"verified":      verified,
		"verified_type": verifyType,
	}
	if info != nil {
		updates["verified_info"] = info
	}
	if verified {
		updates["verified_at"] = gorm.Expr("NOW()")
	}

	return r.db.WithContext(ctx).Model(&models.Tenant{}).Where("tenant_id = ?", tenantID).Updates(updates).Error
}

// ============================================================================
// Organization Repository Implementation
// ============================================================================

type organizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository 创建组织仓储实例
func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) Create(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *organizationRepository) GetByID(ctx context.Context, organizationID string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) GetByTenantID(ctx context.Context, tenantID string) ([]models.Organization, error) {
	var orgs []models.Organization
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&orgs).Error
	return orgs, err
}

func (r *organizationRepository) Update(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

func (r *organizationRepository) Delete(ctx context.Context, organizationID string) error {
	return r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Delete(&models.Organization{}).Error
}

func (r *organizationRepository) GetChildren(ctx context.Context, parentOrgID string) ([]models.Organization, error) {
	var orgs []models.Organization
	err := r.db.WithContext(ctx).Where("parent_organization_id = ?", parentOrgID).Find(&orgs).Error
	return orgs, err
}

// ============================================================================
// Project Repository Implementation
// ============================================================================

type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository 创建项目仓储实例
func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *projectRepository) GetByID(ctx context.Context, projectID string) (*models.Project, error) {
	var project models.Project
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) GetByTenantID(ctx context.Context, tenantID string) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&projects).Error
	return projects, err
}

func (r *projectRepository) GetByOrganizationID(ctx context.Context, organizationID string) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Find(&projects).Error
	return projects, err
}

func (r *projectRepository) Update(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *projectRepository) Delete(ctx context.Context, projectID string) error {
	return r.db.WithContext(ctx).Where("project_id = ?", projectID).Delete(&models.Project{}).Error
}

func (r *projectRepository) List(ctx context.Context, tenantID string, offset, limit int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Project{}).Where("tenant_id = ?", tenantID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset(offset).Limit(limit).Find(&projects).Error
	return projects, total, err
}

// ============================================================================
// User Repository Implementation
// ============================================================================

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByTenantID(ctx context.Context, tenantID string) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&users).Error
	return users, err
}

func (r *userRepository) GetPrimaryUser(ctx context.Context, tenantID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND is_primary = ?", tenantID, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.User{}).Error
}

// ============================================================================
// Verification Repository Implementation
// ============================================================================

type verificationRepository struct {
	db *gorm.DB
}

// NewVerificationRepository 创建认证仓储实例
func NewVerificationRepository(db *gorm.DB) VerificationRepository {
	return &verificationRepository{db: db}
}

func (r *verificationRepository) Create(ctx context.Context, verification *models.Verification) error {
	return r.db.WithContext(ctx).Create(verification).Error
}

func (r *verificationRepository) GetByID(ctx context.Context, verificationID string) (*models.Verification, error) {
	var verification models.Verification
	err := r.db.WithContext(ctx).Where("verification_id = ?", verificationID).First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

func (r *verificationRepository) GetByTenantID(ctx context.Context, tenantID string) ([]models.Verification, error) {
	var verifications []models.Verification
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&verifications).Error
	return verifications, err
}

func (r *verificationRepository) Update(ctx context.Context, verification *models.Verification) error {
	return r.db.WithContext(ctx).Save(verification).Error
}

func (r *verificationRepository) UpdateStatus(ctx context.Context, verificationID string, status models.VerificationStatus, rejectReason *string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == models.VerificationStatusApproved {
		updates["verified_at"] = gorm.Expr("NOW()")
	}
	if rejectReason != nil {
		updates["reject_reason"] = *rejectReason
	}

	return r.db.WithContext(ctx).Model(&models.Verification{}).Where("verification_id = ?", verificationID).Updates(updates).Error
}

func (r *verificationRepository) List(ctx context.Context, status *models.VerificationStatus, offset, limit int) ([]models.Verification, int64, error) {
	var verifications []models.Verification
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Verification{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&verifications).Error
	return verifications, total, err
}
