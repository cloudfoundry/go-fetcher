package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cloudfoundry/go-fetcher/config"
	"github.com/pivotal-golang/lager"
)

type Handler struct {
	config config.Config
	logger lager.Logger
}

func NewHandler(config config.Config, logger lager.Logger) *Handler {
	return &Handler{
		config: config,
		logger: logger,
	}
}

func (h *Handler) GetMeta(writer http.ResponseWriter, request *http.Request) {
	repoName := strings.Split(request.URL.Path, "/")[1]
	logger := h.logger.Session("handler", lager.Data{"repo-name": repoName})

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
				logger.Error("github.response", err)
				http.Error(writer, err.Error(), http.StatusBadGateway)
				return
			}

			if response.StatusCode < 400 {
				location = h.config.OrgList[idx] + repoName
				logger.Debug("Repo found: " + location)
				break
			}
		}
	}

	if location == "" {
		logger.Error("repo.not_found", fmt.Errorf("Repo not in listed orgs"))
		http.Error(writer, "", http.StatusNotFound)
		return
	}

	// do not redirect if the agent is known from the NoRedirect list
	if !contains(h.config.NoRedirectAgents, request.Header.Get("User-Agent")) {
		repoPath := strings.TrimLeft(request.URL.Path, "/")
		// if go-get=1 redirect to godoc.org using an HTML redirect, as expected by go get
		if request.URL.Query().Get("go-get") == "1" {
			logger.Info("redirect.meta", lager.Data{"repoPath": repoPath})
			fmt.Fprintf(writer,
				"<meta http-equiv=\"refresh\" content=\"0; url=https://godoc.org/%s/%s\">",
				h.config.ImportPrefix, repoPath)
		} else {
			logger.Info("redirect.http", lager.Data{"location": location})
			http.Redirect(writer, request, location, http.StatusFound)
		}

		return
	}

	go_import_content := fmt.Sprintf("%s git %s", h.config.ImportPrefix+"/"+repoName, location)
	go_import := fmt.Sprintf("<meta name=\"go-import\" content=\"%s\">", go_import_content)
	logger.Info("meta.go-import", lager.Data{"content": go_import_content})
	fmt.Fprintf(writer, go_import)

	go_source_content := fmt.Sprintf("%s _ %s", h.config.ImportPrefix+"/"+repoName, location)
	go_source := fmt.Sprintf("<meta name=\"go-source\" content=\"%s\">", go_source_content)
	logger.Info("meta.go-source", lager.Data{"content": go_source_content})
	fmt.Fprintf(writer, go_source)
}

func contains(slice []string, object string) bool {
	for _, a := range slice {
		if strings.Contains(object, a) {
			return true
		}
	}
	return false
}
