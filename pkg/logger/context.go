package logger

import (
	"context"

	"github.com/wwnj/happy-billing/pkg/tracing"
	"go.uber.org/zap"
)

// WithContext 创建带上下文的日志记录器
// 自动注入 TraceID、SpanID 和业务上下文
func WithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Logger
	}

	fields := make([]zap.Field, 0, 10)

	// 添加 Trace ID 和 Span ID
	if traceID := tracing.GetTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if spanID := tracing.GetSpanID(ctx); spanID != "" {
		fields = append(fields, zap.String("span_id", spanID))
	}

	// 添加业务上下文
	if tenantID := tracing.GetTenantID(ctx); tenantID != "" {
		fields = append(fields, zap.String("tenant_id", tenantID))
	}
	if userID := tracing.GetUserID(ctx); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}
	if orgID := tracing.GetOrgID(ctx); orgID != "" {
		fields = append(fields, zap.String("org_id", orgID))
	}
	if projectID := tracing.GetProjectID(ctx); projectID != "" {
		fields = append(fields, zap.String("project_id", projectID))
	}
	if orderID := tracing.GetOrderID(ctx); orderID != "" {
		fields = append(fields, zap.String("order_id", orderID))
	}
	if billID := tracing.GetBillID(ctx); billID != "" {
		fields = append(fields, zap.String("bill_id", billID))
	}
	if requestID := tracing.GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	return Logger.With(fields...)
}

// DebugCtx 带上下文的 Debug 日志
func DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Debug(msg, fields...)
}

// InfoCtx 带上下文的 Info 日志
func InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Info(msg, fields...)
}

// WarnCtx 带上下文的 Warn 日志
func WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Warn(msg, fields...)
}

// ErrorCtx 带上下文的 Error 日志
func ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Error(msg, fields...)
}

// FatalCtx 带上下文的 Fatal 日志
func FatalCtx(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Fatal(msg, fields...)
}
