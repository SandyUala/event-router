package clickstream

import (
	"github.com/astronomerio/cs-event-router/kafka"
	"github.com/astronomerio/cs-event-router/pkg/prom"
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
	Topic            string
	MessageHandler   kafka.MessageHandler
	ShutdownChannel  chan struct{}
}

func NewConsumer(opts *ConsumerOptions) (*Consumer, error) {
	consumer, err := confluent.NewConsumer(&confluent.ConfigMap{
		"bootstrap.servers":               opts.BootstrapServers,
		"group.id":                        opts.GroupID,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.auto.commit":              true,
		"statistics.interval.ms":          500,
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
		"Topics":            c.options.Topic,
	}).Debug("Consumer Options")

	err := c.consumer.SubscribeTopics([]string{c.options.Topic}, nil)
	if err != nil {
		logger.Error(err)
		return
	}

	run := true
	for run {
		select {
		case <-c.options.ShutdownChannel:
			logger.Info("Consumer shutting down")
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
				err := c.messageHandler.HandleMessage(e.Value, e.Key)
				// If we receive an error its because we had an issue connecting to houston
				if err != nil {
					logger.Error(err)
					break
				}
			//case confluent.PartitionEOF:
			//	logger.Infof("Reached %v", e)
			case confluent.Error:
				logger.Error(e.Error())
			case *confluent.Stats:
				go prom.HandleKafkaStats(e, "consumer")
				//default:
				//	logger.WithField("type", reflect.TypeOf(e).String()).Errorf("Consumer Ignored Event: %v", e)
			}
		}
	}
	c.Close()
}

func (c *Consumer) Close() {
	c.consumer.Close()
	log.Info("Consumer Closed")
}
