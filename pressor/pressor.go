package pressor

import (
	"dior/cache"
	"dior/option"
	"dior/writer"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type PressAble interface {
	Start() error
}

type Pressor struct {
	opts      *option.Options
	cache     *cache.Cache
	sysSig    chan os.Signal
	writeDone chan int64
	writeCmd  chan int64
	writer    writer.WriteAble
}

func NewPressAble(opts *option.Options) (PressAble, error) {
	writer, err := writer.NewWriter(opts)
	if err != nil {
		return nil, err
	}

	return &Pressor{
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
		fmt.Printf("Writer open fail : %v\n", err)
		os.Exit(1)
	}
	defer this.writer.Close()
CONTROL:
	for {
		select {
		case limit := <-this.writeCmd:
			if limit <= 0 {
				fmt.Printf("receive quit task : %v\n", limit)
				break CONTROL
			} else {
				//fmt.Printf("receive write task : %v\n", limit)
				count, err := this.writeLoop(limit, ticket)
				if err != nil {
					fmt.Printf("write fail : %v\n", err)
					os.Exit(1)
				}
				this.writeDone <- count
			}
		}
	}
	fmt.Printf("do write done\n")
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
	//fmt.Printf("write data count : %v\n", limit)
	return
}

//func (this *Pressor) WriteOnce(data string) {
//	fmt.Printf("It`s a abstract method, you must implements your own method.\n")
//}

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
			fmt.Printf("receive signal : %v\n", sig)
			this.writeCmd <- 0
			break CONTROL
		case t := <-ticker.C:
			if writing {
				delay = true
				fmt.Printf("write delay, start time : %v, Current time: %v, you should reduce speed immediately\n", start, t)
			} else {
				writing = true
				start = time.Now()
				if delay {
					delay = false
					ticker.Reset(time.Second)
					fmt.Printf("reset Ticker\n")
				}
				//fmt.Printf("trigger write task : %v\n", start)
				this.writeCmd <- this.opts.Speed
			}
		case speed := <-this.writeDone:
			writing = false
			fmt.Printf("config speed %v, current speed %v\n", this.opts.Speed, speed)
		}
	}

	return nil
}
