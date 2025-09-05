package storage

import (
	"database/sql"
	"fmt"

	"yaml-backend/pkg/models"

	_ "github.com/mattn/go-sqlite3"
)

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

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}