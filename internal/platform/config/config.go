package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
)

type Config struct {
	HTTPAddr         string
	MySQLHost        string
	MySQLPort        string
	MySQLUser        string
	MySQLPassword    string
	MySQLDatabase    string
	DemoAuthEnabled  bool
	DemoAuthPassword string
	PublicBaseURL    string
	TurnstileSiteKey string
	TurnstileSecret  string
	GoogleClientID   string
	GoogleSecret     string
	FacebookClientID string
	FacebookSecret   string
}

func Load(lookup func(string) (string, bool)) (Config, error) {
	cfg := Config{
		HTTPAddr:         valueOrDefault(lookup, "HTTP_ADDR", "127.0.0.1:3000"),
		MySQLHost:        valueOrDefault(lookup, "MYSQL_HOST", "127.0.0.1"),
		MySQLPort:        valueOrDefault(lookup, "MYSQL_PORT", "3306"),
		MySQLUser:        value(lookup, "MYSQL_USER"),
		MySQLPassword:    value(lookup, "MYSQL_PASSWORD"),
		MySQLDatabase:    value(lookup, "MYSQL_DATABASE"),
		DemoAuthEnabled:  value(lookup, "DEMO_AUTH_ENABLED") == "true",
		DemoAuthPassword: value(lookup, "DEMO_AUTH_PASSWORD"),
		PublicBaseURL:    value(lookup, "PUBLIC_BASE_URL"),
		TurnstileSiteKey: value(lookup, "TURNSTILE_SITE_KEY"),
		TurnstileSecret:  value(lookup, "TURNSTILE_SECRET_KEY"),
		GoogleClientID:   value(lookup, "GOOGLE_OAUTH_CLIENT_ID"),
		GoogleSecret:     value(lookup, "GOOGLE_OAUTH_CLIENT_SECRET"),
		FacebookClientID: value(lookup, "FACEBOOK_OAUTH_CLIENT_ID"),
		FacebookSecret:   value(lookup, "FACEBOOK_OAUTH_CLIENT_SECRET"),
	}

	_, portText, err := net.SplitHostPort(cfg.HTTPAddr)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_ADDR must be a host:port address")
	}
	httpPort, err := strconv.Atoi(portText)
	if err != nil || httpPort < 1 || httpPort > 65535 {
		return Config{}, fmt.Errorf("HTTP_ADDR must use a valid TCP port")
	}
	port, err := strconv.Atoi(cfg.MySQLPort)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("MYSQL_PORT must be a valid TCP port")
	}
	if cfg.MySQLHost == "" {
		return Config{}, fmt.Errorf("MYSQL_HOST is required")
	}
	if cfg.MySQLUser == "" {
		return Config{}, fmt.Errorf("MYSQL_USER is required")
	}
	if cfg.MySQLPassword == "" {
		return Config{}, fmt.Errorf("MYSQL_PASSWORD is required")
	}
	if cfg.MySQLDatabase == "" {
		return Config{}, fmt.Errorf("MYSQL_DATABASE is required")
	}
	if cfg.DemoAuthEnabled {
		host, _, err := net.SplitHostPort(cfg.HTTPAddr)
		ip := net.ParseIP(host)
		if err != nil || (host != "localhost" && (ip == nil || !ip.IsLoopback())) {
			return Config{}, fmt.Errorf("DEMO_AUTH_ENABLED requires HTTP_ADDR to bind to loopback")
		}
		if len(cfg.DemoAuthPassword) < 12 {
			return Config{}, fmt.Errorf("DEMO_AUTH_PASSWORD must contain at least 12 characters when demo auth is enabled")
		}
	}
	if (cfg.TurnstileSiteKey == "") != (cfg.TurnstileSecret == "") {
		return Config{}, fmt.Errorf("TURNSTILE_SITE_KEY and TURNSTILE_SECRET_KEY must be configured together")
	}
	if (cfg.GoogleClientID == "") != (cfg.GoogleSecret == "") {
		return Config{}, fmt.Errorf("GOOGLE_OAUTH_CLIENT_ID and GOOGLE_OAUTH_CLIENT_SECRET must be configured together")
	}
	if (cfg.FacebookClientID == "") != (cfg.FacebookSecret == "") {
		return Config{}, fmt.Errorf("FACEBOOK_OAUTH_CLIENT_ID and FACEBOOK_OAUTH_CLIENT_SECRET must be configured together")
	}
	if cfg.GoogleClientID != "" || cfg.FacebookClientID != "" || cfg.TurnstileSiteKey != "" {
		publicURL, err := url.Parse(cfg.PublicBaseURL)
		localHost := publicURL.Hostname() == "localhost"
		if ip := net.ParseIP(publicURL.Hostname()); ip != nil && ip.IsLoopback() {
			localHost = true
		}
		if err != nil || publicURL.Hostname() == "" || (publicURL.Scheme != "https" && !(publicURL.Scheme == "http" && localHost)) {
			return Config{}, fmt.Errorf("PUBLIC_BASE_URL must use HTTPS except for localhost development")
		}
		if publicURL.Path != "" && publicURL.Path != "/" || publicURL.RawQuery != "" || publicURL.Fragment != "" {
			return Config{}, fmt.Errorf("PUBLIC_BASE_URL must be an origin without path, query, or fragment")
		}
	}
	if (cfg.GoogleClientID != "" || cfg.FacebookClientID != "") && cfg.TurnstileSecret == "" {
		return Config{}, fmt.Errorf("CAPTCHA must be configured when an OAuth provider is enabled")
	}

	return cfg, nil
}

func value(lookup func(string) (string, bool), key string) string {
	result, _ := lookup(key)
	return result
}

func valueOrDefault(lookup func(string) (string, bool), key, fallback string) string {
	if result := value(lookup, key); result != "" {
		return result
	}
	return fallback
}
