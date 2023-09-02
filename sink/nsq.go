package sink

import (
	"dior/component"
	"dior/lg"
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
	component.RegCmpCreator("nsq-sink", newnsqSink)
}

func newnsqSink(opts *option.Options) (component.Component, error) {
	return &nsqSink{
		Asynchronizer:    &component.Asynchronizer{},
		topic:            opts.DstTopic,
		nsqdTCPAddresses: opts.DstNSQDTCPAddresses,
		nsqdLen:          len(opts.DstNSQDTCPAddresses),
		nsqdIndex:        0,
	}, nil
}

func (this *nsqSink) Init() (err error) {
	this.Asynchronizer.Init()

	cfg := nsq.NewConfig()
	this.nsqdLen = len(this.nsqdTCPAddresses)
	this.producers = make([]*nsq.Producer, this.nsqdLen)
	for idx, nsqdTCPAddress := range this.nsqdTCPAddresses {
		this.producers[idx], err = nsq.NewProducer(nsqdTCPAddress, cfg)
		if err != nil {
			lg.DftLgr.Error("nsqSink.Init Producer open err: %v", err)
			return err
		}
	}

	this.Output = func(data []byte) {
		this.producers[this.nsqdIndex].Publish(this.topic, data)
		this.nsqdIndex += 1
	}

	return nil
}

func (this *nsqSink) Stop() {
	this.Asynchronizer.Stop()
	for _, producer := range this.producers {
		producer.Stop()
	}
	lg.DftLgr.Info("nsqSink.Stop done.")
}
