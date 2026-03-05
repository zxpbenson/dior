package source

import (
	"context"
	"dior/component"
	"dior/internal/lg"
	"dior/option"
	"fmt"
	"github.com/nsqio/go-nsq"
)

type NSQSource struct {
	*component.Asynchronizer

	consumer             *nsq.Consumer
	nsqdTCPAddresses     []string
	lookupdHTTPAddresses []string
	topic                string
	channel              string
}

func init() {
	component.RegCmpCreator("nsq-source", newNSQSource)
}

func newNSQSource(name string, opts *option.Options) (component.Component, error) {
	return &NSQSource{
		Asynchronizer:        component.NewAsynchronizer(name),
		nsqdTCPAddresses:     opts.SrcNSQDTCPAddresses,
		lookupdHTTPAddresses: opts.SrcLookupdHTTPAddresses,
		topic:                opts.SrcTopic,
		channel:              opts.SrcChannel,
	}, nil
}

func (s *NSQSource) Init(channel chan []byte) (err error) {
	s.Asynchronizer.Init(channel)
	config := nsq.NewConfig()
	config.MaxInFlight = 5 // 设置最大并发处理数量（控制消费速度）
	s.consumer, err = nsq.NewConsumer(s.topic, s.channel, config)
	if err != nil {
		return err
	}
	s.consumer.AddHandler(nsq.HandlerFunc(s.handlerFunc))
	return nil
}

func (s *NSQSource) handlerFunc(message *nsq.Message) error {
	// 使用select防止channel满时阻塞
	select {
	case s.Channel <- message.Body:
		// 发送成功
		return nil
	default:
		// channel已满，记录警告并重试
		lg.DftLgr.Warn("NSQSource.handlerFunc channel full, message will be requeued")
		// 返回错误让NSQ重新入队
		return fmt.Errorf("channel full, requeue message")
	}
}

func (s *NSQSource) Start(ctx context.Context) {
	s.Asynchronizer.SetState(component.CompStateRunning)

	if len(s.nsqdTCPAddresses) > 0 {
		err := s.consumer.ConnectToNSQDs(s.nsqdTCPAddresses)
		if err != nil {
			lg.DftLgr.Error("NSQSource.Start connect to nsqds fail, nsqds: %v, err: %v", s.nsqdTCPAddresses, err)
		}
		lg.DftLgr.Info("NSQSource.Start connect to nsqds %s", s.nsqdTCPAddresses)
	} else if len(s.lookupdHTTPAddresses) > 0 {
		err := s.consumer.ConnectToNSQLookupds(s.lookupdHTTPAddresses)
		if err != nil {
			lg.DftLgr.Error("NSQSource.Start connect to nsqlookupds fail, nsqlookupds: %v, err: %v", s.lookupdHTTPAddresses, err)
		}
		lg.DftLgr.Info("NSQSource.Start connect to nsqlookupds %s", s.lookupdHTTPAddresses)
	} else {
		lg.DftLgr.Error("NSQSource.Start requires either nsqds or nsqlookupds")
		return
	}

	go s.closeWait(ctx)
}

func (s *NSQSource) closeWait(ctx context.Context) {
	s.Add(1)
	defer func() {
		if err := recover(); err != nil {
			lg.DftLgr.Error("NSQSource.closeWait recover error : %v", err)
		}
		s.Asynchronizer.SetState(component.CompStateStopping)
		s.Asynchronizer.ShowStats()
		s.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			lg.DftLgr.Warn("NSQSource.Start receive quit cmd")
			s.consumer.Stop()
			lg.DftLgr.Warn("NSQSource.Start stop nsq consumer")
			<-s.consumer.StopChan // 阻塞等待 NSQ 消费者完全停止，确保不再写入 Channel
			lg.DftLgr.Warn("NSQSource.Start all nsq consumer routine stopped")
			return
		}
	}
}

func (s *NSQSource) Stop() {
	s.Asynchronizer.Stop()
}
