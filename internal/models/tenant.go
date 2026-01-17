package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// VerifyType 认证类型
type VerifyType string

const (
	VerifyTypeIndividual VerifyType = "INDIVIDUAL" // 个人认证
	VerifyTypeEnterprise VerifyType = "ENTERPRISE" // 企业认证
)

// VerificationStatus 认证状态
type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "PENDING"  // 待审核
	VerificationStatusApproved VerificationStatus = "APPROVED" // 已通过
	VerificationStatusRejected VerificationStatus = "REJECTED" // 已拒绝
)

// OrgType 组织类型
type OrgType string

const (
	OrgTypePersonal      OrgType = "PERSONAL"      // 个人工作区
	OrgTypeHeadquarters  OrgType = "HEADQUARTERS"  // 总部
	OrgTypeBranch        OrgType = "BRANCH"        // 分公司
	OrgTypeDepartment    OrgType = "DEPARTMENT"    // 部门
	OrgTypeSubDepartment OrgType = "SUBDEPARTMENT" // 子部门
)

// VerifiedInfo 认证信息（JSON 字段）
type VerifiedInfo struct {
	RealName    string `json:"real_name,omitempty"`    // 真实姓名
	IDCard      string `json:"id_card,omitempty"`      // 身份证号
	CompanyName string `json:"company_name,omitempty"` // 公司名称
	CreditCode  string `json:"credit_code,omitempty"`  // 统一社会信用代码
}

// Scan 实现 sql.Scanner 接口
func (v *VerifiedInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), v)
	}
	return json.Unmarshal(bytes, v)
}

// Value 实现 driver.Valuer 接口
func (v VerifiedInfo) Value() (driver.Value, error) {
	if v.RealName == "" && v.IDCard == "" && v.CompanyName == "" && v.CreditCode == "" {
		return nil, nil
	}
	return json.Marshal(v)
}

// ============================================================================
// Tenant - 租户
// ============================================================================

// Tenant 租户模型
type Tenant struct {
	ID                int64         `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	TenantID          string        `gorm:"column:tenant_id;uniqueIndex;not null" json:"tenant_id"`
	TenantCode        string        `gorm:"column:tenant_code;uniqueIndex;not null" json:"tenant_code"`
	Name              string        `gorm:"column:name;not null" json:"name"`
	TenantType        TenantType    `gorm:"column:tenant_type;not null" json:"tenant_type"`
	PreferredCurrency string        `gorm:"column:preferred_currency;default:CNY" json:"preferred_currency"`
	Verified          bool          `gorm:"column:verified;default:0" json:"verified"`
	VerifiedType      *VerifyType   `gorm:"column:verified_type" json:"verified_type,omitempty"`
	VerifiedAt        *time.Time    `gorm:"column:verified_at" json:"verified_at,omitempty"`
	VerifiedInfo      *VerifiedInfo `gorm:"column:verified_info;type:json" json:"verified_info,omitempty"`
	Status            Status        `gorm:"column:status;default:1" json:"status"`
	CreatedAt         time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Tenant) TableName() string {
	return "tenants"
}

// ============================================================================
// Organization - 组织
// ============================================================================

// Organization 组织模型
type Organization struct {
	ID                   int64     `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	OrganizationID       string    `gorm:"column:organization_id;uniqueIndex;not null" json:"organization_id"`
	TenantID             string    `gorm:"column:tenant_id;not null;index" json:"tenant_id"`
	ParentOrganizationID *string   `gorm:"column:parent_organization_id;index" json:"parent_organization_id,omitempty"`
	OrgCode              string    `gorm:"column:org_code;uniqueIndex;not null" json:"org_code"`
	Name                 string    `gorm:"column:name;not null" json:"name"`
	OrgType              *OrgType  `gorm:"column:org_type" json:"org_type,omitempty"`
	Level                int       `gorm:"column:level;default:1" json:"level"`
	Status               Status    `gorm:"column:status;default:1" json:"status"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Organization) TableName() string {
	return "organizations"
}

// ============================================================================
// Project - 项目
// ============================================================================

// Project 项目模型
type Project struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	ProjectID      string    `gorm:"column:project_id;uniqueIndex;not null" json:"project_id"`
	TenantID       string    `gorm:"column:tenant_id;not null;index" json:"tenant_id"`
	OrganizationID string    `gorm:"column:organization_id;not null;index" json:"organization_id"`
	ProjectCode    string    `gorm:"column:project_code;uniqueIndex;not null" json:"project_code"`
	Name           string    `gorm:"column:name;not null" json:"name"`
	Description    *string   `gorm:"column:description" json:"description,omitempty"`
	Status         Status    `gorm:"column:status;default:1" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}

