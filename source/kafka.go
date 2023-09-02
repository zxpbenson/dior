package source

import (
	"context"
	"dior/component"
	"dior/kafka"
	"dior/lg"
	"dior/option"
	"errors"
	"github.com/IBM/sarama"
	"log"
)

type KafkaSource struct {
	*component.Asynchronizer

	consumer              *kafka.Consumer
	client                sarama.ConsumerGroup
	kafkaBootstrapServers []string
	topic                 string
	group                 string

	ctx    context.Context
	cancel context.CancelFunc
}

func init() {
	component.RegCmpCreator("kafka-source", newKafkaSource)
}

func newKafkaSource(opts *option.Options) (component.Component, error) {
	return &KafkaSource{
		Asynchronizer:         &component.Asynchronizer{},
		consumer:              &kafka.Consumer{},
		kafkaBootstrapServers: opts.SrcBootstrapServers,
		topic:                 opts.SrcTopic,
		group:                 opts.SrcGroup,
	}, nil
}

func (this *KafkaSource) Init() (err error) {
	this.Asynchronizer.Init()

	this.consumer.Prepare(this.consume)

	version, err := sarama.ParseKafkaVersion("2.1.1")
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	config := sarama.NewConfig()
	config.Version = version
	//config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	//config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	this.client, err = sarama.NewConsumerGroup(this.kafkaBootstrapServers, this.group, config)
	if err != nil {
		lg.DftLgr.Error("KafkaSource.Init fail to start consumer, err:%v\n", err)
		return
	}
	return
}

func (this *KafkaSource) Start() {

	this.ctx, this.cancel = context.WithCancel(context.Background())

	this.Add(1)
	go func() {
		defer this.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := this.client.Consume(this.ctx, []string{this.topic}, this.consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				lg.DftLgr.Error("KafkaSource.Start nameless routine Error from consumer: %v", err)
			}
			// check if context w
			//as cancelled, signaling that the consumer should stop
			if this.ctx.Err() != nil {
				return
			}
			this.consumer.Prepare(this.consume)
		}
	}()

	this.consumer.WaitReady()
}

func (this *KafkaSource) Stop() {
	this.cancel()

	if err := this.client.Close(); err != nil {
		lg.DftLgr.Error("KafkaSource.Stop Error closing client: %v", err)
	}
}

func (this *KafkaSource) consume(msg *sarama.ConsumerMessage) {
	this.Channel <- msg.Value
}
