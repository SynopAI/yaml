package monitor

import (
	"fmt"
	"sync"
	"time"

	"yaml-backend/internal/storage"
	"yaml-backend/pkg/models"
)

type KeyboardMonitor struct {
	storage   *storage.SQLiteStorage
	isRunning bool
}

type AppMonitor struct {
	storage   *storage.SQLiteStorage
	isRunning bool
}

type Manager struct {
	storage         *storage.SQLiteStorage
	keyboardMonitor *KeyboardMonitor
	appMonitor      *AppMonitor
	mu              sync.RWMutex
	isRunning       bool
}

func NewKeyboardMonitor(storage *storage.SQLiteStorage) *KeyboardMonitor {
	return &KeyboardMonitor{
		storage: storage,
	}
}

func NewAppMonitor(storage *storage.SQLiteStorage) *AppMonitor {
	return &AppMonitor{
		storage: storage,
	}
}

func NewManager(storage *storage.SQLiteStorage) *Manager {
	return &Manager{
		storage:         storage,
		keyboardMonitor: NewKeyboardMonitor(storage),
		appMonitor:      NewAppMonitor(storage),
	}
}

func (m *Manager) StartAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return fmt.Errorf("monitors are already running")
	}

	fmt.Println("Starting all monitors...")

	// 启动键盘监控
	if err := m.keyboardMonitor.Start(); err != nil {
		return fmt.Errorf("failed to start keyboard monitor: %w", err)
	}

	// 启动应用监控
	if err := m.appMonitor.Start(); err != nil {
		m.keyboardMonitor.Stop() // 清理已启动的监控
		return fmt.Errorf("failed to start app monitor: %w", err)
	}

	m.isRunning = true
	fmt.Println("All monitors started successfully")
	return nil
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return
	}

	fmt.Println("Stopping all monitors...")

	m.keyboardMonitor.Stop()
	m.appMonitor.Stop()

	m.isRunning = false
	fmt.Println("All monitors stopped")
}

func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}

func (m *Manager) GetStatus() map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]bool{
		"keyboard": m.keyboardMonitor.IsRunning(),
		"app":      m.appMonitor.IsRunning(),
		"overall":  m.isRunning,
	}
}

// KeyboardMonitor methods
func (km *KeyboardMonitor) Start() error {
	if km.isRunning {
		return fmt.Errorf("keyboard monitor is already running")
	}

	km.isRunning = true
	fmt.Println("Keyboard monitor started (mock mode)")

	// 模拟键盘输入数据
	go km.simulateKeyboardInput()
	return nil
}

func (km *KeyboardMonitor) Stop() {
	if !km.isRunning {
		return
	}

	km.isRunning = false
	fmt.Println("Keyboard monitor stopped")
}

func (km *KeyboardMonitor) IsRunning() bool {
	return km.isRunning
}

func (km *KeyboardMonitor) simulateKeyboardInput() {
	for km.isRunning {
		time.Sleep(10 * time.Second)
		if !km.isRunning {
			break
		}

		input := &models.KeyboardInput{
			Text:      "Hello World (simulated)",
			AppName:   "Terminal",
			Timestamp: time.Now(),
		}

		if err := km.storage.SaveKeyboardInput(input); err != nil {
			fmt.Printf("Error saving keyboard input: %v\n", err)
		}
	}
}

// AppMonitor methods
func (am *AppMonitor) Start() error {
	if am.isRunning {
		return fmt.Errorf("app monitor is already running")
	}

	am.isRunning = true
	fmt.Println("App monitor started (mock mode)")

	// 模拟应用切换数据
	go am.simulateAppSwitch()
	return nil
}

func (am *AppMonitor) Stop() {
	if !am.isRunning {
		return
	}

	am.isRunning = false
	fmt.Println("App monitor stopped")
}

func (am *AppMonitor) IsRunning() bool {
	return am.isRunning
}

func (am *AppMonitor) simulateAppSwitch() {
	apps := []string{"Terminal", "Safari", "Xcode", "Finder"}
	index := 0

	for am.isRunning {
		time.Sleep(15 * time.Second)
		if !am.isRunning {
			break
		}

		activity := &models.Activity{
			Type:      models.ActivityTypeApp,
			Content:   fmt.Sprintf("Switched to %s (simulated)", apps[index]),
			AppName:   apps[index],
			Timestamp: time.Now(),
			Duration:  15,
		}

		if err := am.storage.SaveActivity(activity); err != nil {
			fmt.Printf("Error saving app activity: %v\n", err)
		}

		index = (index + 1) % len(apps)
	}
}
