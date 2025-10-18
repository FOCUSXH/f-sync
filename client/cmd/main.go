package main

import (
	"fsync/client/configs"
	"fsync/client/global"
	"fsync/client/logger"
	"log"
)

func main() {
	// 加载配置
	log.Println("加载配置项")
	if err := configs.LoadConfig("client/configs"); err != nil {
		panic(err)
	}
	log.Println("加载到配置", global.Configs)

	// 初始化日志
	if err := logger.InitLogger(); err != nil {
		panic(err)
	}
	defer logger.Sync()
	global.Logger.Info("初始化日志成功")

	// 启动服务
}
