package pkg

import (
	"fmt"

	"github.com/astronomerio/asds/config"
)

func GetRequiredEnvError(env string) error {
	return fmt.Errorf("Env variable %s not set and is required", config.Prefix+"_"+env)
}

func GetRequiredEnvErrorString(prefix, env string) string {
	return GetRequiredEnvError(env).Error()
}
