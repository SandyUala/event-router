package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/astronomerio/clickstream-event-router/pkg"
	"github.com/astronomerio/viper"
	"github.com/sirupsen/logrus"
)

const (
	// Environment Variable Prefix
	Prefix = "ER"

	Debug = "DEBUG"

	BootstrapServers              = "BOOTSTRAP_SERVERS"
	ServePort                     = "SERVE_PORT"
	KafkaGroupID                  = "KAFKA_GROUP_ID"
	KafkaIngestionTopic           = "KAFKA_INGESTION_TOPIC"
	SSEURL                        = "SSE_URL"
	KafkaProducerFlushTimeoutMS   = "KAFKA_PRODUCER_FLUSH_TIMEOUT_MS"
	KafkaProducerMessageTimeoutMS = "KAFKA_PRODUCER_MESSAGE_TIMEOUT_MS"
	MaxRetries                    = "MAX_RETRIES"
	ClickstreamRetryS3Bucket      = "CLICKSTREAM_RETRY_S3_BUCKET"
	ClickstreamRetryS3PathPrefix  = "CLICKSTREAM_RETRY_S3_PATH_PREFIX"
	ClickstreamRetryFlushTimeout  = "CLICKSTREAM_RETRY_TIMEOUT"
	ClickstreamRetryMaxQueue      = "CLICKSTREAM_RETRY_MAX_QUEUE"

	HoustonAPIURL   = "HOUSTON_API_URL"
	HoustonAPIKey   = "HOUSTON_API_KEY"
	HoustonUserName = "HOUSTON_USERNAME"
	HoustonPassword = "HOUSTON_PASSWORD"

	// Non Env Variable configs
	KafakDebug = "KAFKA_DEBUG"
	Retry      = "RETRY"
)

var (
	debug = false

	requiredEnvs = []string{
		BootstrapServers,
		HoustonAPIURL,
		KafkaIngestionTopic,
		KafkaGroupID,
		SSEURL,
	}

	retryRequiredEnvs = []string{
		ClickstreamRetryS3Bucket,
		ClickstreamRetryFlushTimeout,
		ClickstreamRetryMaxQueue,
	}

	allEnvs = []string{
		Debug,
		BootstrapServers,
		ServePort,
		KafkaGroupID,
		KafkaIngestionTopic,
		SSEURL,
		KafkaProducerFlushTimeoutMS,
		KafkaProducerMessageTimeoutMS,
		MaxRetries,
		ClickstreamRetryS3Bucket,
		ClickstreamRetryS3PathPrefix,
		HoustonAPIURL,
		HoustonUserName,
	}
)

type InitOptions struct {
	EnableRetry bool
}

func Initalize(opts *InitOptions) {
	viper.SetEnvPrefix(Prefix)
	viper.AutomaticEnv()

	// Setup default configs
	setDefaults()

	// If retry logic is enabled, additional env vars are required
	if opts.EnableRetry {
		requiredEnvs = append(requiredEnvs, retryRequiredEnvs...)
	}

	// Verify required configs
	if err := verifyRequiredEnvVars(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	// Debug value
	debug = viper.GetBool(Debug)

	// Print out the configs
	printConfigs()

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

func GetInt64(cfg string) int64 {
	return viper.GetInt64(cfg)
}

func SetBool(cfg string, val bool) {
	viper.Set(cfg, val)
}

func setDefaults() {
	viper.SetDefault(Debug, false)
	viper.SetDefault(ServePort, "8080")
	viper.SetDefault(KafkaProducerFlushTimeoutMS, 1000)
	viper.SetDefault(KafkaProducerMessageTimeoutMS, 5000)
	viper.SetDefault(MaxRetries, 2)
}

func verifyRequiredEnvVars() error {
	errs := []string{}
	for _, envVar := range requiredEnvs {
		if len(GetString(envVar)) == 0 {
			errs = append(errs, pkg.GetRequiredEnvErrorString(Prefix, envVar))
		}
	}

	// For Houston, you must have either the API key OR username AND password
	if len(GetString(HoustonAPIKey)) == 0 &&
		len(GetString(HoustonUserName)) == 0 {
		errs = append(errs,
			fmt.Sprintf("%s_%s or %s_%s/%s_%s is required",
				Prefix, HoustonAPIKey, Prefix, HoustonUserName, Prefix, HoustonPassword))
	}

	if len(GetString(HoustonAPIKey)) != 0 &&
		len(GetString(HoustonUserName)) != 0 {
		logrus.Warn(fmt.Sprintf("Both %s_%s and %s_%s provided, using %s_%s",
			Prefix, HoustonUserName, Prefix, HoustonUserName, Prefix, HoustonAPIKey))
	}

	if len(errs) != 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

func IsDebugEnabled() bool {
	return debug
}

func printConfigs() {
	fmt.Println("Initializing with the following configs:")
	fmt.Println("========================================")
	for _, envVar := range allEnvs {
		fmt.Printf("%s_%s = %s\n", Prefix, envVar, GetString(envVar))
	}
	fmt.Println("========================================")
}
