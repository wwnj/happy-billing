package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/pkg/logger"
	"github.com/wwnj/happy-billing/pkg/tracing"
	"go.uber.org/zap"
)

// Logger 日志中间件 - 记录请求日志并自动关联 trace
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 从 context 提取 trace 信息
		ctx := c.Request.Context()
		traceID := tracing.GetTraceID(ctx)
		spanID := tracing.GetSpanID(ctx)

		// 构建日志字段（包含 trace_id 和 span_id）
		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// 添加 trace 信息
		if traceID != "" {
			fields = append(fields, zap.String("trace_id", traceID))
		}
		if spanID != "" {
			fields = append(fields, zap.String("span_id", spanID))
		}

		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		// 使用 otelzap.Logger.Ctx(ctx) 自动关联 trace（同时记录到OpenTelemetry）
		// 但主要的trace_id和span_id已经手动添加到日志字段中了
		// 根据状态码选择日志级别
		if statusCode >= 500 {
			logger.Logger.Ctx(ctx).Error("Server error", fields...)
		} else if statusCode >= 400 {
			logger.Logger.Ctx(ctx).Warn("Client error", fields...)
		} else {
			logger.Logger.Ctx(ctx).Info("Request completed", fields...)
		}
	}
}
