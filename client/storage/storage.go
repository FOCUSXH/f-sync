package storage

import (
	"fsync/client/configs"
	"fsync/client/global"
	"fsync/client/watcher"
	"fsync/pkg/utils"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// StartFileSync 启动文件同步功能
func StartFileSync(cfg *configs.Config) {
	var dir string
	// 如果配置文件中有同步目录，则使用配置文件中的目录
	if cfg.Client != nil && cfg.Client.SyncDir != "" {
		dir = cfg.Client.SyncDir
	} else {
		log.Fatal("错误: 必须在配置文件中指定同步目录")
	}

	createLogDirs(cfg.Logger)
	// 初始化全局 logger
	if err := utils.InitLogger(cfg.Logger); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}

	global.Logger = utils.GetLogger()

	// 在一个goroutine中运行文件监控，避免阻塞主程序
	go func() {
		defer global.Logger.Sync() // 刷新缓冲区

		err := watcher.WatchDirRecursive(dir, func(event fsnotify.Event) {
			// 根据不同的操作类型记录不同的日志
			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				global.Logger.Info("文件创建事件", zap.String("file", event.Name))
			case event.Op&fsnotify.Write == fsnotify.Write:
				global.Logger.Info("文件修改事件", zap.String("file", event.Name))
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				global.Logger.Info("文件删除事件", zap.String("file", event.Name))
			case event.Op&fsnotify.Rename == fsnotify.Rename:
				global.Logger.Info("文件重命名事件", zap.String("file", event.Name))
			case event.Op&fsnotify.Chmod == fsnotify.Chmod:
				global.Logger.Info("文件权限修改事件", zap.String("file", event.Name))
			}
		})
		if err != nil {
			global.Logger.Fatal("监控目录失败:", zap.Error(err))
		}
	}()
}

func createLogDirs(cfg *configs.LoggerConfig) {
	paths := append(cfg.OutputPaths, cfg.ErrorOutputPaths...)
	for _, p := range paths {
		if p != "stdout" && p != "stderr" {
			dir := filepath.Dir(p)
			_ = os.MkdirAll(dir, 0755)
		}
	}
}
