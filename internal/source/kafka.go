package source

import (
	"context"
	"dior/component"
	"dior/internal/kafka"
	"dior/internal/lg"
	"dior/option"
	"errors"
	"github.com/IBM/sarama"
)

type KafkaSource struct {
	*component.Asynchronizer

	consumer              *kafka.Consumer      // init in source.newKafkaSource
	client                sarama.ConsumerGroup // init in source.KafkaSource.Init
	kafkaBootstrapServers []string             // init in source.newKafkaSource
	topic                 string               // init in source.newKafkaSource
	group                 string               // init in source.newKafkaSource

	ctx    context.Context    // init in source.KafkaSource.Start
	cancel context.CancelFunc // init in source.KafkaSource.Start
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

func (s *KafkaSource) Init(channel chan []byte) (err error) {
	s.Asynchronizer.Init(channel)

	s.consumer.Prepare(s.consume)

	version, err := sarama.ParseKafkaVersion("2.1.1")
	if err != nil {
		lg.DftLgr.Error("Error parsing Kafka version: %v", err)
		return err
	}

	config := sarama.NewConfig()
	config.Version = version
	//config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	//config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	s.client, err = sarama.NewConsumerGroup(s.kafkaBootstrapServers, s.group, config)
	if err != nil {
		// Logged by caller if Init returns error
		return err
	}
	return nil
}

func (s *KafkaSource) Start(ctx context.Context) {

	s.ctx, s.cancel = context.WithCancel(ctx)

	s.Add(1)
	go func() {
		defer s.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := s.client.Consume(s.ctx, []string{s.topic}, s.consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				lg.DftLgr.Error("KafkaSource.Start nameless routine Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if s.ctx.Err() != nil {
				return
			}
		}
	}()

	s.consumer.WaitReady(s.ctx)
}

func (s *KafkaSource) Stop() {
	s.cancel()

	if err := s.client.Close(); err != nil {
		lg.DftLgr.Warn("KafkaSource.Stop Error closing client: %v", err)
	}
	lg.DftLgr.Info("KafkaSource.Stop done.")
}

func (s *KafkaSource) consume(msg *sarama.ConsumerMessage) {
	s.Channel <- msg.Value
}
