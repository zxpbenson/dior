package sink

import (
	"context"
	"dior/component"
	"dior/internal/lg"
	"dior/option"
	"sync/atomic"

	"github.com/nsqio/go-nsq"
)

type nsqSink struct {
	*component.Asynchronizer

	producers        []*nsq.Producer
	topic            string
	nsqdTCPAddresses []string
	nsqdLen          int
	nsqdIndex        atomic.Int64 // 使用atomic保证并发安全
}

func init() {
	component.RegCmpCreator("nsq-sink", newNSQSink)
}

func newNSQSink(name string, opts *option.Options) (component.Component, error) {
	return &nsqSink{
		Asynchronizer:    component.NewAsynchronizer(name),
		topic:            opts.DstTopic,
		nsqdTCPAddresses: opts.DstNSQDTCPAddresses,
		nsqdLen:          len(opts.DstNSQDTCPAddresses),
	}, nil
}

func (s *nsqSink) Init(channel chan []byte) (err error) {
	s.Asynchronizer.Init(channel)

	cfg := nsq.NewConfig()
	s.nsqdLen = len(s.nsqdTCPAddresses)
	s.producers = make([]*nsq.Producer, s.nsqdLen)
	for idx, nsqdTCPAddress := range s.nsqdTCPAddresses {
		s.producers[idx], err = nsq.NewProducer(nsqdTCPAddress, cfg)
		if err != nil {
			return err
		}
	}
	s.Output = s.output
	return nil
}

func (s *nsqSink) output(data []byte) {
	// 轮询选择producer（原子操作保证并发安全）
	index := s.nsqdIndex.Add(1) - 1
	producer := s.producers[index%int64(s.nsqdLen)]

	// 发送消息
	err := producer.Publish(s.topic, data)
	if err != nil {
		lg.DftLgr.Error("nsqSink.output publish failed, topic: %s, error: %v", s.topic, err)
		return
	}
	lg.DftLgr.Debug("nsqSink.output publish ok, topic: %s", s.topic)
}

func (s *nsqSink) Start(ctx context.Context) {
	s.Asynchronizer.Start(ctx)
}

func (s *nsqSink) Stop() {
	for _, producer := range s.producers {
		producer.Stop()
	}
	s.Asynchronizer.Stop()
	lg.DftLgr.Info("nsqSink.Stop done.")
}
