package main

import (
	"fmt"
	"flag"
	"log"
	"net/http"
	"os"

  "github.com/cloudfoundry/go-fetcher/handlers"
  "github.com/cloudfoundry/go-fetcher/config"
  "github.com/cloudfoundry/go-fetcher/util"
)

var generate_config = flag.String(
		"generate_config",
		"",
		"Generate deployment configurations",
)

func main() {

	flag.Parse()
	if *generate_config == "true" {
		err := util.GenerateConfig(os.Getenv("ROOT_DIR") + "/util/config.json.template")
		if err != nil {
			log.Fatal(err)
		}
		err = util.GenerateManifest(os.Getenv("ROOT_DIR") + "/util/manifest.yml.template")
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
