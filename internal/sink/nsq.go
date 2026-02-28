package sink

import (
	"context"
	"dior/component"
	"dior/internal/lg"
	"dior/option"
	"github.com/nsqio/go-nsq"
)

type nsqSink struct {
	*component.Asynchronizer

	producers        []*nsq.Producer
	topic            string
	nsqdTCPAddresses []string
	nsqdLen          int
	nsqdIndex        int
}

func init() {
	component.RegCmpCreator("nsq-sink", newNSQSink)
}

func newNSQSink(opts *option.Options) (component.Component, error) {
	return &nsqSink{
		Asynchronizer:    &component.Asynchronizer{},
		topic:            opts.DstTopic,
		nsqdTCPAddresses: opts.DstNSQDTCPAddresses,
		nsqdLen:          len(opts.DstNSQDTCPAddresses),
		nsqdIndex:        0,
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
	s.producers[s.nsqdIndex].Publish(s.topic, data)
	s.nsqdIndex = (s.nsqdIndex + 1) % s.nsqdLen
}

func (s *nsqSink) Start(ctx context.Context) {
	s.Asynchronizer.Start(ctx)
}

func (s *nsqSink) Stop() {
	s.Asynchronizer.Stop()
	for _, producer := range s.producers {
		producer.Stop()
	}
	lg.DftLgr.Info("nsqSink.Stop done.")
}
