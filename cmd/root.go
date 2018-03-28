package cmd

import (
	"github.com/astronomerio/event-router/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// RootCmd is the root cobra command
var RootCmd = &cobra.Command{
	Use: "event-router",
	Run: start,
}

func start(cmd *cobra.Command, args []string) {
	log := logging.GetLogger().WithFields(logrus.Fields{"package": "cmd"})

	log.Info("fuck")
}
