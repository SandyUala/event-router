package logging

import (
	"github.com/astronomerio/event-router/config"
	"github.com/sirupsen/logrus"
)

// Singleton logger for application
var log *logrus.Logger

// Configure logger on startup
func init() {
	cfg := config.Get()
	log = logrus.New()

	if cfg.LogFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	if cfg.DebugMode {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
}

// GetLogger returns the singleton logger
func GetLogger(fields logrus.Fields) *logrus.Entry {
	return log.WithFields(fields)
}
