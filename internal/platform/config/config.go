package config

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr                   string
	MySQLHost                  string
	MySQLPort                  string
	MySQLUser                  string
	MySQLPassword              string
	MySQLDatabase              string
	DemoAuthEnabled            bool
	DemoAuthPassword           string
	StagingNonVerifiedUsername string
	StagingNonVerifiedPassword string
	StagingVerifiedUsername    string
	StagingVerifiedPassword    string
	StagingOfficerUsername     string
	StagingOfficerPassword     string
	StagingAdminUsername       string
	StagingAdminPassword       string
	PublicBaseURL              string
	TurnstileSiteKey           string
	TurnstileSecret            string
	GoogleClientID             string
	GoogleSecret               string
	FacebookClientID           string
	FacebookSecret             string
	ResendAPIKey               string
	ResendFromEmail            string
	BootstrapAdminEmail        string
	BootstrapAdminName         string
	RegencyMasterPath          string
	ScenarioPassword           string
	ScenarioReset              bool
	LoginRateLimitPerMinute    int
	FileStorageDir             string
	// ScenarioUsernames follows the order of the scenario cast: Admin, two
	// Pengurus, then seven Non-verified members.
	ScenarioUsernames []string
}

var scenarioUsernameKeys = []struct{ key, fallback string }{
	{"STAGING_SCENARIO_ADMIN_USERNAME", "admin.uji"},
	{"STAGING_SCENARIO_OFFICER_1_USERNAME", "pengurus.satu"},
	{"STAGING_SCENARIO_OFFICER_2_USERNAME", "pengurus.dua"},
	{"STAGING_SCENARIO_MEMBER_1_USERNAME", "alumni.satu"},
	{"STAGING_SCENARIO_MEMBER_2_USERNAME", "alumni.dua"},
	{"STAGING_SCENARIO_MEMBER_3_USERNAME", "alumni.tiga"},
	{"STAGING_SCENARIO_MEMBER_4_USERNAME", "alumni.empat"},
	{"STAGING_SCENARIO_MEMBER_5_USERNAME", "alumni.lima"},
	{"STAGING_SCENARIO_MEMBER_6_USERNAME", "alumni.enam"},
	{"STAGING_SCENARIO_MEMBER_7_USERNAME", "alumni.tujuh"},
}

