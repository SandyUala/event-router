package prom

import "github.com/prometheus/client_golang/prometheus"

var (
	MessagesConsumed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "messages_consumed",
		Help: "The number of messages consumed",
	})

	MessagesProduced = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "messages_produced",
		Help: "The number of messages produced",
	}, []string{"appId", "integration"})

	SSEClickstreamMessagesReceived = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sse_clickstream_messages_received",
	})
)

func init() {
	prometheus.MustRegister(MessagesConsumed,
		MessagesProduced)
}
