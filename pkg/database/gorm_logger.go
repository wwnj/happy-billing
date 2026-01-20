package database

import (
	"context"
	"fmt"
	"time"

	"github.com/wwnj/happy-billing/pkg/logger"
	"github.com/wwnj/happy-billing/pkg/tracing"
	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger 自定义GORM logger，输出JSON格式日志并包含trace信息
type GormLogger struct {
	LogLevel gormlogger.LogLevel
}

// NewGormLogger 创建GORM logger
func NewGormLogger(level string) gormlogger.Interface {
	return &GormLogger{
		LogLevel: parseGormLogLevel(level),
	}
}

// LogMode 设置日志级别
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 日志
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		fields := l.buildFields(ctx)
		fields = append(fields, zap.String("gorm_msg", fmt.Sprintf(msg, data...)))
		logger.Logger.Ctx(ctx).Info("GORM Info", fields...)
	}
}

// Warn 日志
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		fields := l.buildFields(ctx)
		fields = append(fields, zap.String("gorm_msg", fmt.Sprintf(msg, data...)))
		logger.Logger.Ctx(ctx).Warn("GORM Warn", fields...)
	}
}

// Error 日志
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		fields := l.buildFields(ctx)
		fields = append(fields, zap.String("gorm_msg", fmt.Sprintf(msg, data...)))
		logger.Logger.Ctx(ctx).Error("GORM Error", fields...)
	}
}

// Trace 日志（SQL查询）
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := l.buildFields(ctx)
	fields = append(fields,
		zap.Duration("elapsed", elapsed),
		zap.String("sql", sql),
		zap.Int64("rows", rows),
	)

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error:
		fields = append(fields, zap.Error(err))
		logger.Logger.Ctx(ctx).Error("GORM SQL Error", fields...)
	case elapsed > 200*time.Millisecond && l.LogLevel >= gormlogger.Warn:
		fields = append(fields, zap.String("warn", "slow query"))
		logger.Logger.Ctx(ctx).Warn("GORM Slow SQL", fields...)
	case l.LogLevel >= gormlogger.Info:
		logger.Logger.Ctx(ctx).Info("GORM SQL", fields...)
	}
}

// buildFields 构建包含trace信息的字段
func (l *GormLogger) buildFields(ctx context.Context) []zap.Field {
	fields := []zap.Field{
		zap.String("component", "gorm"),
	}

	// 添加 trace 信息
	if traceID := tracing.GetTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if spanID := tracing.GetSpanID(ctx); spanID != "" {
		fields = append(fields, zap.String("span_id", spanID))
	}

	return fields
}

// parseGormLogLevel 解析GORM日志级别
func parseGormLogLevel(level string) gormlogger.LogLevel {
	switch level {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "warn":
		return gormlogger.Warn
	case "info":
		return gormlogger.Info
	default:
		return gormlogger.Warn
	}
}
