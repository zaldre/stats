package config

import (
	"os"
)

// getEnv retrieves an environment variable or returns a default value
// func getEnv(key string, defaultVal T) T {
// getEnv retrieves an environment variable or returns a default value
func GetEnv(key string, defaultVal string) string {
	var value string
	if value = os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}
