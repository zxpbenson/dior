package source

import (
	"context"
	"dior/component"
	"dior/internal/kafka"
	"dior/internal/lg"
	"dior/option"
	"errors"
	"time"

	"github.com/IBM/sarama"
)

type KafkaSource struct {
	*component.Asynchronizer

	consumer              *kafka.Consumer      // init in source.newKafkaSource
	client                sarama.ConsumerGroup // init in source.KafkaSource.Init
	kafkaBootstrapServers []string             // init in source.newKafkaSource
	topic                 string               // init in source.newKafkaSource
	group                 string               // init in source.newKafkaSource

	// cancel 用于取消消费循环，不保存context
	cancel context.CancelFunc // init in source.KafkaSource.Start
}

func init() {
	component.RegCmpCreator("kafka-source", newKafkaSource)
}

func newKafkaSource(name string, opts *option.Options) (component.Component, error) {
	return &KafkaSource{
		Asynchronizer:         component.NewAsynchronizer(name),
		consumer:              &kafka.Consumer{},
		kafkaBootstrapServers: opts.SrcBootstrapServers,
		topic:                 opts.SrcTopic,
		group:                 opts.SrcGroup,
	}, nil
}

func (s *KafkaSource) Init(channel chan []byte) (err error) {
	s.Asynchronizer.Init(channel)

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
	// 创建可取消的子上下文
	// 注意：cancel保存在结构体中是因为Stop()需要调用它来取消消费循环
	// 这是Go官方允许的模式：保存cancel但不保存context
	ctx, s.cancel = context.WithCancel(ctx)

	s.Add(1)
	s.Asynchronizer.SetState(component.CompStateRunning)
	s.consumer.Prepare(s.consume)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				lg.DftLgr.Error("KafkaSource.Start panic recovered: %v", err)
			}
			s.Asynchronizer.SetState(component.CompStateStopped)
			s.Done()
			lg.DftLgr.Info("KafkaSource.Start goroutine exited")
		}()

		backoff := time.Second
		maxBackoff := 30 * time.Second
		consecutiveErrors := 0
		maxConsecutiveErrors := 10

		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := s.client.Consume(ctx, []string{s.topic}, s.consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					lg.DftLgr.Info("KafkaSource.Start consumer group closed normally")
					return
				}

				consecutiveErrors++
				lg.DftLgr.Error("KafkaSource.Start consume error (attempt %d/%d): %v, retrying after %v",
					consecutiveErrors, maxConsecutiveErrors, err, backoff)

				// 连续错误过多时退出
				if consecutiveErrors >= maxConsecutiveErrors {
					lg.DftLgr.Error("KafkaSource.Start too many consecutive errors, exiting")
					return
				}

				select {
				case <-ctx.Done():
					lg.DftLgr.Info("KafkaSource.Start context cancelled during backoff")
					return
				case <-time.After(backoff):
				}

				// 指数退避
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			} else {
				// 成功消费后重置退避和错误计数
				backoff = time.Second
				consecutiveErrors = 0
			}

			// 检查上下文是否已取消
			if ctx.Err() != nil {
				lg.DftLgr.Info("KafkaSource.Start context cancelled, exiting")
				return
			}
			s.consumer.ResetReady()
		}
	}()
}

func (s *KafkaSource) Stop() {
	s.cancel()

	if err := s.client.Close(); err != nil {
		lg.DftLgr.Warn("KafkaSource.Stop Error closing client: %v", err)
	}
	s.Asynchronizer.Stop()
	lg.DftLgr.Info("KafkaSource.Stop done.")
}

func (s *KafkaSource) consume(msg *sarama.ConsumerMessage) {
	// 使用select防止channel满时阻塞
	select {
	case s.Channel <- msg.Value:
		// 发送成功
	default:
		// channel已满，记录警告并丢弃消息
		lg.DftLgr.Warn("KafkaSource.consume channel full, dropping message (topic: %s, partition: %d, offset: %d)",
			msg.Topic, msg.Partition, msg.Offset)
	}
}
