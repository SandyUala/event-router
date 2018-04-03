package types

// Producer represents a stream producer
type Producer interface {
	HandleMessage()
}
