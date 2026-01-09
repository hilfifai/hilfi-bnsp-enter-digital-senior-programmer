package helper

import (
	"api-jellyfish/internal/pkg/logger/v2"
	"fmt"
	"os"
	"strconv"
)

func GetEnv(key string) string {
	return os.Getenv(key)
}

func GetEnvRequired(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("environment variable %s not set", key)
	}
	return val, nil
}

func GetEnvDefault(key string, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	logger.Log.Debug().Msgf("Environment variable %s not set, using default value: %s\n", key, fallback)
	return fallback
}

func GetEnvAsBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	logger.Log.Debug().Msgf("Environment variable %s not set or invalid, using default value: %t\n", key, defaultVal)
	return defaultVal
}

func GetEnvAsInt(name string, defaultVal int) int {
	if val, ok := os.LookupEnv(name); ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	logger.Log.Debug().Msgf("Environment variable %s not set or invalid, using default value: %d\n", name, defaultVal)
	return defaultVal
}
