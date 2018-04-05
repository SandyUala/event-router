package config

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"
)

// AppConfig is a global instance of Configuration
var (
	AppConfig *Configuration = &Configuration{}
)

// Configuration is a stuct to hold event-router configs
type Configuration struct {
	DebugMode             bool   `mapstructure:"DEBUG_MODE"`
	LogFormat             string `mapstructure:"LOG_FORMAT"`
	APIInterface          string `mapstructure:"API_INTERFACE"`
	APIPort               string `mapstructure:"API_PORT"`
	KafkaBrokers          string `mapstructure:"KAFKA_BROKERS"`
	KafkaGroupID          string `mapstructure:"KAFKA_GROUP_ID"`
	KafkaInputTopic       string `mapstructure:"KAFKA_INPUT_TOPIC"`
	IntegerationConfigDir string `mapstructure:"INTEGRATION_CONFIG_DIR"`
}

func init() {
	appConfig := viper.New()
	appConfig.SetEnvPrefix("ER")
	appConfig.AutomaticEnv()

	appConfig.SetDefault("DEBUG_MODE", false)
	appConfig.SetDefault("LOG_FORMAT", "json")
	appConfig.SetDefault("API_INTERFACE", "0.0.0.0")
	appConfig.SetDefault("API_PORT", "8081")
	appConfig.SetDefault("KAFKA_BROKERS", "")
	appConfig.SetDefault("KAFKA_GROUP_ID", "ap-event-router")
	appConfig.SetDefault("KAFKA_INPUT_TOPIC", "")
	appConfig.SetDefault("INTEGRATION_CONFIG_DIR", "/etc/astronomer/event-router/integrations")

	if err := appConfig.Unmarshal(AppConfig); err != nil {
		fmt.Errorf("Unable to decode into struct, %v", err)
	}
}

// Get returns the config
func Get() *Configuration {
	return AppConfig
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
