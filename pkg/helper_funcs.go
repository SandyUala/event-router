package pkg

import (
	"fmt"
)

func GetRequiredEnvError(prefix, env string) error {
	return fmt.Errorf("%s not set and is required", prefix+"_"+env)
}

func GetRequiredEnvErrorString(prefix, env string) string {
	return GetRequiredEnvError(prefix, env).Error()
}
