package logger

import (
	"fmt"
	"fsync/client/global"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger 根据配置初始化zap日志记录器
func InitLogger() error {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}

	// 处理普通日志输出路径
	outputPaths := make([]string, len(global.Configs.Logger.OutputPaths))
	for i, path := range global.Configs.Logger.OutputPaths {
		if path == "stdout" || path == "stderr" {
			outputPaths[i] = path
		} else {
			// 如果是相对路径，则相对于backend目录
			if !filepath.IsAbs(path) {
				path = filepath.Join(wd, path)
			}
			outputPaths[i] = path

			// 确保日志目录存在
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建日志目录失败 %s: %w", dir, err)
			}
		}
	}

	// 处理错误日志输出路径
	errorOutputPaths := make([]string, len(global.Configs.Logger.ErrorOutputPaths))
	for i, path := range global.Configs.Logger.ErrorOutputPaths {
		if path == "stderr" || path == "stdout" {
			errorOutputPaths[i] = path
		} else {
			// 如果是相对路径，则相对于backend目录
			if !filepath.IsAbs(path) {
				path = filepath.Join(wd, path)
			}
			errorOutputPaths[i] = path

			// 确保日志目录存在
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建错误日志目录失败 %s: %w", dir, err)
			}
		}
	}

	// 解析日志级别
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(global.Configs.Logger.Level)); err != nil {
		return fmt.Errorf("无效的日志级别: %s", global.Configs.Logger.Level)
	}

	// 创建zap配置
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      global.Configs.Logger.Env == "development",
		Encoding:         global.Configs.Logger.Format,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorOutputPaths,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	// 根据配置启用调用者和堆栈跟踪
	if global.Configs.Logger.EnableCaller {
		config.EncoderConfig.CallerKey = "caller"
	}

	if global.Configs.Logger.EnableStacktrace {
		config.EncoderConfig.StacktraceKey = "stacktrace"
	}

	// 如果是控制台格式且在开发环境中，启用彩色编码器
	// if global.Configs.Logger.Format == "console" && global.Configs.App.Env == "development" {
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// }

	// 构建日志记录器
	global.Logger, err = config.Build()
	if err != nil {
		return fmt.Errorf("构建日志记录器失败: %w", err)
	}

	// 将全局日志记录器替换为当前日志记录器
	zap.ReplaceGlobals(global.Logger)
	return nil
}

// Sync 同步日志缓冲区，在程序退出前调用
func Sync() {
	if global.Logger != nil {
		_ = global.Logger.Sync()
	}
}

func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		ip := c.ClientIP()

		// 处理请求
		c.Next()

		// 记录日志
		end := time.Now()
		latency := end.Sub(start)
		statusCode := c.Writer.Status()
		userAgent := c.Request.UserAgent()

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", ip),
			zap.String("user_agent", userAgent),
			zap.Duration("latency", latency),
		}

		// 根据状态码决定日志级别
		if statusCode >= 500 {
			logger.Error("HTTP Request", fields...)
		} else if statusCode >= 400 {
			logger.Warn("HTTP Request", fields...)
		} else {
			logger.Info("HTTP Request", fields...)
		}
	}
}
