package kafka

import (
	"dior/lg"
	"github.com/IBM/sarama"
	"log"
)

type ConsumeFunc func(msg *sarama.ConsumerMessage)

// Consumer represents a Sarama consumer group consumer
// implements sarama.ConsumerGroupHandler
type Consumer struct {
	ready   chan bool
	consume ConsumeFunc
}

func (this *Consumer) Prepare(fun ConsumeFunc) {
	this.consume = fun
	this.ready = make(chan bool)
}

func (this *Consumer) WaitReady() {
	<-this.ready
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (this *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(this.ready)
	lg.DftLgr.Warn("Consumer.Setup done.")
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (this *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	lg.DftLgr.Warn("Consumer.Cleanup done.")
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (this *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}
			if lg.DftLgr.Enable(lg.DEBUG) {
				lg.DftLgr.Debug("Consumer.ConsumeClaim Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
			}
			this.consume(message)

			session.MarkMessage(message, "")
		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/IBM/sarama/issues/1192
		case <-session.Context().Done():
			lg.DftLgr.Warn("Consumer.ConsumeClaim session.Context().Done()")
			return nil
		}
	}
}
