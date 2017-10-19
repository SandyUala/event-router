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
)

func init() {
	RootCmd.AddCommand(StartCmd)
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
	houstonClient := houston.NewHoustonClient(httpClient, config.GetString(config.HoustonAPIURLEnvLabel))

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
	clickstreamProducer, err := clickstream.NewProducer(&clickstream.ProducerConfig{
		BootstrapServers: bootstrapServers,
		Integrations:     integrationClient,
		MessageTimeout:   config.GetInt(config.KafkaProducerMessageTimeoutMSEvnLabel),
		FlushTimeout:     config.GetInt(config.KafkaProducerFlushTimeoutMSEnvLabel),
	})
	if err != nil {
		logger.Panic(err)
	}

	// Clickstream Handler
	clickstreamHandler, err := clickstream.NewConsumer(&clickstream.ConsumerOptions{
		BootstrapServers: bootstrapServers,
		GroupID:          config.GetString(config.GroupIDEnvLabel),
		Topics:           topics,
		MessageHandler:   clickstreamProducer,
	})
	go clickstreamHandler.Run()

	// Start the simple server
	apiClient.Serve(config.GetString(config.ServePortEnvLabel))
}
