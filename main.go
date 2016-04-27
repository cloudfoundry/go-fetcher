package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
  "github.com/cloudfoundry/go-fetcher/handlers"
  "github.com/cloudfoundry/go-fetcher/config"
)

var domain = flag.String(
	"domain",
	"",
	"the address that we expect to receive requests on",
)

var orgFlag = flag.String(
	"orgList",
	"",
	"a comma-separted list of github org urls to redirect import traffic to",
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	config, err := config.Parse("../config.json")
	if err != nil {
		panic(err)
	}
	handler := handlers.NewHandler(*config)
	http.HandleFunc("/", handler.GetMeta)

	fmt.Println("go-fetch-server.ready")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
