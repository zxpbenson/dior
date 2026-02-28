package component

import (
	"context"
	"dior/internal/lg"
	"sync"
)

type Asynchronizer struct {
	control *sync.WaitGroup
	Channel chan []byte
	Output  OutputFunc
	ctx     context.Context
}

func (a *Asynchronizer) Init(channel chan []byte) {
	a.Channel = channel
}

func (a *Asynchronizer) UnderControl(control *sync.WaitGroup) {
	a.control = control
}

func (a *Asynchronizer) work(ctx context.Context) {
	a.control.Add(1)
	defer func() {
		if err := recover(); err != nil {
			lg.DftLgr.Error("Asynchronizer.work recover error : %v", err)
		}
		a.control.Done()
		lg.DftLgr.Warn("Asynchronizer.work goroutine defer func done.")
	}()

	for {
		// Sink 端专注于消费数据。不监听 ctx.Done() 以保证 Graceful Shutdown 能排空剩余数据
		data, ok := <-a.Channel
		if !ok {
			lg.DftLgr.Warn("Asynchronizer.work channel closed")
			return
		}
		a.Output(data)
	}
}

// 不要感到疑惑，在具体组件中，可以选择自行定制Start的具体行为
func (a *Asynchronizer) Start(ctx context.Context) {
	a.ctx = ctx
	go a.work(ctx)
	lg.DftLgr.Info("Asynchronizer.start done.")
}

func (a *Asynchronizer) Stop() {
	// Context cancellation is handled by the parent controller
	lg.DftLgr.Info("Asynchronizer.Stop done.")
}

func (a *Asynchronizer) Add(delta int) {
	a.control.Add(delta)
	lg.DftLgr.Info("Asynchronizer.Add %v", delta)
}

func (a *Asynchronizer) Done() {
	a.control.Done()
	lg.DftLgr.Info("Asynchronizer.Done done.")
}
