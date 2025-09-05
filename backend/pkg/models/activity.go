package models

import (
	"time"
)

// ActivityType 定义活动类型
type ActivityType string

const (
	ActivityTypeKeyboard ActivityType = "keyboard"
	ActivityTypeApp      ActivityType = "app"
	ActivityTypeWeb      ActivityType = "web"
	ActivityTypeClick    ActivityType = "click"
)

// Activity 用户活动记录
type Activity struct {
	ID          int64        `json:"id" db:"id"`
	Type        ActivityType `json:"type" db:"type"`
	Content     string       `json:"content" db:"content"`
	AppName     string       `json:"app_name" db:"app_name"`
	WindowTitle string       `json:"window_title" db:"window_title"`
	URL         string       `json:"url" db:"url"`
	Timestamp   time.Time    `json:"timestamp" db:"timestamp"`
	Duration    int64        `json:"duration" db:"duration"` // 持续时间（秒）
}

// KeyboardInput 键盘输入记录
type KeyboardInput struct {
	ID        int64     `json:"id" db:"id"`
	Text      string    `json:"text" db:"text"`
	AppName   string    `json:"app_name" db:"app_name"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// AppUsage 应用使用记录
type AppUsage struct {
	ID        int64     `json:"id" db:"id"`
	AppName   string    `json:"app_name" db:"app_name"`
	StartTime time.Time `json:"start_time" db:"start_time"`
	EndTime   time.Time `json:"end_time" db:"end_time"`
	Duration  int64     `json:"duration" db:"duration"`
}