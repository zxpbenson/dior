package writer

import (
	"dior/option"
	"fmt"
	"github.com/nsqio/go-nsq"
)

type nsqWriter struct {
	opts   *option.Options
	client *nsq.Producer
}

func init() {
	RegWriteCreator("nsq", newNSQWriter)
}

func newNSQWriter(opts *option.Options) WriteAble {
	return &nsqWriter{
		opts: opts,
	}
}

func (this *nsqWriter) Open() (err error) {
	cfg := nsq.NewConfig()

	this.client, err = nsq.NewProducer(this.opts.NSQDTCPAddresses, cfg)
	if err != nil {
		fmt.Printf("Producer open err: %v\n", err)
		return err
	}
	return nil
}

func (this *nsqWriter) Close() error {
	this.client.Stop()
	return nil
}

func (this *nsqWriter) Write(data string) error {
	//fmt.Printf("write to nsq\n")
	return this.client.Publish(this.opts.Topic, []byte(data))
}
