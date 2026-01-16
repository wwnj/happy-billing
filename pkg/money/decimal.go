package money

import (
	"database/sql/driver"
	"fmt"

	"github.com/shopspring/decimal"
)

// Decimal 金额类型(封装shopspring/decimal)
type Decimal = decimal.Decimal

// 常用常量
var (
	Zero    = decimal.Zero            // 0
	One     = decimal.NewFromInt(1)   // 1
	Hundred = decimal.NewFromInt(100) // 100
)

// NewFromInt 从整数创建金额
func NewFromInt(value int64) Decimal {
	return decimal.NewFromInt(value)
}

// NewFromFloat 从浮点数创建金额(谨慎使用,可能有精度问题)
func NewFromFloat(value float64) Decimal {
	return decimal.NewFromFloat(value)
}

// NewFromString 从字符串创建金额(推荐)
func NewFromString(value string) (Decimal, error) {
	return decimal.NewFromString(value)
}

// MustFromString 从字符串创建金额(失败会panic)
func MustFromString(value string) Decimal {
	d, err := decimal.NewFromString(value)
	if err != nil {
		panic(fmt.Sprintf("invalid decimal string: %s, error: %v", value, err))
	}
	return d
}

// Yuan 从元创建金额(保留4位小数)
func Yuan(yuan float64) Decimal {
	return decimal.NewFromFloat(yuan).Round(4)
}

// Fen 从分创建金额(人民币)
func Fen(fen int64) Decimal {
	return decimal.NewFromInt(fen).Div(Hundred)
}

// ToYuan 转换为元(float64,谨慎使用)
func ToYuan(d Decimal) float64 {
	f, _ := d.Float64()
	return f
}

// ToFen 转换为分(int64,人民币)
func ToFen(d Decimal) int64 {
	return d.Mul(Hundred).IntPart()
}

// Add 加法
func Add(a, b Decimal) Decimal {
	return a.Add(b)
}

// Sub 减法
func Sub(a, b Decimal) Decimal {
	return a.Sub(b)
}

// Mul 乘法
func Mul(a, b Decimal) Decimal {
	return a.Mul(b)
}

// Div 除法
func Div(a, b Decimal) Decimal {
	return a.Div(b)
}

// Sum 求和
func Sum(amounts ...Decimal) Decimal {
	total := Zero
	for _, amount := range amounts {
		total = total.Add(amount)
	}
	return total
}

// Abs 绝对值
func Abs(d Decimal) Decimal {
	return d.Abs()
}

// Neg 取负
func Neg(d Decimal) Decimal {
	return d.Neg()
}

// Round 四舍五入到指定小数位
func Round(d Decimal, places int32) Decimal {
	return d.Round(places)
}

// RoundUp 向上取整到指定小数位
func RoundUp(d Decimal, places int32) Decimal {
	return d.RoundUp(places)
}

// RoundDown 向下取整到指定小数位
func RoundDown(d Decimal, places int32) Decimal {
	return d.RoundDown(places)
}

// IsZero 是否为零
func IsZero(d Decimal) bool {
	return d.IsZero()
}

// IsPositive 是否为正数
func IsPositive(d Decimal) bool {
	return d.GreaterThan(Zero)
}

// IsNegative 是否为负数
func IsNegative(d Decimal) bool {
	return d.LessThan(Zero)
}

// Max 返回最大值
func Max(a, b Decimal) Decimal {
	if a.GreaterThan(b) {
		return a
	}
	return b
}

// Min 返回最小值
func Min(a, b Decimal) Decimal {
	if a.LessThan(b) {
		return a
	}
	return b
}

// Compare 比较两个金额
// 返回: -1 if a < b, 0 if a == b, 1 if a > b
func Compare(a, b Decimal) int {
	return a.Cmp(b)
}

// Equal 判断两个金额是否相等
func Equal(a, b Decimal) bool {
	return a.Equal(b)
}

// GreaterThan 判断a是否大于b
func GreaterThan(a, b Decimal) bool {
	return a.GreaterThan(b)
}

// GreaterThanOrEqual 判断a是否大于等于b
func GreaterThanOrEqual(a, b Decimal) bool {
	return a.GreaterThanOrEqual(b)
}

// LessThan 判断a是否小于b
func LessThan(a, b Decimal) bool {
	return a.LessThan(b)
}

// LessThanOrEqual 判断a是否小于等于b
func LessThanOrEqual(a, b Decimal) bool {
	return a.LessThanOrEqual(b)
}

// Percentage 计算百分比(保留4位小数)
// percentage: 百分比值,如10表示10%
func Percentage(amount Decimal, percentage Decimal) Decimal {
	return amount.Mul(percentage).Div(Hundred).Round(4)
}

// DiscountAmount 计算折扣金额
// discount: 折扣率,如0.8表示8折
func DiscountAmount(amount Decimal, discount Decimal) Decimal {
	return amount.Mul(discount).Round(4)
}

// NullDecimal 可为空的Decimal类型(用于数据库)
type NullDecimal struct {
	Decimal Decimal
	Valid   bool // Valid is true if Decimal is not NULL
}

// Scan 实现sql.Scanner接口
func (nd *NullDecimal) Scan(value interface{}) error {
	if value == nil {
		nd.Decimal, nd.Valid = Zero, false
		return nil
	}
	nd.Valid = true
	return nd.Decimal.Scan(value)
}

// Value 实现driver.Valuer接口
func (nd NullDecimal) Value() (driver.Value, error) {
	if !nd.Valid {
		return nil, nil
	}
	return nd.Decimal.Value()
}
