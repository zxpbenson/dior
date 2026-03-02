package source

import (
	"context"
	"dior/component"
	"dior/internal/cache"
	"dior/internal/lg"
	"dior/option"
	"fmt"
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

func newPressSource(name string, opts *option.Options) (component.Component, error) {
	return &PressSource{
		Asynchronizer: component.NewAsynchronizer(name),
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
	// 自定义Start实现，没有使用 component.Asynchronizer 默认实现
	go s.doSend(ctx)
	go s.doCmd(ctx)
}

func (s *PressSource) doCmd(ctx context.Context) {
	s.Add(1)
	s.Asynchronizer.SetState(component.CompStateRunning)
	defer func() {
		if err := recover(); err != nil {
			lg.DftLgr.Error("PressSource.doCmd recover error : %v", err)
		}
		s.Asynchronizer.SetState(component.CompStateStopped)
		s.Done()
		s.Asynchronizer.ShowStats()
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	writing := false
	var start time.Time

LABEL: // 以1秒为周期，固定发压或者检测发压是否结束
	for {
		select {
		case <-ctx.Done():
			lg.DftLgr.Warn("PressSource.doCmd receive quit cmd")
			break LABEL
		case t := <-ticker.C:
			if writing {
				lg.DftLgr.Warn("PressSource.doCmd write delay, start time : %v, current time: %v, you should reduce speed immediately.", start, t)
				continue LABEL
			}
			start = time.Now()
			lg.DftLgr.Debug("PressSource.doCmd trigger write task and reset Ticker : %v", start)
			writing = true
			s.writeCmd <- s.speed
			ticker.Reset(time.Second)
		case speed := <-s.writeDone:
			writing = false
			lg.DftLgr.Info("PressSource.doCmd config speed %d, current speed %d", s.speed, speed)
		}
	}
	lg.DftLgr.Info("PressSource.doCmd do cmd done")
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
				count, err := s.writeLoop(ctx, limit, ticket)
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

func (s *PressSource) writeLoop(ctx context.Context, limit int64, ticket *cache.Ticket) (count int64, err error) {
	for cnt := int64(1); cnt <= limit; cnt++ {
		data := s.cache.Next(ticket)

		// 使用select防止channel满时阻塞
		select {
		case s.Channel <- data:
			count = cnt
		case <-ctx.Done():
			lg.DftLgr.Warn("PressSource.writeLoop context cancelled, stopping at count %d", count)
			return count, nil
		default:
			// channel已满，记录警告
			lg.DftLgr.Warn("PressSource.writeLoop channel full, stopping at count %d", count)
			return count, fmt.Errorf("channel full")
		}
	}
	lg.DftLgr.Debug("PressSource.writeLoop sent data count: %v", limit)
	return
}

func (s *PressSource) Stop() {
	s.Asynchronizer.Stop()
	lg.DftLgr.Info("PressSource.Stop called, goroutines will exit via context cancellation")
}
