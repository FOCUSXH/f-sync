package watcher

import (
	"fsync/client/global"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// AddRecursive 递归添加目录及其所有子目录到 watcher
func AddRecursive(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			err := watcher.Add(path)
			if err != nil {
				global.Logger.Error("无法监控目录", zap.String("path", path), zap.Error(err))
			} else {
				global.Logger.Info("开始监控目录", zap.String("path", path))
			}
		}
		return nil
	})
}

// WatchDirRecursive 递归监控目录，并自动监控新创建的子目录
func WatchDirRecursive(root string, onChange func(event fsnotify.Event)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// 1. 初始递归添加所有现有目录
	if err := AddRecursive(watcher, root); err != nil {
		return err
	}

	// 2. 监听事件
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			onChange(event)

			if event.Op&fsnotify.Create == fsnotify.Create {
				fi, err := os.Stat(event.Name)
				if err == nil && fi.IsDir() {
					global.Logger.Info("发现新目录，开始监控", zap.String("文件：", event.Name))
					err := AddRecursive(watcher, event.Name)
					if err != nil {
						return err
					}
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}
