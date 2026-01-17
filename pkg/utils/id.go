package utils

import (
	"fmt"
	"sync"
	"time"
)

// IDGenerator 业务 ID 生成器
type IDGenerator struct {
	prefix  string
	counter int
	mu      sync.Mutex
	lastDay string
}

// NewIDGenerator 创建 ID 生成器
func NewIDGenerator(prefix string) *IDGenerator {
	return &IDGenerator{
		prefix:  prefix,
		counter: 0,
		lastDay: "",
	}
}

// Generate 生成业务 ID
// 格式: prefix + YYYYMMDD + 序号
// 例如: T20240117001, ORD20240117001
func (g *IDGenerator) Generate() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	today := time.Now().Format("20060102")

	// 如果是新的一天，重置计数器
	if today != g.lastDay {
		g.counter = 0
		g.lastDay = today
	}

	g.counter++
	return fmt.Sprintf("%s%s%03d", g.prefix, today, g.counter)
}

// 预定义的 ID 生成器
var (
	TenantIDGen   = NewIDGenerator("T")     // 租户ID: T20240117001
	OrgIDGen      = NewIDGenerator("ORG")   // 组织ID: ORG20240117001
	ProjectIDGen  = NewIDGenerator("PRJ")   // 项目ID: PRJ20240117001
	UserIDGen     = NewIDGenerator("U")     // 用户ID: U20240117001
	CategoryIDGen = NewIDGenerator("CAT")   // 分类ID: CAT20240117001
	SPUIDGen      = NewIDGenerator("SPU")   // SPU ID: SPU20240117001
	SKUIDGen      = NewIDGenerator("SKU")   // SKU ID: SKU20240117001
	PriceIDGen    = NewIDGenerator("PRICE") // 价格ID: PRICE20240117001
	OrderIDGen    = NewIDGenerator("ORD")   // 订单ID: ORD20240117001
	BillIDGen     = NewIDGenerator("BILL")  // 账单ID: BILL20240117001
	PaymentIDGen  = NewIDGenerator("PAY")   // 支付ID: PAY20240117001
)
