package mesos

import (
	"fmt"
	"strings"

	"github.com/astronomerio/asds/pkg"
	"github.com/spf13/viper"
)

var (
	mesosURLEnvLabel  = "MESOS_URL"
	mesosRoleEnvLabel = "MESOS_ROLE"
)

type Config struct {
	MesosURL  string
	MesosRole string
}

var config *Config

func GetMesosConfig() (*Config, error) {
	if config == nil {
		if err := setupMesoConfig(); err != nil {
			return nil, err
		}
	}
	return config, nil
}

func setupMesoConfig() error {
	setDefaults()
	var errors []string
	mesosURL := viper.GetString(mesosURLEnvLabel)
	if len(mesosURL) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(mesosURLEnvLabel))
	}

	mesosRole := viper.GetString(mesosRoleEnvLabel)

	if len(errors) != 0 {
		return fmt.Errorf(strings.Join(errors, ", "))
	}

	config = &Config{
		MesosURL:  mesosURL,
		MesosRole: mesosRole,
	}

	return nil
}

func setDefaults() {
	viper.SetDefault(mesosRoleEnvLabel, "saas")
}
