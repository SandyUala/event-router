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
	cassandraEnabled           bool
	cassandraReplicationFactor int
	cassandraTable             string
	runID                      int64
	cassandraKeyspace          string
)

func init() {
	RootCmd.AddCommand(MockCmd)
	MockCmd.Flags().StringVar(&cassandraServers, "cassandra-servers", "", "comma separated list of cassandra servers")
	MockCmd.Flags().BoolVar(&cassandraEnabled, "enable-cassandra", false, "enable cassandra for recording message ids")
	MockCmd.Flags().IntVar(&cassandraReplicationFactor, "replication-factor", 1, "cassandra replication factor")
	MockCmd.Flags().StringVar(&cassandraTable, "cassandra-table", "", "casandra table")
	MockCmd.Flags().Int64Var(&runID, "run-id", 0, "run id")
	MockCmd.Flags().StringVar(&cassandraKeyspace, "cassandra-keyspace", "mock", "cassandra keyspace")
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
	var err error
	if cassandraEnabled {
		if len(cassandraServers) == 0 {
			logger.Error("--cassandra-servers required when cassandra is enabled")
			os.Exit(1)
		}
		if len(cassandraTable) == 0 {
			logger.Error("--cassandra-table required when cassandra is enabled")
			os.Exit(1)
		}
		if runID == 0 {
			logger.Error("--run-id required when cassandra is enabled")
			os.Exit(1)
		}
		cassandraClient, err = cassandra.NewClient(&cassandra.Configs{
			MessageTableName:  cassandraTable,
			ReplicationFactor: cassandraReplicationFactor,
			Servers:           strings.Split(cassandraServers, ","),
			RunID:             runID,
			Keyspace:          cassandraKeyspace,
		})
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}
	}

	// Create our clickstreamProducer
	clickstreamOptions := &clickstream.ProducerConfig{
		BootstrapServers: bootstrapServers,
		Integrations:     integration,
		MessageTimeout:   config.GetInt(config.KafkaProducerMessageTimeoutMSEvnLabel),
		Cassandra:        cassandraClient,
		CassandraEnabled: cassandraEnabled,
		RetryS3Bucket:    config.GetString(config.ClickstreamRetryS3BucketEnvLabel),
		RetryTopic:       config.GetString(config.ClickstreamRetryTopicEnvLabel),
		S3PathPrefix:     config.GetString(config.S3PathPrefixEnvLabel),
	}
	clickstreamProducer, err := clickstream.NewProducer(clickstreamOptions)
	if err != nil {
		logger.Panic(err)
	}

	// Create clickstream handler
	clickstreamHandler, err := clickstream.NewConsumer(&clickstream.ConsumerOptions{
		BootstrapServers: bootstrapServers,
		GroupID:          config.GetString(config.GroupIDEnvLabel),
		Topics:           topics,
		MessageHandler:   clickstreamProducer,
	})
	logger.Info("Starting Clickstream Handler")
	go clickstreamHandler.Run()

	// Create clickstream retry producer
	clickstreamRetryProducer, err := clickstream.NewRetryProducer(clickstreamOptions, config.GetInt(config.MaxRetriesEnvLabel))
	if err != nil {
		logger.Panic(err)
	}

	// Create clickstream retry handler
	clickstreamRetryHandler, err := clickstream.NewConsumer(&clickstream.ConsumerOptions{
		BootstrapServers: bootstrapServers,
		GroupID:          config.GetString(config.GroupIDEnvLabel),
		Topics:           []string{clickstreamOptions.RetryTopic},
		MessageHandler:   clickstreamRetryProducer,
	})
	logger.Info("Starting Clickstream Retry Handler")
	go clickstreamRetryHandler.Run()

	// Start the simple server
	logger.Info("Starting HTTP Server")
	if err := apiClient.Serve(config.GetString(config.ServePortEnvLabel)); err != nil {
		logger.Panic(err)
	}
	logger.Debug("Exiting event-router")
}
