package config

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"
)

// AppConfig is the gloabl application configuration
var AppConfig = &Configuration{}

// Configuration is a stuct to hold event-router configs
type Configuration struct {
	DebugMode                   bool   `mapstructure:"DEBUG_MODE"`
	LogFormat                   string `mapstructure:"LOG_FORMAT"`
	APIInterface                string `mapstructure:"API_INTERFACE"`
	APIPort                     string `mapstructure:"API_PORT"`
	KafkaBrokers                string `mapstructure:"KAFKA_BROKERS"`
	KafkaConsumerTopic          string `mapstructure:"KAFKA_CONSUMER_TOPIC"`
	KafkaConsumerGroupID        string `mapstructure:"KAFKA_CONSUMER_GROUP_ID"`
	KafkaProducerMessageTimeout int    `mapstructure:"KAFKA_PRODUCER_MESSAGE_TIMEOUT"`
	KafkaProducerFlushTimeout   int    `mapstructure:"KAFKA_PRODUCER_FLUSH_TIMEOUT"`
	IntegerationConfigDir       string `mapstructure:"INTEGRATION_CONFIG_DIR"`
}

func init() {
	appViper := viper.New()
	appViper.SetEnvPrefix("ER")
	appViper.AutomaticEnv()

	appViper.SetDefault("DEBUG_MODE", false)
	appViper.SetDefault("LOG_FORMAT", "json")
	appViper.SetDefault("API_INTERFACE", "0.0.0.0")
	appViper.SetDefault("API_PORT", "8081")
	appViper.SetDefault("KAFKA_BROKERS", "")
	appViper.SetDefault("KAFKA_CONSUMER_GROUP_ID", "ap-event-router")
	appViper.SetDefault("KAFKA_CONSUMER_TOPIC", "")
	appViper.SetDefault("KAFKA_PRODUCER_MESSAGE_TIMEOUT", 5000)
	appViper.SetDefault("KAFKA_PRODUCER_FLUSH_TIMEOUT", 1000)
	appViper.SetDefault("INTEGRATION_CONFIG_DIR", "/etc/astronomer/event-router/integrations")

	if err := appViper.Unmarshal(AppConfig); err != nil {
		fmt.Errorf("Unable to decode into struct, %v", err)
	}
}

// Print prints the configuration to stdout
func (c *Configuration) Print() {
	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	fmt.Println("=============== Configuration ===============")
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fmt.Printf("%s %s = %v\n", t.Field(i).Name, f.Type(), f.Interface())
	}
	fmt.Println("=============================================")
}
