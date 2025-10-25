package config

import (
	"fmt"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	MediaDirs  []string
	SabAPIKey  string
	SabHost    string
	Uptime     string
	LogLevel   string
	StatsFile  string
	WebTimeout int
	SabPort    int
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		MediaDirs:  []string{"/mnt/core/pub/cloud/tv", "/mnt/core/pub/cloud/movies"},
		SabAPIKey:  GetEnv("SABAPIKEY", "YOURKEY"),
		SabHost:    GetEnv("SABHOST", "https://sab.zaldre.com"),
		Uptime:     GetEnv("UPTIME", "https://app.statuscake.com/button/index.php?Track=lmmBTReo4c&Days=30&Design=2"),
		LogLevel:   GetEnv("LOGLEVEL", "Normal"),
		StatsFile:  GetEnv("STATSFILE", "/container/data/stats/index.html"),
		WebTimeout: getIntEnv("WEBTIMEOUT", 15),
		SabPort:    getIntEnv("SABPORT", 443),
	}
}

// getIntEnv gets an integer environment variable with a default value
func getIntEnv(key string, defaultValue int) int {
	value := GetEnv(key, fmt.Sprintf("%d", defaultValue))
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}
