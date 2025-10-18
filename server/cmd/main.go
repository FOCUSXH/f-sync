package main

import (
	"fsync/server/configs"
	"fsync/server/global"
	"fsync/server/internal/db"
	"fsync/server/internal/routers"
	"fsync/server/logger"
	"log"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	if err := configs.LoadConfig("server/configs"); err != nil {
		log.Panic("加载配置项错误")
		panic(err)
	}
	log.Println("加载到配置", global.Configs)

	// 初始化日志
	if err := logger.InitLogger(); err != nil {
		log.Panic("初始化日志失败")
		panic(err)
	}
	global.Logger.Info("初始化日志成功")
	defer logger.Sync()

	// 初始化数据库
	if err := db.InitDB(); err != nil {
		global.Logger.Panic("初始化数据库失败")
		panic(err)
	}
	global.Logger.Info("初始化数据库成功")

	// 初始化路由
	r := routers.InitRouter()

	// 启动服务
	if err := r.Run(global.Configs.Server.Host + global.Configs.Server.Port); err != nil {
		global.Logger.Panic("启动服务失败")
		panic(err)
	}
	global.Logger.Info("启动服务成功:", zap.String("addr", global.Configs.Server.Host+global.Configs.Server.Port))
}
