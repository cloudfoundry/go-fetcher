package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/go-fetcher/cache"
	"github.com/cloudfoundry/go-fetcher/config"
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
			logger.Debug("index-page", lager.Data{"location": request.URL.Path})
			indexHtmlPath, err := filepath.Abs(h.config.IndexPath)

			if err != nil {
				logger.Error("index-page", fmt.Errorf("could not get absolute path of IndexPath"))
			}
			http.ServeFile(writer, request, indexHtmlPath)
			return
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

	repoPath := strings.TrimLeft(request.URL.Path, "/")

	if request.URL.Query().Get("go-get") == "1" {
		// Always emit go-import and go-source meta tags so `go get` can resolve the import path.
		goImportContent := fmt.Sprintf("%s git %s", h.config.ImportPrefix+"/"+repoName, location)
		goImport := fmt.Sprintf("<meta name=\"go-import\" content=\"%s\">", goImportContent)
		logger.Debug("meta.go-import", lager.Data{"content": goImportContent})
		fmt.Fprintf(writer, goImport) //nolint:errcheck,staticcheck

		goSourceContent := fmt.Sprintf("%s _ %s", h.config.ImportPrefix+"/"+repoName, location)
		goSource := fmt.Sprintf("<meta name=\"go-source\" content=\"%s\">", goSourceContent)
		logger.Debug("meta.go-source", lager.Data{"content": goSourceContent})
		fmt.Fprintf(writer, goSource) //nolint:errcheck,staticcheck

		// Also add a browser redirect to pkg.go.dev for human visitors.
		if !contains(h.config.NoRedirectAgents, request.Header.Get("User-Agent")) {
			logger.Debug("redirect.meta", lager.Data{"path": repoPath})
			if _, err := fmt.Fprintf(writer,
				"<meta http-equiv=\"refresh\" content=\"0; url=https://pkg.go.dev/%s/%s\">",
				h.config.ImportPrefix, repoPath); err != nil {
				logger.Error("redirect.meta", err)
			}
		} else {
			logger.Debug("redirect.http", lager.Data{"location": location})
			http.Redirect(writer, request, location, http.StatusFound)
		}

		return
	}

	// do not redirect if the agent is known from the NoRedirect list
	if !contains(h.config.NoRedirectAgents, request.Header.Get("User-Agent")) {
		logger.Debug("redirect.http", lager.Data{"location": location})
		http.Redirect(writer, request, location, http.StatusFound)
		return
	}

	goImportContent := fmt.Sprintf("%s git %s", h.config.ImportPrefix+"/"+repoName, location)
	// Go 1.25+ recognizes an optional fourth field naming the subdirectory that
	// holds the module's go.mod, allowing a module to live below the repo root.
	if subdir := h.config.Subdirs[repoName]; subdir != "" {
		goImportContent = fmt.Sprintf("%s %s", goImportContent, subdir)
	}
	goImport := fmt.Sprintf("<meta name=\"go-import\" content=\"%s\">", goImportContent)
	logger.Debug("meta.go-import", lager.Data{"content": goImportContent})
	if _, err := fmt.Fprint(writer, goImport); err != nil {
		logger.Error("meta.go-import", err)
	}

	goSourceContent := fmt.Sprintf("%s _ %s", h.config.ImportPrefix+"/"+repoName, location)
	goSource := fmt.Sprintf("<meta name=\"go-source\" content=\"%s\">", goSourceContent)
	logger.Debug("meta.go-source", lager.Data{"content": goSourceContent})
	if _, err := fmt.Fprint(writer, goSource); err != nil {
		logger.Error("meta.go-source", err)
	}
}

func contains(slice []string, object string) bool {
	for _, a := range slice {
		if strings.Contains(object, a) {
			return true
		}
	}
	return false
}
