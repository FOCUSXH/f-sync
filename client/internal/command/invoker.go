// client/internal/command/invoker.go
package command

import (
	"sync"

	"go.uber.org/zap"
)

// CommandQueue 定义命令队列结构
type CommandQueue struct {
	commands chan Command
	workers  int
	logger   *zap.Logger
	// 用于优雅关闭
	wg   sync.WaitGroup
	quit chan struct{}
}

// NewCommandQueue 创建一个新的命令队列
func NewCommandQueue(logger *zap.Logger, bufferSize int, numWorkers int) *CommandQueue {
	cq := &CommandQueue{
		commands: make(chan Command, bufferSize), // 带缓冲的 channel
		workers:  numWorkers,
		logger:   logger,
		quit:     make(chan struct{}),
	}

	// 启动工作协程
	for i := 0; i < numWorkers; i++ {
		cq.wg.Add(1)
		go cq.worker(i)
	}

	return cq
}

// worker 工作协程，负责从队列中取出命令并执行
func (cq *CommandQueue) worker(workerID int) {
	defer cq.wg.Done()
	for {
		select {
		case cmd := <-cq.commands: // 从 channel 中接收命令
			if cmd == nil { // 检查是否收到关闭信号（发送 nil 作为关闭信号）
				cq.logger.Info("工作线程接收到关闭信号", zap.Int("worker_id", workerID))
				return
			}
			cq.logger.Info("工作线程开始执行命令", zap.Int("worker_id", workerID), zap.String("command_desc", cmd.GetDescription()))
			if err := cmd.Execute(); err != nil {
				cq.logger.Error("工作线程执行命令失败", zap.Int("worker_id", workerID), zap.Error(err))
			} else {
				cq.logger.Info("工作线程已完成命令执行", zap.Int("worker_id", workerID), zap.String("command_desc", cmd.GetDescription()))
			}
		case <-cq.quit: // 接收到退出信号
			cq.logger.Info("Worker shutting down", zap.Int("worker_id", workerID))
			return
		}
	}
}

// Enqueue 将命令添加到队列中（异步）
func (cq *CommandQueue) Enqueue(cmd Command) {
	select {
	case cq.commands <- cmd:
		cq.logger.Info("工作指令已加入队列", zap.String("command_desc", cmd.GetDescription()))
	default:
		// 如果队列满了，可以选择丢弃命令或返回错误
		// 这里选择丢弃并记录日志
		cq.logger.Warn("队列已满，丢弃执行命令", zap.String("command_desc", cmd.GetDescription()))
	}
}

// Stop 优雅地停止命令队列
func (cq *CommandQueue) Stop() {
	close(cq.quit) // 发送退出信号给所有 worker
	// 发送 nil 命令给每个 worker，作为一种更明确的关闭信号（可选，quit channel 通常足够）
	for i := 0; i < cq.workers; i++ {
		cq.Enqueue(nil) // 发送 nil 作为关闭信号
	}
	cq.wg.Wait()       // 等待所有 worker 结束
	close(cq.commands) // 关闭命令 channel
	cq.logger.Info("Command Queue stopped gracefully")
}

// GetCommandChannel 获取底层的命令 channel (如果需要从外部发送命令)
func (cq *CommandQueue) GetCommandChannel() chan<- Command {
	return cq.commands
}