// ============================================================================
// User - 用户
// ============================================================================

// User 用户模型
type User struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	UserID       string    `gorm:"column:user_id;uniqueIndex;not null" json:"user_id"`
	TenantID     string    `gorm:"column:tenant_id;not null;index" json:"tenant_id"`
	IsPrimary    bool      `gorm:"column:is_primary;default:0" json:"is_primary"`
	Username     *string   `gorm:"column:username" json:"username,omitempty"`
	RealName     *string   `gorm:"column:real_name" json:"real_name,omitempty"`
	IDCard       *string   `gorm:"column:id_card" json:"id_card,omitempty"`
	Email        *string   `gorm:"column:email;index" json:"email,omitempty"`
	Phone        *string   `gorm:"column:phone;index" json:"phone,omitempty"`
	PasswordHash *string   `gorm:"column:password_hash" json:"-"` // 密码哈希不返回给前端
	Status       Status    `gorm:"column:status;default:1" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// ============================================================================
// Verification - 认证
// ============================================================================

// Verification 认证模型
type Verification struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	VerificationID string     `gorm:"column:verification_id;uniqueIndex;not null" json:"verification_id"`
	TenantID       string     `gorm:"column:tenant_id;not null;index" json:"tenant_id"`
	UserID         string     `gorm:"column:user_id;not null;index" json:"user_id"`
	VerifyType     VerifyType `gorm:"column:verify_type;not null" json:"verify_type"`

	// 个人认证字段
	RealName    *string `gorm:"column:real_name" json:"real_name,omitempty"`
	IDCard      *string `gorm:"column:id_card" json:"id_card,omitempty"`
	IDCardFront *string `gorm:"column:id_card_front" json:"id_card_front,omitempty"`
	IDCardBack  *string `gorm:"column:id_card_back" json:"id_card_back,omitempty"`

	// 企业认证字段
	CompanyName *string `gorm:"column:company_name" json:"company_name,omitempty"`
	CreditCode  *string `gorm:"column:credit_code" json:"credit_code,omitempty"`
	LicenseURL  *string `gorm:"column:license_url" json:"license_url,omitempty"`
	LegalPerson *string `gorm:"column:legal_person" json:"legal_person,omitempty"`

	Status       VerificationStatus `gorm:"column:status;not null;default:PENDING;index" json:"status"`
	RejectReason *string            `gorm:"column:reject_reason" json:"reject_reason,omitempty"`
	VerifiedAt   *time.Time         `gorm:"column:verified_at" json:"verified_at,omitempty"`
	VerifiedBy   *string            `gorm:"column:verified_by" json:"verified_by,omitempty"`
	CreatedAt    time.Time          `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time          `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Verification) TableName() string {
	return "verifications"
}

// ============================================================================
// 请求/响应 DTO
// ============================================================================

// RegisterIndividualRequest 个人用户注册请求
type RegisterIndividualRequest struct {
	Name  string `json:"name" binding:"required"`        // 姓名
	Email string `json:"email" binding:"required,email"` // 邮箱
	Phone string `json:"phone" binding:"required"`       // 手机号
}

// RegisterEnterpriseRequest 企业用户注册请求
type RegisterEnterpriseRequest struct {
	CompanyName string `json:"company_name" binding:"required"` // 公司名称
	Email       string `json:"email" binding:"required,email"`  // 邮箱
	Phone       string `json:"phone" binding:"required"`        // 手机号
}

// TenantResponse 租户响应（包含关联信息）
type TenantResponse struct {
	Tenant         *Tenant        `json:"tenant"`
	Organizations  []Organization `json:"organizations,omitempty"`
	DefaultProject *Project       `json:"default_project,omitempty"`
	PrimaryUser    *User          `json:"primary_user,omitempty"`
}
