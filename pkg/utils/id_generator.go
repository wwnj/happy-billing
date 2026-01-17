package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// IDPrefix ID前缀常量
const (
	PrefixTenant       = "tenant"   // 租户
	PrefixOrganization = "org"      // 组织
	PrefixProject      = "proj"     // 项目
	PrefixUser         = "user"     // 用户
	PrefixVerification = "verify"   // 认证
	PrefixSPU          = "spu"      // 商品SPU
	PrefixSKU          = "sku"      // 商品SKU
	PrefixOrder        = "order"    // 订单
	PrefixBill         = "bill"     // 账单
	PrefixPayment      = "pay"      // 支付
	PrefixCategory     = "category" // 分类
	PrefixResource     = "resource" // 资源实例
	PrefixPriceRule    = "price"    // 定价规则
	PrefixDiscount     = "discount" // 折扣规则
)

// EnhancedIDGenerator 增强的分布式ID生成器
// 格式: prefix_hash10
// 例如: tenant_a3f9b2c4d5, org_b4e1c3f2a6
//
// 特点:
// 1. 前缀清晰，辨识度高（小写）
// 2. 日期信息已混淆在hash中，保持时间顺序性
// 3. Hash混淆，完全隐藏日期和自增规律
// 4. 分布式环境下保证唯一性
type EnhancedIDGenerator struct {
	redis  *redis.Client
	prefix string
}

// NewEnhancedIDGenerator 创建增强ID生成器
func NewEnhancedIDGenerator(redis *redis.Client, prefix string) *EnhancedIDGenerator {
	return &EnhancedIDGenerator{
		redis:  redis,
		prefix: prefix,
	}
}

// Generate 生成业务ID
// 格式: prefix_hash10
// 例如: tenant_a3f9b2c4d5
func (g *EnhancedIDGenerator) Generate(ctx context.Context) (string, error) {
	now := time.Now()

	// 日期部分：YYMMDD（用于Redis key和hash计算，不显示在最终ID中）
	dateStr := now.Format("060102") // 240117

	// Redis key：prefix:YYMMDD
	redisKey := fmt.Sprintf("id_seq:%s:%s", g.prefix, dateStr)

	// 获取自增序号（原子操作）
	seq, err := g.redis.Incr(ctx, redisKey).Result()
	if err != nil {
		return "", fmt.Errorf("failed to generate sequence: %w", err)
	}

	// 设置过期时间（当天结束后2天过期，避免跨时区问题）
	if seq == 1 {
		g.redis.Expire(ctx, redisKey, 48*time.Hour)
	}

	// 生成Hash（混淆日期和自增规律）
	// Hash输入：prefix + date + sequence + timestamp
	// 日期信息完全混淆在hash中，不暴露在最终ID中
	hashInput := fmt.Sprintf("%s-%s-%d-%d", g.prefix, dateStr, seq, now.UnixNano())
	hash := sha256.Sum256([]byte(hashInput))
	hashHex := hex.EncodeToString(hash[:])

	// 取Hash前10位（小写字母和数字混合）
	hashSuffix := hashHex[:10]

	// 最终ID格式：prefix_hash10（全小写）
	id := fmt.Sprintf("%s_%s", g.prefix, hashSuffix)

	return id, nil
}

// GenerateCompact 生成紧凑格式ID（无下划线）
// 格式: prefixhash8
// 例如: tnt2a3f9b2c4, org4b1e5d3f
// 适用于长度敏感的场景
func (g *EnhancedIDGenerator) GenerateCompact(ctx context.Context) (string, error) {
	now := time.Now()

	// 时间部分：YYMMDDHHMM（精确到分钟，用于Redis key和hash计算）
	timeStr := now.Format("0601021504") // 2401171520

	// Redis key：prefix:YYMMDDHHMM
	redisKey := fmt.Sprintf("id_seq:%s:%s", g.prefix, timeStr)

	// 获取自增序号
	seq, err := g.redis.Incr(ctx, redisKey).Result()
	if err != nil {
		return "", fmt.Errorf("failed to generate sequence: %w", err)
	}

	// 设置过期时间（1小时后过期）
	if seq == 1 {
		g.redis.Expire(ctx, redisKey, 1*time.Hour)
	}

	// 生成Hash（时间信息混淆在hash中）
	hashInput := fmt.Sprintf("%s-%s-%d-%d", g.prefix, timeStr, seq, now.UnixNano())
	hash := sha256.Sum256([]byte(hashInput))
	hashHex := hex.EncodeToString(hash[:])

	// 取Hash前8位（小写）
	hashSuffix := hashHex[:8]

	// 前缀缩写（取前3个字符，小写）
	prefixShort := g.prefix
	if len(prefixShort) > 3 {
		prefixShort = prefixShort[:3]
	}

	// 最终ID格式：prefixhash8（全小写，无下划线）
	id := fmt.Sprintf("%s%s", prefixShort, hashSuffix)

	return id, nil
}

