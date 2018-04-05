package cmd

import (
	"io"
	"os"
	"os/signal"
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
	log := logging.GetLogger(logrus.Fields{"package": "api"})

	config.AppConfig.Print()

	// Listen for system signals to shutdown and close our shutdown channel
	shutdownChan := make(chan struct{})
	go func() {
		sc := make(chan os.Signal)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP)
		<-sc
		close(shutdownChan)
	}()

	// Create a stream consumer
	consumer, err := kafka.NewConsumer(&kafka.ConsumerConfig{
		BootstrapServers: config.AppConfig.KafkaBrokers,
		GroupID:          config.AppConfig.KafkaGroupID,
		Topic:            config.AppConfig.KafkaInputTopic,
		DebugMode:        config.AppConfig.DebugMode,
		ShutdownChannel:  shutdownChan,
	})
	if err != nil {
		log.Fatal(err)
	}

	// file, err := os.Create("./records.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer file.Close()

	producer, err := kafka.NewProducer(&kafka.ProducerConfig{
		BootstrapServers: config.AppConfig.KafkaBrokers,
		DebugMode:        config.AppConfig.DebugMode,
		ShutdownChannel:  shutdownChan,
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if _, err = io.Copy(producer, consumer); err != nil {
			log.Fatal(err)
		}
	}()

	apiServerConfig := &api.ServerConfig{
		APIInterface: config.AppConfig.APIInterface,
		APIPort:      config.AppConfig.APIPort,
	}

	apiServer := api.NewServer().
		WithConfig(apiServerConfig).
		WithRouteHandler(v1.NewPrometheusHandler())

	apiServer.Run(shutdownChan)
	log.Info("Finished")
}
