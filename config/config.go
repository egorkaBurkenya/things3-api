package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Token         string
	Port          string
	Host          string
	LogLevel      string
	ThingsURLToken string
}

func Load() (*Config, error) {
	loadDotEnv()

	token := os.Getenv("THINGS_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("THINGS_API_TOKEN environment variable is required")
	}

	port := os.Getenv("THINGS_API_PORT")
	if port == "" {
		port = "7420"
	}

	host := os.Getenv("THINGS_API_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	thingsURLToken := os.Getenv("THINGS_URL_TOKEN")

	return &Config{
		Token:          token,
		Port:           port,
		Host:           host,
		LogLevel:       logLevel,
		ThingsURLToken: thingsURLToken,
	}, nil
}

func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}

func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
