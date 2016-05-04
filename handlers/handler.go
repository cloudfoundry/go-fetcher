package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cloudfoundry/go-fetcher/config"
)

type Handler struct {
	config config.Config
}

func NewHandler(config config.Config) *Handler {
	return &Handler{
		config: config,
	}
}

func (h *Handler) GetMeta(writer http.ResponseWriter, request *http.Request) {
	repoName := strings.Split(request.URL.Path, "/")[1]
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	location := ""
	for k := range h.config.Overrides {
		if k == repoName {
			location = h.config.Overrides[k]
		}
	}

	if location == "" {
		for idx := range h.config.OrgList {
			response, err := http.Head(h.config.OrgList[idx] + repoName)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadGateway)
				return
			}

			if response.StatusCode < 400 {
				location = h.config.OrgList[idx] + repoName
				break
			}
		}
	}

	if location == "" {
		http.Error(writer, "", http.StatusNotFound)
		return
	}

	// do not redirect if the agent is known from the NoRedirect list
	if !contains(h.config.NoRedirectAgents, request.Header.Get("User-Agent")) {
		repoPath := strings.TrimLeft(request.URL.Path, "/")
		// if go-get=1 redirect to godoc.org using an HTML redirect, as expected by go get
		if request.URL.Query().Get("go-get") == "1" {
			fmt.Fprintf(writer,
				"<meta http-equiv=\"refresh\" content=\"0; url=https://godoc.org/%s/%s\">",
				h.config.ImportPrefix, repoPath)
		} else {
			http.Redirect(writer, request, location, http.StatusFound)
		}

		return
	}

	fmt.Fprintf(writer, "<meta name=\"go-import\" content=\"%s git %s\">",
		h.config.ImportPrefix+"/"+repoName,
		location,
	)

	fmt.Fprintf(writer, "<meta name=\"go-source\" content=\"%s _ %s\">",
		h.config.ImportPrefix+"/"+repoName,
		location,
	)
}

func contains(slice []string, object string) bool {
	for _, a := range slice {
		if strings.Contains(object, a) {
			return true
		}
	}
	return false
}
