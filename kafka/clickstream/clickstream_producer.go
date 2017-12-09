package clickstream

import (
	"reflect"

	"fmt"

	"github.com/astronomerio/event-router/cassandra"
	"github.com/astronomerio/event-router/config"
	"github.com/astronomerio/event-router/deadletterqueue"
	"github.com/astronomerio/event-router/integrations"
	"github.com/astronomerio/event-router/pkg/prom"
	confluent "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type Producer struct {
	producer     *confluent.Producer
	integrations integrations.Integrations
	config       *ProducerConfig
}

type ProducerConfig struct {
	BootstrapServers string
	GroupID          string
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

//easyjson:json
type Message struct {
	AppId        string          `json:"appId"`
	MessageID    string          `json:"messageId"`
	Integrations map[string]bool `json:"integrations"`
}

// NewProducer returns a new Kafka Producer setup with the given ProducerConfig
func NewProducer(cfg *ProducerConfig) (*Producer, error) {
	cfgMap := &confluent.ConfigMap{
		"bootstrap.servers":       cfg.BootstrapServers,
		"message.timeout.ms":      cfg.MessageTimeout,
		"request.required.acks":   -1,
		"go.produce.channel.size": 1000,
	}
	if config.GetBool(config.KafakDebug) {
		cfgMap.SetKey("debug", "protocol,topic,msg")
	}
	p, err := confluent.NewProducer(cfgMap)
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

// HandleMessage takes in a Kafka Message and Key as byte slices.
func (c *Producer) HandleMessage(message []byte, key []byte) error {
	logger := log.WithField("function", "HandleMessage")
	// Get the appId
	dat := &Message{}
	if err := dat.UnmarshalJSON(message); err != nil {
		//if err := json.Unmarshal(message, &dat); err != nil {
		logger.Error("Error unmarshaling message json: " + err.Error())
		return errors.Wrap(err, "Error unmarshaling json")
	}
	prom.BytesConsumed.With(prometheus.Labels{"appId": dat.AppId}).Add(float64(len(message)))
	prom.MessagesConsumed.With(prometheus.Labels{"appId": dat.AppId}).Inc()
	// Get enabled integrations for the app id
	integrations, err := c.integrations.GetIntegrations(dat.AppId)
	if err != nil {
		// We had an error connecting to Houston.  Send the message to the retry topic
		c.producer.ProduceChannel() <- &confluent.Message{
			TopicPartition: confluent.TopicPartition{
				Topic:     &c.config.RetryTopic,
				Partition: confluent.PartitionAny,
			},
			Key:   key,
			Value: message,
		}
		return errors.Wrap(err, "Error getting integrations for appID "+dat.AppId)
	}
	//logger.WithFields(logrus.Fields{"intsLen": len(*integrations), "appId": dat.AppId}).Debug("Checking integrations")
	if integrations == nil || len(*integrations) == 0 {
		// No integrations found
		prom.MessagesWithNoIntegration.With(prometheus.Labels{"appId": dat.AppId}).Inc()
		return nil
	}
	for name, integration := range *integrations {
		// If the integration name is in the list of integrations from the message,
		// and the message has it as false, don't send.  If there is no value, don't send
		if ok, intEnabled := dat.Integrations[name]; ok && !intEnabled {
			continue
		}
		c.producer.ProduceChannel() <- &confluent.Message{
			TopicPartition: confluent.TopicPartition{
				Topic:     &integration,
				Partition: confluent.PartitionAny,
			},
			Key:   key,
			Value: message,
		}
		prom.MessagesProduced.With(prometheus.Labels{"appId": dat.AppId, "integration": integration}).Inc()
		prom.BytesProduced.With(prometheus.Labels{"appId": dat.AppId, "integration": integration}).Add(float64(len(message)))
		if c.config.CassandraEnabled {
			// Send the message ID to cassandra
			if err := c.config.Cassandra.InsertMessage(dat.MessageID, integration+"-sent"); err != nil {
				logger.Error(err)
			}
		}
	}
	return nil
}

// handleEvents is an internal event handler that listens to the producers events
// Runs in a loop waiting for the shutdown channel to close it.
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
			case *confluent.Message:
				// If there was an error producing the message, handle it!
				if e.TopicPartition.Error != nil {
					logger.Errorf("Delivery failed: %v", e.TopicPartition.Error)
					dat := &Message{}
					if err := dat.UnmarshalJSON(e.Value); err != nil {
						//if err := json.Unmarshal(e.Value, &dat); err != nil {
						logger.Error("Error unmarshaling message json: " + err.Error())
						return
					}
					prom.MessagesProducedFailed.With(prometheus.Labels{"integration": *e.TopicPartition.Topic, "appId": dat.AppId}).Inc()
					if config.GetBool(config.Retry) {
						// Send to retry cache
						c.config.DeadletterClient.AddToQueue(&deadletterqueue.QueueObject{
							Key:  fmt.Sprintf("%s/%s", dat.AppId, e.TopicPartition.Topic),
							Data: e.Value,
						})
					}
				}
			case *confluent.Error:
				logger.Error(e.Error())
			case *confluent.Stats:
				go prom.HandleKafkaStats(e, "producer")

			default:
				logger.WithField("type", reflect.TypeOf(e).String()).Errorf("Producer Ignored Event: %v", e)
			}
		}
	}
	logger.Debug("exiting handleEvents")
}

// Close will close the Kafka Producer, flush any remaining messages, and report on any messages that failed to flush
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
