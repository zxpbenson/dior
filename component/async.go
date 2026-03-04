package component

import (
	"context"
	"dior/internal/lg"
	"fmt"
	"sync"
	"sync/atomic"
)

// ComponentState 组件运行状态
type ComponentState int32

const (
	CompStateIdle ComponentState = iota
	CompStateRunning
	CompStateStopping
	CompStateStopped
)

func (s ComponentState) String() string {
	switch s {
	case CompStateIdle:
		return "Idle"
	case CompStateRunning:
		return "Running"
	case CompStateStopping:
		return "Stopping"
	case CompStateStopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}

// ErrorHandler 组件错误处理函数类型
type ErrorHandler func(err error)

// Asynchronizer 提供异步处理能力的基础组件
// 职责：
// 1. 管理数据channel的读取
// 2. 通过Output函数处理数据
// 3. 支持优雅关闭和强制退出
// 4. 提供错误通知机制
type Asynchronizer struct {
	control *sync.WaitGroup
	Channel chan []byte
	Output  OutputFunc

	name    string       // component name
	state   atomic.Int32 // 状态管理
	onError ErrorHandler // 错误处理

	// 统计信息
	processedCount atomic.Int64
	errorCount     atomic.Int64
}

// NewAsynchronizer 创建新的Asynchronizer实例
func NewAsynchronizer(name string) *Asynchronizer {
	a := &Asynchronizer{name: name}
	a.SetState(CompStateIdle)
	return a
}

// Init 初始化Asynchronizer
func (a *Asynchronizer) Init(channel chan []byte) {
	a.Channel = channel
}

// UnderControl 注册到WaitGroup以支持优雅关闭
func (a *Asynchronizer) UnderControl(control *sync.WaitGroup) {
	a.control = control
}

// SetErrorHandler 设置错误处理函数
func (a *Asynchronizer) SetErrorHandler(handler ErrorHandler) {
	a.onError = handler
}

// GetState 获取当前状态
func (a *Asynchronizer) GetState() ComponentState {
	return ComponentState(a.state.Load())
}

// SetState 获取当前状态
func (a *Asynchronizer) SetState(state ComponentState) {
	a.state.Store(int32(state))
}

func (a *Asynchronizer) AddProcessedCount(delta int64) {
	a.processedCount.Add(delta)
}

func (a *Asynchronizer) AddErrorCount(delta int64) {
	a.errorCount.Add(delta)
}

// GetStats 获取统计信息
func (a *Asynchronizer) GetStats() (processed, errors int64) {
	return a.processedCount.Load(), a.errorCount.Load()
}

// 打印统计信息
func (a *Asynchronizer) ShowStats() {
	lg.DftLgr.Info("Asynchronizer.work stopped for %s, processed=%d, errors=%d",
		a.name, a.processedCount.Load(), a.errorCount.Load())
}

// work 是Sink组件的核心工作循环
// 设计原则：
// - 不主动监听ctx.Done()，确保能排空channel中的剩余数据
// - 通过channel关闭来触发正常退出
// - 通过panic恢复来处理异常
// 注意：调用方需在启动goroutine前调用Add(1)，避免与Wait()产生竞争
func (a *Asynchronizer) work(ctx context.Context) {
	a.SetState(CompStateRunning)
	defer func() {
		if err := recover(); err != nil {
			lg.DftLgr.Error("Asynchronizer.work panic recovered: %v", err)
			a.errorCount.Add(1)
			if a.onError != nil {
				a.onError(fmt.Errorf("panic: %v", err))
			}
		}
		a.SetState(CompStateStopped)
		a.control.Done()
		a.ShowStats()
	}()

	for {
		// Sink端专注于消费数据
		// 不监听ctx.Done()以保证Graceful Shutdown能排空剩余数据
		select {
		case data, ok := <-a.Channel:
			if !ok {
				lg.DftLgr.Info("Asynchronizer.work channel closed, exiting gracefully")
				return
			}
			a.processData(data)
		case <-ctx.Done():
			// 即使收到取消信号，也要尝试排空channel
			lg.DftLgr.Warn("Asynchronizer.work context cancelled, draining channel...")
			a.drainChannel()
			return
		}
	}
}

// processData 处理单条数据
func (a *Asynchronizer) processData(data []byte) {
	defer func() {
		if err := recover(); err != nil {
			lg.DftLgr.Error("Asynchronizer.processData panic: %v", err)
			a.errorCount.Add(1)
		}
	}()

	if a.Output != nil {
		if err := a.Output(data); err == nil {
			a.processedCount.Add(1)
		} else {
			panic(err)
		}
	}
}

// drainChannel 排空channel中的剩余数据
func (a *Asynchronizer) drainChannel() {
	for {
		select {
		case data, ok := <-a.Channel:
			if !ok {
				lg.DftLgr.Warn("Asynchronizer.drainChannel channel closed and empty, exiting gracefully")
				return
			}
			a.processData(data)
		default:
			// channel已空
			lg.DftLgr.Warn("Asynchronizer.drainChannel channel empty, exiting gracefully")
			return
		}
	}
}

// Start 启动异步处理goroutine
// 注意：具体组件可以自行定制Start的行为
func (a *Asynchronizer) Start(ctx context.Context) {
	a.control.Add(1) // 在启动goroutine前调用Add，避免与Wait()产生竞争
	go a.work(ctx)
	lg.DftLgr.Info("Asynchronizer.Start done, state=%s", a.GetState())
}

// Stop 停止异步处理
// 注意：Asynchronizer本身不执行停止操作，由Controller通过关闭channel来触发退出
func (a *Asynchronizer) Stop() {
	a.SetState(CompStateStopping)
	lg.DftLgr.Info("Asynchronizer.Stop called, state=%s", a.GetState())
}

// Add 增加WaitGroup计数
func (a *Asynchronizer) Add(delta int) {
	a.control.Add(delta)
	lg.DftLgr.Debug("Asynchronizer.Add(%d)", delta)
}

// Done 减少WaitGroup计数
func (a *Asynchronizer) Done() {
	a.control.Done()
	lg.DftLgr.Debug("Asynchronizer.Done()")
}
