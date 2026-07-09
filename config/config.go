package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"code.cloudfoundry.org/lager"
)

const (
	DEBUG = "debug"
	INFO  = "info"
	ERROR = "error"
	FATAL = "fatal"
)

type Config struct {
	LogLevel         string
	ImportPrefix     string
	OrgList          []string
	NoRedirectAgents []string
	Overrides        map[string]string
	// Subdirs maps a repo name to the directory within its repository that
	// contains the Go module's root (its go.mod). When set, it is emitted as
	// the optional fourth field of the go-import meta tag, which the go command
	// recognizes as of Go 1.25 to fetch a module from a subdirectory.
	Subdirs              map[string]string
	GithubAPIKey         string
	GithubStatusEndpoint string
	GithubURL            string
	IndexPath            string
}

func (c *Config) GetLogLevel() lager.LogLevel {
	var minLagerLogLevel lager.LogLevel
	switch c.LogLevel {
	case DEBUG:
		minLagerLogLevel = lager.DEBUG
	case INFO:
		minLagerLogLevel = lager.INFO
	case ERROR:
		minLagerLogLevel = lager.ERROR
	case FATAL:
		minLagerLogLevel = lager.FATAL
	default:
		panic(fmt.Errorf("unknown log level: %s", c.LogLevel))
	}
	return minLagerLogLevel
}

func Parse(configPath string) (*Config, error) {
	jsonBlob, err := ioutil.ReadFile(configPath)

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
