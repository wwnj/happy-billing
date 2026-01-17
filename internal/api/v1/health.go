package v1

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wwnj/happy-billing/internal/api/response"
	"github.com/wwnj/happy-billing/pkg/database"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Services  map[string]ServiceInfo `json:"services"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Status  string `json:"status"` // ok, error
	Message string `json:"message,omitempty"`
}

// Health 健康检查
// @Summary 健康检查
// @Description 检查服务及依赖服务的健康状态
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	services := make(map[string]ServiceInfo)

	// 检查 MySQL
	mysqlStatus := ServiceInfo{Status: "ok"}
	if db := database.GetMySQL(); db != nil {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			mysqlStatus.Status = "error"
			mysqlStatus.Message = "connection failed"
		}
	} else {
		mysqlStatus.Status = "error"
		mysqlStatus.Message = "not initialized"
	}
	services["mysql"] = mysqlStatus

	// 检查 Redis
	redisStatus := ServiceInfo{Status: "ok"}
	if redis := database.GetRedis(); redis != nil {
		if err := redis.Ping(c).Err(); err != nil {
			redisStatus.Status = "error"
			redisStatus.Message = "connection failed"
		}
	} else {
		redisStatus.Status = "error"
		redisStatus.Message = "not initialized"
	}
	services["redis"] = redisStatus

	// 检查 ClickHouse (可选)
	if ch := database.GetClickHouse(); ch != nil {
		chStatus := ServiceInfo{Status: "ok"}
		if err := ch.Ping(); err != nil {
			chStatus.Status = "error"
			chStatus.Message = "connection failed"
		}
		services["clickhouse"] = chStatus
	}

	// 判断整体状态
	status := "ok"
	for _, svc := range services {
		if svc.Status == "error" {
			status = "degraded"
			break
		}
	}

	resp := HealthResponse{
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
		Services:  services,
	}

	response.Success(c, resp)
}

// Ping 简单的 ping 检查
// @Summary Ping
// @Description 简单的存活检查
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /ping [get]
func (h *HealthHandler) Ping(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "pong",
		"time":    time.Now().Format(time.RFC3339),
	})
}
