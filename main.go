package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
  "github.com/cloudfoundry/go-fetcher/handlers"
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

	flag.Parse()
	orgList := strings.Split(*orgFlag, ",")


	handler := handlers.NewHandler(domain, orgList)
	http.HandleFunc("/", handler.GetMeta)

	fmt.Println("go-fetch-server.ready")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
