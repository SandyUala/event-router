package clickstream

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/astronomerio/clickstream-event-router/kafka"
	"github.com/astronomerio/clickstream-event-router/pkg/prom"
	confluent "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("package", "clickstream")
)

type Consumer struct {
	options        *ConsumerOptions
	messageHandler kafka.MessageHandler
	consumer       *confluent.Consumer
}

type ConsumerOptions struct {
	BootstrapServers string
	GroupID          string
	Topics           []string
	MessageHandler   kafka.MessageHandler
}

func NewConsumer(opts *ConsumerOptions) (*Consumer, error) {
	consumer, err := confluent.NewConsumer(&confluent.ConfigMap{
		"bootstrap.servers":               opts.BootstrapServers,
		"group.id":                        opts.GroupID,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.auto.commit":              true,
		"default.topic.config":            confluent.ConfigMap{"auto.offset.reset": "earliest"}})

	if err != nil {
		return nil, errors.Wrap(err, "Error creating consumer")
	}
	return &Consumer{
		options:        opts,
		messageHandler: opts.MessageHandler,
		consumer:       consumer,
	}, nil
}

func (c *Consumer) Run() {
	logger := log.WithField("function", "Run")
	logger.Info("Starting Kafka Consumer")
	logger.WithFields(logrus.Fields{
		"Bootstrap Servers": c.options.BootstrapServers,
		"GroupID":           c.options.GroupID,
		"Topics":            c.options.Topics,
	}).Debug("Consumer Options")

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	err := c.consumer.SubscribeTopics(c.options.Topics, nil)
	if err != nil {
		logger.Error(err)
		return
	}

	run := true
	for run {
		select {
		case sig := <-sigchan:
			logger.Infof("Consumer caught signal %v: terminating", sig)
			run = false

		case ev := <-c.consumer.Events():
			switch e := ev.(type) {
			case confluent.AssignedPartitions:
				logger.Infof("Assigning Partition %v", e)
				c.consumer.Assign(e.Partitions)
			case confluent.RevokedPartitions:
				logger.Infof("Revoking Partition %v", e)
				c.consumer.Unassign()
			case *confluent.Message:
				go c.messageHandler.HandleMessage(e.Value, e.Key)
				prom.MessagesConsumed.Inc()
			case confluent.PartitionEOF:
				logger.Infof("Reached %v", e)
			case confluent.Error:
				logger.Error(e.Error())
			}
		}
	}

	c.consumer.Close()
	logger.Info("Consumer Closed")
}
