package cmd

import (
	"github.com/astronomerio/event-router/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use: "start",
	Run: start,
}

func init() {
	RootCmd.AddCommand(startCmd)
}

func start(cmd *cobra.Command, args []string) {
	log := logging.GetLogger().WithFields(logrus.Fields{"package": "cmd"})
	log.Info("fuck")
}