func Load(lookup func(string) (string, bool)) (Config, error) {
	cfg := Config{
		HTTPAddr:                   valueOrDefault(lookup, "HTTP_ADDR", "127.0.0.1:3000"),
		MySQLHost:                  valueOrDefault(lookup, "MYSQL_HOST", "127.0.0.1"),
		MySQLPort:                  valueOrDefault(lookup, "MYSQL_PORT", "3306"),
		MySQLUser:                  value(lookup, "MYSQL_USER"),
		MySQLPassword:              value(lookup, "MYSQL_PASSWORD"),
		MySQLDatabase:              value(lookup, "MYSQL_DATABASE"),
		DemoAuthEnabled:            value(lookup, "DEMO_AUTH_ENABLED") == "true",
		DemoAuthPassword:           value(lookup, "DEMO_AUTH_PASSWORD"),
		StagingNonVerifiedUsername: value(lookup, "STAGING_AUTH_NON_VERIFIED_USERNAME"),
		StagingNonVerifiedPassword: value(lookup, "STAGING_AUTH_NON_VERIFIED_PASSWORD"),
		StagingVerifiedUsername:    value(lookup, "STAGING_AUTH_VERIFIED_USERNAME"),
		StagingVerifiedPassword:    value(lookup, "STAGING_AUTH_VERIFIED_PASSWORD"),
		StagingOfficerUsername:     value(lookup, "STAGING_AUTH_OFFICER_USERNAME"),
		StagingOfficerPassword:     value(lookup, "STAGING_AUTH_OFFICER_PASSWORD"),
		StagingAdminUsername:       value(lookup, "STAGING_AUTH_ADMIN_USERNAME"),
		StagingAdminPassword:       value(lookup, "STAGING_AUTH_ADMIN_PASSWORD"),
		PublicBaseURL:              value(lookup, "PUBLIC_BASE_URL"),
		TurnstileSiteKey:           value(lookup, "TURNSTILE_SITE_KEY"),
		TurnstileSecret:            value(lookup, "TURNSTILE_SECRET_KEY"),
		GoogleClientID:             value(lookup, "GOOGLE_OAUTH_CLIENT_ID"),
		GoogleSecret:               value(lookup, "GOOGLE_OAUTH_CLIENT_SECRET"),
		FacebookClientID:           value(lookup, "FACEBOOK_OAUTH_CLIENT_ID"),
		FacebookSecret:             value(lookup, "FACEBOOK_OAUTH_CLIENT_SECRET"),
		ResendAPIKey:               value(lookup, "RESEND_API_KEY"),
		ResendFromEmail:            value(lookup, "RESEND_FROM_EMAIL"),
		BootstrapAdminEmail:        strings.ToLower(strings.TrimSpace(value(lookup, "BOOTSTRAP_ADMIN_EMAIL"))),
		BootstrapAdminName:         strings.TrimSpace(value(lookup, "BOOTSTRAP_ADMIN_NAME")),
		RegencyMasterPath:          strings.TrimSpace(value(lookup, "REGENCY_MASTER_PATH")),
		FileStorageDir:             strings.TrimSpace(value(lookup, "FILE_STORAGE_DIR")),
		ScenarioPassword:           value(lookup, "STAGING_SCENARIO_PASSWORD"),
		ScenarioReset:              value(lookup, "STAGING_SCENARIO_RESET") == "true",
	}

	cfg.LoginRateLimitPerMinute = 5
	if raw := value(lookup, "LOGIN_RATE_LIMIT_PER_MINUTE"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 600 {
			return Config{}, fmt.Errorf("LOGIN_RATE_LIMIT_PER_MINUTE must be between 1 and 600")
		}
		cfg.LoginRateLimitPerMinute = limit
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
	stagingCredentials := []struct {
		username string
		password string
	}{
		{cfg.StagingNonVerifiedUsername, cfg.StagingNonVerifiedPassword},
		{cfg.StagingVerifiedUsername, cfg.StagingVerifiedPassword},
		{cfg.StagingOfficerUsername, cfg.StagingOfficerPassword},
		{cfg.StagingAdminUsername, cfg.StagingAdminPassword},
	}
	configuredStagingCredentials := 0
	seenUsernames := make(map[string]struct{}, len(stagingCredentials))
	for _, credential := range stagingCredentials {
		usernameSet := credential.username != ""
		passwordSet := credential.password != ""
		if usernameSet != passwordSet {
			return Config{}, fmt.Errorf("each staging auth username must be configured with its password")
		}
		if !usernameSet {
			continue
		}
		configuredStagingCredentials++
		if !validStagingUsername(credential.username) {
			return Config{}, fmt.Errorf("staging auth usernames must be 3-64 ASCII letters, digits, dots, underscores, or hyphens")
		}
		if len(credential.password) < 12 {
			return Config{}, fmt.Errorf("staging auth passwords must contain at least 12 characters")
		}
		username := strings.ToLower(credential.username)
		if _, exists := seenUsernames[username]; exists {
			return Config{}, fmt.Errorf("staging auth usernames must be unique")
		}
		seenUsernames[username] = struct{}{}
	}
	if configuredStagingCredentials != 0 && configuredStagingCredentials != len(stagingCredentials) {
		return Config{}, fmt.Errorf("all four staging auth username/password pairs must be configured together")
	}
	seenScenario := make(map[string]struct{}, len(scenarioUsernameKeys))
	for _, entry := range scenarioUsernameKeys {
		username := strings.ToLower(strings.TrimSpace(valueOrDefault(lookup, entry.key, entry.fallback)))
		if !validStagingUsername(username) {
			return Config{}, fmt.Errorf("%s must be 3-64 ASCII letters, digits, dots, underscores, or hyphens", entry.key)
		}
		if _, duplicate := seenScenario[username]; duplicate {
			return Config{}, fmt.Errorf("scenario account usernames must be unique")
		}
		seenScenario[username] = struct{}{}
		cfg.ScenarioUsernames = append(cfg.ScenarioUsernames, username)
	}
	if cfg.ScenarioPassword != "" && len(cfg.ScenarioPassword) < 12 {
		return Config{}, fmt.Errorf("STAGING_SCENARIO_PASSWORD must contain at least 12 characters")
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
	if (cfg.ResendAPIKey == "") != (cfg.ResendFromEmail == "") {
		return Config{}, fmt.Errorf("RESEND_API_KEY and RESEND_FROM_EMAIL must be configured together")
	}
	if cfg.ResendFromEmail != "" {
		from, err := mail.ParseAddress(cfg.ResendFromEmail)
		if err != nil || from.Address != cfg.ResendFromEmail {
			return Config{}, fmt.Errorf("RESEND_FROM_EMAIL must be a valid email address")
		}
	}
	if (cfg.BootstrapAdminEmail == "") != (cfg.BootstrapAdminName == "") {
		return Config{}, fmt.Errorf("BOOTSTRAP_ADMIN_EMAIL and BOOTSTRAP_ADMIN_NAME must be configured together")
	}
	if cfg.BootstrapAdminEmail != "" {
		parsed, err := mail.ParseAddress(cfg.BootstrapAdminEmail)
		if err != nil || parsed.Address != cfg.BootstrapAdminEmail {
			return Config{}, fmt.Errorf("BOOTSTRAP_ADMIN_EMAIL must be a valid email address")
		}
		if cfg.ResendAPIKey == "" {
			return Config{}, fmt.Errorf("email OTP must be configured before bootstrapping an admin")
		}
		if len(cfg.BootstrapAdminName) > 160 {
			return Config{}, fmt.Errorf("BOOTSTRAP_ADMIN_NAME must be at most 160 characters")
		}
	}
	if cfg.GoogleClientID != "" || cfg.FacebookClientID != "" || cfg.TurnstileSiteKey != "" || cfg.ResendAPIKey != "" {
		if cfg.PublicBaseURL == "" {
			return Config{}, fmt.Errorf("PUBLIC_BASE_URL is required when OAuth, CAPTCHA, or email OTP is enabled")
		}
		publicURL, err := url.Parse(cfg.PublicBaseURL)
		if err != nil || publicURL == nil {
			return Config{}, fmt.Errorf("PUBLIC_BASE_URL must be a valid URL")
		}
		localHost := publicURL.Hostname() == "localhost"
		if ip := net.ParseIP(publicURL.Hostname()); ip != nil && ip.IsLoopback() {
			localHost = true
		}
		if publicURL.Hostname() == "" || (publicURL.Scheme != "https" && !(publicURL.Scheme == "http" && localHost)) {
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

func validStagingUsername(username string) bool {
	if len(username) < 3 || len(username) > 64 {
		return false
	}
	for _, character := range username {
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '.' || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
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
