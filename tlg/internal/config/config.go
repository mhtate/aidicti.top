package config

import (
	"os"
	"time"
)

type Config struct {
	Port        string
	CacheExpiry time.Duration
}

func LoadConfig() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		CacheExpiry: getEnvAsDuration("CACHE_EXPIRY", 5*time.Minute),
	}
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
