package config

import (
	"fmt"
	"log"
	"reflect"

	"github.com/spf13/viper"
)

// Configuration is a stuct to hold event-router configs
type Configuration struct {
	DebugMode bool
	LogFormat string
}

// AppConfig is a global instance of Configuration
var AppConfig Configuration

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	AppConfig = Configuration{}

	setDefaults()
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Failed reading config file: %s\n", err)
	}

	viper.SetEnvPrefix("event-router")
	viper.AutomaticEnv()

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
}

func setDefaults() {
	viper.SetDefault("DebugMode", false)
	viper.SetDefault("LogFormat", "json")
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
