package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	pkgdomain "github.com/wwnj/happy-billing/internal/domain/package"
	"github.com/wwnj/happy-billing/pkg/money"
)

// PackageDO 套餐包数据对象
type PackageDO struct {
	ID             string     `gorm:"primaryKey;column:id;type:varchar(64)"`
	PackageNo      string     `gorm:"uniqueIndex;column:package_no;type:varchar(64);not null"`
	TenantID       string     `gorm:"index:idx_tenant_user;column:tenant_id;type:varchar(64);not null"`
	UserID         string     `gorm:"index:idx_tenant_user;column:user_id;type:varchar(64);not null"`
	Type           string     `gorm:"index;column:type;type:varchar(20);not null"`
	Status         string     `gorm:"index;column:status;type:varchar(20);not null"`
	Name           string     `gorm:"column:name;type:varchar(255)"`
	Description    string     `gorm:"column:description;type:text"`
	TotalQuota     string     `gorm:"column:total_quota;type:decimal(20,4);not null"`
	UsedQuota      string     `gorm:"column:used_quota;type:decimal(20,4);not null;default:0"`
	RemainingQuota string     `gorm:"column:remaining_quota;type:decimal(20,4);not null"`
	QuotaUnit      string     `gorm:"column:quota_unit;type:varchar(20);not null"`
	Price          string     `gorm:"column:price;type:decimal(20,4);not null"`
	Currency       string     `gorm:"column:currency;type:varchar(10);not null;default:'CNY'"`
	PurchasedAt    time.Time  `gorm:"column:purchased_at;type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	ValidFrom      time.Time  `gorm:"column:valid_from;type:timestamp;not null"`
	ValidTo        time.Time  `gorm:"index;column:valid_to;type:timestamp;not null"`
	ExhaustedAt    *time.Time `gorm:"column:exhausted_at;type:timestamp"`
	CancelledAt    *time.Time `gorm:"column:cancelled_at;type:timestamp"`
	Metadata       JSONMap    `gorm:"column:metadata;type:jsonb"`
	CreatedAt      time.Time  `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName 指定表名
func (PackageDO) TableName() string {
	return "packages"
}

// JSONMap 用于JSONB字段
type JSONMap map[string]interface{}

// Scan 实现sql.Scanner接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		*j = make(map[string]interface{})
		return nil
	}

	result := make(map[string]interface{})
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		*j = make(map[string]interface{})
		return err
	}

	*j = result
	return nil
}

// Value 实现driver.Valuer接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// ToDomain 转换为领域模型
func (do *PackageDO) ToDomain() (*pkgdomain.Package, error) {
	totalQuota, err := money.NewFromString(do.TotalQuota)
	if err != nil {
		return nil, err
	}

	usedQuota, err := money.NewFromString(do.UsedQuota)
	if err != nil {
		return nil, err
	}

	remainingQuota, err := money.NewFromString(do.RemainingQuota)
	if err != nil {
		return nil, err
	}

	price, err := money.NewFromString(do.Price)
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]interface{})
	if do.Metadata != nil {
		metadata = do.Metadata
	}

	pkg := &pkgdomain.Package{
		ID:             do.ID,
		PackageNo:      do.PackageNo,
		TenantID:       do.TenantID,
		UserID:         do.UserID,
		Type:           pkgdomain.PackageType(do.Type),
		Status:         pkgdomain.PackageStatus(do.Status),
		Name:           do.Name,
		Description:    do.Description,
		TotalQuota:     totalQuota,
		UsedQuota:      usedQuota,
		RemainingQuota: remainingQuota,
		QuotaUnit:      do.QuotaUnit,
		Price:          price,
		Currency:       do.Currency,
		PurchasedAt:    do.PurchasedAt,
		ValidFrom:      do.ValidFrom,
		ValidTo:        do.ValidTo,
		ExhaustedAt:    do.ExhaustedAt,
		CancelledAt:    do.CancelledAt,
		Metadata:       metadata,
	}

	return pkg, nil
}

// FromDomain 从领域模型转换
func (do *PackageDO) FromDomain(pkg *pkgdomain.Package) {
	do.ID = pkg.ID
	do.PackageNo = pkg.PackageNo
	do.TenantID = pkg.TenantID
	do.UserID = pkg.UserID
	do.Type = string(pkg.Type)
	do.Status = string(pkg.Status)
	do.Name = pkg.Name
	do.Description = pkg.Description
	do.TotalQuota = pkg.TotalQuota.String()
	do.UsedQuota = pkg.UsedQuota.String()
	do.RemainingQuota = pkg.RemainingQuota.String()
	do.QuotaUnit = pkg.QuotaUnit
	do.Price = pkg.Price.String()
	do.Currency = pkg.Currency
	do.PurchasedAt = pkg.PurchasedAt
	do.ValidFrom = pkg.ValidFrom
	do.ValidTo = pkg.ValidTo
	do.ExhaustedAt = pkg.ExhaustedAt
	do.CancelledAt = pkg.CancelledAt

	if pkg.Metadata != nil {
		do.Metadata = pkg.Metadata
	} else {
		do.Metadata = make(map[string]interface{})
	}
}
