package pkgdomain

import "github.com/wwnj/happy-billing/pkg/errors"

// PackageStatus 套餐包状态
type PackageStatus string

const (
	// PackageStatusActive 激活状态（可用）
	PackageStatusActive PackageStatus = "ACTIVE"
	// PackageStatusExpired 已过期
	PackageStatusExpired PackageStatus = "EXPIRED"
	// PackageStatusExhausted 已耗尽
	PackageStatusExhausted PackageStatus = "EXHAUSTED"
	// PackageStatusCancelled 已取消
	PackageStatusCancelled PackageStatus = "CANCELLED"
)

// Validate 验证套餐包状态
func (s PackageStatus) Validate() error {
	switch s {
	case PackageStatusActive, PackageStatusExpired, PackageStatusExhausted, PackageStatusCancelled:
		return nil
	default:
		return errors.New(errors.CodeInvalidParam, "invalid package status")
	}
}

// IsActive 是否激活状态
func (s PackageStatus) IsActive() bool {
	return s == PackageStatusActive
}

// PackageType 套餐包类型
type PackageType string

const (
	// PackageTypeGPU GPU算力包
	PackageTypeGPU PackageType = "GPU"
	// PackageTypeStorage 存储包
	PackageTypeStorage PackageType = "STORAGE"
	// PackageTypeToken Token包
	PackageTypeToken PackageType = "TOKEN"
	// PackageTypeTraffic 流量包
	PackageTypeTraffic PackageType = "TRAFFIC"
)

// Validate 验证套餐包类型
func (t PackageType) Validate() error {
	switch t {
	case PackageTypeGPU, PackageTypeStorage, PackageTypeToken, PackageTypeTraffic:
		return nil
	default:
		return errors.New(errors.CodeInvalidParam, "invalid package type")
	}
}
