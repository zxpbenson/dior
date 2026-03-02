package sink

import (
	"context"
	"dior/component"
	"dior/internal/lg"
	"dior/option"
	"github.com/IBM/sarama"
	"time"
)

type kafkaSink struct {
	*component.Asynchronizer

	producer              sarama.SyncProducer
	kafkaBootstrapServers []string
	topic                 string
}

func init() {
	component.RegCmpCreator("kafka-sink", newkafkaSink)
}

func newkafkaSink(name string, opts *option.Options) (component.Component, error) {
	return &kafkaSink{
		Asynchronizer:         component.NewAsynchronizer(name),
		kafkaBootstrapServers: opts.DstBootstrapServers,
		topic:                 opts.DstTopic,
	}, nil
}

func (s *kafkaSink) Init(channel chan []byte) (err error) {
	s.Asynchronizer.Init(channel)

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follower都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner //写到随机分区中，我们默认设置32个分区
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回

	// 连接kafka
	s.producer, err = sarama.NewSyncProducer(s.kafkaBootstrapServers, config)
	if err != nil {
		return err
	}

	s.Output = s.produce
	return nil
}

func (s *kafkaSink) produce(data []byte) {
	// 构造一个消息
	msg := &sarama.ProducerMessage{}
	msg.Topic = s.topic
	msg.Value = sarama.ByteEncoder(data)

	// 发送消息，带重试机制
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		partitionId, offset, err := s.producer.SendMessage(msg)
		if err == nil {
			lg.DftLgr.Debug("kafkaSink.produce send msg ok, pid: %v offset: %v", partitionId, offset)
			return
		}

		lastErr = err
		lg.DftLgr.Warn("kafkaSink.produce send msg failed (attempt %d/%d): %v", attempt, maxRetries, err)

		// 最后一次重试不需要等待
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
		}
	}

	// 所有重试都失败
	lg.DftLgr.Error("kafkaSink.produce send msg failed after %d attempts, last error: %v", maxRetries, lastErr)
}

func (s *kafkaSink) Start(ctx context.Context) {
	s.Asynchronizer.Start(ctx)
}

func (s *kafkaSink) Stop() {
	s.producer.Close()
	s.Asynchronizer.Stop()
	lg.DftLgr.Info("kafkaSink.Stop done.")
}
