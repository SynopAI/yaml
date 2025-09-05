package api

import (
	"net/http"
	"strconv"

	"yaml-backend/internal/ai"
	"yaml-backend/internal/monitor"
	"yaml-backend/internal/storage"
	"yaml-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	storage   *storage.SQLiteStorage
	monitor   *monitor.Manager
	aiService *ai.AIService
}

func NewHandler(storage *storage.SQLiteStorage, monitor *monitor.Manager, aiService *ai.AIService) *Handler {
	return &Handler{
		storage:   storage,
		monitor:   monitor,
		aiService: aiService,
	}
}

// GetActivities 获取最近的活动记录
func (h *Handler) GetActivities(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	activities, err := h.storage.GetRecentActivities(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"count":      len(activities),
	})
}

// PostActivity 添加新的活动记录
func (h *Handler) PostActivity(c *gin.Context) {
	var activity models.Activity
	if err := c.ShouldBindJSON(&activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storage.SaveActivity(&activity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Activity saved successfully"})
}

// PostKeyboardInput 添加键盘输入记录
func (h *Handler) PostKeyboardInput(c *gin.Context) {
	var input models.KeyboardInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storage.SaveKeyboardInput(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Keyboard input saved successfully"})
}

// GetHealth 健康检查
func (h *Handler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "yaml-backend",
	})
}

// StartMonitoring 启动监控
func (h *Handler) StartMonitoring(c *gin.Context) {
	if err := h.monitor.StartAll(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Monitoring started successfully"})
}

// StopMonitoring 停止监控
func (h *Handler) StopMonitoring(c *gin.Context) {
	h.monitor.StopAll()
	c.JSON(http.StatusOK, gin.H{"message": "Monitoring stopped successfully"})
}

// GetMonitorStatus 获取监控状态
func (h *Handler) GetMonitorStatus(c *gin.Context) {
	status := h.monitor.GetStatus()
	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"running": h.monitor.IsRunning(),
	})
}

// GenerateActivitySummary 生成活动总结
func (h *Handler) GenerateActivitySummary(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	summary, err := h.aiService.GenerateActivitySummary(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GenerateKeyboardSummary 生成键盘输入总结
func (h *Handler) GenerateKeyboardSummary(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "15")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	summary, err := h.aiService.GenerateKeyboardSummary(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetAISummaries 获取AI总结历史
func (h *Handler) GetAISummaries(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	summaries, err := h.aiService.GetRecentSummaries(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summaries": summaries,
		"count":     len(summaries),
	})
}

// StreamActivitySummary 流式生成活动总结
func (h *Handler) StreamActivitySummary(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 获取最近的活动数据
	activities, err := h.storage.GetRecentActivities(limit)
	if err != nil {
		c.SSEvent("error", gin.H{"error": err.Error()})
		return
	}

	if len(activities) == 0 {
		c.SSEvent("data", "暂无活动数据可供分析")
		c.SSEvent("done", "")
		return
	}

	// 调用AI流式生成总结
	resultChan, errorChan := h.aiService.StreamActivitySummary(activities)

	// 处理流式响应
	for {
		select {
		case chunk, ok := <-resultChan:
			if !ok {
				c.SSEvent("done", "")
				return
			}
			c.SSEvent("data", chunk)
			c.Writer.Flush()
		case err := <-errorChan:
			if err != nil {
				c.SSEvent("error", gin.H{"error": err.Error()})
				return
			}
		}
	}
}