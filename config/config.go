package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/astronomerio/event-router/pkg"
	"github.com/astronomerio/viper"
	"github.com/sirupsen/logrus"
)

const (
	// Environment Variable Prefix
	Prefix = "ER"

	DebugEnvLabel = "DEBUG"

	BootstrapServersEnvLabel = "BOOTSTRAP_SERVERS"
	ApplicationIDEnvLabel    = "APPLICATION_ID"
	ServePortEnvLabel        = "SERVE_PORT"
	GroupIDEnvLabel          = "GROUP_ID"
	TopicEnvLabel            = "TOPIC"

	HoustonAPIURLEnvLabel   = "HOUSTON_API_URL"
	HoustonAPIKeyEnvLabel   = "HOUSTON_API_KEY"
	HoustonUserNameEnvLabel = "HOUSTON_USER_NAME"
	HoustonPasswordEnvLabel = "HOUSTON_PASSWORD"
)

var (
	debug = false

	requiredEnvs = []string{
		BootstrapServersEnvLabel,
		ApplicationIDEnvLabel,
		HoustonAPIURLEnvLabel,
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

func GetString(cfg string) string {
	return viper.GetString(cfg)
}

func GetBool(cfg string) bool {
	return viper.GetBool(cfg)
}

func setDefaults() {
	viper.SetDefault(DebugEnvLabel, false)
	viper.SetDefault(ServePortEnvLabel, "8080")
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