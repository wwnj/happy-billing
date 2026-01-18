package models

import "time"

// ============================================================================
// 认证相关 DTO
// ============================================================================

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string    `json:"token"`      // JWT token
	UserID    string    `json:"user_id"`    // 用户ID
	TenantID  string    `json:"tenant_id"`  // 租户ID
	Username  string    `json:"username"`   // 用户名
	RealName  string    `json:"real_name"`  // 真实姓名
	Email     string    `json:"email"`      // 邮箱
	Phone     string    `json:"phone"`      // 手机号
	ExpiresAt time.Time `json:"expires_at"` // 过期时间
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	UserID    string `json:"user_id"`    // 用户ID
	TenantID  string `json:"tenant_id"`  // 租户ID
	Username  string `json:"username"`   // 用户名
	RealName  string `json:"real_name"`  // 真实姓名
	Email     string `json:"email"`      // 邮箱
	Phone     string `json:"phone"`      // 手机号
	IsPrimary bool   `json:"is_primary"` // 是否主账号
	Status    Status `json:"status"`     // 状态
}
