package kafka

import (
	"encoding/json"

	v1types "github.com/astronomerio/event-api/types/v1"
	"github.com/astronomerio/event-router/config"
	"github.com/astronomerio/event-router/logging"
	confluent "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	DebugMode        bool
}

// NewProducer creates a new Producer
func NewProducer(cfg *ProducerConfig) (*Producer, error) {
	// Create producer config map
	cfgMap := &confluent.ConfigMap{
		"bootstrap.servers":       cfg.BootstrapServers,
		"message.timeout.ms":      cfg.MessageTimeout,
		"request.required.acks":   -1,
		"go.produce.channel.size": 1000,
	}

	// Set Kafka debugging if in DebugMode
	if cfg.DebugMode == true {
		cfgMap.SetKey("debug", "protocol,topic,msg")
	}

	// Create the new producer
	p, err := confluent.NewProducer(cfgMap)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating producer")
	}

	// Create the new Producer
	producer := &Producer{
		config:   cfg,
		producer: p,
	}

	// Fork off to handle kafka events
	go producer.handleEvents()

	// Return the new Producer
	return producer, nil
}

// Write forwards events to destination topics
func (p *Producer) Write(d []byte) (int, error) {
	log := logging.GetLogger(logrus.Fields{"package": "kafka"})

	// Unmarshal to type
	ev := v1types.Event{}
	err := json.Unmarshal(d, &ev)
	if err != nil {
		return 0, err
	}

	integrations := config.IntegrationConfig.EnabledIntegrations(ev.GetWriteKey())
	for _, integration := range integrations {
		name := integration
		log.Infof("Pushing to %s", name)
		p.producer.ProduceChannel() <- &confluent.Message{
			TopicPartition: confluent.TopicPartition{
				Topic:     &name,
				Partition: confluent.PartitionAny,
			},
			Key:   []byte(ev.GetMessageID()),
			Value: d,
		}
	}

	return len(d), nil
}

func (p *Producer) handleEvents() {
	log := logging.GetLogger(logrus.Fields{"package": "kafka"})

	for {
		select {
		case ev := <-p.producer.Events():
			switch e := ev.(type) {
			case *confluent.Message:
				if e.TopicPartition.Error != nil {
					log.Errorf("Delivery failed: %v", e.TopicPartition.Error)
				} else {
					log.Infof("Delivered message to topic %v\n", e.TopicPartition)
				}
			case *confluent.Error:
				log.Error(e.Error())
				// case *confluent.Stats:
			}
		}
	}
}

// Close cleans up and shutsdown the producer
func (p *Producer) Close() {
	log := logging.GetLogger(logrus.Fields{"package": "kafka"})

	p.producer.Close()
	log.Info("Producer has been closed")
}
