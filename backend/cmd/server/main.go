package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"yaml-backend/internal/ai"
	"yaml-backend/internal/api"
	"yaml-backend/internal/monitor"
	"yaml-backend/internal/storage"
)

func main() {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	// 创建应用数据目录
	dataDir := filepath.Join(homeDir, ".yaml")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// 数据库文件路径
	dbPath := filepath.Join(dataDir, "yaml.db")
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
	apiKey := "sk-JIyFjsX1HIuusXty13315a05E29440D88369B8797159E3A4"
	baseURL := "https://aihubmix.com/gemini"
	aiService := ai.NewAIService(storage, apiKey, baseURL)

	// 设置路由
	router := api.SetupRoutes(storage, monitorManager, aiService)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
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