package cmd

import (
	"strings"

	"github.com/astronomerio/event-router/api"
	"github.com/astronomerio/event-router/api/v1"
	"github.com/astronomerio/event-router/config"
	"github.com/astronomerio/event-router/kafka/clickstream"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	log     = logrus.WithField("package", "cmd")
	RootCmd = &cobra.Command{
		Use:   "event-router",
		Short: "event-router will route incoming events from analytics.js to the correct integration",
		Run:   start,
	}
)

func Execute() {
	RootCmd.Execute()
}

func start(cmd *cobra.Command, args []string) {
	logger := log.WithField("function", "start")
	logger.Info("Starting event-router")

	// Create our simple web server
	apiClient := api.NewClient()
	apiClient.AppendRouteHandler(v1.NewPromHandler())
	// Setup api debug level (for gin logging)
	api.Debug = config.GetBool(config.DebugEnvLabel)

	bootstrapServers := strings.Split(config.GetString(config.BootstrapServersEnvLabel), ",")
	topics := strings.Split(config.GetString(config.TopicEnvLabel), ",")

	// Create our clickstreamProducer
	clickstreamProducer, err := clickstream.NewProducer(&clickstream.ProducerOptions{
		BootstrapServers: bootstrapServers,
	})
	if err != nil {
		logger.Panic(err)
	}

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
