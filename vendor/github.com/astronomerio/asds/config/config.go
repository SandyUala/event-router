package config

import (
	"errors"
	"fmt"

	"strings"

	"github.com/spf13/viper"
)

var (
	// Prefix Environment Prefix
	Prefix                   = "ASDS"
	logDebug                 = "LOG_DEBUG"
	logFormat                = "LOG_FORMAT"
	listenPort               = "PORT"
	dockerAirflowImageTagENV = "DOCKER_AIRFLOW_IMAGE_TAG"
)

// Config struct holds the provisioner configs
type Config struct {
	LogDebug              bool
	LogFormat             string
	Port                  string
	MarathonURL           string
	DockerAirflowImageTag string
}

// GetConfigs returns the basic configs
func GetConfigs() (*Config, error) {
	viper.SetEnvPrefix(Prefix)
	viper.AutomaticEnv()
	setupDefaults()

	// Log Debug Statements
	logDebug := viper.GetBool(logDebug)

	// Log Format
	logFormat := viper.GetString(logFormat)
	if err := validateLogFormat(logFormat); err != nil {
		return nil, err
	}

	// Docker Airflow Image Tag
	dockerAirflowImageTag := viper.GetString(dockerAirflowImageTagENV)
	if len(dockerAirflowImageTag) == 0 {
		return nil, getError(dockerAirflowImageTagENV)
	}

	// Port
	port := viper.Get(listenPort).(string)

	config := &Config{
		LogDebug:  logDebug,
		LogFormat: logFormat,
		Port:      port,
		DockerAirflowImageTag: dockerAirflowImageTag,
	}
	return config, nil
}

func setupDefaults() {
	viper.SetDefault(logDebug, true)
	viper.SetDefault(logFormat, "stdout")
	viper.SetDefault(listenPort, "8081")
}

func validateLogFormat(logFormat string) error {
	switch strings.ToLower(logFormat) {
	case "stdout":
		return nil
	case "json":
		return nil
	case "logstash":
		return errors.New("log format 'logstash' is not impemented yet, sorry")
	default:
		return fmt.Errorf("log format '%s' is not valid.  Valid options are 'stdout', 'json', 'logstash'", logFormat)
	}
}

func getError(env string) error {
	return fmt.Errorf("Env variable %s not set and is required", Prefix+"_"+env)
}
