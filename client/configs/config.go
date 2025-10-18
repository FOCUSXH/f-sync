// client/configs/config.go
package configs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoggerConfig 对应 YAML 中的日志配置
type LoggerConfig struct {
	Level            string   `yaml:"level"`            // debug, info, warn, error
	Encoding         string   `yaml:"encoding"`         // console, json
	OutputPaths      []string `yaml:"outputPaths"`      // stdout, 文件路径等
	ErrorOutputPaths []string `yaml:"errorOutputPaths"` // stderr, 错误日志路径
	MaxSize          int      `yaml:"maxSize"`          // MB
	MaxBackups       int      `yaml:"maxBackups"`
	MaxAge           int      `yaml:"maxAge"`
	Compress         bool     `yaml:"compress"`
}

// ClientConfig 客户端配置
type ClientConfig struct {
	SyncDir    string `yaml:"syncDir"`    // 同步目录
	ServerAddr string `yaml:"serverAddr"` // 服务器地址
	TokenDir   string `yaml:"tokenDir"`   // Token存储目录
	Protocol   string `yaml:"protocol"`   // 协议 (http 或 https)
}

// Config 包含所有配置项
type Config struct {
	Logger *LoggerConfig `yaml:"logger"`
	Client *ClientConfig `yaml:"client"`
}

// LoadLoggerConfig 从指定路径加载 YAML 配置
func LoadLoggerConfig(filepath string) (*LoggerConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var config LoggerConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfig 从指定路径加载完整配置
func LoadConfig(configFilepath string) (*Config, error) {
	data, err := os.ReadFile(configFilepath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// 展开 TokenDir 中的 ~ 符号
	if config.Client != nil && strings.HasPrefix(config.Client.TokenDir, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			config.Client.TokenDir = filepath.Join(homeDir, config.Client.TokenDir[1:])
		}
	}

	return &config, nil
}

// GetTokenFilePath 获取Token文件路径
func (c *Config) GetTokenFilePath() string {
	if c.Client != nil && c.Client.TokenDir != "" {
		return filepath.Join(c.Client.TokenDir, ".fsync_token")
	}
	// 默认存储在用户主目录下
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// 如果无法获取用户主目录，则存储在当前目录
		return ".fsync_token"
	}
	return filepath.Join(homeDir, ".fsync_token")
}

// GetServerURL 获取服务端完整URL
func (c *Config) GetServerURL() string {
	// 如果地址已经包含协议，则直接返回
	if strings.HasPrefix(c.Client.ServerAddr, "http://") || strings.HasPrefix(c.Client.ServerAddr, "https://") {
		return c.Client.ServerAddr
	}

	// 根据配置的协议来构建URL
	protocol := "http"
	if c.Client.Protocol == "https" {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s", protocol, c.Client.ServerAddr)
}