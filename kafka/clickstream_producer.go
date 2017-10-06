package kafka

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

type ClickstreamProducer struct {
	producer *kafka.Producer
}

type ClickstreamProducerOptions struct {
	HoustonURL       string
	BootstrapServers string
}

func NewClickstreamProducer(opts *ClickstreamProducerOptions) (*ClickstreamProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": opts.BootstrapServers,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error creating kafka messageHandler")
	}
	cs := &ClickstreamProducer{
		producer: p,
	}
	// Handle messageHandler events in a goroutine
	go cs.handleEvents()
	return cs, nil
}

func (c *ClickstreamProducer) HandleMessage(message *sarama.ConsumerMessage) {

}

func (c *ClickstreamProducer) handleEvents() {
	logger := log.WithField("function", "handleEvents")
	for e := range c.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				logger.Errorf("Delivery failed: %v\n", m.TopicPartition.Error)
			}
			return

		default:
			fmt.Printf("Ignored event: %s\n", ev)
		}
	}
}

func (c *ClickstreamProducer) Close() {
	log.WithField("function", "close").Info("Closing messageHandler")
	c.producer.Close()
}
