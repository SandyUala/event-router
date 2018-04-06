package kafka

import (
	"io"

	"github.com/astronomerio/event-router/logging"
	confluent "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	DebugMode        bool
	KafkaDebugMode   bool
}

// NewConsumer creates a new kafka consumer
func NewConsumer(cfg *ConsumerConfig) (*Consumer, error) {
	// Create consumer config map
	cfgMap := &confluent.ConfigMap{
		"bootstrap.servers":        cfg.BootstrapServers,
		"group.id":                 cfg.GroupID,
		"session.timeout.ms":       6000,
		"go.events.channel.enable": true,
		"enable.auto.commit":       true,
		"statistics.interval.ms":   500,
		"default.topic.config":     confluent.ConfigMap{"auto.offset.reset": "earliest"},
		// "go.application.rebalance.enable": true,
	}

	// Set Kafka debugging if in DebugMode
	if cfg.KafkaDebugMode == true {
		cfgMap.SetKey("debug", "protocol,topic,msg")
	}

	// Create the new consumer
	c, err := confluent.NewConsumer(cfgMap)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating consumer")
	}

	// Start the subscription
	if err := c.SubscribeTopics([]string{cfg.Topic}, nil); err != nil {
		return nil, errors.Wrap(err, "Error subscribing to topic ")
	}

	// Create the new Consumer
	consumer := &Consumer{
		config:   cfg,
		consumer: c,
	}

	// Return the new Consumer
	return consumer, nil
}

// Read subscribes to topics and receives messages
func (c *Consumer) Read(d []byte) (int, error) {
	log := logging.GetLogger(logrus.Fields{"package": "kafka"})

	select {
	case <-c.config.ShutdownChannel:
		log.Info("Kafka consumer received shutdown signal")
		return 0, io.EOF
	case ev := <-c.consumer.Events():
		switch e := ev.(type) {
		case *confluent.Message:
			copy(d, e.Value)
			return len(e.Value), nil
		case confluent.Error:
			log.Error(e)
		case confluent.AssignedPartitions, confluent.RevokedPartitions:
			log.Info(e)
		case confluent.OffsetsCommitted, confluent.PartitionEOF:
			log.Debug(e)
			// case *confluent.Stats:
			// 	log.Debug(e)
		}
	}
	return 0, nil
}

// Close cleans up and shutsdown the consumer
func (c *Consumer) Close() {
	log := logging.GetLogger(logrus.Fields{"package": "kafka"})

	c.consumer.Close()
	log.Info("Consumer has been closed")
}
