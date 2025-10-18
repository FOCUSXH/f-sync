package routers

import (
	"fmt"
	"fsync/server/global"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InitRouter 初始化gin路由
func InitRouter() *gin.Engine {
	// 根据环境设置gin模式
	if global.Configs.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建gin实例
	r := gin.New()

	// 使用自定义日志中间件
	r.Use(loggerMiddleware())

	// 添加健康检查路由
	r.GET("/health", healthCheck)

	return r
}

// loggerMiddleware 自定义日志中间件
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 记录请求信息
		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		// 使用自定义日志记录器记录请求信息
		global.Logger.Info("请求处理完成",
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("client_ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("comment", comment),
		)
	}
}

// healthCheck 健康检查接口
func healthCheck(c *gin.Context) {
	global.Logger.Info("健康检查接口被调用")
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "服务运行正常",
	})
}

// GetAddr 获取服务监听地址
func GetAddr() string {
	return fmt.Sprintf("%s:%d", global.Configs.Server.Host, global.Configs.Server.Port)
}
