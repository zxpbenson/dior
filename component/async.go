package component

import (
	"dior/lg"
	"sync"
	"time"
)

type Asynchronizer struct {
	control   *sync.WaitGroup
	Channel   chan []byte
	CloseChan chan interface{}
	Output    OutputFunc
}

func (this *Asynchronizer) Init() {
	//this.control = &sync.WaitGroup{}
	//this.Channel = make(chan []byte, 100)
	this.CloseChan = make(chan interface{})
}

func (this *Asynchronizer) UnderControl(control *sync.WaitGroup) {
	this.control = control
}

func (this *Asynchronizer) work() {
	this.control.Add(1)
	defer func() {
		if err := recover(); err != nil {
			lg.DftLgr.Error("Asynchronizer.work recover error : %v", err)
		}
		this.control.Done()
		lg.DftLgr.Warn("Asynchronizer.work goroutine defer func done.")
	}()

LABEL:
	for {
		select {
		case _, ok := <-this.CloseChan:
			if !ok {
				lg.DftLgr.Warn("Asynchronizer.work receive quit cmd")
				break LABEL
			}
		case data := <-this.Channel:
			this.Output(data)
		}
	}
	lg.DftLgr.Info("Asynchronizer.Work do write done")
}

func (this *Asynchronizer) Start() {
	go this.work()
	lg.DftLgr.Info("Asynchronizer.start done.")
}

func (this *Asynchronizer) Stop() {
	close(this.CloseChan) //通知 work goroutine 结束
	lg.DftLgr.Info("Asynchronizer.Stop done.")
}

func (this *Asynchronizer) SetChannel(channel chan []byte) {
	this.Channel = channel
	lg.DftLgr.Info("Asynchronizer.SetDataChan done.")
}

func (this *Asynchronizer) WaitEmpty() {
	for {
		len := len(this.Channel)
		if len == 0 {
			break
		}
		lg.DftLgr.Warn("Asynchronizer.WaitEmpty remain data : %v", len)
		time.Sleep(time.Second)
	}
	lg.DftLgr.Info("Asynchronizer.WaitEmpty done.")
}

func (this *Asynchronizer) Add(delta int) {
	this.control.Add(delta)
	lg.DftLgr.Info("Asynchronizer.Add %v", delta)
}

func (this *Asynchronizer) Done() {
	this.control.Done()
	lg.DftLgr.Info("Asynchronizer.Done done.")
}
