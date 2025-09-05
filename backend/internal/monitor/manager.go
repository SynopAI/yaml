package monitor

import (
	"fmt"
	"sync"

	"yaml-backend/internal/storage"
)

type Manager struct {
	storage         *storage.SQLiteStorage
	keyboardMonitor *KeyboardMonitor
	appMonitor      *AppMonitor
	mu              sync.RWMutex
	isRunning       bool
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