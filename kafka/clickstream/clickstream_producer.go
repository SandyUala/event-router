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
	"github.com/sirupsen/logrus"
)

type Producer struct {
	producer     *kafka.Producer
	integrations *integrations.Client
}

type ProducerOptions struct {
	BootstrapServers string
	Integrations     *integrations.Client
}

type Message struct {
	AppId        string          `json:"appId"`
	Integrations map[string]bool `json:"integrations"`
}

func NewProducer(opts *ProducerOptions) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": opts.BootstrapServers,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error creating kafka messageHandler")
	}
	cs := &Producer{
		producer:     p,
		integrations: opts.Integrations,
	}
	// Handle messageHandler events in a goroutine
	go cs.handleEvents()
	return cs, nil
}

func (c *Producer) HandleMessage(message []byte, key []byte) {
	logger := log.WithField("function", "HandleMessage")
	logger.Debug("Entered HandleMessage")
	// Get the appId
	dat := &Message{}
	if err := json.Unmarshal(message, &dat); err != nil {
		logger.Error("Error unmarshaling message json: " + err.Error())
		return
	}
	ints := c.integrations.GetIntegrations(dat.AppId)
	if ints == nil {
		logger.Errorf("No integrations returned for appId %s", dat.AppId)
		return
	}
	for name, integration := range ints {
		// If the integration name is in the list of integrations from the message,
		// and the message has it as false, don't send.  If there is no value, don't send
		if ok, intEnabled := dat.Integrations[name]; ok && !intEnabled {
			continue
		}
		logger.WithFields(logrus.Fields{
			"appId":       dat.AppId,
			"integration": integration,
		}).Debug("Sending message to integration")
		c.producer.ProduceChannel() <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &integration,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(key),
			Value: message,
		}
		go prom.MessagesProduced.With(prometheus.Labels{"appId": dat.AppId, "integration": integration}).Inc()
	}
	logger.Debug("Exiting HandleMessage")
}

func (c *Producer) handleEvents() {
	logger := log.WithField("function", "handleEvents")
	logger.Debug("Entered handleEvents")
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	for run == true {
		select {
		case sig := <-sigchan:
			logger.Infof("Producer caught signal %v: terminating", sig)
			c.Close()
			run = false
		case ev := <-c.producer.Events():
			switch e := ev.(type) {
			case *kafka.Message:
				m := e
				if m.TopicPartition.Error != nil {
					logger.Errorf("Delivery failed: %v", m.TopicPartition.Error)
				}
				return

			default:
				fmt.Printf("Ignored event: %s\n", e)
			}
		}
	}
	logger.Debug("exiting handleEvents")
}

func (c *Producer) Close() {
	log.WithField("function", "close").Info("Closing messageHandler")
	c.producer.Flush(100)
	c.producer.Close()
	log.Info("Producer Closed")
}
