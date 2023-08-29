package writer

import (
	"dior/lg"
	"dior/option"
	"github.com/nsqio/go-nsq"
)

type nsqWriter struct {
	opts   *option.Options
	client *nsq.Producer
	logger lg.Logger
}

func init() {
	RegWriteCreator("nsq", newNSQWriter)
}

func newNSQWriter(opts *option.Options) (WriteAble, error) {
	logger, err := lg.NewLogger(opts.LogPrefix, opts.LogLevel)
	if err != nil {
		return nil, err
	}
	return &nsqWriter{
		opts:   opts,
		logger: logger,
	}, nil
}

func (this *nsqWriter) Open() (err error) {
	cfg := nsq.NewConfig()

	this.client, err = nsq.NewProducer(this.opts.NSQDTCPAddresses, cfg)
	if err != nil {
		this.logger.Error("Producer open err: %v", err)
		return err
	}
	return nil
}

func (this *nsqWriter) Close() error {
	this.client.Stop()
	return nil
}

func (this *nsqWriter) Write(data string) error {
	this.logger.Debug("write to nsq : %v", data)
	return this.client.Publish(this.opts.Topic, []byte(data))
}
