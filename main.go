package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"

	"github.com/cloudfoundry/go-fetcher/cache"
	"github.com/cloudfoundry/go-fetcher/config"
	"github.com/cloudfoundry/go-fetcher/handlers"
	"github.com/cloudfoundry/go-fetcher/util"
	"github.com/google/go-github/github"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"

	"code.cloudfoundry.org/clock"
)

var generateConfig = flag.Bool(
	"generateConfig",
	false,
	"Generate deployment configurations",
)

func main() {
	// if the flag `generate_config` is set to true, run the code to generate
	// config.json and manifest.yml from the provided templates
	flag.Parse()

	programLevel := new(slog.LevelVar)

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	logger := slog.New(logHandler).With("component", "go-fetcher")

	if *generateConfig {
		configFile := "config.json"
		err := util.GenerateConfig(configFile)
		if err != nil {
			log.Fatal(err)
		}

		configFile = "manifest.yml"
		err = util.GenerateManifest(configFile)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	configFile := os.Getenv("CONFIG")
	cfg, err := config.Parse(configFile)
	if err != nil {
		logger.Error("config.parse", "error", fmt.Errorf("unable to parse CONFIG='%s': %w", configFile, err))
	}

	logLevel, err := cfg.GetLogLevel()
	if err != nil {
		logger.Warn("config.log-level", "error", err)
	}
	programLevel.Set(logLevel)

	port := os.Getenv("PORT")
	if port == "" {
		logger.Error("server.failed", "error", fmt.Errorf("$PORT must be set"))
	}

	clck := clock.NewClock()
	locationCache := cache.NewLocationCache(logger.With("component", "cache"), clck)
	handler := handlers.NewHandler(logger, *cfg, locationCache)
	http.HandleFunc("/", handler.GetMeta)

	var tc *http.Client
	if cfg.GithubAPIKey != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: cfg.GithubAPIKey},
		)
		tc = oauth2.NewClient(context.Background(), ts)
	}

	client := github.NewClient(tc)
	githubURL, err := url.Parse(fmt.Sprintf("%s/", strings.TrimSuffix(cfg.GithubURL, "/")))
	if err != nil {
		log.Fatal(err)
	}
	client.BaseURL = githubURL

	httpServer := http_server.New(":"+port, http.DefaultServeMux)
	cacheLoader := cache.NewCacheLoader(
		logger.With("component", "cache-loader"),
		cfg.OrgList,
		locationCache,
		client.Repositories,
		clck,
	)

	members := grouper.Members{
		{Name: "cache-loader", Runner: cacheLoader},
		{Name: "http-server", Runner: httpServer},
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err = <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", "error", err)
		os.Exit(1)
	}

	logger.Info("exited")
}
