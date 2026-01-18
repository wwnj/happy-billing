package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务接口
type AuthService interface {
	// Login 用户登录
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	// GetUserInfo 获取用户信息
	GetUserInfo(ctx context.Context, userID string) (*models.UserInfoResponse, error)
}

// authService 认证服务实现
type authService struct {
	db *gorm.DB
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB) AuthService {
	return &authService{
		db: db,
	}
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// 1. 查找用户
	var user models.User
	err := s.db.WithContext(ctx).
		Where("username = ? AND status = ?", req.Username, models.StatusEnabled).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrUnauthorized, "用户名或密码错误")
		}
		return nil, err
	}

	// 2. 验证密码
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, errors.New(errors.ErrUnauthorized, "用户未设置密码，请联系管理员")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New(errors.ErrUnauthorized, "用户名或密码错误")
	}

	// 3. 生成Token (简单的随机token，生产环境应该使用JWT)
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// 4. 构造响应
	expiresAt := time.Now().Add(24 * time.Hour) // Token 24小时有效期

	realName := ""
	if user.RealName != nil {
		realName = *user.RealName
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	phone := ""
	if user.Phone != nil {
		phone = *user.Phone
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	return &models.LoginResponse{
		Token:     token,
		UserID:    user.UserID,
		TenantID:  user.TenantID,
		Username:  username,
		RealName:  realName,
		Email:     email,
		Phone:     phone,
		ExpiresAt: expiresAt,
	}, nil
}

// GetUserInfo 获取用户信息
func (s *authService) GetUserInfo(ctx context.Context, userID string) (*models.UserInfoResponse, error) {
	var user models.User
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrUserNotFound, "用户不存在")
		}
		return nil, err
	}

	realName := ""
	if user.RealName != nil {
		realName = *user.RealName
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	phone := ""
	if user.Phone != nil {
		phone = *user.Phone
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	return &models.UserInfoResponse{
		UserID:    user.UserID,
		TenantID:  user.TenantID,
		Username:  username,
		RealName:  realName,
		Email:     email,
		Phone:     phone,
		IsPrimary: user.IsPrimary,
		Status:    user.Status,
	}, nil
}

// generateToken 生成随机token (简化实现，生产环境应该使用JWT)
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashPassword 加密密码 (辅助函数，用于创建测试用户)
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
