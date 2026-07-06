package config

import (
	"testing"
	"time"
)

func TestLoadDefaultsAndEnvOverrides(t *testing.T) {
	t.Setenv("APP_PORT", "18080")
	t.Setenv("PROXY_TIMEOUT_SEC", "2")
	t.Setenv("MOCK_FAIL", "true")
	t.Setenv("RATE_LIMIT_KEY_QPS", "7.5")

	cfg := Load()

	if cfg.AppPort != "18080" {
		t.Fatalf("AppPort = %q", cfg.AppPort)
	}
	if cfg.ProxyTimeout != 2*time.Second {
		t.Fatalf("ProxyTimeout = %v", cfg.ProxyTimeout)
	}
	if !cfg.MockFail {
		t.Fatal("MockFail should be true")
	}
	if cfg.RateLimitKeyQPS != 7.5 {
		t.Fatalf("RateLimitKeyQPS = %v", cfg.RateLimitKeyQPS)
	}
}
