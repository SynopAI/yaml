package api

import (
	"net/http"
	"strconv"

	"yaml-backend/internal/monitor"
	"yaml-backend/internal/storage"
	"yaml-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	storage *storage.SQLiteStorage
	monitor *monitor.Manager
}

func NewHandler(storage *storage.SQLiteStorage, monitor *monitor.Manager) *Handler {
	return &Handler{
		storage: storage,
		monitor: monitor,
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