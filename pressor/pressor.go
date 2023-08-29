package pressor

import (
	"dior/cache"
	"dior/lg"
	"dior/option"
	"dior/writer"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type PressAble interface {
	Start() error
}

type Pressor struct {
	logger    lg.Logger
	opts      *option.Options
	cache     *cache.Cache
	sysSig    chan os.Signal
	writeDone chan int64
	writeCmd  chan int64
	writer    writer.WriteAble
}

func NewPressor(opts *option.Options) (PressAble, error) {
	writer, err := writer.NewWriter(opts)
	if err != nil {
		return nil, err
	}
	logger, err := lg.NewLogger(opts.LogPrefix, opts.LogLevel)
	if err != nil {
		return nil, err
	}
	return &Pressor{
		logger:    logger,
		cache:     cache.NewCache(),
		writeDone: make(chan int64),
		writeCmd:  make(chan int64, 10),
		sysSig:    make(chan os.Signal, 1),
		opts:      opts,
		writer:    writer,
	}, nil
}

func (this *Pressor) doWrite() {
	ticket := cache.NewTicket()
	err := this.writer.Open()
	if err != nil {
		this.logger.Fatal("Writer open fail : %v", err)
	}
	defer this.writer.Close()
CONTROL:
	for {
		select {
		case limit := <-this.writeCmd:
			if limit <= 0 {
				this.logger.Warn("receive quit task : %v", limit)
				break CONTROL
			} else {
				this.logger.Debug("receive write task : %v", limit)
				count, err := this.writeLoop(limit, ticket)
				if err != nil {
					this.logger.Fatal("write fail : %v", err)
				}
				this.writeDone <- count
			}
		}
	}
	this.logger.Info("do write done")
}

func (this *Pressor) writeLoop(limit int64, ticket *cache.Ticket) (count int64, err error) {
	for cnt := int64(1); cnt <= limit; cnt++ {
		one := this.cache.Next(ticket)
		err = this.writer.Write(one)
		if err != nil {
			return
		}
		count = cnt
	}
	this.logger.Debug("write data count : %v", limit)
	return
}

func (this *Pressor) Start() error {

	this.cache.Load(this.opts.DataFile)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	signal.Notify(this.sysSig, os.Interrupt, syscall.SIGINT, syscall.SIGQUIT)
	defer signal.Stop(this.sysSig)

	writing, delay := false, false
	var start time.Time

	go this.doWrite()

CONTROL:
	for {
		select {
		case sig := <-this.sysSig:
			this.logger.Warn("receive signal : %v", sig)
			this.writeCmd <- 0
			break CONTROL
		case t := <-ticker.C:
			if writing {
				delay = true
				this.logger.Warn("write delay, start time : %v, Current time: %v, you should reduce speed immediately.", start, t)
			} else {
				writing = true
				start = time.Now()
				if delay {
					delay = false
					ticker.Reset(time.Second)
					this.logger.Warn("reset Ticker : %v", start)
				}
				this.logger.Debug("trigger write task : %v", start)
				this.writeCmd <- this.opts.Speed
			}
		case speed := <-this.writeDone:
			writing = false
			this.logger.Info("config speed %d, current speed %d", this.opts.Speed, speed)
		}
	}

	return nil
}
