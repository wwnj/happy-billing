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
	"github.com/wwnj/happy-billing/internal/api/router"
	v1 "github.com/wwnj/happy-billing/internal/api/v1"
	"github.com/wwnj/happy-billing/internal/repository"
	"github.com/wwnj/happy-billing/internal/service"
	"github.com/wwnj/happy-billing/pkg/config"
	"github.com/wwnj/happy-billing/pkg/database"
	"github.com/wwnj/happy-billing/pkg/logger"
	"github.com/wwnj/happy-billing/pkg/tracing"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(&logger.Config{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		Output:     cfg.Log.Output,
		FilePath:   cfg.Log.FilePath,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	}); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Happy Billing API Server Starting...")

	// 初始化链路追踪
	if err := tracing.Init(&tracing.Config{
		Enabled:     cfg.Tracing.Enabled,
		ServiceName: cfg.Tracing.ServiceName,
		Endpoint:    cfg.Tracing.Endpoint,
		SampleRate:  cfg.Tracing.SampleRate,
	}); err != nil {
		logger.Fatalf("Failed to init tracing: %v", err)
	}
	if cfg.Tracing.Enabled {
		logger.Infof("Tracing enabled, endpoint: %s, sample rate: %.2f", cfg.Tracing.Endpoint, cfg.Tracing.SampleRate)
	}

	// 初始化数据库
	if err := database.InitAll(cfg); err != nil {
		logger.Fatalf("Failed to init databases: %v", err)
	}
	defer database.CloseAll()

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化仓储层
	tenantRepo := repository.NewTenantRepository(database.GetMySQL())
	orgRepo := repository.NewOrganizationRepository(database.GetMySQL())
	projectRepo := repository.NewProjectRepository(database.GetMySQL())
	userRepo := repository.NewUserRepository(database.GetMySQL())

	// 初始化服务层
	tenantService := service.NewTenantService(
		database.GetMySQL(),
		database.GetRedis(),
		tenantRepo,
		orgRepo,
		projectRepo,
		userRepo,
	)

	// 初始化处理器
	healthHandler := v1.NewHealthHandler()
	tenantHandler := v1.NewTenantHandler(tenantService)

	// 设置路由
	r := router.SetupRouter(&router.Handlers{
		Health: healthHandler,
		Tenant: tenantHandler,
	})

	// 创建 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 启动服务器
	go func() {
		logger.Infof("HTTP Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 关闭追踪系统
	if err := tracing.Shutdown(ctx); err != nil {
		logger.Errorf("Failed to shutdown tracing: %v", err)
	}

	// 优雅关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
