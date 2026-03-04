package component

import (
	"context"
	"dior/internal/lg"
	"dior/option"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// State 表示Controller的运行状态
type State int

const (
	StateInitialized State = iota
	StateStarting
	StateRunning
	StateStopping
	StateStopped
)

func (s State) String() string {
	switch s {
	case StateInitialized:
		return "Initialized"
	case StateStarting:
		return "Starting"
	case StateRunning:
		return "Running"
	case StateStopping:
		return "Stopping"
	case StateStopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}

// Controllable 定义可被控制的组件接口
type Controllable interface {
	UnderControl(control *sync.WaitGroup)
}

// Controller 管理Source和Sink组件的生命周期
// 职责：
// 1. 创建和管理数据传输channel
// 2. 协调Source和Sink的启动和停止
// 3. 处理系统信号实现优雅关闭
// 4. 确保数据完整性和资源正确释放
type Controller struct {
	source  Component       //
	sink    Component       //
	sysSig  chan os.Signal  // init in controller.NewController
	srcWG   *sync.WaitGroup // init in controller.NewController
	sinkWG  *sync.WaitGroup // init in controller.NewController
	channel chan []byte     //

	// 状态管理
	state   State        // init in controller.NewController
	stateMu sync.RWMutex //

	// cancel用于取消消费循环
	// 注意：只保存cancel不保存context是Go官方允许的模式
	cancel context.CancelFunc

	// 配置
	stopTimeout time.Duration // 停止超时时间 init in controller.NewController
}

// NewController 创建新的Controller实例
func NewController() *Controller {
	return &Controller{
		srcWG:       &sync.WaitGroup{},
		sinkWG:      &sync.WaitGroup{},
		sysSig:      make(chan os.Signal, 1),
		state:       StateInitialized,
		stopTimeout: 30 * time.Second, // 默认30秒超时
	}
}

// SetStopTimeout 设置停止超时时间
func (c *Controller) SetStopTimeout(timeout time.Duration) {
	c.stopTimeout = timeout
}

// GetState 获取当前状态
func (c *Controller) GetState() State {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state
}

func (c *Controller) setState(state State) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	lg.DftLgr.Info("Controller state change: %s -> %s", c.state, state)
	c.state = state
}

// AddComponents 根据配置创建并添加Source和Sink组件
func (c *Controller) AddComponents(opts *option.Options) error {
	source, err := NewComponent(opts.Src+"-source", opts)
	if err != nil {
		return fmt.Errorf("Controller.AddComponents create source fail: %w", err)
	}
	sink, err := NewComponent(opts.Dst+"-sink", opts)
	if err != nil {
		return fmt.Errorf("Controller.AddComponents create sink fail: %w", err)
	}
	channel := make(chan []byte, opts.ChanSize)
	c.addIO(source, channel, sink)
	return nil
}

func (c *Controller) addIO(source Component, channel chan []byte, sink Component) {
	c.source = source
	c.sink = sink
	c.channel = channel
	source.UnderControl(c.srcWG)
	sink.UnderControl(c.sinkWG)
}

// Init 初始化所有组件
// 按照Sink先于Source的顺序初始化，确保数据通路就绪
func (c *Controller) Init() error {
	// 先初始化Sink，确保消费者就绪
	if err := c.sink.Init(c.channel); err != nil {
		return fmt.Errorf("controller init sink fail: %w", err)
	}

	// 再初始化Source
	if err := c.source.Init(c.channel); err != nil {
		return fmt.Errorf("controller init source fail: %w", err)
	}
	return nil
}

// Start 启动所有组件并阻塞直到收到停止信号
// 启动顺序：先启动Sink确保消费者就绪，再启动Source开始生产数据
func (c *Controller) Start() {
	c.setState(StateStarting)

	// 创建可取消的上下文
	// 注意：ctx作为局部变量传递，cancel保存在结构体中供Stop()使用
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	// 先启动Sink，确保消费者就绪
	c.sink.Start(ctx)
	lg.DftLgr.Info("Controller started sink")

	// 再启动Source
	c.source.Start(ctx)
	lg.DftLgr.Info("Controller started source")

	c.setState(StateRunning)

	// 监听系统信号
	signal.Notify(c.sysSig, os.Interrupt, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	// 阻塞等待停止信号
	sig := <-c.sysSig
	signal.Stop(c.sysSig)
	lg.DftLgr.Warn("Controller receive signal: %v", sig)

	// 执行优雅关闭
	c.gracefulShutdown()
}

// gracefulShutdown 执行优雅关闭流程
// 关闭顺序：
// 1. 停止Source（停止生产数据）
// 2. 等待Source完全停止
// 3. 关闭Channel（通知Sink没有更多数据）
// 4. 等待Sink排空Channel中的数据
// 5. 停止Sink（释放资源）
func (c *Controller) gracefulShutdown() {
	c.setState(StateStopping)

	// 创建超时上下文，防止停止流程卡住
	// 在goroutine中执行source.Stop()，支持超时
	ctx, cancel := context.WithTimeout(context.Background(), c.stopTimeout)
	defer cancel()

	// ====== Phase 1: 停止Source ======
	// 发送取消信号
	c.cancel()
	lg.DftLgr.Warn("Controller phase 1: trigger global cancellation and waiting for source goroutines...")

	sourceWaitDone := make(chan struct{})
	go func() {
		defer close(sourceWaitDone)
		c.srcWG.Wait()
	}()

	select {
	case <-sourceWaitDone:
		lg.DftLgr.Info("Controller source goroutines stopped")
	case <-ctx.Done():
		lg.DftLgr.Warn("Controller wait for source timeout, proceeding anyway")
	}

	// ====== Phase 2: 等待Source完全停止 ======
	lg.DftLgr.Warn("Controller phase 2: stopping source...")
	sourceStopDone := make(chan struct{})
	go func() {
		defer close(sourceStopDone)
		c.source.Stop()
	}()

	select {
	case <-sourceStopDone:
		lg.DftLgr.Info("Controller source.Stop() completed")
	case <-ctx.Done():
		lg.DftLgr.Warn("Controller source.Stop() timeout, proceeding anyway")
	}

	// ====== Phase 3: 关闭Channel ======
	lg.DftLgr.Warn("Controller phase 3: closing data channel...")
	close(c.channel)
	lg.DftLgr.Info("Controller data channel closed")

	// ====== Phase 4: 等待Sink排空数据 ======
	lg.DftLgr.Warn("Controller phase 4: waiting for sink to drain...")
	sinkWaitDone := make(chan struct{})
	go func() {
		defer close(sinkWaitDone)
		c.sinkWG.Wait()
	}()

	select {
	case <-sinkWaitDone:
		lg.DftLgr.Info("Controller sink drained all data")
	case <-ctx.Done():
		lg.DftLgr.Warn("Controller wait for sink drain timeout, proceeding anyway")
	}

	// ====== Phase 5: 停止Sink ======
	lg.DftLgr.Warn("Controller phase 5: stopping sink...")
	sinkStopDone := make(chan struct{})
	go func() {
		defer close(sinkStopDone)
		c.sink.Stop()
	}()

	select {
	case <-sinkStopDone:
		lg.DftLgr.Info("Controller sink.Stop() completed")
	case <-ctx.Done():
		lg.DftLgr.Warn("Controller sink.Stop() timeout")
	}

	c.setState(StateStopped)
	lg.DftLgr.Warn("Controller graceful shutdown completed")
}

// Stop 主动停止Controller（用于程序化停止，非信号触发）
// 此方法暂时用不到
func (c *Controller) Stop() {
	if c.GetState() != StateRunning {
		lg.DftLgr.Warn("Controller state shutdown %i", c.GetState())
		return
	}

	// 触发停止信号
	c.sysSig <- syscall.SIGTERM
}
