package main

import (
	"flag"
	"fmt"
	"fsync/client/configs"
	"fsync/client/services"
	"fsync/client/storage"
	"log"
	"os"
)

func main() {
	var (
		loginFlag    bool
		logoutFlag   bool
		registerFlag bool
	)

	flag.BoolVar(&loginFlag, "login", false, "用户登录")
	flag.BoolVar(&logoutFlag, "logout", false, "用户登出")
	flag.BoolVar(&registerFlag, "register", false, "用户注册")

	flag.Usage = func() {
		_, _ = os.Stderr.WriteString("用法: client [选项]\n")
		_, _ = os.Stderr.WriteString("选项:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// 加载完整配置
	cfg, err := configs.LoadConfig("client/configs/configs.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 使用switch case处理命令行参数
	switch {
	case logoutFlag:
		// 处理登出
		if err := services.ClearTokens(cfg); err != nil {
			log.Fatal("登出失败:", err)
		}
		fmt.Println("登出成功")
		return

	case registerFlag:
		// 处理注册
		// 获取带确认的用户凭据
		username, password, err := services.GetCredentialsWithValidation()
		if err != nil {
			log.Fatal("获取用户凭据失败:", err)
		}

		if err := services.Register(cfg, username, password); err != nil {
			log.Fatal("注册失败:", err)
		}
		fmt.Println("注册成功，请使用 -login 参数登录")
		return

	case loginFlag:
		// 处理登录
		username, password, err := services.GetCredentials()
		if err != nil {
			log.Fatal("获取用户凭据失败:", err)
		}

		_, err = services.Login(cfg, username, password)
		if err != nil {
			log.Fatal("登录失败:", err)
		}
		fmt.Println("登录成功")

		// 登录成功后启动文件同步
		storage.StartFileSync(cfg)
		select {}

	default:
		_, err = services.LoadTokens(cfg)
		if err != nil {
			fmt.Println("未检测到登录信息，请先登录或注册")
			os.Exit(1)
		}

		// 检查服务器证书
		if err := services.CheckAndDownloadCertificate(cfg); err != nil {
			log.Fatal("检查服务器证书失败:", err)
		}

		// 启动文件同步
		storage.StartFileSync(cfg)
		// 阻止主程序退出，保持文件监控运行
		select {}
	}
}
