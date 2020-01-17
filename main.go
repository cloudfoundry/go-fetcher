package main

import (
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
	"code.cloudfoundry.org/lager"
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
		templateFile := os.Getenv("ROOT_DIR") + "/util/config.json.template"
		configFile := os.Getenv("ROOT_DIR") + "/config.json"
		err := util.GenerateConfig(templateFile, configFile)
		if err != nil {
			log.Fatal(err)
		}

		templateFile = os.Getenv("ROOT_DIR") + "/util/manifest.yml.template"
		configFile = os.Getenv("ROOT_DIR") + "/manifest.yml"
		err = util.GenerateManifest(templateFile, configFile)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	configFile := os.Getenv("CONFIG")
	config, err := config.Parse(configFile)

	if err != nil {
		panic("config file error: " + err.Error())
	}

	logger := lager.NewLogger("go-fetcher")
	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), config.GetLogLevel())
	logger.RegisterSink(sink)

	port := os.Getenv("PORT")
	if port == "" {
		logger.Error("server.failed", fmt.Errorf("$PORT must be set"))
	}

	clock := clock.NewClock()
	locationCache := cache.NewLocationCache(logger.Session("cache"), clock)
	handler := handlers.NewHandler(logger, *config, locationCache)
	http.HandleFunc("/", handler.GetMeta)

	var tc *http.Client
	if config.GithubAPIKey != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.GithubAPIKey},
		)
		tc = oauth2.NewClient(oauth2.NoContext, ts)
	}

	client := github.NewClient(tc)
	githubURL, err := url.Parse(fmt.Sprintf("%s/", strings.TrimSuffix(config.GithubURL, "/")))
	if err != nil {
		log.Fatal(err)
	}
	client.BaseURL = githubURL

	httpServer := http_server.New(":"+port, http.DefaultServeMux)
	cacheLoader := cache.NewCacheLoader(
		logger.Session("cache-loader"),
		config.GithubURL, config.OrgList, locationCache, client.Repositories, clock,
	)

	members := grouper.Members{
		{"cache-loader", cacheLoader},
		{"http-server", httpServer},
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
