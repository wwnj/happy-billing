package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	balanceCmd "github.com/wwnj/happy-billing/internal/application/command/balance"
	balanceQuery "github.com/wwnj/happy-billing/internal/application/query/balance"
	"github.com/wwnj/happy-billing/internal/infrastructure/config"
	"github.com/wwnj/happy-billing/internal/infrastructure/lock"
	"github.com/wwnj/happy-billing/internal/infrastructure/logger"
	"github.com/wwnj/happy-billing/internal/infrastructure/persistence/postgres"
	httpHandler "github.com/wwnj/happy-billing/internal/interfaces/http"
	"github.com/wwnj/happy-billing/pkg/idempotent"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 2. 初始化日志
	loggerConfig := &logger.Config{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		Output:     cfg.Logger.Output,
		FilePath:   cfg.Logger.FilePath,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
	}
	if err := logger.Init(loggerConfig); err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}
	log := logger.GetLogger()
	defer log.Sync()

	log.Info("Starting Happy Billing API Server",
		zap.String("app", cfg.App.Name),
		zap.String("version", cfg.App.Version),
		zap.String("environment", cfg.App.Environment),
	)

	// 3. 初始化数据库
	db, err := postgres.InitDB(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to init database", zap.Error(err))
	}
	defer postgres.CloseDB()

	log.Info("Database connected",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
	)

	// 4. 初始化Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})
	defer redisClient.Close()

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect Redis", zap.Error(err))
	}

	log.Info("Redis connected",
		zap.String("host", cfg.Redis.Host),
		zap.Int("port", cfg.Redis.Port),
	)

	// 5. 创建基础设施组件
	lockClient := lock.NewRedisLock(redisClient)
	idempotencyChecker := idempotent.NewRedisIdempotencyChecker(redisClient)

	// 6. 创建仓储
	accountRepo := postgres.NewAccountRepository(db, lockClient)
	txRepo := postgres.NewTransactionRepository(db)

	// 7. 创建应用服务
	chargeService := balanceCmd.NewChargeService(accountRepo, txRepo, lockClient, idempotencyChecker)
	deductService := balanceCmd.NewDeductService(accountRepo, txRepo, lockClient, idempotencyChecker)
	balanceQueryService := balanceQuery.NewBalanceQuery(accountRepo, txRepo)

	// 8. 创建HTTP处理器
	balanceHandler := httpHandler.NewBalanceHandler(chargeService, deductService, balanceQueryService)

	// 9. 创建Gin路由
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware())
	router.Use(CORSMiddleware())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 注册路由
	v1 := router.Group("/api/v1")
	balanceHandler.RegisterRoutes(v1)

	// 10. 启动HTTP服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// 优雅关闭
	go func() {
		log.Info("HTTP server starting", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// 优雅关闭，最多等待30秒
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	log := logger.GetLogger()

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
