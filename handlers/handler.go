package handlers

import (
	"embed"
	_ "embed"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/cloudfoundry/go-fetcher/cache"
	"github.com/cloudfoundry/go-fetcher/config"
)

//go:embed index.html
var embeddedFiles embed.FS

type Handler struct {
	config        config.Config
	logger        *slog.Logger
	locationCache *cache.LocationCache
}

func NewHandler(logger *slog.Logger, config config.Config, locationCache *cache.LocationCache) *Handler {
	return &Handler{
		config:        config,
		logger:        logger,
		locationCache: locationCache,
	}
}

func (h *Handler) GetMeta(writer http.ResponseWriter, request *http.Request) {
	start := time.Now()

	repoName := strings.Split(request.URL.Path, "/")[1]
	logger := h.logger.With("action", "handler.getmeta", "repo-name", repoName)

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Handle index requests (/, /index.htm, index.html)
	if slices.Contains([]string{"/", "/index.htm", "/index.html"}, request.URL.Path) {
		logger.Debug("index-page", "location", request.URL.Path)

		serveIndex(writer, request)
		return
	}

	location := ""

	repoOverride, overrideExists := h.config.Overrides[repoName]
	if overrideExists {
		location = repoOverride.Repository
		logger.Debug("override", "location", location)
	}

	if location == "" {
		if loc, ok := h.locationCache.Lookup(repoName); ok {
			location = loc
			logger.Debug("cache-hit", "location", location)
		}
	}

	if location == "" {
		logger.Error("not-found", "error", fmt.Errorf("repo not in cache or override list"))
		http.Error(writer, "", http.StatusNotFound)
		return
	}

	defer func() {
		logger.Info("served", "duration", fmt.Sprint(time.Since(start)), "location", location)
	}()

	repoPath := strings.TrimLeft(request.URL.Path, "/")

	if request.URL.Query().Get("go-get") == "1" {
		// Always emit go-import and go-source meta tags so `go get` can resolve the import path.
		goImportContent := fmt.Sprintf("%s/%s git %s", h.config.ImportPrefix, repoName, location)
		goImport := fmt.Sprintf(`<meta name="go-import" content="%s">`, html.EscapeString(goImportContent))
		logger.Debug("meta.go-import", "content", goImportContent)
		fmt.Fprint(writer, goImport) //nolint:errcheck,staticcheck,govet

		goSourceContent := fmt.Sprintf("%s/%s _ %s", h.config.ImportPrefix, repoName, location)
		goSource := fmt.Sprintf(`<meta name="go-source" content="%s">`, html.EscapeString(goSourceContent))
		logger.Debug("meta.go-source", "content", goSourceContent)
		fmt.Fprint(writer, goSource) //nolint:errcheck,staticcheck,govet

		// Also add a browser redirect to pkg.go.dev for human visitors.
		if !slices.Contains(h.config.NoRedirectAgents, request.Header.Get("User-Agent")) {
			logger.Debug("redirect.meta", "path", repoPath)
			if _, err := fmt.Fprintf(writer,
				`<meta http-equiv="refresh" content="0; url=https://pkg.go.dev/%s/%s">`,
				html.EscapeString(h.config.ImportPrefix), html.EscapeString(repoPath)); err != nil {
				logger.Error("redirect.meta", "error", err)
			}
		} else {
			logger.Debug("redirect.http", "location", location)
			http.Redirect(writer, request, location, http.StatusFound)
		}

		return
	}

	// do not redirect if the agent is known from the NoRedirect list
	if !slices.Contains(h.config.NoRedirectAgents, request.Header.Get("User-Agent")) {
		logger.Debug("redirect.http", "location", location)
		http.Redirect(writer, request, location, http.StatusFound)
		return
	}

	goImportContent := fmt.Sprintf("%s git %s", h.config.ImportPrefix+"/"+repoName, location)
	// Go 1.25+ recognizes an optional fourth field naming the subdirectory that
	// holds the module's go.mod, allowing a module to live below the repo root.
	if repoOverride.Path != "" {
		goImportContent = fmt.Sprintf("%s %s", goImportContent, repoOverride.Path)
	}
	goImport := fmt.Sprintf(`<meta name="go-import" content="%s">`, html.EscapeString(goImportContent))
	logger.Debug("meta.go-import", "content", goImportContent)
	if _, err := fmt.Fprint(writer, goImport); err != nil {
		logger.Error("meta.go-import", "error", err)
	}

	goSourceContent := fmt.Sprintf("%s _ %s", h.config.ImportPrefix+"/"+repoName, location)
	goSource := fmt.Sprintf(`<meta name="go-source" content="%s">`, html.EscapeString(goSourceContent))
	logger.Debug("meta.go-source", "content", goSourceContent)
	if _, err := fmt.Fprint(writer, goSource); err != nil {
		logger.Error("meta.go-source", "error", err)
	}
}

func serveIndex(writer http.ResponseWriter, request *http.Request) {
	file, err := embeddedFiles.Open("index.html")
	if err != nil {
		http.Error(writer, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close() //nolint:errcheck

	stat, err := file.Stat()
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.ServeContent(writer, request, stat.Name(), stat.ModTime(), file.(io.ReadSeeker))
}
