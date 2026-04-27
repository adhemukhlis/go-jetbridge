package config

import (
	"os"
	"strconv"
)

// GetEnv retrieves the value of the environment variable named by the key.
// It returns the value, which will be the default value if the variable is not present.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// GetEnvInt retrieves the value of the environment variable named by the key and parses it as an int.
// It returns the value, which will be the default value if the variable is not present or cannot be parsed.
func GetEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
