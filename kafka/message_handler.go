package kafka

import "github.com/Shopify/sarama"

type MessageHandler interface {
	HandleMessage(message *sarama.ConsumerMessage)
	Close()
}
