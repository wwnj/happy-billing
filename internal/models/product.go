package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ProductType 产品类型
type ProductType string

const (
	ProductTypeGPU         ProductType = "GPU"          // GPU计算
	ProductTypeCPU         ProductType = "CPU"          // CPU计算
	ProductTypeStorage     ProductType = "STORAGE"      // 存储
	ProductTypeLLMToken    ProductType = "LLM_TOKEN"    // 大语言模型Token
	ProductTypeObjectStore ProductType = "OBJECT_STORE" // 对象存储
	ProductTypeBlockStore  ProductType = "BLOCK_STORE"  // 块存储
)

// StockType 库存类型
type StockType string

const (
	StockTypeAvailable StockType = "AVAILABLE" // 有货
	StockTypeSoldOut   StockType = "SOLD_OUT"  // 售罄
)

// SpecTemplate SPU规格模板（JSON 字段）
type SpecTemplate map[string][]string

// Scan 实现 sql.Scanner 接口
func (s *SpecTemplate) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), s)
	}
	return json.Unmarshal(bytes, s)
}

// Value 实现 driver.Valuer 接口
func (s SpecTemplate) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// SpecValues SKU规格值（JSON 字段）
type SpecValues map[string]string

// Scan 实现 sql.Scanner 接口
func (s *SpecValues) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), s)
	}
	return json.Unmarshal(bytes, s)
}

// Value 实现 driver.Valuer 接口
func (s SpecValues) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// ============================================================================
// ProductCategory - 产品分类
// ============================================================================

// ProductCategory 产品分类模型
type ProductCategory struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	CategoryID   string    `gorm:"column:category_id;uniqueIndex;not null" json:"category_id"`
	CategoryCode string    `gorm:"column:category_code;uniqueIndex;not null" json:"category_code"`
	CategoryName string    `gorm:"column:category_name;not null" json:"category_name"`
	ParentID     *int64    `gorm:"column:parent_id;index" json:"parent_id,omitempty"`
	Level        *int8     `gorm:"column:level" json:"level,omitempty"`
	SortOrder    int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	Icon         *string   `gorm:"column:icon" json:"icon,omitempty"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (ProductCategory) TableName() string {
	return "product_categories"
}

// ============================================================================
// ProductSpu - 标准产品单元/产品族
// ============================================================================

// ProductSpu SPU模型（标准产品单元）
type ProductSpu struct {
	ID           int64         `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	SpuID        string        `gorm:"column:spu_id;uniqueIndex;not null" json:"spu_id"`
	SpuCode      string        `gorm:"column:spu_code;uniqueIndex;not null" json:"spu_code"`
	SpuName      string        `gorm:"column:spu_name;not null" json:"spu_name"`
	CategoryID   int64         `gorm:"column:category_id;not null;index" json:"category_id"`
	ProductType  ProductType   `gorm:"column:product_type;not null;index" json:"product_type"`
	Brand        *string       `gorm:"column:brand" json:"brand,omitempty"`
	Description  *string       `gorm:"column:description" json:"description,omitempty"`
	BillingUnit  MeteringUnit  `gorm:"column:billing_unit;not null" json:"billing_unit"`
	SpecTemplate *SpecTemplate `gorm:"column:spec_template;type:json" json:"spec_template,omitempty"`
	Status       Status        `gorm:"column:status;default:1;index" json:"status"`
	CreatedAt    time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (ProductSpu) TableName() string {
	return "product_spu"
}

// ============================================================================
// ProductSku - 库存保管单元/可售卖规格
// ============================================================================

// ProductSku SKU模型（库存保管单元）
type ProductSku struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	SkuID      string     `gorm:"column:sku_id;uniqueIndex;not null" json:"sku_id"`
	SkuCode    string     `gorm:"column:sku_code;uniqueIndex;not null" json:"sku_code"`
	SpuID      int64      `gorm:"column:spu_id;not null;index" json:"spu_id"`
	SpuCode    string     `gorm:"column:spu_code;not null;index" json:"spu_code"`
	SkuName    string     `gorm:"column:sku_name;not null" json:"sku_name"`
	SpecValues SpecValues `gorm:"column:spec_values;type:json;not null" json:"spec_values"`
	Region     *string    `gorm:"column:region;index" json:"region,omitempty"`
	StockType  *StockType `gorm:"column:stock_type" json:"stock_type,omitempty"`
	Status     Status     `gorm:"column:status;default:1;index" json:"status"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (ProductSku) TableName() string {
	return "product_sku"
}

// ============================================================================
// 请求/响应 DTO
// ============================================================================

// CreateCategoryRequest 创建产品分类请求
type CreateCategoryRequest struct {
	CategoryCode string  `json:"category_code" binding:"required"`
	CategoryName string  `json:"category_name" binding:"required"`
	ParentID     *int64  `json:"parent_id,omitempty"`
	Level        *int8   `json:"level,omitempty"`
	SortOrder    int     `json:"sort_order"`
	Icon         *string `json:"icon,omitempty"`
}

// CreateSpuRequest 创建SPU请求
type CreateSpuRequest struct {
	SpuCode      string              `json:"spu_code" binding:"required"`
	SpuName      string              `json:"spu_name" binding:"required"`
	CategoryID   int64               `json:"category_id" binding:"required"`
	ProductType  ProductType         `json:"product_type" binding:"required"`
	Brand        *string             `json:"brand,omitempty"`
	Description  *string             `json:"description,omitempty"`
	BillingUnit  MeteringUnit        `json:"billing_unit" binding:"required"`
	SpecTemplate map[string][]string `json:"spec_template,omitempty"`
}

// CreateSkuRequest 创建SKU请求
type CreateSkuRequest struct {
	SkuCode    string            `json:"sku_code" binding:"required"`
	SpuID      int64             `json:"spu_id" binding:"required"`
	SkuName    string            `json:"sku_name" binding:"required"`
	SpecValues map[string]string `json:"spec_values" binding:"required"`
	Region     *string           `json:"region,omitempty"`
	StockType  *StockType        `json:"stock_type,omitempty"`
}

// ProductSpuResponse SPU响应（包含分类信息）
type ProductSpuResponse struct {
	ProductSpu
	Category *ProductCategory `json:"category,omitempty"`
}

// ProductSkuResponse SKU响应（包含SPU和分类信息）
type ProductSkuResponse struct {
	ProductSku
	Spu      *ProductSpu      `json:"spu,omitempty"`
	Category *ProductCategory `json:"category,omitempty"`
}

// CategoryTreeNode 分类树节点
type CategoryTreeNode struct {
	ProductCategory
	Children []CategoryTreeNode `json:"children,omitempty"`
}

// SpuListQueryRequest SPU列表查询请求
type SpuListQueryRequest struct {
	Pagination
	CategoryID  *int64       `json:"category_id" form:"category_id"`
	ProductType *ProductType `json:"product_type" form:"product_type"`
	Brand       *string      `json:"brand" form:"brand"`
	Status      *Status      `json:"status" form:"status"`
	Keyword     *string      `json:"keyword" form:"keyword"` // 搜索关键词（匹配名称/编码）
}

// SkuListQueryRequest SKU列表查询请求
type SkuListQueryRequest struct {
	Pagination
	SpuID     *int64     `json:"spu_id" form:"spu_id"`
	Region    *string    `json:"region" form:"region"`
	StockType *StockType `json:"stock_type" form:"stock_type"`
	Status    *Status    `json:"status" form:"status"`
	Keyword   *string    `json:"keyword" form:"keyword"` // 搜索关键词（匹配名称/编码）
}
