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
}
