package clickstream

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"encoding/json"

	"github.com/astronomerio/event-router/s3"
	confluent "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

type RetryProducer struct {
	producer  *confluent.Producer
	config    *ProducerConfig
	maxRetrys int
	s3Client  s3.S3Client
}

type RetryMessage struct {
	RetryCount  int             `json:"retryCount"`
	Integration string          `json:"integration"`
	Message     json.RawMessage `json:"message"`
}

func NewRetryProducer(config *ProducerConfig, maxRetrys int, s3Client s3.S3Client) (*RetryProducer, error) {
	p, err := confluent.NewProducer(&confluent.ConfigMap{
		"bootstrap.servers":  config.BootstrapServers,
		"message.timeout.ms": config.MessageTimeout,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error creating kafka messageHandler")
	}
	cs := &RetryProducer{
		producer:  p,
		config:    config,
		maxRetrys: maxRetrys,
		s3Client:  s3Client,
	}
	// Handle messageHandler events in a goroutine
	go cs.handleEvents()
	return cs, nil
}

func (c *RetryProducer) HandleMessage(message []byte, key []byte) {
	logger := log.WithField("function", "HandleMessage")
	dat := &RetryMessage{}
	if err := json.Unmarshal(message, &dat); err != nil {
		logger.Error("Error unmarshaling retry message json: " + err.Error())
		return
	}
	// If we hit the max retries, send to S3
	if dat.RetryCount > c.maxRetrys {
		c.sendToS3(dat)
	}

	// Produce the message
	c.producer.ProduceChannel() <- &confluent.Message{
		TopicPartition: confluent.TopicPartition{
			Topic:     &dat.Integration,
			Partition: confluent.PartitionAny,
		},
		Key:   key,
		Value: dat.Message,
	}

}

func (c *RetryProducer) sendToS3(rertryMessagae *RetryMessage) {
	logger := log.WithField("function", "sendToS3")
	logger.Debug("Entered sendToS3")
	// Get message ID
	dat := &Message{}
	if err := json.Unmarshal(rertryMessagae.Message, &dat); err != nil {
		logger.Error("Error unmarshaling message json: " + err.Error())
		return
	}
	// Create the S3 key
	key := ""
	if len(c.config.S3PathPrefix) > 0 {
		key = c.config.S3PathPrefix + "/"
	}
	key += dat.MessageID
	// Send to s3
	if err := c.s3Client.SendToS3(&c.config.RetryS3Bucket, &key, rertryMessagae.Message); err != nil {
		logger.Error("Error sending message to S3: " + err.Error())
	}
}

func (c *RetryProducer) handleEvents() {
	logger := log.WithField("function", "handleEvents")
	logger.Debug("Entered handleEvents")

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	run := true
	for run {
		select {
		case sig := <-sigchan:
			logger.Infof("Retry Producer caught signal %v: terminating", sig)
			c.Close()
			run = false
		case ev := <-c.producer.Events():
			switch e := ev.(type) {
			case *confluent.Message:
				m := e
				if m.TopicPartition.Error != nil {
					logger.Errorf("Delivery failed: %v", m.TopicPartition.Error)
					// Retry the message
					dat := &RetryMessage{}
					if err := json.Unmarshal(m.Value, &dat); err != nil {
						logger.Error("Error unmarshaling retry message json: " + err.Error())
						return
					}
					// We are retrying an existing message, increment the retry count
					dat.RetryCount++
					// So we can get into a loop where sending to the retry channel fails.
					// If that happens, persist the retry message in s3
					if dat.RetryCount > c.maxRetrys {
						c.sendToS3(dat)
						continue
					}
					c.producer.ProduceChannel() <- &confluent.Message{
						TopicPartition: confluent.TopicPartition{
							Topic:     &c.config.RetryTopic,
							Partition: confluent.PartitionAny,
						},
						Key:   m.Key,
						Value: *dat.ToBytes(),
					}
				}
			case *confluent.Error:
				logger.Error(e.Error())

			default:
				fmt.Printf("Ignored event: %s\n", e)
			}
		}
	}
	logger.Debug("exiting handleEvents")
}

func (c *RetryProducer) Close() {
	logger := log.WithField("function", "Close")
	logger.Info("Closing messageHandler")
	messagesLeftToFlushed := c.producer.Flush(c.config.FlushTimeout)
	if messagesLeftToFlushed != 0 {
		logger.Errorf("Failed to flush %d messages after %d ms", messagesLeftToFlushed, c.config.FlushTimeout)
	}
	c.producer.Close()
	if c.config.CassandraEnabled {
		c.config.Cassandra.Close()
	}
	logger.Info("Producer Closed")
}

func (m *RetryMessage) ToBytes() *[]byte {
	dat, err := json.Marshal(m)
	if err != nil {
		log.Error(err)
		return nil
	}
	return &dat
}
