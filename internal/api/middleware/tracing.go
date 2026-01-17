package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wwnj/happy-billing/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// Tracing OpenTelemetry 追踪中间件
func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 HTTP Header 中提取 Trace Context
		ctx := c.Request.Context()

		// 创建 Span
		spanName := c.Request.Method + " " + c.FullPath()
		if c.FullPath() == "" {
			spanName = c.Request.Method + " " + c.Request.URL.Path
		}

		ctx, span := tracing.StartSpan(ctx, spanName)
		defer span.End()

		// 设置 HTTP 相关属性
		span.SetAttributes(
			semconv.HTTPMethod(c.Request.Method),
			semconv.HTTPRoute(c.FullPath()),
			semconv.HTTPTarget(c.Request.URL.String()),
			semconv.HTTPScheme(c.Request.URL.Scheme),
			semconv.HTTPUserAgent(c.Request.UserAgent()),
			semconv.HTTPClientIP(c.ClientIP()),
			attribute.String("http.host", c.Request.Host),
		)

		// 生成或提取 Request ID
		requestID := c.GetHeader(tracing.HeaderRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx = tracing.SetRequestID(ctx, requestID)
		c.Header(tracing.HeaderRequestID, requestID)

		// 从 HTTP Header 提取业务上下文并染色
		businessCtx := &tracing.BusinessContext{
			TenantID:  c.GetHeader(tracing.HeaderTenantID),
			UserID:    c.GetHeader(tracing.HeaderUserID),
			OrgID:     c.GetHeader(tracing.HeaderOrgID),
			ProjectID: c.GetHeader(tracing.HeaderProjectID),
			RequestID: requestID,
		}
		ctx = tracing.SetBusinessContext(ctx, businessCtx)

		// 将 Trace ID 和 Span ID 添加到响应头（方便调试）
		c.Header("X-Trace-ID", tracing.GetTraceID(ctx))
		c.Header("X-Span-ID", tracing.GetSpanID(ctx))

		// 更新请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 继续处理请求
		c.Next()

		// 请求处理完成后，记录状态码
		statusCode := c.Writer.Status()
		span.SetAttributes(semconv.HTTPStatusCode(statusCode))

		// 如果有错误，标记 Span 为错误状态
		if statusCode >= 400 {
			span.SetStatus(codes.Error, c.Errors.String())
			if len(c.Errors) > 0 {
				span.RecordError(c.Errors.Last())
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}
}
