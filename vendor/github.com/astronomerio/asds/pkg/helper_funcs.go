package pkg

import (
	"fmt"

	"github.com/astronomerio/asds/config"
	"crypto/rand"
	"encoding/base64"
)

func GetRequiredEnvError(env string) error {
	return fmt.Errorf("Env variable %s not set and is required", config.Prefix+"_"+env)
}

func GetRequiredEnvErrorString(env string) string {
	return GetRequiredEnvError(env).Error()
}

func GetFernet() (string, error){
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
