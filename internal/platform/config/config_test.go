package config

import (
	"strings"
	"testing"
)

func TestLoadRequiresMySQLCredentials(t *testing.T) {
	_, err := Load(env(map[string]string{
		"MYSQL_USER":     "app",
		"MYSQL_DATABASE": "alumni",
	}))
	if err == nil || !strings.Contains(err.Error(), "MYSQL_PASSWORD is required") {
		t.Fatalf("expected missing password error, got %v", err)
	}
}

func TestLoadUsesSafeLocalDefaults(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"MYSQL_USER":     "app",
		"MYSQL_PASSWORD": "test-password",
		"MYSQL_DATABASE": "alumni",
	}))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:3000" || cfg.MySQLHost != "127.0.0.1" || cfg.MySQLPort != "3306" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadRejectsInvalidHTTPAddress(t *testing.T) {
	_, err := Load(env(map[string]string{
		"HTTP_ADDR":      ":not-a-port",
		"MYSQL_USER":     "app",
		"MYSQL_PASSWORD": "test-password",
		"MYSQL_DATABASE": "alumni",
	}))
	if err == nil || !strings.Contains(err.Error(), "HTTP_ADDR must use a valid TCP port") {
		t.Fatalf("expected invalid HTTP address error, got %v", err)
	}
}

func TestLoadRejectsInvalidMySQLPort(t *testing.T) {
	_, err := Load(env(map[string]string{
		"MYSQL_PORT":     "70000",
		"MYSQL_USER":     "app",
		"MYSQL_PASSWORD": "test-password",
		"MYSQL_DATABASE": "alumni",
	}))
	if err == nil || !strings.Contains(err.Error(), "MYSQL_PORT must be a valid TCP port") {
		t.Fatalf("expected invalid MySQL port error, got %v", err)
	}
}

func TestLoadAllowsLocalDemoAuthOnlyOnLoopback(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"HTTP_ADDR":          "127.0.0.1:3000",
		"MYSQL_USER":         "app",
		"MYSQL_PASSWORD":     "test-password",
		"MYSQL_DATABASE":     "alumni",
		"DEMO_AUTH_ENABLED":  "true",
		"DEMO_AUTH_PASSWORD": "local-test-password",
	}))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.DemoAuthEnabled {
		t.Fatal("DemoAuthEnabled = false, want true")
	}
}

func TestLoadRejectsPublicDemoAuth(t *testing.T) {
	_, err := Load(env(map[string]string{
		"HTTP_ADDR":          "0.0.0.0:3000",
		"MYSQL_USER":         "app",
		"MYSQL_PASSWORD":     "test-password",
		"MYSQL_DATABASE":     "alumni",
		"DEMO_AUTH_ENABLED":  "true",
		"DEMO_AUTH_PASSWORD": "local-test-password",
	}))
	if err == nil || !strings.Contains(err.Error(), "requires HTTP_ADDR to bind to loopback") {
		t.Fatalf("expected public demo auth to be rejected, got %v", err)
	}
}

func TestLoadRequiresTurnstileForConfiguredOAuth(t *testing.T) {
	_, err := Load(env(map[string]string{
		"MYSQL_USER":                 "app",
		"MYSQL_PASSWORD":             "test-password",
		"MYSQL_DATABASE":             "alumni",
		"PUBLIC_BASE_URL":            "https://alumni.example",
		"GOOGLE_OAUTH_CLIENT_ID":     "client-id",
		"GOOGLE_OAUTH_CLIENT_SECRET": "client-secret",
	}))
	if err == nil || !strings.Contains(err.Error(), "CAPTCHA must be configured") {
		t.Fatalf("expected OAuth without CAPTCHA to be rejected, got %v", err)
	}
}

func env(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
