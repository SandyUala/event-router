package clickstream

import (
	"testing"

	"github.com/astronomerio/cs-event-router/integrations"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var jsonBlob = `
{
	"appId":"test",
	"messageId":"testMessage"
}`

func TestProduce(t *testing.T) {
	p, _ := NewProducer(&ProducerConfig{
		Integrations: &integrations.MockClient{},
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
