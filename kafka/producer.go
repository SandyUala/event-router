package kafka

import (
	"encoding/json"

	v1types "github.com/astronomerio/event-api/types/v1"
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
	ShutdownChannel  chan struct{}
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

	return &Producer{
		config:   cfg,
		producer: p,
	}, nil
}

// Write forwards events to destination topics
func (p *Producer) Write(d []byte) (int, error) {
	log := logging.GetLogger(logrus.Fields{"package": "api"})

	// Unmarshal to type
	ev := v1types.Event{}
	err := json.Unmarshal(d, &ev)
	if err != nil {
		return 0, err
	}

	log.Infof("%+v", ev.GetWriteKey())
	return len(d), nil
}
