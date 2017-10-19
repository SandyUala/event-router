package clickstream

import (
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func TestConsumer(t *testing.T) {

	opts := &ConsumerOptions{
		GroupID: "test",
	}
	c, _ := NewConsumer(opts)

	topic := "testTopic"
	// Test Message

	km := kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic: &topic,
		},
		Value: []byte(jsonBlob),
		Key:   []byte("key"),
	}

	c.consumer.Events() <- &km
}
