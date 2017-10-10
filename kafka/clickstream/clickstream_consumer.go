package clickstream

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/astronomerio/event-router/kafka"
	"github.com/astronomerio/event-router/pkg/prom"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("package", "kafka")
)

type Consumer struct {
	config         *cluster.Config
	options        *ConsumerOptions
	messageHandler kafka.MessageHandler
	consumer       *cluster.Consumer
}

type ConsumerOptions struct {
	BootstrapServers []string
	GroupID          string
	Topics           []string
	MessageHandler   kafka.MessageHandler
}

func NewConsumer(opts *ConsumerOptions) (*Consumer, error) {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true

	return &Consumer{
		config:         config,
		options:        opts,
		messageHandler: opts.MessageHandler,
	}, nil
}

func (c *Consumer) Run() {
	logger := log.WithField("function", "Run")
	logger.Info("Starting Kafka Consumer")
	consumer, err := cluster.NewConsumer(c.options.BootstrapServers, c.options.GroupID, c.options.Topics, c.config)
	if err != nil {
		logger.Error(err)
		return
	}
	c.consumer = consumer
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for ntf := range consumer.Notifications() {
			log.Info("Notification: %+v\n", ntf)
		}
	}()

	go func() {
		for err := range consumer.Errors() {
			logger.Errorf("Error in kafka consumer: %+v\n", err)
		}
	}()

	run := true
	for run == true {
		select {
		case sig := <-sigchan:
			log.Info("Caught signal %v: terminating\n", sig)
			run = false
		case msg := <-consumer.Messages():
			go func() {
				c.messageHandler.HandleMessage(msg.Value, msg.Key)
				consumer.MarkOffset(msg, "")
				prom.MessagesConsumed.Inc()
			}()
		}
	}

	c.Close()
}

func (c *Consumer) Close() {
	logger := log.WithField("function", "Close")
	logger.Warn("Closing Consumer")
	if c.consumer == nil {
		logger.Error("Consumer not running, close should not have been called!")
		return
	}
	if err := c.consumer.CommitOffsets(); err != nil {
		logger.WithField("error", err).Error("Error commiting offsets")
	}
	c.consumer.Close()
	c.messageHandler.Close()
}
