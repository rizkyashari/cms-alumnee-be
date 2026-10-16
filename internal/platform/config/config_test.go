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

func TestLoadRequiresCompleteEmailOTPAndBootstrapConfiguration(t *testing.T) {
	base := map[string]string{
		"MYSQL_USER": "app", "MYSQL_PASSWORD": "test-password", "MYSQL_DATABASE": "alumni",
		"RESEND_API_KEY": "re_test_key",
	}
	if _, err := Load(env(base)); err == nil || !strings.Contains(err.Error(), "RESEND_API_KEY and RESEND_FROM_EMAIL") {
		t.Fatalf("expected incomplete email sender error, got %v", err)
	}
	valid := map[string]string{
		"MYSQL_USER": "app", "MYSQL_PASSWORD": "test-password", "MYSQL_DATABASE": "alumni",
		"RESEND_API_KEY": "re_test_key", "RESEND_FROM_EMAIL": "otp@example.test",
		"PUBLIC_BASE_URL":       "https://alumni.example",
		"BOOTSTRAP_ADMIN_EMAIL": "admin@example.test", "BOOTSTRAP_ADMIN_NAME": "Initial Admin",
	}
	cfg, err := Load(env(valid))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.BootstrapAdminEmail != "admin@example.test" || cfg.BootstrapAdminName != "Initial Admin" {
		t.Fatalf("bootstrap admin configuration not loaded: %+v", cfg)
	}
}

func TestLoadRequiresPublicOriginForEmailOTP(t *testing.T) {
	_, err := Load(env(map[string]string{
		"MYSQL_USER": "app", "MYSQL_PASSWORD": "test-password", "MYSQL_DATABASE": "alumni",
		"RESEND_API_KEY": "re_test_key", "RESEND_FROM_EMAIL": "otp@example.test",
	}))
	if err == nil || !strings.Contains(err.Error(), "PUBLIC_BASE_URL is required") {
		t.Fatalf("expected public-origin configuration error, got %v", err)
	}
}

func TestLoadRejectsMalformedPublicBaseURL(t *testing.T) {
	_, err := Load(env(map[string]string{
		"MYSQL_USER": "app", "MYSQL_PASSWORD": "test-password", "MYSQL_DATABASE": "alumni",
		"RESEND_API_KEY": "re_test_key", "RESEND_FROM_EMAIL": "otp@example.test",
		"PUBLIC_BASE_URL": "%",
	}))
	if err == nil || !strings.Contains(err.Error(), "PUBLIC_BASE_URL must be a valid URL") {
		t.Fatalf("expected malformed public URL error, got %v", err)
	}
}

func TestLoadRejectsAdminBootstrapWithoutEmailProvider(t *testing.T) {
	_, err := Load(env(map[string]string{
		"MYSQL_USER": "app", "MYSQL_PASSWORD": "test-password", "MYSQL_DATABASE": "alumni",
		"BOOTSTRAP_ADMIN_EMAIL": "admin@example.test", "BOOTSTRAP_ADMIN_NAME": "Initial Admin",
	}))
	if err == nil || !strings.Contains(err.Error(), "email OTP must be configured") {
		t.Fatalf("expected bootstrap email-provider error, got %v", err)
	}
}

func TestLoadAllowsFourStagingAccounts(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"MYSQL_USER": "app", "MYSQL_PASSWORD": "test-password", "MYSQL_DATABASE": "alumni",
		"STAGING_AUTH_NON_VERIFIED_USERNAME": "guest", "STAGING_AUTH_NON_VERIFIED_PASSWORD": "guest-password-123",
		"STAGING_AUTH_VERIFIED_USERNAME": "member", "STAGING_AUTH_VERIFIED_PASSWORD": "member-password-123",
		"STAGING_AUTH_OFFICER_USERNAME": "officer", "STAGING_AUTH_OFFICER_PASSWORD": "officer-password-123",
		"STAGING_AUTH_ADMIN_USERNAME": "admin", "STAGING_AUTH_ADMIN_PASSWORD": "admin-password-123",
	}))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.StagingAdminUsername != "admin" || cfg.StagingAdminPassword != "admin-password-123" {
		t.Fatal("staging account credentials were not loaded")
	}
}

func TestLoadRejectsPartialStagingAccounts(t *testing.T) {
	_, err := Load(env(map[string]string{
		"MYSQL_USER": "app", "MYSQL_PASSWORD": "test-password", "MYSQL_DATABASE": "alumni",
		"STAGING_AUTH_ADMIN_USERNAME": "admin",
	}))
	if err == nil || !strings.Contains(err.Error(), "configured with its password") {
		t.Fatalf("expected incomplete account error, got %v", err)
	}
}

func TestLoadRejectsDuplicateOrWeakStagingCredentials(t *testing.T) {
	base := map[string]string{
		"MYSQL_USER": "app", "MYSQL_PASSWORD": "test-password", "MYSQL_DATABASE": "alumni",
		"STAGING_AUTH_NON_VERIFIED_USERNAME": "same", "STAGING_AUTH_NON_VERIFIED_PASSWORD": "guest-password-123",
		"STAGING_AUTH_VERIFIED_USERNAME": "member", "STAGING_AUTH_VERIFIED_PASSWORD": "member-password-123",
		"STAGING_AUTH_OFFICER_USERNAME": "officer", "STAGING_AUTH_OFFICER_PASSWORD": "officer-password-123",
		"STAGING_AUTH_ADMIN_USERNAME": "admin", "STAGING_AUTH_ADMIN_PASSWORD": "admin-password-123",
	}
	duplicate := make(map[string]string, len(base))
	for key, value := range base {
		duplicate[key] = value
	}
	duplicate["STAGING_AUTH_ADMIN_USERNAME"] = "SAME"
	if _, err := Load(env(duplicate)); err == nil || !strings.Contains(err.Error(), "must be unique") {
		t.Fatalf("expected duplicate username error, got %v", err)
	}

	weak := make(map[string]string, len(base))
	for key, value := range base {
		weak[key] = value
	}
	weak["STAGING_AUTH_ADMIN_PASSWORD"] = "short"
	if _, err := Load(env(weak)); err == nil || !strings.Contains(err.Error(), "at least 12 characters") {
		t.Fatalf("expected weak password error, got %v", err)
	}
}

func env(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
