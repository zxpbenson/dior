package source

import (
	"context"
	"dior/component"
	"dior/internal/lg"
	"dior/option"
	"github.com/nsqio/go-nsq"
)

type NSQSource struct {
	*component.Asynchronizer

	consumer    *nsq.Consumer
	nsqds       []string
	nsqLookupds []string
	topic       string
	channel     string
}

func init() {
	component.RegCmpCreator("nsq-source", newNSQSource)
}

func newNSQSource(opts *option.Options) (component.Component, error) {
	return &NSQSource{
		Asynchronizer: &component.Asynchronizer{},
		nsqds:         opts.SrcNSQDTCPAddresses,
		nsqLookupds:   opts.SrcLookupdTCPAddresses,
		topic:         opts.SrcTopic,
		channel:       opts.SrcChannel,
	}, nil
}

func (s *NSQSource) Init(channel chan []byte) (err error) {
	s.Asynchronizer.Init(channel)
	config := nsq.NewConfig()
	s.consumer, err = nsq.NewConsumer(s.topic, s.channel, config)
	if err != nil {
		return err
	}
	s.consumer.AddHandler(nsq.HandlerFunc(s.handlerFunc))
	return nil
}

func (s *NSQSource) Start(ctx context.Context) {
	if len(s.nsqds) > 0 {
		err := s.consumer.ConnectToNSQDs(s.nsqds)
		if err != nil {
			lg.DftLgr.Error("NSQSource.Init connect to nsqds fail, nsqds : %v, err : %v", s.nsqds, err)
		}
		return
	}

	if len(s.nsqLookupds) > 0 {
		err := s.consumer.ConnectToNSQLookupds(s.nsqLookupds)
		if err != nil {
			lg.DftLgr.Error("NSQSource.Init connect to nsqlookupds fail, nsqlookupds : %v, err : %v", s.nsqLookupds, err)
		}
		return
	}

	lg.DftLgr.Error("NSQSource.Init required nsqds or nsqlookupds")
}

func (s *NSQSource) Stop() {
	s.consumer.Stop()
	<-s.consumer.StopChan // 阻塞等待 NSQ 消费者完全停止，确保不再写入 Channel
	//s.Asynchronizer.Stop()//Reader没起新goroutine不需要调用wg.Done
}

func (s *NSQSource) handlerFunc(message *nsq.Message) error {
	s.Channel <- message.Body
	return nil
}
