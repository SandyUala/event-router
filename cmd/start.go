package cmd

import (
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/astronomerio/event-router/api"
	"github.com/astronomerio/event-router/api/v1"
	"github.com/astronomerio/event-router/config"
	"github.com/astronomerio/event-router/kafka"
	"github.com/astronomerio/event-router/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	startCmd = &cobra.Command{
		Use: "start",
		Run: start,
	}
)

func init() {
	RootCmd.AddCommand(startCmd)
}

func start(cmd *cobra.Command, args []string) {
	log := logging.GetLogger(logrus.Fields{"package": "cmd"})
	config.AppConfig.Print()

	// Create a waitgroup to ensure a clean shutdown.
	var wg sync.WaitGroup

	// Listen for system signals to shutdown and close our shutdown channel
	shutdownChan := make(chan struct{})
	go func() {
		sc := make(chan os.Signal)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP)
		<-sc
		close(shutdownChan)
	}()

	// Run the event pipeline
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Create a stream consumer
		consumer, err := kafka.NewConsumer(&kafka.ConsumerConfig{
			BootstrapServers: config.AppConfig.KafkaBrokers,
			GroupID:          config.AppConfig.KafkaConsumerGroupID,
			Topic:            config.AppConfig.KafkaConsumerTopic,
			DebugMode:        config.AppConfig.DebugMode,
			ShutdownChannel:  shutdownChan,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer consumer.Close()

		// Create a stream producer
		producer, err := kafka.NewProducer(&kafka.ProducerConfig{
			BootstrapServers: config.AppConfig.KafkaBrokers,
			DebugMode:        config.AppConfig.DebugMode,
			MessageTimeout:   config.AppConfig.KafkaProducerMessageTimeout,
			FlushTimeout:     config.AppConfig.KafkaProducerFlushTimeout,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer producer.Close()

		// Block and continuously pipe data from the consumer to the producer.
		// Unblocks when io.EOF is returned from consumer.
		if _, err = io.Copy(producer, consumer); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Create API config
		apiServerConfig := &api.ServerConfig{
			APIInterface: config.AppConfig.APIInterface,
			APIPort:      config.AppConfig.APIPort,
		}

		// Create our API server
		apiServer := api.NewServer(apiServerConfig).
			WithRouteHandler(v1.NewPrometheusHandler())

		defer apiServer.Close()
		apiServer.Serve(shutdownChan)
	}()

	wg.Wait()
	log.Info("Finished")
}
