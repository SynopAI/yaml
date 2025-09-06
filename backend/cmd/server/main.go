package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"yaml-backend/internal/ai"
	"yaml-backend/internal/api"
	"yaml-backend/internal/monitor"
	"yaml-backend/internal/storage"
	"yaml-backend/pkg/config"
)

func main() {
	// 加载配置文件
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		// 默认配置文件路径
		configPath = "config/config.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// 获取数据库路径
	dbPath, err := cfg.GetDatabasePath()
	if err != nil {
		log.Fatal("Failed to get database path:", err)
	}
	fmt.Printf("Database path: %s\n", dbPath)

	// 初始化数据库
	storage, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}
	defer storage.Close()

	// 创建监控管理器
	monitorManager := monitor.NewManager(storage)

	// 创建AI服务
	aiService := ai.NewAIService(storage, cfg.AI.Gemini.APIKey, cfg.AI.Gemini.BaseURL, cfg.GetAITimeout())

	// 设置路由
	router := api.SetupRoutes(storage, monitorManager, aiService, cfg)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}

	// 设置优雅关闭
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nShutting down gracefully...")
		monitorManager.StopAll()
		os.Exit(0)
	}()

	fmt.Printf("Starting YAML Backend Server on port %s...\n", port)
	fmt.Printf("API endpoints available at: http://localhost:%s/api/v1\n", port)
	fmt.Printf("Health check: http://localhost:%s/api/v1/health\n", port)
	fmt.Printf("Monitor control: http://localhost:%s/api/v1/monitor/\n", port)
	fmt.Printf("AI Summary: http://localhost:%s/api/v1/ai/\n", port)
	fmt.Println("\nNote: Make sure to grant Accessibility permissions in System Preferences")
	fmt.Println("System Preferences > Security & Privacy > Privacy > Accessibility")
	fmt.Println("\nAI Summary endpoints:")
	fmt.Printf("- POST /api/v1/ai/summary/activity - 生成活动总结\n")
	fmt.Printf("- POST /api/v1/ai/summary/keyboard - 生成键盘输入总结\n")
	fmt.Printf("- GET /api/v1/ai/summaries - 获取历史总结\n")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
