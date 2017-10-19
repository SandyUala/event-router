package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/astronomerio/event-router/cmd"
	"github.com/astronomerio/event-router/pkg"
	"github.com/astronomerio/viper"
	"github.com/sirupsen/logrus"
)

const (
	// Environment Variable Prefix
	Prefix = "ER"

	DebugEnvLabel = "DEBUG"

	BootstrapServersEnvLabel              = "BOOTSTRAP_SERVERS"
	ServePortEnvLabel                     = "SERVE_PORT"
	GroupIDEnvLabel                       = "GROUP_ID"
	TopicEnvLabel                         = "TOPIC"
	SSEURLEnvLabel                        = "SSE_URL"
	SSEAuthEnvLabel                       = "SSE_AUTH"
	KafkaProducerFlushTimeoutMSEnvLabel   = "KAFKA_PRODUCER_FLUSH_TIMEOUT_MS"
	KafkaProducerMessageTimeoutMSEvnLabel = "KAFKA_PRODUCER_MESSAGE_TIMEOUT_MS"
	MaxRetriesEnvLabel                    = "MAX_RETRIES"
	ClickstreamRetryTopicEnvLabel         = "CLICKSTREAM_RETRY_TOPIC"
	ClickstreamRetryS3BucketEnvLabel      = "CLICKSTREAM_RETRY_S3_BUCKET"
	S3PathPrefixEnvLabel                  = "S3_PATH_PREFIX"

	HoustonAPIURLEnvLabel   = "HOUSTON_API_URL"
	HoustonAPIKeyEnvLabel   = "HOUSTON_API_KEY"
	HoustonUserNameEnvLabel = "HOUSTON_USER_NAME"
	HoustonPasswordEnvLabel = "HOUSTON_PASSWORD"
)

var (
	debug = false

	requiredEnvs = []string{
		BootstrapServersEnvLabel,
		HoustonAPIURLEnvLabel,
		TopicEnvLabel,
		GroupIDEnvLabel,
		SSEURLEnvLabel,
		SSEAuthEnvLabel,
	}

	retryRequiredEnvs = []string{
		ClickstreamRetryTopicEnvLabel,
		ClickstreamRetryS3BucketEnvLabel,
	}
)

func Initalize() {
	viper.SetEnvPrefix(Prefix)
	viper.AutomaticEnv()

	// Setup default configs
	setDefaults()

	// If retry logic is enabled, additional env vars are required
	if cmd.EnableRetry {
		requiredEnvs = append(requiredEnvs, retryRequiredEnvs...)
	}

	// Verify required configs
	if err := verifyRequiredEnvVars(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	// Debug value
	debug = viper.GetBool(DebugEnvLabel)

}

func GetString(cfg string) string {
	return viper.GetString(cfg)
}

func GetBool(cfg string) bool {
	return viper.GetBool(cfg)
}

func GetInt(cfg string) int {
	return viper.GetInt(cfg)
}

func setDefaults() {
	viper.SetDefault(DebugEnvLabel, false)
	viper.SetDefault(ServePortEnvLabel, "8080")
	viper.SetDefault(KafkaProducerFlushTimeoutMSEnvLabel, 1000)
	viper.SetDefault(KafkaProducerMessageTimeoutMSEvnLabel, 5000)
	viper.SetDefault(MaxRetriesEnvLabel, 5)
}

func verifyRequiredEnvVars() error {
	errs := []string{}
	for _, envVar := range requiredEnvs {
		if len(GetString(envVar)) == 0 {
			errs = append(errs, pkg.GetRequiredEnvErrorString(Prefix, envVar))
		}
	}

	// For Houston, you must have either the API key OR username AND password
	if len(GetString(HoustonAPIKeyEnvLabel)) == 0 &&
		len(GetString(HoustonUserNameEnvLabel)) == 0 {
		errs = append(errs,
			fmt.Sprintf("%s_%s or %s_%s/%s_%s is required",
				Prefix, HoustonAPIKeyEnvLabel, Prefix, HoustonUserNameEnvLabel, Prefix, HoustonPasswordEnvLabel))
	}

	if len(GetString(HoustonAPIKeyEnvLabel)) != 0 &&
		len(GetString(HoustonUserNameEnvLabel)) != 0 {
		logrus.Warn(fmt.Sprintf("Both %s_%s and %s_%s provided, using %s_%s",
			Prefix, HoustonUserNameEnvLabel, Prefix, HoustonUserNameEnvLabel, Prefix, HoustonAPIKeyEnvLabel))
	}

	if len(errs) != 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

func IsDebugEnabled() bool {
	return debug
}
