// routers/router.go
package routers

import (
	"fsync/server/global"
	chat_handler "fsync/server/internal/modules/chat/handler"
	user_handler "fsync/server/internal/modules/user/handler"
	"fsync/server/models"
	"fsync/server/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	if global.Configs.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	r.Use(logger.GinLogger(global.Logger))
	r.Use(gin.Recovery())

	// 注册路由组
	registerChatRoutes(r)
	registerUserRoutes(r)

	r.GET("/health", healthCheck)
	return r
}

// registerChatRoutes 注册 chat 相关路由
func registerChatRoutes(r *gin.Engine) {
	chatGroup := r.Group("/chat")
	{
		chatGroup.POST("/chat", chat_handler.GetChatHistory)
	}
}

// registerUserRoutes 注册 user 相关路由
func registerUserRoutes(r *gin.Engine) {
	userGroup := r.Group("/user")
	{
		userGroup.POST("/register", user_handler.Register)
		userGroup.POST("/login", user_handler.Login)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "healthy",
		Data: nil,
	})
}
