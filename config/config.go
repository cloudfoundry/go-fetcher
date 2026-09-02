package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

const (
	DEBUG = "debug"
	INFO  = "info"
	WARN  = "warn"
	ERROR = "error"
)

type Config struct {
	GithubAPI        string
	GithubAPIKey     string `json:"-"` // never load from JSON
	ImportPrefix     string `json:"-"` // never load from JSON
	LogLevel         string
	OrgList          []string
	NoRedirectAgents []string
	Overrides        map[string]Override
}

type Override struct {
	Repository string
	Path       string
}

func (c *Config) GetLogLevel() (slog.Level, error) {
	switch c.LogLevel {
	case DEBUG:
		return slog.LevelDebug, nil
	case INFO:
		return slog.LevelInfo, nil
	case WARN:
		return slog.LevelWarn, nil
	case ERROR:
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level '%s', falling back to 'slog.LevelInfo'", c.LogLevel)
	}
}

func Parse(configPath string) (*Config, error) {
	jsonBlob, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(jsonBlob, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
