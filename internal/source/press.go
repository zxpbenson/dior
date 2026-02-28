package source

import (
	"context"
	"dior/component"
	"dior/internal/cache"
	"dior/internal/lg"
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

func (s *PressSource) Init(channel chan []byte) error {
	s.Asynchronizer.Init(channel)
	return s.cache.Load(s.dataFile)
}

func (s *PressSource) Start(ctx context.Context) {
	//s.Asynchronizer.Start(ctx)//自己实现了定制化的Start
	go s.doSend(ctx)
	go s.generate(ctx)
}

func (s *PressSource) Stop() {
	s.Asynchronizer.Stop()
}

func (s *PressSource) generate(ctx context.Context) {
	s.Add(1)
	defer func() {
		if err := recover(); err != nil {
			lg.DftLgr.Error("PressSource.generate recover error : %v", err)
		}
		s.Done()
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	writing := false
	var start time.Time

LABEL: // 以1秒为周期，固定发压或者检测发压是否结束
	for {
		select {
		case <-ctx.Done():
			lg.DftLgr.Warn("PressSource.generate receive quit cmd")
			break LABEL
		case t := <-ticker.C:
			if writing {
				lg.DftLgr.Warn("PressSource.generate write delay, start time : %v, current time: %v, you should reduce speed immediately.", start, t)
				continue LABEL
			}
			start = time.Now()
			lg.DftLgr.Debug("PressSource.generate trigger write task and reset Ticker : %v", start)
			writing = true
			s.writeCmd <- s.speed
			ticker.Reset(time.Second)
		case speed := <-s.writeDone:
			writing = false
			lg.DftLgr.Info("PressSource.generate config speed %d, current speed %d", s.speed, speed)
		}
	}
	lg.DftLgr.Info("PressSource.generate do send done")
}

func (s *PressSource) doSend(ctx context.Context) {
	s.Add(1)
	defer func() {
		if err := recover(); err != nil {
			lg.DftLgr.Error("PressSource.doSend recover error : %v", err)
		}
		s.Done()
	}()

	ticket := cache.NewTicket()

LABEL:
	for {
		select {
		case <-ctx.Done():
			lg.DftLgr.Warn("PressSource.doSend receive quit cmd")
			break LABEL
		case limit := <-s.writeCmd:
			if limit > 0 {
				lg.DftLgr.Debug("PressSource.doSend receive send cmd : %v", limit)
				count, err := s.writeLoop(limit, ticket)
				if err != nil {
					lg.DftLgr.Error("PressSource.doSend send fail : %v", err)
				}
				// 使用 select 防止 generate 提前退出导致这里死锁
				select {
				case s.writeDone <- count:
				case <-ctx.Done():
					lg.DftLgr.Warn("PressSource.doSend receive quit cmd while reporting done")
					break LABEL
				}
			}
		}
	}
	lg.DftLgr.Info("PressSource.doSend do send done")
}

func (s *PressSource) writeLoop(limit int64, ticket *cache.Ticket) (count int64, err error) {
	for cnt := int64(1); cnt <= limit; cnt++ {
		data := s.cache.Next(ticket)
		s.Channel <- data
		count = cnt
	}
	lg.DftLgr.Debug("PressSource.doSend send data count : %v", limit)
	return
}
