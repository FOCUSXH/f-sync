package storage

import (
	"fmt"
	"fsync/client/global"
	"fsync/client/internal/watcher"
	"os"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// StartFileSync 启动文件同步功能
func StartFileSync() error {
	var dir string
	// 如果配置文件中有同步目录，则使用配置文件中的目录
	if global.Configs.Client.SyncDir != "" {
		dir = global.Configs.Client.SyncDir
		info, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("目标目录不存在")
			}
			return err
		}
		global.Logger.Info("同步目录", zap.Any("目录信息", info))
	} else {
		return fmt.Errorf("错误: 必须在配置文件中指定同步目录")
	}

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
			global.Logger.Error("监控目录失败:", zap.Error(err))
		}
	}()
	return nil
}
