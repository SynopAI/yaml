package api

import (
	"yaml-backend/internal/ai"
	"yaml-backend/internal/monitor"
	"yaml-backend/internal/storage"
	"yaml-backend/pkg/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(storage *storage.SQLiteStorage, monitorManager *monitor.Manager, aiService *ai.AIService, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// CORS 配置
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.API.CORSOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(corsConfig))

	// 创建处理器
	handler := NewHandler(storage, monitorManager, aiService)

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 健康检查
		api.GET("/health", handler.GetHealth)

		// 活动记录相关
		api.GET("/activities", handler.GetActivities)
		api.POST("/activities", handler.PostActivity)

		// 键盘输入相关
		api.GET("/keyboard", handler.GetKeyboardInputs)
		api.POST("/keyboard", handler.PostKeyboardInput)

		// 监控相关
		api.POST("/monitor/start", handler.StartMonitoring)
		api.POST("/monitor/stop", handler.StopMonitoring)
		api.GET("/monitor/status", handler.GetMonitorStatus)

		// AI总结相关
		api.POST("/ai/summary/activity", handler.GenerateActivitySummary)
		api.POST("/ai/summary/keyboard", handler.GenerateKeyboardSummary)
		api.GET("/ai/summaries", handler.GetAISummaries)
		// 流式AI总结
		api.GET("/ai/stream/activity", handler.StreamActivitySummary)
	}

	return r
}