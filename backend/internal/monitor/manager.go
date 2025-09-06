package monitor

import (
	"fmt"
	"sync"

	"yaml-backend/internal/storage"
)

type Manager struct {
	storage     *storage.SQLiteStorage
	realManager *RealMonitorManager
	mu          sync.RWMutex
	isRunning   bool
}

func NewManager(storage *storage.SQLiteStorage) *Manager {
	return &Manager{
		storage:     storage,
		realManager: NewRealMonitorManager(storage),
	}
}

func (m *Manager) StartAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return fmt.Errorf("monitors are already running")
	}

	fmt.Println("Starting real monitors...")
	if err := m.realManager.StartAll(); err != nil {
		return fmt.Errorf("failed to start real monitors: %w", err)
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
	m.realManager.StopAll()
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

	status := m.realManager.GetStatus()
	status["mode"] = true // 始终为真实监控模式
	return status
}
