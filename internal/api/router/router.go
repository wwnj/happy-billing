package router

import (
	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/internal/api/middleware"
	v1 "github.com/wwnj/happy-billing/internal/api/v1"
)

// Handlers 包含所有处理器
type Handlers struct {
	Health *v1.HealthHandler
	Tenant *v1.TenantHandler
}

// SetupRouter 设置路由
func SetupRouter(handlers *Handlers) *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(middleware.Tracing()) // 追踪中间件

	// 健康检查（不需要认证）
	r.GET("/health", handlers.Health.Health)
	r.GET("/ping", handlers.Health.Ping)

	// API v1 路由组
	apiV1 := r.Group("/api/v1")
	{
		// 租户注册（公开接口）
		apiV1.POST("/tenants/register/individual", handlers.Tenant.RegisterIndividual)
		apiV1.POST("/tenants/register/enterprise", handlers.Tenant.RegisterEnterprise)

		// 租户管理
		tenants := apiV1.Group("/tenants")
		{
			tenants.GET("", handlers.Tenant.ListTenants)
			tenants.GET("/:tenant_id", handlers.Tenant.GetTenant)
			tenants.GET("/:tenant_id/organizations", handlers.Tenant.GetOrganizationsByTenant)
			tenants.GET("/:tenant_id/projects", handlers.Tenant.GetProjectsByTenant)
			tenants.GET("/:tenant_id/users", handlers.Tenant.GetUsersByTenant)
		}

		// 组织管理
		apiV1.POST("/organizations", handlers.Tenant.CreateOrganization)

		// 项目管理
		apiV1.POST("/projects", handlers.Tenant.CreateProject)

		// 用户管理
		apiV1.POST("/users", handlers.Tenant.CreateUser)
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
