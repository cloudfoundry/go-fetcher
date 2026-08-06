package main

import (
	"context"
	"flag"
	"fmt"
	"log"
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
	"code.cloudfoundry.org/lager/v3"
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

	if *generateConfig {
		templateFile := "util/config.json.template"
		configFile := "config.json"
		err := util.GenerateConfig(templateFile, configFile)
		if err != nil {
			log.Fatal(err)
		}

		templateFile = "util/manifest.yml.template"
		configFile = "manifest.yml"
		err = util.GenerateManifest(templateFile, configFile)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	configFile := os.Getenv("CONFIG")
	cfg, err := config.Parse(configFile)

	if err != nil {
		panic("config file error: " + err.Error())
	}

	logger := lager.NewLogger("go-fetcher")
	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), cfg.GetLogLevel())
	logger.RegisterSink(sink)

	port := os.Getenv("PORT")
	if port == "" {
		logger.Error("server.failed", fmt.Errorf("$PORT must be set"))
	}

	clck := clock.NewClock()
	locationCache := cache.NewLocationCache(logger.Session("cache"), clck)
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
		logger.Session("cache-loader"),
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
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
}
