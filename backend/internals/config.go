package internals

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Port            string
	DBPath          string
	HistoryLimit    int
	MaxContentSize  int64
	CleanupInterval time.Duration
	SnippetTTL      time.Duration
	MaxConnections  int
}

// LoadConfig reads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		DBPath:          getEnv("DB_PATH", "./clipsync.db"),
		HistoryLimit:    getEnvInt("HISTORY_LIMIT", 50),
		MaxContentSize:  getEnvInt64("MAX_CONTENT_SIZE", 10*1024*1024), // 10MB
		CleanupInterval: getEnvDuration("CLEANUP_INTERVAL", 1*time.Hour),
		SnippetTTL:      getEnvDuration("SNIPPET_TTL", 24*time.Hour),
		MaxConnections:  getEnvInt("MAX_CONNECTIONS", 10000),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
