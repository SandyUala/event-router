package clickstream

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"encoding/json"

	"github.com/astronomerio/event-router/integrations"
	"github.com/astronomerio/event-router/pkg/prom"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type Producer struct {
	producer     *kafka.Producer
	integrations *integrations.Client
}

type ProducerOptions struct {
	BootstrapServers []string
}

type Message struct {
	appId        string          `json:"appId"`
	integrations map[string]bool `json:"integrations"`
}

func NewProducer(opts *ProducerOptions) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": opts.BootstrapServers,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error creating kafka messageHandler")
	}
	cs := &Producer{
		producer: p,
	}
	// Handle messageHandler events in a goroutine
	go cs.handleEvents()
	return cs, nil
}

func (c *Producer) HandleMessage(message []byte, key []byte) {
	logger := log.WithField("function", "HandleMessage")
	// Get the appId
	dat := &Message{}
	if err := json.Unmarshal(message, &dat); err != nil {
		logger.Error("Error unmarshaling message json: " + err.Error())
		return
	}
	ints := c.integrations.GetIntegrations(dat.appId)
	if ints == nil {
		logger.Errorf("No integrations returned for appId %s", dat.appId)
		return
	}
	for _, integration := range ints {
		c.producer.ProduceChannel() <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &integration,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(key),
			Value: message,
		}
		go prom.MessagesProduced.With(prometheus.Labels{"appId": dat.appId, "integration": integration}).Inc()
	}
}

func (c *Producer) handleEvents() {
	logger := log.WithField("function", "handleEvents")
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for true {
			select {
			case sig := <-sigchan:
				logger.Infof("Producer caught signal %v: terminating\n", sig)
				c.Close()
			}
		}
	}()

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

func (c *Producer) Close() {
	log.WithField("function", "close").Info("Closing messageHandler")
	c.producer.Flush(1000)
	c.producer.Close()
}