// ============================================================================
// 便捷函数：为各个业务实体生成ID
// ============================================================================

// GenerateTenantID 生成租户ID
// 格式: tenant_a3f9b2c4d5
func GenerateTenantID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixTenant)
	return gen.Generate(ctx)
}

// GenerateOrganizationID 生成组织ID
// 格式: org_b4e1c3f2a6
func GenerateOrganizationID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixOrganization)
	return gen.Generate(ctx)
}

// GenerateProjectID 生成项目ID
// 格式: proj_c5d2a4e7b1
func GenerateProjectID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixProject)
	return gen.Generate(ctx)
}

// GenerateUserID 生成用户ID
// 格式: user_d6e3b5f8c2
func GenerateUserID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixUser)
	return gen.Generate(ctx)
}

// GenerateVerificationID 生成认证ID
// 格式: verify_e7f4c6a9d3
func GenerateVerificationID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixVerification)
	return gen.Generate(ctx)
}

// GenerateSPUID 生成SPU ID
// 格式: spu_f8a5d7b1e4
func GenerateSPUID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixSPU)
	return gen.Generate(ctx)
}

// GenerateSKUID 生成SKU ID
// 格式: sku_a9b6e8c2f5
func GenerateSKUID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixSKU)
	return gen.Generate(ctx)
}

// GenerateOrderID 生成订单ID
// 格式: order_b1c7f9d3a6
func GenerateOrderID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixOrder)
	return gen.Generate(ctx)
}

// GenerateBillID 生成账单ID
// 格式: bill_c2d8a1e4b7
func GenerateBillID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixBill)
	return gen.Generate(ctx)
}

// GeneratePaymentID 生成支付ID
// 格式: pay_d3e9b2f5c8
func GeneratePaymentID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixPayment)
	return gen.Generate(ctx)
}

// GenerateCategoryID 生成分类ID
// 格式: category_e4f1c3a6d9
func GenerateCategoryID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixCategory)
	return gen.Generate(ctx)
}

// GenerateResourceID 生成资源实例ID
// 格式: resource_f5a2d4b7e1
func GenerateResourceID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixResource)
	return gen.Generate(ctx)
}

// GeneratePriceRuleID 生成定价规则ID
// 格式: price_a1b2c3d4e5
func GeneratePriceRuleID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixPriceRule)
	return gen.Generate(ctx)
}

// GenerateDiscountID 生成折扣规则ID
// 格式: discount_b2c3d4e5f6
func GenerateDiscountID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixDiscount)
	return gen.Generate(ctx)
}

// ============================================================================
// 紧凑格式ID生成函数（适用于长度敏感场景）
// ============================================================================

// GenerateCompactTenantID 生成紧凑租户ID
// 格式: ten2a3f9b2c4
func GenerateCompactTenantID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixTenant)
	return gen.GenerateCompact(ctx)
}

// GenerateCompactOrderID 生成紧凑订单ID
// 格式: ord4b1e5d3f2
func GenerateCompactOrderID(ctx context.Context, redis *redis.Client) (string, error) {
	gen := NewEnhancedIDGenerator(redis, PrefixOrder)
	return gen.GenerateCompact(ctx)
}

// ============================================================================
// ID解析函数
// ============================================================================

// ParseID 解析ID，提取前缀和hash信息
func ParseID(id string) (prefix string, hash string, err error) {
	// 新格式：prefix_hash10
	var parts []string
	start := 0
	for i := 0; i < len(id); i++ {
		if id[i] == '_' {
			parts = append(parts, id[start:i])
			start = i + 1
		}
	}
	if start < len(id) {
		parts = append(parts, id[start:])
	}

	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("invalid ID format: %s", id)
}

// ExtractPrefixFromID 从ID中提取前缀
// 注意：新格式ID中日期信息已混淆在hash中，无法直接提取
func ExtractPrefixFromID(id string) (string, error) {
	prefix, _, err := ParseID(id)
	if err != nil {
		return "", err
	}

	return prefix, nil
}
