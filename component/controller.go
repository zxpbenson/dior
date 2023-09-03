package component

import (
	"dior/lg"
	"dior/option"
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
	control *sync.WaitGroup
}

func NewController() *Controller {
	return &Controller{
		control: &sync.WaitGroup{},
		sysSig:  make(chan os.Signal, 1),
	}
}

func (this *Controller) AddComponents(opts *option.Options) {
	source, err := NewComponent(opts.Src+"-source", opts)
	if err != nil {
		lg.DftLgr.Fatal("Controller.AddComponents create source fail : %v", err)
	}
	sink, err := NewComponent(opts.Dst+"-sink", opts)
	if err != nil {
		lg.DftLgr.Fatal("Controller.AddComponents create sink fail : %v", err)
	}
	channel := make(chan []byte, opts.ChanSize)
	this.AddIO(source, channel, sink)
}

func (this *Controller) AddIO(source Component, channel chan []byte, sink Component) {
	this.source = source
	this.sink = sink
	source.SetChannel(channel)
	sink.SetChannel(channel)
	source.UnderControl(this.control)
	sink.UnderControl(this.control)
}

func (this *Controller) Init() {
	err := this.sink.Init()
	if err != nil {
		lg.DftLgr.Fatal("Controller init sink fail : %v", err)
	}

	err = this.source.Init()
	if err != nil {
		lg.DftLgr.Fatal("Controller init source fail : %v", err)
	}
}

func (this *Controller) Start() {

	this.sink.Start()
	this.source.Start()

	signal.Notify(this.sysSig, os.Interrupt, syscall.SIGINT, syscall.SIGQUIT)

CONTROL:
	for {
		select {
		case sig := <-this.sysSig:
			lg.DftLgr.Warn("Controller receive signal from system : %v", sig)
			signal.Stop(this.sysSig)
			this.sink.Stop()
			this.source.Stop()
			break CONTROL
		}
	}

	lg.DftLgr.Warn("Controller wait for source / sink close done...")
	this.control.Wait()
	lg.DftLgr.Warn("Controller has closed all source / sink.")
}
