package cmd

import (
	"strings"

	"github.com/astronomerio/event-router/api"
	"github.com/astronomerio/event-router/api/v1"
	"github.com/astronomerio/event-router/config"
	"github.com/astronomerio/event-router/houston"
	"github.com/astronomerio/event-router/integrations"
	"github.com/astronomerio/event-router/kafka/clickstream"
	"github.com/astronomerio/event-router/pkg"
	"github.com/astronomerio/event-router/sse"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	StartCmd = &cobra.Command{
		Use:   "start",
		Short: "start event-router",
		Long:  "Starts the event-router",
		Run:   start,
	}

	EnableRetry = false
)

func init() {
	RootCmd.AddCommand(StartCmd)
	StartCmd.Flags().BoolVar(&EnableRetry, "retry", false, "enables retry logic")
}

func start(cmd *cobra.Command, args []string) {
	// Setup debug logging first
	if config.IsDebugEnabled() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logger := log.WithField("function", "start")
	logger.Info("Starting event-router")

	// Create our simple web server
	apiClient := api.NewClient()
	apiClient.AppendRouteHandler(v1.NewPromHandler())
	// Setup api debug level (for gin logging)
	api.Debug = config.GetBool(config.DebugEnvLabel)

	bootstrapServers := config.GetString(config.BootstrapServersEnvLabel)
	topics := strings.Split(config.GetString(config.TopicEnvLabel), ",")

	// HTTP Client
	httpClient := pkg.NewHTTPClient()

	// Houston Client
	houstonClient := houston.NewHoustonClient(httpClient)

	// Integration Client
	integrationClient := integrations.NewClient(houstonClient)

	// SSE Client
	if !DisableSSE {
		sseClient := sse.NewSSEClient(config.GetString(config.SSEURLEnvLabel),
			config.GetString(config.SSEAuthEnvLabel))
		// Register our integrations event listener with the SSE Client
		sseClient.Subscribe("appChanges", integrationClient.EventListener)
	}

	// Create our clickstreamProducer
	clickstreamProducerOptions := &clickstream.ProducerConfig{
		BootstrapServers: bootstrapServers,
		Integrations:     integrationClient,
		MessageTimeout:   config.GetInt(config.KafkaProducerMessageTimeoutMSEvnLabel),
		FlushTimeout:     config.GetInt(config.KafkaProducerFlushTimeoutMSEnvLabel),
		RetryS3Bucket:    config.GetString(config.ClickstreamRetryS3BucketEnvLabel),
		RetryTopic:       config.GetString(config.ClickstreamRetryTopicEnvLabel),
		S3PathPrefix:     config.GetString(config.S3PathPrefixEnvLabel),
	}
	clickstreamProducer, err := clickstream.NewProducer(clickstreamProducerOptions)
	if err != nil {
		logger.Panic(err)
	}

	// Clickstream Consumer
	clickstreamConsumer, err := clickstream.NewConsumer(&clickstream.ConsumerOptions{
		BootstrapServers: bootstrapServers,
		GroupID:          config.GetString(config.GroupIDEnvLabel),
		Topics:           topics,
		MessageHandler:   clickstreamProducer,
	})
	go clickstreamConsumer.Run()

	// If Retry is enabled, start the consumer and producer
	if EnableRetry {
		// Create clickstream retry producer
		clickstreamRetryProducer, err := clickstream.NewRetryProducer(clickstreamProducerOptions, config.GetInt(config.MaxRetriesEnvLabel))
		if err != nil {
			logger.Panic(err)
		}

		// Create clickstream retry consumer
		clickstreamRetryConsumer, err := clickstream.NewConsumer(&clickstream.ConsumerOptions{
			BootstrapServers: bootstrapServers,
			GroupID:          config.GetString(config.GroupIDEnvLabel),
			Topics:           []string{clickstreamProducerOptions.RetryTopic},
			MessageHandler:   clickstreamRetryProducer,
		})
		logger.Info("Starting Clickstream Retry Handler")
		go clickstreamRetryConsumer.Run()
	}

	// Start the simple server
	logger.Info("Starting HTTP Server")
	if err := apiClient.Serve(config.GetString(config.ServePortEnvLabel)); err != nil {
		logger.Panic(err)
	}
	logger.Debug("Exiting event-router")
}
