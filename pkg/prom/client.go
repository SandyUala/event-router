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
)

func init() {
	prometheus.MustRegister(MessagesConsumed,
		MessagesProduced)
}
