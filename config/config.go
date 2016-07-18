package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pivotal-golang/lager"
)

const (
	DEBUG = "debug"
	INFO  = "info"
	ERROR = "error"
	FATAL = "fatal"
)

type Config struct {
	LogLevel             string
	ImportPrefix         string
	OrgList              []string
	NoRedirectAgents     []string
	Overrides            map[string]string
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
