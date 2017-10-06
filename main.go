package main

import (
	"os"

	"github.com/astronomerio/event-router/cmd"
	"github.com/astronomerio/event-router/config"
	"github.com/sirupsen/logrus"
)

func main() {
	// Setup debug logging first
	if config.IsDebugEnabled() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if err := cmd.RootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
