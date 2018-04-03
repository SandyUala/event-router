package types

// Consumer represents a stream consumer
type Consumer interface {
	Run()
	Close()
}
