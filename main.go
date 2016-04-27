package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
  "path/filepath"

  "github.com/cloudfoundry/go-fetcher/handlers"
  "github.com/cloudfoundry/go-fetcher/config"
)



func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

  files, _ := filepath.Glob("*")
	configFile := os.Getenv("CONFIG")
	if configFile == "" {
		configFile = "app/config.json"
	}
	config, err := config.Parse(configFile)

	if err != nil {
    fmt.Println(files)
		panic(err)
	}
	handler := handlers.NewHandler(*config)
	http.HandleFunc("/", handler.GetMeta)

	fmt.Println("go-fetch-server.ready")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
