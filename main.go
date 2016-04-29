package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cloudfoundry/go-fetcher/config"
	"github.com/cloudfoundry/go-fetcher/handlers"
	"github.com/cloudfoundry/go-fetcher/util"
)

var generate_config = flag.String(
	"generate_config",
	"",
	"Generate deployment configurations",
)

func main() {

	// if the flag `generate_config` is set to true, run the code to generate
	// config.json and manifest.yml from the provided templates
	flag.Parse()
	if *generate_config == "true" {
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

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	configFile := os.Getenv("CONFIG")
	config, err := config.Parse(configFile)

	if err != nil {
		log.Fatal("config file error: ", err)
	}
	handler := handlers.NewHandler(*config)
	http.HandleFunc("/", handler.GetMeta)

	fmt.Println("go-fetch-server.ready")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
