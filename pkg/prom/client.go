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

	MessagesProducedFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "messages_produced_failed",
		Help: "The number of messages produced that failed to send",
	})

	MessagesRetried = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "messages_retried",
		Help: "The number of clickstream message that retried to produce",
	}, []string{"appId", "integration"})
)

func init() {
	prometheus.MustRegister(
		MessagesConsumed,
		MessagesProduced,
		SSEClickstreamMessagesReceived,
		MessagesProducedFailed,
		MessagesRetried,
	)
}
