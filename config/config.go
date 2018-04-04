package config

import (
	"fmt"
	"log"
	"reflect"

	"github.com/spf13/viper"
)

// Configuration is a stuct to hold event-router configs
type Configuration struct {
	DebugMode       bool   `mapstructure:"DEBUG_MODE"`
	LogFormat       string `mapstructure:"LOG_FORMAT"`
	APIInterface    string `mapstructure:"API_INTERFACE"`
	APIPort         string `mapstructure:"API_PORT"`
	KafkaBrokers    string `mapstructure:"KAFKA_BROKERS"`
	KafkaGroupID    string `mapstructure:"KAFKA_GROUP_ID"`
	KafkaInputTopic string `mapstructure:"KAFKA_INPUT_TOPIC"`
}

// AppConfig is a global instance of Configuration
var AppConfig Configuration

func init() {
	viper.SetEnvPrefix("ER")
	viper.AutomaticEnv()

	viper.SetDefault("DEBUG_MODE", false)
	viper.SetDefault("LOG_FORMAT", "json")
	viper.SetDefault("API_INTERFACE", "0.0.0.0")
	viper.SetDefault("API_PORT", "8082")
	viper.SetDefault("KAFKA_BROKERS", "")
	viper.SetDefault("KAFKA_GROUP_ID", "ap-event-router")
	viper.SetDefault("KAFKA_INPUT_TOPIC", "")

	AppConfig = Configuration{}
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
}

// Get returns the config
func Get() *Configuration {
	return &AppConfig
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
