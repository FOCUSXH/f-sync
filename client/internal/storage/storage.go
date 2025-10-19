// client/internal/storage/storage.go
package storage

import (
	"fmt"
	"fsync/client/global"
	"fsync/client/internal/command"
	"fsync/client/internal/watcher"
	"os"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)


var commandManager *command.CommandManager

// StartFileSync 启动文件同步功能
func StartFileSync() error {
	var dir string
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


	// 创建命令管理器，包含异步队列, bufferSize: 队列大小，numWorkers: 工作协程数量
	commandManager = command.NewCommandManager(global.Logger, 100, 2)

	go func() {
		defer global.Logger.Sync() // 刷新缓冲区
		defer commandManager.Stop() // 确保在监控退出时停止队列

		err := watcher.WatchDirRecursive(dir, func(event fsnotify.Event) {
			var cmd command.Command
			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				// 检查是文件还是目录
				fileInfo, err := os.Stat(event.Name)
				if err != nil {
					global.Logger.Error("无法获取文件信息", zap.String("file", event.Name), zap.Error(err))
					return
				}
				
				if fileInfo.IsDir() {
					cmd = &command.FileCommand{
						Action:      "create_dir",
						FilePath:    event.Name,
						Description: fmt.Sprintf("创建目录: %s", event.Name),
						Logger:      global.Logger,
					}
				} else {
					cmd = &command.FileCommand{
						Action:      "create_file",
						FilePath:    event.Name,
						Description: fmt.Sprintf("创建文件: %s", event.Name),
						Logger:      global.Logger,
					}
				}
			case event.Op&fsnotify.Write == fsnotify.Write:
				
				cmd = &command.FileCommand{
					Action:      "write",
					FilePath:    event.Name,
					Description: fmt.Sprintf("修改文件: %s", event.Name),
					Logger:      global.Logger,
				}
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				cmd = &command.FileCommand{
					Action:      "remove",
					FilePath:    event.Name,
					Description: fmt.Sprintf("删除文件: %s", event.Name),
					Logger:      global.Logger,
				}
			case event.Op&fsnotify.Rename == fsnotify.Rename:
				cmd = &command.FileCommand{
					Action:      "rename",
					FilePath:    event.Name,
					Description: fmt.Sprintf("重命名文件: %s", event.Name),
					Logger:      global.Logger,
				}
			case event.Op&fsnotify.Chmod == fsnotify.Chmod:
				cmd = &command.FileCommand{
					Action:      "chmod",
					FilePath:    event.Name,
					Description: fmt.Sprintf("修改文件权限: %s", event.Name),
					Logger:      global.Logger,
				}
			}
			
			// 将命令添加到管理器内部异步执行
			if cmd != nil {
				commandManager.AddCommand(cmd)
			}
		})
		if err != nil {
			global.Logger.Error("监控目录失败:", zap.Error(err))
		}
	}()
	return nil
}

// 提供一个外部接口来撤销上一个操作
func UndoLastAction() error {
	if commandManager != nil {
		return commandManager.UndoLast()
	}
	return fmt.Errorf("command manager not initialized")
}

// 提供一个外部接口来撤销所有操作
func UndoAllActions() error {
	if commandManager != nil {
		return commandManager.UndoAll()
	}
	return fmt.Errorf("command manager not initialized")
}