package writer

import (
	"dior/option"
	"fmt"

	"github.com/IBM/sarama"
)

type kafkaWriter struct {
	opts   *option.Options
	client sarama.SyncProducer
}

func init() {
	RegWriteCreator("kafka", newKafkaWriter)
}

func newKafkaWriter(opts *option.Options) WriteAble {
	return &kafkaWriter{
		opts: opts,
	}
}

func (this *kafkaWriter) Open() (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follower都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner //写到随机分区中，我们默认设置32个分区
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回

	// 连接kafka
	this.client, err = sarama.NewSyncProducer(this.opts.KafkaBootstrapServer, config)
	if err != nil {
		fmt.Printf("Producer open err : %v\n", err)
		return err
	}
	return nil
}

func (this *kafkaWriter) Close() error {
	return this.client.Close()
}

func (this *kafkaWriter) Write(data string) error {
	//fmt.Printf("write to kafka\n")

	// 构造一个消息
	msg := &sarama.ProducerMessage{}
	msg.Topic = this.opts.Topic
	msg.Value = sarama.StringEncoder(data)

	// 发送消息
	//partitionId, offset, err := this.client.SendMessage(msg)
	_, _, err := this.client.SendMessage(msg)
	if err != nil {
		fmt.Printf("send msg failed, err : %v\n", err)
		return err
	}
	//fmt.Printf("pid:%v offset:%v\n", partitionId, offset)
	return nil
}
