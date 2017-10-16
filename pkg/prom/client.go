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
		Help: "The number of clickstream message received by SSE client",
	})

	MessagesProducedFailuer = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "message_produced_failure",
		Help: "The number of messages produced that failed to send",
	})
)

func init() {
	prometheus.MustRegister(
		MessagesConsumed,
		MessagesProduced,
		SSEClickstreamMessagesReceived,
		MessagesProducedFailuer,
	)
}
