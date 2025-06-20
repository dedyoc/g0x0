package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL   string
	StoragePath   string
	MaxFileSize   int64
	MaxExpiration time.Duration
	MinExpiration time.Duration
	SecretBytes   int
	URLAlphabet   string
}

func Load() *Config {
	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://g0x0:g0x0@localhost:5432/g0x0?sslmode=disable"),
		StoragePath:   getEnv("STORAGE_PATH", "./uploads"),
		MaxFileSize:   getEnvInt64("MAX_FILE_SIZE", 256*1024*1024),        // 256MB
		MaxExpiration: getEnvDuration("MAX_EXPIRATION", 365*24*time.Hour), // 1 year
		MinExpiration: getEnvDuration("MIN_EXPIRATION", 30*24*time.Hour),  // 30 days
		SecretBytes:   getEnvInt("SECRET_BYTES", 16),
		URLAlphabet:   getEnv("URL_ALPHABET", "DEQhd2uFteibPwq0SWBInTpA_jcZL5GKz3YCR14Ulk87Jors9vNHgfaOmMXy6Vx-"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
