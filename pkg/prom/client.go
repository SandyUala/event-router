package prom

import "github.com/prometheus/client_golang/prometheus"

var (
	MessagesConsumed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "event_router_messages_consumed",
		Help: "The number of messages consumed",
	}, []string{"appId"})

	MessagesProduced = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "event_router_messages_produced",
		Help: "The number of messages produced",
	}, []string{"appId", "integration"})

	SSEClickstreamMessagesReceived = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "event_router_sse_clickstream_messages_received",
		Help: "The number of clickstream message received by SSE client",
	})

	MessagesProducedFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "event_router_messages_produced_failed",
		Help: "The number of messages produced that failed to send",
	})

	MessagesRetried = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "event_router_messages_retried",
		Help: "The number of clickstream message that retried to produce",
	}, []string{"appId", "integration"})

	BytesConsumed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "event_router_bytes_consumed",
		Help: "The number of bytes consumed",
	}, []string{"appId"})

	BytesProduced = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "event_router_bytes_produced",
		Help: "The number of bytes produced",
	}, []string{"appId", "integration"})

	BytesRetried = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "event_router_bytes_retried",
		Help: "The number of bytes retried",
	}, []string{"appId", "integration"})

	MessagesUploadedToS3 = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "event_router_messages_uploaded_to_s3",
		Help: "The number of messages uploaded to S3",
	}, []string{"appId", "integration"})

	BytesUploadedToS3 = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "event_router_bytes_uploaded_to_s3",
		Help: "The number of bytes uploaded to S3",
	}, []string{"appId", "integration"})

	MessagesWithNoIntegration = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "event_router_messages_no_integrations",
		Help: "The number of messages with no integrations enabled",
	}, []string{"appId"})
)

func init() {
	prometheus.MustRegister(
		MessagesConsumed,
		MessagesProduced,
		SSEClickstreamMessagesReceived,
		MessagesProducedFailed,
		MessagesRetried,
		BytesConsumed,
		BytesProduced,
		BytesRetried,
		MessagesUploadedToS3,
		BytesUploadedToS3,
		MessagesWithNoIntegration,
	)
}
