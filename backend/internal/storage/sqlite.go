package storage

import (
	"database/sql"
	"fmt"
	"time"

	"yaml-backend/pkg/models"

	_ "github.com/mattn/go-sqlite3"
)

// SummaryResult AI总结结果（为了避免循环导入）
type SummaryResult struct {
	ID        int64     `json:"id" db:"id"`
	Type      string    `json:"type" db:"type"`
	Summary   string    `json:"summary" db:"summary"`
	DataCount int       `json:"data_count" db:"data_count"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return storage, nil
}

func (s *SQLiteStorage) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS activities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			content TEXT,
			app_name TEXT,
			window_title TEXT,
			url TEXT,
			timestamp DATETIME NOT NULL,
			duration INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS keyboard_inputs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT NOT NULL,
			app_name TEXT,
			timestamp DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS app_usage (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			duration INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS ai_summaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			summary TEXT NOT NULL,
			data_count INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL
		)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	return nil
}

func (s *SQLiteStorage) SaveActivity(activity *models.Activity) error {
	query := `INSERT INTO activities (type, content, app_name, window_title, url, timestamp, duration) 
			   VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	_, err := s.db.Exec(query, activity.Type, activity.Content, activity.AppName, 
		activity.WindowTitle, activity.URL, activity.Timestamp, activity.Duration)
	return err
}

func (s *SQLiteStorage) SaveKeyboardInput(input *models.KeyboardInput) error {
	query := `INSERT INTO keyboard_inputs (text, app_name, timestamp) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, input.Text, input.AppName, input.Timestamp)
	return err
}

func (s *SQLiteStorage) GetRecentActivities(limit int) ([]*models.Activity, error) {
	query := `SELECT id, type, content, app_name, window_title, url, timestamp, duration 
			   FROM activities ORDER BY timestamp DESC LIMIT ?`
	
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		activity := &models.Activity{}
		err := rows.Scan(&activity.ID, &activity.Type, &activity.Content, 
			&activity.AppName, &activity.WindowTitle, &activity.URL, 
			&activity.Timestamp, &activity.Duration)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

func (s *SQLiteStorage) GetRecentKeyboardInputs(limit int) ([]*models.KeyboardInput, error) {
	query := `SELECT id, text, app_name, timestamp 
			   FROM keyboard_inputs ORDER BY timestamp DESC LIMIT ?`
	
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inputs []*models.KeyboardInput
	for rows.Next() {
		input := &models.KeyboardInput{}
		err := rows.Scan(&input.ID, &input.Text, &input.AppName, &input.Timestamp)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}

	return inputs, nil
}

func (s *SQLiteStorage) SaveSummary(summary *SummaryResult) error {
	query := `INSERT INTO ai_summaries (type, summary, data_count, created_at) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, summary.Type, summary.Summary, summary.DataCount, summary.CreatedAt)
	return err
}

func (s *SQLiteStorage) GetRecentSummaries(limit int) ([]*SummaryResult, error) {
	query := `SELECT id, type, summary, data_count, created_at 
			   FROM ai_summaries ORDER BY created_at DESC LIMIT ?`
	
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*SummaryResult
	for rows.Next() {
		summary := &SummaryResult{}
		err := rows.Scan(&summary.ID, &summary.Type, &summary.Summary, 
			&summary.DataCount, &summary.CreatedAt)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// GetActivityCount 获取活动记录总数
func (s *SQLiteStorage) GetActivityCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM activities").Scan(&count)
	return count, err
}

// GetKeyboardInputCount 获取键盘输入总数
func (s *SQLiteStorage) GetKeyboardInputCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM keyboard_inputs").Scan(&count)
	return count, err
}

// GetMostActiveApp 获取最活跃的应用
func (s *SQLiteStorage) GetMostActiveApp() (string, error) {
	var appName string
	err := s.db.QueryRow(`
		SELECT app_name 
		FROM activities 
		WHERE app_name IS NOT NULL AND app_name != '' 
		GROUP BY app_name 
		ORDER BY COUNT(*) DESC 
		LIMIT 1
	`).Scan(&appName)
	
	if err == sql.ErrNoRows {
		return "-", nil
	}
	return appName, err
}