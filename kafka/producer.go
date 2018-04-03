package kafka

import (
	confluent "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

// Producer is a kafka producer
type Producer struct {
	config   *ProducerConfig
	producer *confluent.Producer
}

// ProducerConfig is a kafka producer config
type ProducerConfig struct {
	BootstrapServers string
	MessageTimeout   int
}

// NewProducer creates a new Producer
func NewProducer(cfg *ProducerConfig) (*Producer, error) {
	producer, err := confluent.NewProducer(&confluent.ConfigMap{
		"bootstrap.servers":       cfg.BootstrapServers,
		"message.timeout.ms":      cfg.MessageTimeout,
		"request.required.acks":   -1,
		"go.produce.channel.size": 1000,
	})

	if err != nil {
		return nil, errors.Wrap(err, "Error creating producer")
	}

	return &Producer{
		config:   cfg,
		producer: producer,
	}, nil
}
