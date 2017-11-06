package cmd

import (
	"os"
	"runtime"

	"time"

	"os/signal"
	"syscall"

	"runtime/pprof"

	"runtime/trace"

	"github.com/astronomerio/clickstream-event-router/api"
	"github.com/astronomerio/clickstream-event-router/api/v1"
	"github.com/astronomerio/clickstream-event-router/config"
	"github.com/astronomerio/clickstream-event-router/houston"
	"github.com/astronomerio/clickstream-event-router/integrations"
	"github.com/astronomerio/clickstream-event-router/kafka/clickstream"
	"github.com/astronomerio/clickstream-event-router/pkg"
	"github.com/astronomerio/clickstream-event-router/s3"
	"github.com/astronomerio/clickstream-event-router/sse"
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

	DisableCacheTTL = false
	EnableRetry     = false
	StartPPROF      = ""
	KafkaDebug      = false
)

func init() {
	RootCmd.AddCommand(StartCmd)
	StartCmd.Flags().BoolVar(&EnableRetry, "retry", false, "enables retry logic")
	StartCmd.Flags().BoolVar(&DisableCacheTTL, "disable-cache-ttl", false, "disables cache ttl")
	StartCmd.Flags().StringVarP(&StartPPROF, "pprof", "p", "", "enable pprof and set file location")
	StartCmd.Flags().BoolVar(&KafkaDebug, "kafka-debug", false, "enable kafka debuging")
}

func start(cmd *cobra.Command, args []string) {
	// Setup debug logging first
	if config.IsDebugEnabled() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logger := log.WithField("function", "start")
	logger.Info("Starting event-router")

	if len(StartPPROF) != 0 {
		if StartPPROF[len(StartPPROF)-1] == '/' {
			StartPPROF = StartPPROF[:len(StartPPROF)-1]
		}
		logger.Info("Enabling Profiling")
		// CPU Profile
		f, err := os.Create(StartPPROF + "/cpuprofile.pprof")
		if err != nil {
			logger.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			logger.Fatal(err)
		}
		defer pprof.StopCPUProfile()

		// Trace
		t, err := os.Create(StartPPROF + "/clickstream-event-router.trace")
		if err != nil {
			logger.Fatal(err)
		}
		if err := trace.Start(t); err != nil {
			logger.Fatal(err)
		}
		defer trace.Stop()
	}

	// Create our simple web server
	apiClient := api.NewClient()
	apiClient.AppendRouteHandler(v1.NewPromHandler())
	// Setup api debug level (for gin logging)
	api.Debug = config.GetBool(config.Debug)

	// Disable cache TTL if flag passed
	if DisableCacheTTL {
		config.SetBool(config.DisableCacheTTL, true)
	}

	bootstrapServers := config.GetString(config.BootstrapServers)
	topic := config.GetString(config.KafkaIngestionTopic)
	config.SetBool(config.KafakDebug, KafkaDebug)
	config.SetBool(config.Retry, EnableRetry)

	// Shutdown Channel
	shutdownChannel := make(chan struct{})

	// HTTP Client
	httpClient := pkg.NewHTTPClient()

	// Houston Client
	houstonClient := houston.NewHoustonClient(httpClient, config.GetString(config.HoustonAPIURL))

	// Integration Client
	integrationClient := integrations.NewClient(houstonClient, shutdownChannel)

	// SSE Client
	if !DisableSSE {
		sseClient, err := sse.NewSSEClient(config.GetString(config.SSEURL), houstonClient, shutdownChannel)
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}
		// Register our integrations event listener with the SSE Client
		sseClient.Subscribe("clickstream", integrationClient.EventListener)
	}

	// Create our clickstreamProducer
	clickstreamProducerOptions := &clickstream.ProducerConfig{
		BootstrapServers: bootstrapServers,
		Integrations:     integrationClient,
		MessageTimeout:   config.GetInt(config.KafkaProducerMessageTimeoutMS),
		FlushTimeout:     config.GetInt(config.KafkaProducerFlushTimeoutMS),
		RetryS3Bucket:    config.GetString(config.ClickstreamRetryS3Bucket),
		RetryTopic:       config.GetString(config.ClickstreamRetryTopic),
		S3PathPrefix:     config.GetString(config.ClickstreamRetryS3PathPrefix),
		MasterTopic:      topic,
		ShutdownChannel:  shutdownChannel,
	}
	clickstreamProducer, err := clickstream.NewProducer(clickstreamProducerOptions)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	// Clickstream Consumer
	clickstreamConsumer, err := clickstream.NewConsumer(&clickstream.ConsumerOptions{
		BootstrapServers: bootstrapServers,
		GroupID:          config.GetString(config.KafkaGroupID),
		Topic:            topic,
		MessageHandler:   clickstreamProducer,
		ShutdownChannel:  shutdownChannel,
	})
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	shouldShutdown := false
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		close(shutdownChannel)
		shouldShutdown = true
	}()

	go func() {
		for {
			if shouldShutdown {
				logger.Debug("Shutting down consumer")
				return
			}
			// Try to get an auth token.  Used as a healthcheck
			_, err := houstonClient.GetAuthorizationToken()
			if err != nil {
				logger.Error(err)
				// Sleep for 5 seconds
				time.Sleep(time.Second * 5)
			} else {
				clickstreamConsumer.Run()
			}
		}
	}()

	// If Retry is enabled, start the consumer and producer
	if EnableRetry {
		s3Client, err := s3.NewClient()
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}
		// Create clickstream retry producer
		clickstreamRetryProducer, err := clickstream.NewRetryProducer(clickstreamProducerOptions, config.GetInt(config.MaxRetries), s3Client)
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}

		// Create clickstream retry consumer
		clickstreamRetryConsumer, err := clickstream.NewConsumer(&clickstream.ConsumerOptions{
			BootstrapServers: bootstrapServers,
			GroupID:          config.GetString(config.KafkaGroupID),
			Topic:            clickstreamProducerOptions.RetryTopic,
			MessageHandler:   clickstreamRetryProducer,
		})
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}
		logger.Info("Starting Clickstream Retry Handler")
		go clickstreamRetryConsumer.Run()
	}

	// Start the simple server
	logger.Info("Starting HTTP Server")
	if err := apiClient.Serve(config.GetString(config.ServePort)); err != nil {
		logger.Error(err)
	}
	if len(StartPPROF) != 0 {
		// Write out memory heap
		f, err := os.Create(StartPPROF + "/memoryprofile.pprof")
		if err != nil {
			logger.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			logger.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}
	logger.Debug("Exiting event-router")
}
