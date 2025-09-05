package monitor

import (
	"fmt"
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