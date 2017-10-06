package cmd

import (
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
}
