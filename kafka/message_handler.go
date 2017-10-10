package kafka

type MessageHandler interface {
	HandleMessage(message []byte, key []byte)
	Close()
}
