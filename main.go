package main

import (
	"os"

	"github.com/astronomerio/cs-event-router/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
