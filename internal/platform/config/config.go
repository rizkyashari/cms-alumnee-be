package config

import (
	"fmt"
	"net"
	"strconv"
)

type Config struct {
	HTTPAddr      string
	MySQLHost     string
	MySQLPort     string
	MySQLUser     string
	MySQLPassword string
	MySQLDatabase string
}

func Load(lookup func(string) (string, bool)) (Config, error) {
	cfg := Config{
		HTTPAddr:      valueOrDefault(lookup, "HTTP_ADDR", "127.0.0.1:3000"),
		MySQLHost:     valueOrDefault(lookup, "MYSQL_HOST", "127.0.0.1"),
		MySQLPort:     valueOrDefault(lookup, "MYSQL_PORT", "3306"),
		MySQLUser:     value(lookup, "MYSQL_USER"),
		MySQLPassword: value(lookup, "MYSQL_PASSWORD"),
		MySQLDatabase: value(lookup, "MYSQL_DATABASE"),
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
