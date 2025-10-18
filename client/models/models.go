package models

// Config 是整个应用的配置根结构体
type Config struct {
	Logger LoggerConfig `mapstructure:"logger"`
	Client ClientConfig `mapstructure:"client"`
}

// LoggerConfig Zap 日志配置
type LoggerConfig struct {
	Level            string   `mapstructure:"level"`
	Env              string   `mapstructure:"env"`
	Format           string   `mapstructure:"format"` // json 或 console
	OutputPaths      []string `mapstructure:"output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
	EnableCaller     bool     `mapstructure:"enable_caller"`
	EnableStacktrace bool     `mapstructure:"enable_stacktrace"`
}

type ClientConfig struct {
	SyncDir    string `mapstructure:"sync_dir"`
	ServerAddr string `mapstructure:"server_addr"`
	TokenDir   string `mapstructure:"token_dir"`
	Protocol   string `mapstructure:"protocol"`
}
