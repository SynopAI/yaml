package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	AI       AIConfig       `yaml:"ai"`
	Monitor  MonitorConfig  `yaml:"monitor"`
	API      APIConfig      `yaml:"api"`
	Logging  LoggingConfig  `yaml:"logging"`
	Frontend FrontendConfig `yaml:"frontend"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type          string `yaml:"type"`
	Filename      string `yaml:"filename"`
	DataDir       string `yaml:"data_dir"`
	RetentionDays int    `yaml:"retention_days"`
}

// AIConfig AI服务配置
type AIConfig struct {
	Gemini GeminiConfig `yaml:"gemini"`
}

// GeminiConfig Gemini API配置
type GeminiConfig struct {
	APIKey         string           `yaml:"api_key"`
	BaseURL        string           `yaml:"base_url"`
	TimeoutSeconds int              `yaml:"timeout_seconds"`
	Generation     GenerationConfig `yaml:"generation"`
}

// GenerationConfig 生成配置
type GenerationConfig struct {
	Temperature     float64 `yaml:"temperature"`
	MaxOutputTokens int     `yaml:"max_output_tokens"`
	TopP            float64 `yaml:"top_p"`
	TopK            int     `yaml:"top_k"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	CollectionInterval int `yaml:"collection_interval"`
	KeyboardBufferSize int `yaml:"keyboard_buffer_size"`
	AppSwitchInterval  int `yaml:"app_switch_interval"`
}

// APIConfig API配置
type APIConfig struct {
	CORSOrigins  []string `yaml:"cors_origins"`
	DefaultLimit int      `yaml:"default_limit"`
	MaxLimit     int      `yaml:"max_limit"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	FileOutput bool   `yaml:"file_output"`
	FilePath   string `yaml:"file_path"`
}

// FrontendConfig 前端配置
type FrontendConfig struct {
	WebPort    int    `yaml:"web_port"`
	APIBaseURL string `yaml:"api_base_url"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 如果没有指定配置文件路径，使用默认路径
	if configPath == "" {
		execDir, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed to get executable directory: %w", err)
		}
		configPath = filepath.Join(filepath.Dir(execDir), "config", "config.yaml")
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// 解析YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}

	if c.Database.Filename == "" {
		return fmt.Errorf("database filename cannot be empty")
	}

	if c.AI.Gemini.APIKey == "" {
		return fmt.Errorf("AI API key cannot be empty")
	}

	if c.AI.Gemini.BaseURL == "" {
		return fmt.Errorf("AI base URL cannot be empty")
	}

	return nil
}

// GetDatabasePath 获取数据库完整路径
func (c *Config) GetDatabasePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	dataDir := filepath.Join(homeDir, c.Database.DataDir)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create data directory: %w", err)
	}

	return filepath.Join(dataDir, c.Database.Filename), nil
}

// GetServerAddress 获取服务器地址
func (c *Config) GetServerAddress() string {
	return c.Server.Host + ":" + c.Server.Port
}

// GetAITimeout 获取AI服务超时时间
func (c *Config) GetAITimeout() time.Duration {
	return time.Duration(c.AI.Gemini.TimeoutSeconds) * time.Second
}

// GetMonitorInterval 获取监控间隔
func (c *Config) GetMonitorInterval() time.Duration {
	return time.Duration(c.Monitor.CollectionInterval) * time.Second
}

// GetAppSwitchInterval 获取应用切换检测间隔
func (c *Config) GetAppSwitchInterval() time.Duration {
	return time.Duration(c.Monitor.AppSwitchInterval) * time.Millisecond
}
