package utils

import (
	"math"
	"time"
)

// TimeUtils 时间工具
type TimeUtils struct{}

var Time = &TimeUtils{}

// StartOfHour 获取小时开始时间
func (t *TimeUtils) StartOfHour(tm time.Time) time.Time {
	return time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), 0, 0, 0, tm.Location())
}

// EndOfHour 获取小时结束时间
func (t *TimeUtils) EndOfHour(tm time.Time) time.Time {
	return time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), 59, 59, 999999999, tm.Location())
}

// StartOfDay 获取当天开始时间
func (t *TimeUtils) StartOfDay(tm time.Time) time.Time {
	return time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, tm.Location())
}

// EndOfDay 获取当天结束时间
func (t *TimeUtils) EndOfDay(tm time.Time) time.Time {
	return time.Date(tm.Year(), tm.Month(), tm.Day(), 23, 59, 59, 999999999, tm.Location())
}

// StartOfMonth 获取月初时间
func (t *TimeUtils) StartOfMonth(tm time.Time) time.Time {
	return time.Date(tm.Year(), tm.Month(), 1, 0, 0, 0, 0, tm.Location())
}

// EndOfMonth 获取月末时间
func (t *TimeUtils) EndOfMonth(tm time.Time) time.Time {
	return t.StartOfMonth(tm).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// DurationSeconds 计算两个时间之间的秒数
func (t *TimeUtils) DurationSeconds(start, end time.Time) int64 {
	return int64(end.Sub(start).Seconds())
}

// DurationHours 计算两个时间之间的小时数（向上取整）
func (t *TimeUtils) DurationHours(start, end time.Time) int64 {
	seconds := t.DurationSeconds(start, end)
	return int64(math.Ceil(float64(seconds) / 3600.0))
}

// FormatDateTime 格式化日期时间
func (t *TimeUtils) FormatDateTime(tm time.Time) string {
	return tm.Format("2006-01-02 15:04:05")
}

// FormatDate 格式化日期
func (t *TimeUtils) FormatDate(tm time.Time) string {
	return tm.Format("2006-01-02")
}

// ParseDateTime 解析日期时间字符串
func (t *TimeUtils) ParseDateTime(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
}

// ParseDate 解析日期字符串
func (t *TimeUtils) ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, time.Local)
}
