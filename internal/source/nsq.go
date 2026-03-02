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
		Asynchronizer: component.NewAsynchronizer(),
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
	// 修复：移除多余的return，确保后续代码执行
	if len(s.nsqds) > 0 {
		err := s.consumer.ConnectToNSQDs(s.nsqds)
		if err != nil {
			lg.DftLgr.Error("NSQSource.Start connect to nsqds fail, nsqds: %v, err: %v", s.nsqds, err)
		}
	} else if len(s.nsqLookupds) > 0 {
		err := s.consumer.ConnectToNSQLookupds(s.nsqLookupds)
		if err != nil {
			lg.DftLgr.Error("NSQSource.Start connect to nsqlookupds fail, nsqlookupds: %v, err: %v", s.nsqLookupds, err)
		}
	} else {
		lg.DftLgr.Error("NSQSource.Start requires either nsqds or nsqlookupds")
		return
	}

	// 启动异步处理
	s.Asynchronizer.Start(ctx)
}

func (s *NSQSource) Stop() {
	s.consumer.Stop()
	<-s.consumer.StopChan // 阻塞等待 NSQ 消费者完全停止，确保不再写入 Channel
	//s.Asynchronizer.Stop()//Reader没起新goroutine不需要调用wg.Done
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
