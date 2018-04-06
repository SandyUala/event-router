package config

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"
)

// IntegrationConfig is the collective configuration of all integrations
var IntegrationConfig = &IntegrationConfiguration{}

// IntegrationConfiguration is a stuct to hold event-router configs
type IntegrationConfiguration struct {
	Applications []Application `yaml:"applications"`
}

// Print prints the configuration to stdout
func (ic *IntegrationConfiguration) Print() {
	v := reflect.ValueOf(ic).Elem()
	t := v.Type()

	fmt.Println("========= Integration Configuration =========")
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fmt.Printf("%s %s = %v\n", t.Field(i).Name, f.Type(), f.Interface())
	}
	fmt.Println("=============================================")
}

// EnabledIntegrations returns the enabled integrations for a given writeKey
func (ic *IntegrationConfiguration) EnabledIntegrations(writeKey string) (names []string) {
	for _, app := range ic.Applications {
		if app.WriteKey == writeKey {
			for _, integration := range app.Integrations {
				names = append(names, integration.Name)
			}
		}
	}
	return
}

// Application is a configured application
type Application struct {
	WriteKey     string        `yaml:"writeKey"`
	Integrations []Integration `yaml:"integrations"`
}

// Integration is a group of settings for an integration
type Integration struct {
	Name   string                 `yaml:"name"`
	Config map[string]interface{} `yaml:"config"`
}

func init() {
	intViper := viper.New()
	intViper.SetConfigName("integrations")
	intViper.SetConfigType("yaml")
	intViper.AutomaticEnv()
	intViper.AddConfigPath(AppConfig.IntegerationConfigDir)

	if err := intViper.ReadInConfig(); err != nil {
		fmt.Printf("Failed reading config file: %s\n", err)
	}

	if err := intViper.Unmarshal(IntegrationConfig); err != nil {
		fmt.Errorf("Unable to decode into struct, %v", err)
	}
}
