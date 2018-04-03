package kafka

import (
	"github.com/astronomerio/event-router/logging"
	confluent "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	log = logging.GetLogger().WithFields(logrus.Fields{"package": "kafka"})
)

// Consumer is a kafka consumer
type Consumer struct {
	config   *ConsumerConfig
	consumer *confluent.Consumer
}

// ConsumerConfig is a kafka consumer configuration
type ConsumerConfig struct {
	BootstrapServers string
	GroupID          string
	Topic            string
	ShutdownChannel  chan struct{}
}

// NewConsumer creates a new kafka consumer
func NewConsumer(cfg *ConsumerConfig) (*Consumer, error) {
	consumer, err := confluent.NewConsumer(&confluent.ConfigMap{
		"bootstrap.servers":               cfg.BootstrapServers,
		"group.id":                        cfg.GroupID,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.auto.commit":              true,
		"statistics.interval.ms":          500,
		"default.topic.config":            confluent.ConfigMap{"auto.offset.reset": "earliest"},
	})

	if err != nil {
		return nil, errors.Wrap(err, "Error creating consumer")
	}

	return &Consumer{
		config:   cfg,
		consumer: consumer,
	}, nil
}
