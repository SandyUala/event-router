package clickstream

import (
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type MockIntegrationClient struct {
}

func (c *MockIntegrationClient) GetIntegrations(appId string) *map[string]string {
	return &map[string]string{
		"S3 Event": "s3-event",
	}
}

func (c *MockIntegrationClient) UpdateIntegrationsForApp(appId string) error {
	return nil
}

func (c *MockIntegrationClient) EventListener(eventRaw, dataRaw []byte) {

}

var jsonBlob = `
{
	"appId":"test",
	"messageId":"testMessage"
}`

func TestProduce(t *testing.T) {
	p, _ := NewProducer(&ProducerConfig{
		Integrations: &MockIntegrationClient{},
	})

	message := []byte(jsonBlob)
	key := []byte("key")

	p.HandleMessage(message, key)

	// Send a test message down the producer events channel to simulate an error
	topic := "testTopic"
	kafkaKessage := &kafka.Message{
		Value: []byte(jsonBlob),
		Key:   key,
		TopicPartition: kafka.TopicPartition{
			Topic: &topic,
		},
	}
	p.producer.Events() <- kafkaKessage

}
