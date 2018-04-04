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
	DebugMode        bool
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
	if cfg.DebugMode == true {
		cfgMap.SetKey("debug", "protocol,topic,msg")
	}

	// Create the new consumer
	consumer, err := confluent.NewConsumer(cfgMap)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating consumer")
	}

	return &Consumer{
		config:   cfg,
		consumer: consumer,
	}, nil
}

// Run subscribes to topics and receives messages
func (c *Consumer) Run() {
	log.Info("Starting Kafka consumer")

	// Start the subscription
	if err := c.consumer.SubscribeTopics([]string{c.config.Topic}, nil); err != nil {
		log.Fatal("Error subscribing to topic ", err)
	}

	// Close consumer when we exit
	defer c.Close()

	// Loop until we are notified on the ShutdownChannel
	for {
		select {
		case <-c.config.ShutdownChannel:
			log.Info("Kafka consumer shutting down")
			break
		case ev := <-c.consumer.Events():
			switch e := ev.(type) {
			case *confluent.Message:
				c.config.MessageHandler.HandleMessage(e.Value, e.Key)
			}
		}
	}
}

// Close cleans up and shutsdown the consumer
func (c *Consumer) Close() {
	c.consumer.Close()
	log.info("Consumer has been closed")
}
