package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math"
	"strconv"
)

// MoneyUtils 金额工具（分为单位）
type MoneyUtils struct{}

var Money = &MoneyUtils{}

// Yuan2Fen 元转分
func (m *MoneyUtils) Yuan2Fen(yuan float64) int64 {
	return int64(math.Round(yuan * 100))
}

// Fen2Yuan 分转元
func (m *MoneyUtils) Fen2Yuan(fen int64) float64 {
	return float64(fen) / 100.0
}

// FormatYuan 格式化金额为元（保留2位小数）
func (m *MoneyUtils) FormatYuan(fen int64) string {
	yuan := m.Fen2Yuan(fen)
	return strconv.FormatFloat(yuan, 'f', 2, 64)
}

// Round 四舍五入到指定小数位
func (m *MoneyUtils) Round(value float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(value*multiplier) / multiplier
}

// StringUtils 字符串工具
type StringUtils struct{}

var String = &StringUtils{}

// MD5 计算字符串的 MD5 值
func (s *StringUtils) MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// IsEmpty 判断字符串是否为空
func (s *StringUtils) IsEmpty(str string) bool {
	return len(str) == 0
}

// IsNotEmpty 判断字符串是否非空
func (s *StringUtils) IsNotEmpty(str string) bool {
	return len(str) > 0
}

// DefaultIfEmpty 如果为空则返回默认值
func (s *StringUtils) DefaultIfEmpty(str, defaultStr string) string {
	if s.IsEmpty(str) {
		return defaultStr
	}
	return str
}

// SliceUtils 切片工具
type SliceUtils struct{}

var Slice = &SliceUtils{}

// ContainsString 判断字符串切片是否包含指定元素
func (s *SliceUtils) ContainsString(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// ContainsInt 判断整数切片是否包含指定元素
func (s *SliceUtils) ContainsInt(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// RemoveDuplicateStrings 去除字符串切片中的重复元素
func (s *SliceUtils) RemoveDuplicateStrings(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}
	for _, item := range slice {
		if _, ok := keys[item]; !ok {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}
