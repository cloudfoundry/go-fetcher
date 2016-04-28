package handlers

import (
  "net/http"
	"strings"
	"fmt"
	"github.com/cloudfoundry/go-fetcher/config"
)

type handler struct {
	config config.Config
}

func NewHandler(config config.Config) *handler {
  return &handler{
		config: config,
	}
}

func (h *handler) GetMeta(writer http.ResponseWriter, request *http.Request) {
		repoName := strings.Split(request.URL.Path, "/")[1]
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(writer, "<meta name=\"go-import\" content=\"%s git %s\">",
			h.config.Host + "/" + repoName,
			h.config.OrgList[0] + repoName,
		)

		fmt.Fprintf(writer, "<meta name=\"go-source\" content=\"%s _ %s\">",
			h.config.Host + "/" + repoName,
			h.config.OrgList[0] + repoName,
		)

		// do not redirect if the agent is known from the NoRedirect list
	  if contains(h.config.NoRedirectAgents, request.Header.Get("User-Agent")){
			return
		}

		repoPath := strings.TrimRight(strings.TrimLeft(request.URL.Path, "/"), "?")
		fmt.Fprintf(writer, "<meta http-equiv=\"refresh\" content=\"0; url=https://godoc.org/%s/%s\">", h.config.Host, repoPath)
}


func contains(slice []string, object string) bool {
	for _, a := range slice {
		if a == object {
			return true
		}
	}
	return false
}
