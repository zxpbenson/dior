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
)

type Controllable interface {
	UnderControl(control *sync.WaitGroup)
}

type Controller struct {
	source  Component
	sink    Component
	sysSig  chan os.Signal
	srcWG   *sync.WaitGroup
	sinkWG  *sync.WaitGroup
	channel chan []byte
}

func NewController() *Controller {
	return &Controller{
		srcWG:  &sync.WaitGroup{},
		sinkWG: &sync.WaitGroup{},
		sysSig: make(chan os.Signal, 1),
	}
}

func (c *Controller) AddComponents(opts *option.Options) error {
	source, err := NewComponent(opts.Src+"-source", opts)
	if err != nil {
		return fmt.Errorf("Controller.AddComponents create source fail : %w", err)
	}
	sink, err := NewComponent(opts.Dst+"-sink", opts)
	if err != nil {
		return fmt.Errorf("Controller.AddComponents create sink fail : %w", err)
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

func (c *Controller) Init() error {
	if err := c.sink.Init(c.channel); err != nil {
		return fmt.Errorf("Controller init sink fail : %w", err)
	}

	if err := c.source.Init(c.channel); err != nil {
		return fmt.Errorf("Controller init source fail : %w", err)
	}
	return nil
}

func (c *Controller) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	//defer cancel() // 重复

	c.sink.Start(ctx)
	c.source.Start(ctx)

	signal.Notify(c.sysSig, os.Interrupt, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case sig := <-c.sysSig:
		lg.DftLgr.Warn("Controller receive signal from system : %v", sig)
		signal.Stop(c.sysSig)
		
		// 1. 通知并强制关闭 Source
		cancel()        // 通过 ctx 取消通知 Source
		c.source.Stop() // 显式调用 Stop (某些组件如 NSQ 依赖此方法阻塞等待结束)
	}

	lg.DftLgr.Warn("Controller wait for source close done...")
	c.srcWG.Wait()

	// 2. Source 已完全停止，不会再产生数据，此时安全关闭 Channel
	close(c.channel)
	lg.DftLgr.Warn("Controller closed data channel.")

	// 3. 通知 Sink 准备结束，并等待其排空 Channel 中的剩余数据
	lg.DftLgr.Warn("Controller wait for sink close done...")
	c.sink.Stop()
	c.sinkWG.Wait()

	lg.DftLgr.Warn("Controller has safely closed all source / sink.")
}
