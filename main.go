package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		repoName := strings.Split(r.URL.Path, "/")[1]
		fmt.Fprintf(w, "<meta name=\"go-import\" content=\"%s git %s\">",
			*domain + "/" + repoName,
			orgList[0] + repoName,
		)
	})

	fmt.Println("go-fetch-server.ready")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
