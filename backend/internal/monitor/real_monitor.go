package monitor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"yaml-backend/internal/storage"
	"yaml-backend/pkg/models"
)

// RealMonitorEvent 表示从Swift监控程序接收的事件
type RealMonitorEvent struct {
	Type      string    `json:"type"`
	Text      string    `json:"text,omitempty"`
	AppName   string    `json:"app_name"`
	BundleID  string    `json:"bundle_id,omitempty"`
	Timestamp string    `json:"timestamp"`
	KeyCode   int       `json:"key_code,omitempty"`
	Modifiers uint64    `json:"modifiers,omitempty"`
	PID       int32     `json:"pid,omitempty"`
}

// RealKeyboardMonitor 真实键盘监控器
type RealKeyboardMonitor struct {
	storage   *storage.SQLiteStorage
	isRunning bool
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// RealAppMonitor 真实应用监控器
type RealAppMonitor struct {
	storage   *storage.SQLiteStorage
	isRunning bool
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// RealMonitorManager 真实监控管理器
type RealMonitorManager struct {
	storage         *storage.SQLiteStorage
	keyboardMonitor *RealKeyboardMonitor
	appMonitor      *RealAppMonitor
	swiftProcess    *exec.Cmd
	mu              sync.RWMutex
	isRunning       bool
	cancel          context.CancelFunc
}

// NewRealKeyboardMonitor 创建真实键盘监控器
func NewRealKeyboardMonitor(storage *storage.SQLiteStorage) *RealKeyboardMonitor {
	return &RealKeyboardMonitor{
		storage: storage,
	}
}

// NewRealAppMonitor 创建真实应用监控器
func NewRealAppMonitor(storage *storage.SQLiteStorage) *RealAppMonitor {
	return &RealAppMonitor{
		storage: storage,
	}
}

// NewRealMonitorManager 创建真实监控管理器
func NewRealMonitorManager(storage *storage.SQLiteStorage) *RealMonitorManager {
	return &RealMonitorManager{
		storage:         storage,
		keyboardMonitor: NewRealKeyboardMonitor(storage),
		appMonitor:      NewRealAppMonitor(storage),
	}
}

// Start 启动真实键盘监控
func (rkm *RealKeyboardMonitor) Start() error {
	rkm.mu.Lock()
	defer rkm.mu.Unlock()

	if rkm.isRunning {
		return fmt.Errorf("real keyboard monitor is already running")
	}

	rkm.isRunning = true
	fmt.Println("Real keyboard monitor started")
	return nil
}

// Stop 停止真实键盘监控
func (rkm *RealKeyboardMonitor) Stop() {
	rkm.mu.Lock()
	defer rkm.mu.Unlock()

	if !rkm.isRunning {
		return
	}

	if rkm.cancel != nil {
		rkm.cancel()
	}

	rkm.isRunning = false
	fmt.Println("Real keyboard monitor stopped")
}

// IsRunning 检查键盘监控是否运行
func (rkm *RealKeyboardMonitor) IsRunning() bool {
	rkm.mu.RLock()
	defer rkm.mu.RUnlock()
	return rkm.isRunning
}

// Start 启动真实应用监控
func (ram *RealAppMonitor) Start() error {
	ram.mu.Lock()
	defer ram.mu.Unlock()

	if ram.isRunning {
		return fmt.Errorf("real app monitor is already running")
	}

	ram.isRunning = true
	fmt.Println("Real app monitor started")
	return nil
}

// Stop 停止真实应用监控
func (ram *RealAppMonitor) Stop() {
	ram.mu.Lock()
	defer ram.mu.Unlock()

	if !ram.isRunning {
		return
	}

	if ram.cancel != nil {
		ram.cancel()
	}

	ram.isRunning = false
	fmt.Println("Real app monitor stopped")
}

// IsRunning 检查应用监控是否运行
func (ram *RealAppMonitor) IsRunning() bool {
	ram.mu.RLock()
	defer ram.mu.RUnlock()
	return ram.isRunning
}

// StartAll 启动所有真实监控
func (rmm *RealMonitorManager) StartAll() error {
	fmt.Println("[DEBUG] StartAll called")
	rmm.mu.Lock()
	defer rmm.mu.Unlock()

	if rmm.isRunning {
		fmt.Println("[DEBUG] Real monitors are already running")
		return fmt.Errorf("real monitors are already running")
	}

	fmt.Println("[DEBUG] Starting all real monitors...")

	// 编译Swift监控程序
	fmt.Println("[DEBUG] Compiling Swift monitor...")
	if err := rmm.compileSwiftMonitor(); err != nil {
		fmt.Printf("[ERROR] Failed to compile Swift monitor: %v\n", err)
		return fmt.Errorf("failed to compile Swift monitor: %w", err)
	}
	fmt.Println("[DEBUG] Swift monitor compiled successfully")

	// 启动Swift监控进程
	fmt.Println("[DEBUG] Creating context for Swift process...")
	ctx, cancel := context.WithCancel(context.Background())
	rmm.cancel = cancel

	fmt.Println("[DEBUG] Starting Swift monitor process...")
	if err := rmm.startSwiftProcess(ctx); err != nil {
		fmt.Printf("[ERROR] Failed to start Swift monitor process: %v\n", err)
		return fmt.Errorf("failed to start Swift monitor process: %w", err)
	}
	fmt.Println("[DEBUG] Swift monitor process started")

	// 启动各个监控器
	fmt.Println("[DEBUG] Starting keyboard monitor...")
	if err := rmm.keyboardMonitor.Start(); err != nil {
		fmt.Printf("[ERROR] Failed to start keyboard monitor: %v\n", err)
		return fmt.Errorf("failed to start keyboard monitor: %w", err)
	}
	fmt.Println("[DEBUG] Keyboard monitor started")

	fmt.Println("[DEBUG] Starting app monitor...")
	if err := rmm.appMonitor.Start(); err != nil {
		fmt.Printf("[ERROR] Failed to start app monitor: %v\n", err)
		rmm.keyboardMonitor.Stop()
		return fmt.Errorf("failed to start app monitor: %w", err)
	}
	fmt.Println("[DEBUG] App monitor started")

	rmm.isRunning = true
	fmt.Println("[SUCCESS] All real monitors started successfully")
	return nil
}

// StopAll 停止所有真实监控
func (rmm *RealMonitorManager) StopAll() {
	rmm.mu.Lock()
	defer rmm.mu.Unlock()

	if !rmm.isRunning {
		return
	}

	fmt.Println("Stopping all real monitors...")

	// 停止各个监控器
	rmm.keyboardMonitor.Stop()
	rmm.appMonitor.Stop()

	// 停止Swift进程
	if rmm.cancel != nil {
		rmm.cancel()
	}

	if rmm.swiftProcess != nil && rmm.swiftProcess.Process != nil {
		rmm.swiftProcess.Process.Kill()
	}

	rmm.isRunning = false
	fmt.Println("All real monitors stopped")
}

// IsRunning 检查监控管理器是否运行
func (rmm *RealMonitorManager) IsRunning() bool {
	rmm.mu.RLock()
	defer rmm.mu.RUnlock()
	return rmm.isRunning
}

// GetStatus 获取监控状态
func (rmm *RealMonitorManager) GetStatus() map[string]bool {
	rmm.mu.RLock()
	defer rmm.mu.RUnlock()

	return map[string]bool{
		"keyboard": rmm.keyboardMonitor.IsRunning(),
		"app":      rmm.appMonitor.IsRunning(),
		"overall":  rmm.isRunning,
	}
}

// compileSwiftMonitor 编译Swift监控程序
func (rmm *RealMonitorManager) compileSwiftMonitor() error {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// 构建Swift源文件路径（相对于当前工作目录）
	swiftFile := filepath.Join(wd, "internal", "monitor", "real_monitor.swift")
	outputFile := filepath.Join(wd, "internal", "monitor", "real_monitor")

	// 检查Swift文件是否存在
	if _, err := os.Stat(swiftFile); os.IsNotExist(err) {
		return fmt.Errorf("Swift monitor file not found: %s", swiftFile)
	}

	// 检查是否已经编译过
	if _, err := os.Stat(outputFile); err == nil {
		fmt.Println("Swift monitor already compiled, skipping compilation")
		return nil
	}

	// 编译Swift程序
	cmd := exec.Command("swiftc", "-o", outputFile, swiftFile)
	cmd.Dir = wd

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Swift compilation failed: %s\nOutput: %s", err, string(output))
	}

	fmt.Println("Swift monitor compiled successfully")
	return nil
}

// startSwiftProcess 启动Swift监控进程
func (rmm *RealMonitorManager) startSwiftProcess(ctx context.Context) error {
	fmt.Println("[DEBUG] Starting Swift process setup...")
	
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("[ERROR] Failed to get working directory: %v\n", err)
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	fmt.Printf("[DEBUG] Working directory: %s\n", wd)

	monitorPath := filepath.Join(wd, "internal", "monitor", "real_monitor")
	fmt.Printf("[DEBUG] Swift monitor path: %s\n", monitorPath)

	// 检查编译后的文件是否存在
	if _, err := os.Stat(monitorPath); os.IsNotExist(err) {
		fmt.Printf("[ERROR] Swift monitor not found: %s\n", monitorPath)
		return fmt.Errorf("compiled Swift monitor not found: %s", monitorPath)
	}
	fmt.Println("[DEBUG] Swift monitor executable found")

	// 启动Swift监控进程
	fmt.Println("[DEBUG] Creating Swift command...")
	rmm.swiftProcess = exec.CommandContext(ctx, monitorPath)
	rmm.swiftProcess.Dir = wd
	
	// 设置环境变量，确保输出不被缓冲
	rmm.swiftProcess.Env = append(os.Environ(), "NSUnbufferedIO=YES")
	fmt.Println("[DEBUG] Environment variables set")

	// 获取输出管道
	fmt.Println("[DEBUG] Setting up stdout pipe...")
	stdout, err := rmm.swiftProcess.StdoutPipe()
	if err != nil {
		fmt.Printf("[ERROR] Failed to get stdout pipe: %v\n", err)
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	fmt.Println("[DEBUG] Setting up stderr pipe...")
	stderr, err := rmm.swiftProcess.StderrPipe()
	if err != nil {
		fmt.Printf("[ERROR] Failed to get stderr pipe: %v\n", err)
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// 启动进程
	fmt.Println("[DEBUG] Starting Swift process...")
	if err := rmm.swiftProcess.Start(); err != nil {
		fmt.Printf("[ERROR] Failed to start Swift process: %v\n", err)
		return fmt.Errorf("failed to start Swift process: %w", err)
	}
	
	fmt.Printf("[DEBUG] Swift process started with PID: %d\n", rmm.swiftProcess.Process.Pid)

	// 启动输出处理协程
	fmt.Println("[DEBUG] Starting output handlers...")
	go rmm.handleSwiftOutput(stdout)
	go rmm.handleSwiftErrors(stderr)

	fmt.Println("[SUCCESS] Swift monitor process started successfully")
	return nil
}

// handleSwiftOutput 处理Swift程序的输出
func (rmm *RealMonitorManager) handleSwiftOutput(stdout io.ReadCloser) {
	fmt.Println("[DEBUG] Starting Swift output handler")
	scanner := bufio.NewScanner(stdout)
	lineCount := 0
	
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		fmt.Printf("[DEBUG] Swift output line %d: %s\n", lineCount, line)
		
		// 查找事件数据
		if strings.HasPrefix(line, "YAML_EVENT: ") {
			eventJSON := strings.TrimPrefix(line, "YAML_EVENT: ")
			fmt.Printf("[DEBUG] Found YAML event: %s\n", eventJSON)
			rmm.processEvent(eventJSON)
		} else {
			// 普通日志输出
			fmt.Printf("Swift Monitor: %s\n", line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[ERROR] Error reading Swift output: %v\n", err)
	}
	
	fmt.Printf("[DEBUG] Swift output handler finished, processed %d lines\n", lineCount)
}

// handleSwiftErrors 处理Swift程序的错误输出
func (rmm *RealMonitorManager) handleSwiftErrors(stderr io.ReadCloser) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("Swift Monitor Error: %s\n", line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading Swift errors: %v\n", err)
	}
}

// processEvent 处理从Swift程序接收的事件
func (rmm *RealMonitorManager) processEvent(eventJSON string) {
	fmt.Printf("[DEBUG] Processing event JSON: %s\n", eventJSON)
	
	var event RealMonitorEvent
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		fmt.Printf("[ERROR] Error parsing event JSON: %v\n", err)
		fmt.Printf("[ERROR] Raw JSON: %s\n", eventJSON)
		return
	}

	fmt.Printf("[DEBUG] Parsed event: Type=%s, AppName=%s\n", event.Type, event.AppName)

	// 解析时间戳
	timestamp, err := time.Parse(time.RFC3339, event.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	// 根据事件类型处理
	switch event.Type {
	case "keyboard":
		rmm.handleKeyboardEvent(event, timestamp)
	case "app_activation", "app_launch", "app_termination":
		rmm.handleAppEvent(event, timestamp)
	default:
		fmt.Printf("Unknown event type: %s\n", event.Type)
	}
}

// handleKeyboardEvent 处理键盘事件
func (rmm *RealMonitorManager) handleKeyboardEvent(event RealMonitorEvent, timestamp time.Time) {
	fmt.Printf("[DEBUG] Handling keyboard event: Text=%s, AppName=%s, Timestamp=%s\n", event.Text, event.AppName, timestamp.Format(time.RFC3339))
	
	input := &models.KeyboardInput{
		Text:      event.Text,
		AppName:   event.AppName,
		Timestamp: timestamp,
	}

	fmt.Printf("[DEBUG] Saving keyboard input to database...\n")
	if err := rmm.storage.SaveKeyboardInput(input); err != nil {
		fmt.Printf("[ERROR] Error saving keyboard input: %v\n", err)
	} else {
		fmt.Printf("[SUCCESS] Keyboard input saved successfully\n")
	}
}

// handleAppEvent 处理应用事件
func (rmm *RealMonitorManager) handleAppEvent(event RealMonitorEvent, timestamp time.Time) {
	fmt.Printf("[DEBUG] Handling app event: Type=%s, AppName=%s, Timestamp=%s\n", event.Type, event.AppName, timestamp.Format(time.RFC3339))
	
	activityType := models.ActivityTypeApp
	content := fmt.Sprintf("%s: %s", event.Type, event.AppName)

	fmt.Printf("[DEBUG] Creating activity: Type=%s, Content=%s\n", activityType, content)
	
	activity := &models.Activity{
		Type:      activityType,
		Content:   content,
		AppName:   event.AppName,
		Timestamp: timestamp,
		Duration:  0, // 实时事件，持续时间为0
	}

	fmt.Printf("[DEBUG] Saving app activity to database...\n")
	if err := rmm.storage.SaveActivity(activity); err != nil {
		fmt.Printf("[ERROR] Error saving app activity: %v\n", err)
	} else {
		fmt.Printf("[SUCCESS] App activity saved successfully\n")
	}
}