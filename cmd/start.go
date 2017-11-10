package cmd

import (
	"os"

	"time"

	"os/signal"
	"syscall"

	"runtime/pprof"

	"runtime/trace"

	"github.com/astronomerio/clickstream-event-router/api"
	"github.com/astronomerio/clickstream-event-router/api/v1"
	"github.com/astronomerio/clickstream-event-router/config"
	"github.com/astronomerio/clickstream-event-router/deadletterqueue"
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
	StartPPROF      = false
	StartTrace      = ""
	StartProfile    = ""
	KafkaDebug      = false
)

func init() {
	RootCmd.AddCommand(StartCmd)
	StartCmd.Flags().BoolVar(&EnableRetry, "retry", false, "enables retry logic")
	StartCmd.Flags().BoolVar(&DisableCacheTTL, "disable-cache-ttl", false, "disables cache ttl")
	StartCmd.Flags().BoolVar(&StartPPROF, "pprof", false, "enables pprof")
	StartCmd.Flags().StringVarP(&StartProfile, "profile", "p", "", "enable cpu profile and set file location")
	StartCmd.Flags().StringVarP(&StartTrace, "trace", "t", "", "enable trace and set file location")
	StartCmd.Flags().BoolVar(&KafkaDebug, "kafka-debug", false, "enable kafka debuging")
}

func start(cmd *cobra.Command, args []string) {
	// Setup debug logging first
	if config.IsDebugEnabled() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logger := log.WithField("function", "start")
	logger.Info("Starting event-router")

	if len(StartProfile) != 0 {
		if StartProfile[len(StartProfile)-1] == '/' {
			StartProfile = StartProfile[:len(StartProfile)-1]
		}
		logger.Info("Enabling Profiling")
		// CPU Profile
		f, err := os.Create(StartProfile + "/cpuprofile.pprof")
		if err != nil {
			logger.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			logger.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}

	if len(StartTrace) != 0 {
		if StartTrace[len(StartTrace)-1] == '/' {
			StartTrace = StartTrace[:len(StartTrace)-1]
		}
		logger.Info("Enabling Tracing")
		// Trace
		t, err := os.Create(StartTrace + "/clickstream-event-router.trace")
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
	apiClient.AppendRouteHandler(v1.NewIntegrationsHandler(integrationClient))

	// SSE Client
	if !DisableSSE {
		sseClient, err := sse.NewSSEClient(config.GetString(config.SSEURL), houstonClient, shutdownChannel)
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}
		// Register our integrations event listener with the SSE Client
		sseClient.Subscribe("clickstream", integrationClient.SSEEventListener)
	}

	// Create our clickstreamProducer
	clickstreamProducerOptions := &clickstream.ProducerConfig{
		BootstrapServers: bootstrapServers,
		Integrations:     integrationClient,
		MessageTimeout:   config.GetInt(config.KafkaProducerMessageTimeoutMS),
		FlushTimeout:     config.GetInt(config.KafkaProducerFlushTimeoutMS),
		RetryTopic:       config.GetString(config.RetryTopic),
		RetryS3Bucket:    config.GetString(config.ClickstreamRetryS3Bucket),
		S3PathPrefix:     config.GetString(config.ClickstreamRetryS3PathPrefix),
		MasterTopic:      topic,
		ShutdownChannel:  shutdownChannel,
	}
	if EnableRetry {
		s3Client, err := s3.NewClient()
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}
		dlc := deadletterqueue.NewClient(
			&deadletterqueue.ClientConfig{
				S3Bucket:        config.GetString(config.ClickstreamRetryS3Bucket),
				ShutdownChannel: shutdownChannel,
				FlushTimeout:    config.GetInt64(config.ClickstreamRetryFlushTimeout),
				QueueSize:       config.GetInt64(config.ClickstreamRetryMaxQueue),
			}, s3Client)
		clickstreamProducerOptions.DeadletterClient = dlc
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

	// Start the simple server
	logger.Info("Starting HTTP Server")
	if err := apiClient.Serve(config.GetString(config.ServePort), StartPPROF); err != nil {
		logger.Error(err)
	}
	logger.Debug("Exiting event-router")
}
