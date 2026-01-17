package router

import (
	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/internal/api/middleware"
	v1 "github.com/wwnj/happy-billing/internal/api/v1"
)

// Handlers 包含所有处理器
type Handlers struct {
	Health   *v1.HealthHandler
	Tenant   *v1.TenantHandler
	Product  *v1.ProductHandler
	Pricing  *v1.PricingHandler
	Order    *v1.OrderHandler
	Bill     *v1.BillHandler
	Payment  *v1.PaymentHandler
	Currency *v1.CurrencyHandler
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

		// 产品分类管理
		categories := apiV1.Group("/products/categories")
		{
			categories.POST("", handlers.Product.CreateCategory)
			categories.GET("", handlers.Product.ListCategories)
			categories.GET("/tree", handlers.Product.GetCategoryTree)
			categories.GET("/:category_id", handlers.Product.GetCategory)
		}

		// SPU管理
		spu := apiV1.Group("/products/spu")
		{
			spu.POST("", handlers.Product.CreateSpu)
			spu.GET("", handlers.Product.ListSpu)
			spu.GET("/:spu_id", handlers.Product.GetSpu)
			spu.GET("/:spu_id/skus", handlers.Product.GetSkusBySpu)
		}

		// SKU管理
		sku := apiV1.Group("/products/sku")
		{
			sku.POST("", handlers.Product.CreateSku)
			sku.GET("", handlers.Product.ListSku)
			sku.GET("/:sku_id", handlers.Product.GetSku)
		}

		// 定价规则管理
		priceRules := apiV1.Group("/price-rules")
		{
			priceRules.POST("", handlers.Pricing.CreatePriceRule)
			priceRules.GET("", handlers.Pricing.ListPriceRules)
			priceRules.GET("/:rule_id", handlers.Pricing.GetPriceRule)
		}

		// 折扣规则管理
		discountRules := apiV1.Group("/discount-rules")
		{
			discountRules.POST("", handlers.Pricing.CreateDiscountRule)
			discountRules.GET("", handlers.Pricing.ListDiscountRules)
			discountRules.GET("/:discount_id", handlers.Pricing.GetDiscountRule)
		}

		// 价格查询和计算
		apiV1.POST("/pricing/query", handlers.Pricing.QueryPrice)
		apiV1.POST("/pricing/calculate", handlers.Pricing.CalculatePrice)

		// 订单管理
		orders := apiV1.Group("/orders")
		{
			orders.POST("", handlers.Order.CreateOrder)
			orders.GET("", handlers.Order.ListOrders)
			orders.GET("/:order_id", handlers.Order.GetOrder)
			orders.POST("/:order_id/cancel", handlers.Order.CancelOrder)
			orders.GET("/:order_id/items", handlers.Order.GetOrderItems)
			orders.GET("/:order_id/bills", handlers.Bill.GetOrderBills)
		}

		// 账单管理
		bills := apiV1.Group("/bills")
		{
			bills.GET("", handlers.Bill.ListBills)
			bills.GET("/:bill_id", handlers.Bill.GetBill)
		}

		// 支付管理
		payments := apiV1.Group("/payments")
		{
			payments.POST("", handlers.Payment.CreatePayment)
			payments.GET("/:payment_id", handlers.Payment.GetPayment)
		}

		// 账户余额管理
		apiV1.GET("/tenants/:tenant_id/balance", handlers.Payment.GetBalance)
		apiV1.POST("/tenants/:tenant_id/balance/recharge", handlers.Payment.Recharge)

		// 货币管理
		exchangeRates := apiV1.Group("/exchange-rates")
		{
			exchangeRates.POST("", handlers.Currency.CreateExchangeRate)
			exchangeRates.GET("", handlers.Currency.ListExchangeRates)
			exchangeRates.GET("/query", handlers.Currency.GetExchangeRate)
		}

		// 货币转换
		apiV1.POST("/currency/convert", handlers.Currency.ConvertCurrency)
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
