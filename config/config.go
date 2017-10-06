package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/astronomerio/asds/pkg"
	"github.com/astronomerio/viper"
	"github.com/sirupsen/logrus"
)

const (
	// Environment Variable Prefix
	Prefix = "ER"

	DebugEnvLabel = "DEBUG"

	BootstrapServersEnvLabel = "BOOTSTRAP_SERVERS"
	ApplicationIDEnvLabel    = "APPLICATION_ID"
)

var (
	debug = false

	requiredEnvs = []string{
		BootstrapServersEnvLabel,
		ApplicationIDEnvLabel,
	}
)

func init() {
	viper.SetEnvPrefix(Prefix)
	viper.AutomaticEnv()

	// Setup default configs
	setDefaults()

	// Verify required configs
	if err := verifyRequiredEnvVars(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	// Debug value
	debug = viper.GetBool(DebugEnvLabel)

}

func GetConfig(cfg string) string {
	return viper.GetString(cfg)
}

func setDefaults() {
	viper.SetDefault(DebugEnvLabel, false)
}

func verifyRequiredEnvVars() error {
	errs := []string{}
	for _, envVar := range requiredEnvs {
		if len(GetConfig(envVar)) == 0 {
			errs = append(errs, pkg.GetRequiredEnvErrorString(envVar))
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf(strings.Join(errs, ", "))
	}
	return nil
}

func IsDebugEnabled() bool {
	return debug
}
