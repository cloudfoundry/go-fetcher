package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloudfoundry/go-fetcher/cache"
	"github.com/cloudfoundry/go-fetcher/config"
	"github.com/pivotal-golang/lager"
)

type Handler struct {
	config        config.Config
	logger        lager.Logger
	locationCache *cache.LocationCache
}

func NewHandler(logger lager.Logger, config config.Config, locationCache *cache.LocationCache) *Handler {
	return &Handler{
		config:        config,
		logger:        logger,
		locationCache: locationCache,
	}
}

func (h *Handler) GetMeta(writer http.ResponseWriter, request *http.Request) {
	start := time.Now()

	repoName := strings.Split(request.URL.Path, "/")[1]
	logger := h.logger.Session("handler.getmeta", lager.Data{"repo-name": repoName})

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Handle index requests (/, /index.htm, index.html)
	for _, path := range []string{"/", "/index.htm", "/index.html"} {
		if request.URL.Path == path {
			logger.Debug("redirect.index", lager.Data{"location": request.URL.Path, "destination": h.config.IndexRedirect})
			http.Redirect(writer, request, h.config.IndexRedirect, http.StatusFound)
		}
	}

	location := ""
	for k := range h.config.Overrides {
		if k == repoName {
			location = h.config.Overrides[k]
			logger.Debug("override", lager.Data{"location": location})
		}
	}

	if location == "" {
		if loc, ok := h.locationCache.Lookup(repoName); ok {
			location = loc
			logger.Debug("cache-hit", lager.Data{"location": location})
		}
	}

	if location == "" {
		logger.Error("not-found", fmt.Errorf("repo not in cache or override list"))
		http.Error(writer, "", http.StatusNotFound)
		return
	}

	defer func() {
		logger.Info("served", lager.Data{"duration": fmt.Sprint(time.Since(start)), "location": location})
	}()

	// do not redirect if the agent is known from the NoRedirect list
	if !contains(h.config.NoRedirectAgents, request.Header.Get("User-Agent")) {
		repoPath := strings.TrimLeft(request.URL.Path, "/")
		// if go-get=1 redirect to godoc.org using an HTML redirect, as expected by go get
		if request.URL.Query().Get("go-get") == "1" {
			logger.Debug("redirect.meta", lager.Data{"path": repoPath})
			fmt.Fprintf(writer,
				"<meta http-equiv=\"refresh\" content=\"0; url=https://godoc.org/%s/%s\">",
				h.config.ImportPrefix, repoPath)
		} else {
			logger.Debug("redirect.http", lager.Data{"location": location})
			http.Redirect(writer, request, location, http.StatusFound)
		}

		return
	}

	goImportContent := fmt.Sprintf("%s git %s", h.config.ImportPrefix+"/"+repoName, location)
	goImport := fmt.Sprintf("<meta name=\"go-import\" content=\"%s\">", goImportContent)
	logger.Debug("meta.go-import", lager.Data{"content": goImportContent})
	fmt.Fprintf(writer, goImport)

	goSourceContent := fmt.Sprintf("%s _ %s", h.config.ImportPrefix+"/"+repoName, location)
	goSource := fmt.Sprintf("<meta name=\"go-source\" content=\"%s\">", goSourceContent)
	logger.Debug("meta.go-source", lager.Data{"content": goSourceContent})
	fmt.Fprintf(writer, goSource)
}

func contains(slice []string, object string) bool {
	for _, a := range slice {
		if strings.Contains(object, a) {
			return true
		}
	}
	return false
}
