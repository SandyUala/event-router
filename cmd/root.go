package cmd

import (
	"github.com/astronomerio/clickstream-event-router/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	log     = logrus.WithField("package", "cmd")
	RootCmd = &cobra.Command{
		Use:   "event-router",
		Short: "event-router will route incoming events from analytics.js to the correct integration",
	}

	DisableSSE bool
)

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().BoolVar(&DisableSSE, "disable-sse", false, "disables SSE client")
}

func initConfig() {
	config.Initalize(
		&config.InitOptions{
			EnableRetry: EnableRetry,
		},
	)
}
