package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

  "github.com/cloudfoundry/go-fetcher/handlers"
  "github.com/cloudfoundry/go-fetcher/config"
)



func main() {
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
