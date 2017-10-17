package cmd

import (
	"strings"

	"os"

	"github.com/astronomerio/event-router/api"
	"github.com/astronomerio/event-router/api/v1"
	"github.com/astronomerio/event-router/cassandra"
	"github.com/astronomerio/event-router/config"
	"github.com/astronomerio/event-router/houston"
	"github.com/astronomerio/event-router/integrations"
	"github.com/astronomerio/event-router/kafka/clickstream"
	"github.com/astronomerio/event-router/sse"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	MockCmd = &cobra.Command{
		Use:   "mock",
		Short: "run even-router with a mock houston, takes a list of integrations to enabled",
		Run:   mock,
	}

	cassandraServers           string
	enableCassandra            bool
	cassandraReplicationFactor int
	cassandraTable             string
)

func init() {
	RootCmd.AddCommand(MockCmd)
	MockCmd.Flags().StringVar(&cassandraServers, "cassandra-servers", "", "comma separated list of cassandra servers")
	MockCmd.Flags().BoolVar(&enableCassandra, "enable-cassandra", false, "enable cassandra for recording message ids")
	MockCmd.Flags().IntVar(&cassandraReplicationFactor, "replication-factor", 1, "cassandra replication factor")
	MockCmd.Flags().StringVar(&cassandraTable, "cassandra-table", "", "casandra table")
}

func mock(cmd *cobra.Command, args []string) {
	// Setup debug logging first
	if config.IsDebugEnabled() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logger := log.WithField("function", "start")
	logger.Info("Starting mock event-router")

	logger.Warn("================================================================")
	logger.Warn("WARNING!  THIS IS A MOCK CLIENT.")
	logger.Warn("It will connect to kafka and process data on the topics, " +
		"but it will not connect to Houston or receive app change events!")
	logger.Warn("================================================================")

	// Create our simple web server
	apiClient := api.NewClient()
	apiClient.AppendRouteHandler(v1.NewPromHandler())
	// Setup api debug level (for gin logging)
	api.Debug = config.GetBool(config.DebugEnvLabel)

	bootstrapServers := config.GetString(config.BootstrapServersEnvLabel)
	topics := strings.Split(config.GetString(config.TopicEnvLabel), ",")

	// Parse the args
	if len(args) == 0 {
		logger.Error("Arguments must be supplied in the format '<name>:<code>,<name>:<code>'\nExample:  'S3 Event Logs:s3-event-logs")
		os.Exit(1)
	}

	values := make(map[string]string)

	for _, ints := range strings.Split(args[0], ",") {
		value := strings.Split(ints, ":")
		values[value[0]] = value[1]
	}

	mockHoustonClient := &houston.MockClient{
		Integrations: values,
	}
	integration := integrations.NewClient(mockHoustonClient)

	// SSE Client
	if !DisableSSE {
		sseClient := sse.NewSSEClient(config.GetString(config.SSEURLEnvLabel),
			config.GetString(config.SSEAuthEnvLabel))
		sseClient.Subscribe("clickstream", integration.EventListener)
	}

	// If we are persisting to cassandra, create the client and pass it to the producer
	var cassandraClient *cassandra.Client
	if enableCassandra {
		cassandraClient, err := cassandra.NewCilent(&cassandra.Configs{
			MessageTableName:  cassandraTable,
			ReplicationFactor: cassandraReplicationFactor,
			Servers:           strings.Split(cassandraServers, ","),
		})
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}
	}

	// Create our clickstreamProducer
	clickstreamProducer, err := clickstream.NewProducer(&clickstream.ProducerOptions{
		BootstrapServers: bootstrapServers,
		Integrations:     integration,
		MessageTimeout:   config.GetInt(config.KafkaProducerMessageTimeoutMSEvnLabel),
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
	logger.Info("Starting Clickstream Handler")
	go clickstreamHandler.Run()

	// Start the simple server
	logger.Info("Starting HTTP Server")
	if err := apiClient.Serve(config.GetString(config.ServePortEnvLabel)); err != nil {
		logger.Panic(err)
	}
	logger.Debug("Exiting event-router")
}
