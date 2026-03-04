package source

import (
	"context"
	"dior/component"
	"dior/internal/lg"
	"dior/option"
	"errors"
	"time"

	"github.com/IBM/sarama"
)

type KafkaSource struct {
	*component.Asynchronizer

	client                sarama.ConsumerGroup // init in source.KafkaSource.Init
	kafkaBootstrapServers []string             // init in source.newKafkaSource
	topic                 string               // init in source.newKafkaSource
	group                 string               // init in source.newKafkaSource
}

func init() {
	component.RegCmpCreator("kafka-source", newKafkaSource)
}

func newKafkaSource(name string, opts *option.Options) (component.Component, error) {
	return &KafkaSource{
		Asynchronizer:         component.NewAsynchronizer(name),
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
	s.Add(1)
	s.Asynchronizer.SetState(component.CompStateRunning)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				lg.DftLgr.Error("KafkaSource.Start panic recovered: %v", err)
			}
			s.Asynchronizer.SetState(component.CompStateStopping)
			s.Asynchronizer.ShowStats()
			s.Done()
		}()

		backoff := time.Second
		maxBackoff := 30 * time.Second
		consecutiveErrors := 0
		maxConsecutiveErrors := 10

		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := s.client.Consume(ctx, []string{s.topic}, s); err != nil {
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
		}
	}()
}

func (s *KafkaSource) Stop() {
	if err := s.client.Close(); err != nil {
		lg.DftLgr.Warn("KafkaSource.Stop Error closing client: %v", err)
	}
	s.Asynchronizer.Stop()
	lg.DftLgr.Info("KafkaSource.Stop done.")
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (s *KafkaSource) Setup(sarama.ConsumerGroupSession) error {
	lg.DftLgr.Warn("KafkaSource.Setup done.")
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (s *KafkaSource) Cleanup(sarama.ConsumerGroupSession) error {
	lg.DftLgr.Warn("KafkaSource.Cleanup done.")
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (s *KafkaSource) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				lg.DftLgr.Warn("message channel was closed")
				return nil
			}
			if lg.DftLgr.Enable(lg.DEBUG) {
				lg.DftLgr.Debug("KafkaSource.ConsumeClaim Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
			}
			// 只有在消息成功发送到 channel 时才提交偏移量
			select {
			case s.Channel <- message.Value:
				session.MarkMessage(message, "")
				s.Asynchronizer.AddProcessedCount(1)
			default:
				// channel已满，记录警告但不提交偏移量，Kafka会重新投递
				lg.DftLgr.Warn("KafkaSource.consume channel full, message will be redelivered (topic: %s, partition: %d, offset: %d)",
					message.Topic, message.Partition, message.Offset)
			}
		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/IBM/sarama/issues/1192
		case <-session.Context().Done():
			lg.DftLgr.Warn("KafkaSource.ConsumeClaim session.Context().Done()")
			return nil
		}
	}
}
