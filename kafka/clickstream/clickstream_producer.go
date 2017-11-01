package clickstream

import (
	"encoding/json"

	"reflect"

	"fmt"

	"github.com/astronomerio/clickstream-event-router/cassandra"
	"github.com/astronomerio/clickstream-event-router/config"
	"github.com/astronomerio/clickstream-event-router/deadletterqueue"
	"github.com/astronomerio/clickstream-event-router/integrations"
	"github.com/astronomerio/clickstream-event-router/pkg/prom"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type Producer struct {
	producer     *kafka.Producer
	integrations integrations.Integrations
	config       *ProducerConfig
}

type ProducerConfig struct {
	BootstrapServers string
	Integrations     integrations.Integrations
	MessageTimeout   int
	FlushTimeout     int
	Cassandra        *cassandra.Client
	CassandraEnabled bool
	RetryTopic       string
	RetryS3Bucket    string
	S3PathPrefix     string
	MasterTopic      string
	ShutdownChannel  chan struct{}
	DeadletterClient *deadletterqueue.Client
}

type Message struct {
	AppId        string          `json:"appId"`
	MessageID    string          `json:"messageId"`
	Integrations map[string]bool `json:"integrations"`
}

func NewProducer(cfg *ProducerConfig) (*Producer, error) {
	cfgMap := &kafka.ConfigMap{
		"bootstrap.servers":     cfg.BootstrapServers,
		"message.timeout.ms":    cfg.MessageTimeout,
		"request.required.acks": -1,
	}
	if config.GetBool(config.KafakDebug) {
		cfgMap.SetKey("debug", "protocol,topic,msg")
	}
	p, err := kafka.NewProducer(cfgMap)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating kafka messageHandler")
	}
	cs := &Producer{
		producer:     p,
		integrations: cfg.Integrations,
		config:       cfg,
	}
	// Handle messageHandler events in a goroutine
	go cs.handleEvents()
	return cs, nil
}

func (c *Producer) HandleMessage(message []byte, key []byte) error {
	logger := log.WithField("function", "HandleMessage")
	// Get the appId
	dat := &Message{}
	if err := json.Unmarshal(message, &dat); err != nil {
		logger.Error("Error unmarshaling message json: " + err.Error())
		return errors.Wrap(err, "Error unmarshaling json")
	}
	ints, err := c.integrations.GetIntegrations(dat.AppId)
	if err != nil {
		// We had an error connecting to Houston.  Put the message back onto the main
		// topic so we don't lose it
		c.producer.ProduceChannel() <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &c.config.MasterTopic,
				Partition: kafka.PartitionAny,
			},
			Key:   key,
			Value: message,
		}
		return errors.Wrap(err, "Error getting integrations for appID "+dat.AppId)
	}
	prom.BytesConsumed.With(prometheus.Labels{"appId": dat.AppId}).Add(float64(len(message)))
	prom.MessagesConsumed.With(prometheus.Labels{"appId": dat.AppId}).Inc()
	//logger.WithFields(logrus.Fields{"intsLen": len(*ints), "appId": dat.AppId}).Debug("Checking integrations")
	if ints == nil || len(*ints) == 0 {
		// No integrations found
		prom.MessagesWithNoIntegration.With(prometheus.Labels{"appId": dat.AppId}).Inc()
		return nil
	}
	for name, integration := range *ints {
		// If the integration name is in the list of integrations from the message,
		// and the message has it as false, don't send.  If there is no value, don't send
		if ok, intEnabled := dat.Integrations[name]; ok && !intEnabled {
			continue
		}
		c.producer.ProduceChannel() <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &integration,
				Partition: kafka.PartitionAny,
			},
			Key:   key,
			Value: message,
		}
		if c.config.CassandraEnabled {
			// Send the message ID to cassandra
			if err := c.config.Cassandra.InsertMessage(dat.MessageID, integration+"-sent"); err != nil {
				logger.Error(err)
			}
		}
	}
	return nil
}

func (c *Producer) handleEvents() {
	logger := log.WithField("function", "handleEvents")
	logger.Debug("Entered handleEvents")
	run := true
	for run {
		select {
		case <-c.config.ShutdownChannel:
			logger.Info("Producer shutting down")
			c.Close()
			run = false
		case ev := <-c.producer.Events():
			switch e := ev.(type) {
			case *kafka.Message:
				m := e
				if m.TopicPartition.Error != nil {
					logger.Errorf("Delivery failed: %v", m.TopicPartition.Error)
					prom.MessagesProducedFailed.Inc()
					dat := &Message{}
					if err := json.Unmarshal(m.Value, &dat); err != nil {
						logger.Error("Error unmarshaling message json: " + err.Error())
						return
					}
					if config.GetBool(config.Retry) {
						// Send to retry cache
						c.config.DeadletterClient.AddToQueue(&deadletterqueue.QueueObject{
							Key:  fmt.Sprintf("%s/%s", dat.AppId, m.TopicPartition.Topic),
							Data: m.Value,
						})
					}
				} else {
					go func() {
						dat := &Message{}
						if err := json.Unmarshal(m.Value, &dat); err != nil {
							logger.Error("Error unmarshaling message json: " + err.Error())
							return
						}
						prom.MessagesProduced.With(prometheus.Labels{"appId": dat.AppId, "integration": *m.TopicPartition.Topic}).Inc()
						prom.BytesProduced.With(prometheus.Labels{"appId": dat.AppId, "integration": *m.TopicPartition.Topic}).Add(float64(len(m.Value)))
						if c.config.CassandraEnabled {
							// Send the message ID to cassandra
							if err := c.config.Cassandra.InsertMessage(dat.MessageID, *m.TopicPartition.Topic+"-confirmed"); err != nil {
								logger.Error(err)
							}
						}
					}()
				}
			case *kafka.Error:
				logger.Error(e.Error())

			default:
				logger.WithField("type", reflect.TypeOf(e).String()).Errorf("Ignored Event: %v", e)
			}
		}
	}
	logger.Debug("exiting handleEvents")
}

func (c *Producer) Close() {
	logger := log.WithField("function", "Close")
	logger.Info("Closing messageHandler")
	messagesLeftToFlushed := c.producer.Flush(c.config.FlushTimeout)
	if messagesLeftToFlushed != 0 {
		logger.Errorf("Failed to flush %d messages after %d ms", messagesLeftToFlushed, c.config.FlushTimeout)
	} else {
		logger.Info("All messages flushed")
	}
	c.producer.Close()
	if c.config.CassandraEnabled {
		c.config.Cassandra.Close()
	}
	logger.Info("Producer Closed")
}
