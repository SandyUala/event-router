package kafka

type Consumer interface {
	Run()
	Close()
}
