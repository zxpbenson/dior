package source

import (
	"dior/cache"
	"dior/component"
	"dior/lg"
	"dior/option"
	"time"
)

type PressSource struct {
	*component.Asynchronizer

	cache     *cache.Cache
	dataFile  string
	writeDone chan int64
	writeCmd  chan int64
	speed     int64
}

func init() {
	component.RegCmpCreator("press-source", newPressSource)
}

func newPressSource(opts *option.Options) (component.Component, error) {
	return &PressSource{
		Asynchronizer: &component.Asynchronizer{},
		cache:         cache.NewCache(opts.SrcScannerBufSizeMb),
		dataFile:      opts.SrcFile,
		writeDone:     make(chan int64),
		writeCmd:      make(chan int64, 10),
		speed:         opts.SrcSpeed,
	}, nil
}

func (this *PressSource) Init() error {
	this.Asynchronizer.Init()
	return this.cache.Load(this.dataFile)
}

func (this *PressSource) Start() {
	//this.Asynchronizer.Start()//自己实现了定制化的Start
	go this.doSend()
	go this.generate()
}

func (this *PressSource) Stop() {
	this.Asynchronizer.Stop()
}

func (this *PressSource) generate() {
	this.Add(1)
	defer func() {
		if err := recover(); err != nil {
			lg.DftLgr.Error("PressSource.generate recover error : %v", err)
		}
		this.Done()
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	writing, delay := false, false
	var start time.Time

LABEL:
	for {
		select {
		case _, ok := <-this.CloseChan:
			if !ok {
				lg.DftLgr.Warn("PressSource.generate receive quit cmd")
				break LABEL
			}
		case t := <-ticker.C:
			if writing {
				delay = true
				lg.DftLgr.Warn("PressSource.generate write delay, start time : %v, current time: %v, you should reduce speed immediately.", start, t)
			} else {
				writing = true
				start = time.Now()
				if delay {
					delay = false
					ticker.Reset(time.Second)
					lg.DftLgr.Warn("PressSource.generate reset Ticker : %v", start)
				}
				lg.DftLgr.Debug("PressSource.generate trigger write task : %v", start)
				this.writeCmd <- this.speed
			}
		case speed := <-this.writeDone:
			writing = false
			lg.DftLgr.Info("PressSource.generate config speed %d, current speed %d", this.speed, speed)
		}
	}
	lg.DftLgr.Info("PressSource.generate do send done")
}

func (this *PressSource) doSend() {
	this.Add(1)
	defer func() {
		if err := recover(); err != nil {
			lg.DftLgr.Error("PressSource.doSend recover error : %v", err)
		}
		this.Done()
	}()

	ticket := cache.NewTicket()

LABEL:
	for {
		select {
		case _, ok := <-this.CloseChan:
			if !ok {
				lg.DftLgr.Warn("PressSource.doSend receive quit cmd")
				break LABEL
			}
		case limit := <-this.writeCmd:
			if limit > 0 {
				lg.DftLgr.Debug("PressSource.doSend receive send cmd : %v", limit)
				count, err := this.writeLoop(limit, ticket)
				if err != nil {
					lg.DftLgr.Fatal("PressSource.doSend send fail : %v", err)
				}
				this.writeDone <- count
			}
		}
	}
	lg.DftLgr.Info("PressSource.doSend do send done")
}

func (this *PressSource) writeLoop(limit int64, ticket *cache.Ticket) (count int64, err error) {
	for cnt := int64(1); cnt <= limit; cnt++ {
		data := this.cache.Next(ticket)
		this.Channel <- data
		count = cnt
	}
	lg.DftLgr.Debug("PressSource.doSend send data count : %v", limit)
	return
}
