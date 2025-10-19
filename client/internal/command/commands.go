// client/internal/command/commands.go
package command

import (
	"fsync/client/global"

	"go.uber.org/zap"
)

// FileCommand 文件操作命令结构体
type FileCommand struct {
	Action      string
	FilePath    string
	Description string
	Logger      *zap.Logger
}

// Execute 执行命令
func (fc *FileCommand) Execute() error {
	fc.Logger.Info("执行文件操作命令",
		zap.String("action", fc.Action),
		zap.String("file", fc.FilePath))

	switch fc.Action {
	case "create_dir":
		global.Logger.Info("创建目录")

	case "create_file":
		global.Logger.Info("创建文件")

	case "write":
		global.Logger.Info("修改文件")

	case "remove":
		global.Logger.Info("删除文件")

	case "rename":
		global.Logger.Info("重命名文件")

	case "chmod":
		global.Logger.Info("修改文件权限")

	}
	return nil
}

// Undo 撤销命令
func (fc *FileCommand) Undo() error {
	// 这里应该实现撤销操作
	fc.Logger.Info("撤销文件操作命令",
		zap.String("action", fc.Action),
		zap.String("file", fc.FilePath))
	return nil
}

// GetDescription 获取命令描述
func (fc *FileCommand) GetDescription() string {
	return fc.Description
}

// NewCommandManager 创建命令管理器实例
func NewCommandManager(logger *zap.Logger, queueBufferSize int, numWorkers int) *CommandManager {
	commandQueue := NewCommandQueue(logger, queueBufferSize, numWorkers)
	return &CommandManager{
		Commands:     make([]Command, 0),
		CommandQueue: commandQueue,
		Logger:       logger,
	}
}

// AddCommand 添加命令
func (cm *CommandManager) AddCommand(cmd Command) {
	// 为了支持撤销，需要将命令也存储在切片中
	cm.mutex.Lock()
	cm.Commands = append(cm.Commands, cmd)
	cm.mutex.Unlock()

	// 将命令添加到异步队列中执行
	cm.CommandQueue.Enqueue(cmd)
	cm.Logger.Info("添加命令到命令队列并准备异步执行", zap.String("description", cmd.GetDescription()))
}

// UndoLast 撤销最后一个命令
func (cm *CommandManager) UndoLast() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if len(cm.Commands) == 0 {
		cm.Logger.Info("没有可撤销的命令")
		return nil // 或者返回错误
	}

	lastCmd := cm.Commands[len(cm.Commands)-1]
	cm.Logger.Info("撤销最后一个命令", zap.String("description", lastCmd.GetDescription()))

	if err := lastCmd.Undo(); err != nil {
		cm.Logger.Error("撤销命令失败", zap.Error(err))
		return err
	}

	// 从切片中移除已撤销的命令
	cm.Commands = cm.Commands[:len(cm.Commands)-1]
	return nil
}

// UndoAll 撤销所有命令
func (cm *CommandManager) UndoAll() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 从后往前撤销，符合栈的 LIFO 原则
	for i := len(cm.Commands) - 1; i >= 0; i-- {
		cmd := cm.Commands[i]
		cm.Logger.Info("撤销命令", zap.String("description", cmd.GetDescription()))

		if err := cmd.Undo(); err != nil {
			cm.Logger.Error("撤销命令失败", zap.Error(err))
			// 可以选择继续撤销其他命令或停止
			// return err // 如果需要失败时停止
		}
	}

	// 清空所有命令
	cm.Commands = cm.Commands[:0]
	return nil
}

// Stop 优雅地停止命令管理器（停止内部的队列）
func (cm *CommandManager) Stop() {
	cm.CommandQueue.Stop()
	cm.Logger.Info("Command Manager stopped gracefully")
}
