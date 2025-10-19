// client/internal/command/command.go
package command

import (
	"sync"

	"go.uber.org/zap"
)

// Command 定义命令接口
type Command interface {
	Execute() error
	Undo() error
	GetDescription() string
}

// CommandManager 命令管理器
type CommandManager struct {
	// Commands []Command // 保留这个切片用于撤销操作
	Commands     []Command     // 用于存储所有命令以支持撤销
	CommandQueue *CommandQueue // 用于异步执行
	Logger       *zap.Logger
	mutex        sync.Mutex // 保护 Commands 切片的并发访问
}
