package router

import (
	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/internal/api/middleware"
	v1 "github.com/wwnj/happy-billing/internal/api/v1"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(middleware.Tracing()) // 追踪中间件

	// 健康检查（不需要认证）
	healthHandler := v1.NewHealthHandler()
	r.GET("/health", healthHandler.Health)
	r.GET("/ping", healthHandler.Ping)

	// API v1 路由组
	apiV1 := r.Group("/api/v1")
	{
		// 待添加：租户、产品、订单等业务路由
		_ = apiV1
	}

	return r
}

// corsMiddleware CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, Accept, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
