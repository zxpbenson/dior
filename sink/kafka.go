package sink

import (
	"dior/component"
	"dior/lg"
	"dior/option"
	"github.com/IBM/sarama"
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

func newkafkaSink(opts *option.Options) (component.Component, error) {
	return &kafkaSink{
		Asynchronizer:         &component.Asynchronizer{},
		kafkaBootstrapServers: opts.DstBootstrapServers,
		topic:                 opts.DstTopic,
	}, nil
}

func (this *kafkaSink) Init() (err error) {
	this.Asynchronizer.Init()

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follower都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner //写到随机分区中，我们默认设置32个分区
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回

	// 连接kafka
	this.producer, err = sarama.NewSyncProducer(this.kafkaBootstrapServers, config)
	if err != nil {
		//lg.DftLgr.Error("kafkaSink init err : %v", err)
		return err
	}

	this.Output = this.produce
	return nil
}

func (this *kafkaSink) Stop() {
	this.Asynchronizer.Stop()
	this.producer.Close()
	lg.DftLgr.Info("kafkaSink.Stop done.")
}

func (this *kafkaSink) produce(data []byte) {
	// 构造一个消息
	msg := &sarama.ProducerMessage{}
	msg.Topic = this.topic
	msg.Value = sarama.ByteEncoder(data)

	// 发送消息
	partitionId, offset, err := this.producer.SendMessage(msg)
	if err != nil {
		lg.DftLgr.Error("kafkaSink.produce send msg failed, err : %v", err)
		return
	}
	lg.DftLgr.Debug("kafkaSink.produce send msg ok, pid : %v offset : %v", partitionId, offset)
}
