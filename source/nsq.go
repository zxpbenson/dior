package source

import (
	"dior/component"
	"dior/lg"
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

func (this *NSQSource) Init() (err error) {
	//this.Asynchronizer.Init()//Reader没起新goroutine不需要调用=
	config := nsq.NewConfig()
	this.consumer, err = nsq.NewConsumer(this.topic, this.channel, config)
	if err != nil {
		lg.DftLgr.Fatal("NSQSource.Init fail : %v", err)
	}
	this.consumer.AddHandler(nsq.HandlerFunc(this.handlerFunc))
	return nil
}

func (this *NSQSource) Start() {
	if len(this.nsqds) > 0 {
		err := this.consumer.ConnectToNSQDs(this.nsqds)
		if err != nil {
			lg.DftLgr.Fatal("NSQSource.Init connect to nsqds fail, nsqds : %v, err : %v", this.nsqds, err)
		}
	} else if len(this.nsqLookupds) > 0 {
		err := this.consumer.ConnectToNSQLookupds(this.nsqLookupds)
		if err != nil {
			lg.DftLgr.Fatal("NSQSource.Init connect to nsqlookupds fail, nsqlookupds : %v, err : %v", this.nsqLookupds, err)
		}
	} else {
		lg.DftLgr.Fatal("NSQSource.Init required nsqds or nsqlookupds")
	}
}

func (this *NSQSource) Stop() {
	this.consumer.Stop()
	//this.Asynchronizer.Stop()//Reader没起新goroutine不需要调用wg.Done
}

func (this *NSQSource) handlerFunc(message *nsq.Message) error {
	this.Channel <- message.Body
	return nil
}
