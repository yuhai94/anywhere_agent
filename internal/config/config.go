package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config 存储所有配置项
type Config struct {
	V2Ray  V2RayConfig  `yaml:"v2ray"`
	API    APIConfig    `yaml:"api"`
	Checks ChecksConfig `yaml:"checks"`
	Log    LogConfig    `yaml:"log"`
}

// V2RayConfig V2Ray相关配置
type V2RayConfig struct {
	Port      int    `yaml:"port"`
	UUID      string `yaml:"uuid"`
	AccessLog string `yaml:"access_log"`
}

// APIConfig API服务相关配置
type APIConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

// ChecksConfig 检查相关配置
type ChecksConfig struct {
	TrafficInterval int `yaml:"traffic_interval"`
	IdleTimeout     int `yaml:"idle_timeout"`
}

// LogConfig 日志相关配置
type LogConfig struct {
	Level      string `yaml:"level"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

// AppConfig 全局配置实例
var AppConfig Config

// LoadConfig 加载配置文件
func LoadConfig() error {
	// 检查配置文件是否存在
	if _, err := os.Stat(CLIConfig.ConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", CLIConfig.ConfigFile)
	}

	// 读取配置文件
	configData, err := os.ReadFile(CLIConfig.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析YAML配置
	if err := yaml.Unmarshal(configData, &AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证配置完整性
	if err := validateConfig(); err != nil {
		return err
	}

	// 创建日志目录（如果不存在）
	if err := os.MkdirAll(CLIConfig.LogDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", CLIConfig.LogDir, err)
	}

	return nil
}

// validateConfig 验证配置完整性
func validateConfig() error {
	// 验证V2Ray配置
	if AppConfig.V2Ray.Port == 0 {
		return fmt.Errorf("v2ray.port is required")
	}
	if AppConfig.V2Ray.UUID == "" {
		return fmt.Errorf("v2ray.uuid is required")
	}
	if AppConfig.V2Ray.AccessLog == "" {
		return fmt.Errorf("v2ray.access_log is required")
	}

	// 验证API配置
	if AppConfig.API.Address == "" {
		return fmt.Errorf("api.address is required")
	}
	if AppConfig.API.Port == 0 {
		return fmt.Errorf("api.port is required")
	}

	// 验证Checks配置
	if AppConfig.Checks.TrafficInterval == 0 {
		return fmt.Errorf("checks.traffic_interval is required")
	}
	if AppConfig.Checks.IdleTimeout == 0 {
		return fmt.Errorf("checks.idle_timeout is required")
	}

	// 验证Log配置
	if AppConfig.Log.Level == "" {
		return fmt.Errorf("log.level is required")
	}
	if AppConfig.Log.MaxSize == 0 {
		return fmt.Errorf("log.max_size is required")
	}
	if AppConfig.Log.MaxBackups == 0 {
		return fmt.Errorf("log.max_backups is required")
	}
	if AppConfig.Log.MaxAge == 0 {
		return fmt.Errorf("log.max_age is required")
	}

	return nil
}
