package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppPort            string
	AdminToken         string
	MySQLDSN           string
	RedisAddr          string
	RedisDB            int
	ProxyTimeout       time.Duration
	MockLatency        time.Duration
	MockFail           bool
	KeyCacheTTL        time.Duration
	RateLimitGlobalQPS float64
	RateLimitKeyQPS    float64
	RateLimitTenantQPS float64
	GinMode            string
}

func Load() Config {
	return Config{
		AppPort:            getEnv("APP_PORT", "8080"),
		AdminToken:         getEnv("ADMIN_TOKEN", "admin-dev-token"),
		MySQLDSN:           getEnv("MYSQL_DSN", "root:root@tcp(localhost:3306)/ai_gateway?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisDB:            getEnvInt("REDIS_DB", 0),
		ProxyTimeout:       time.Duration(getEnvInt("PROXY_TIMEOUT_SEC", 30)) * time.Second,
		MockLatency:        time.Duration(getEnvInt("MOCK_LATENCY_MS", 0)) * time.Millisecond,
		MockFail:           getEnvBool("MOCK_FAIL", false),
		KeyCacheTTL:        time.Duration(getEnvInt("KEY_CACHE_TTL_SEC", 300)) * time.Second,
		RateLimitGlobalQPS: getEnvFloat("RATE_LIMIT_GLOBAL_QPS", 100),
		RateLimitKeyQPS:    getEnvFloat("RATE_LIMIT_KEY_QPS", 20),
		RateLimitTenantQPS: getEnvFloat("RATE_LIMIT_TENANT_QPS", 50),
		GinMode:            getEnv("GIN_MODE", "release"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvFloat(key string, fallback float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
