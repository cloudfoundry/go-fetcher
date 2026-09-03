package handlers_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/clock"
	"github.com/cloudfoundry/go-fetcher/cache"
	"github.com/cloudfoundry/go-fetcher/config"
	"github.com/cloudfoundry/go-fetcher/handlers"
)

var _ = Describe("Handler", func() {
	var (
		handler       *handlers.Handler
		req           *http.Request
		res           *httptest.ResponseRecorder
		ginkgoLogger  *slog.Logger
		locationCache *cache.LocationCache
		cfg           config.Config
	)

	BeforeEach(func() {
		cfg = config.Config{
			GithubAPI:        "https://example.com",
			ImportPrefix:     "import-prefix",
			NoRedirectAgents: []string{"NoRedirect"},
			OrgList:          []string{"org1", "org2"},
			Overrides: map[string]config.Override{
				"overridden": {
					Repository: "https://override.example.com/other-org/overridden",
					Path:       "override/path",
				},
			},
		}

		ginkgoHandler := slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{Level: slog.LevelDebug})
		ginkgoLogger = slog.New(ginkgoHandler)

		clock := clock.NewClock()
		locationCache = cache.NewLocationCache(ginkgoLogger, clock)

		handler = handlers.NewHandler(ginkgoLogger, cfg, locationCache)
	})

	Describe("Index", func() {
		var indexHtml []byte

		JustBeforeEach(func() {
			res = httptest.NewRecorder()
			handler.GetMeta(res, req)
			var readErr error
			indexHtml, readErr = os.ReadFile("index.html")
			Expect(readErr).NotTo(HaveOccurred())
		})

		Context("when a default URL is requested", func() {
			BeforeEach(func() {
				var err error
				req, err = http.NewRequest("GET", "/", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("/ serves the static index page", func() {
				Expect(res.Code).To(Equal(http.StatusOK))
				body := res.Body.String()
				Expect(body).To(Equal(string(indexHtml)))
			})

			It("/index.htm serves the static index page", func() {
				Expect(res.Code).To(Equal(http.StatusOK))
				body := res.Body.String()
				Expect(body).To(Equal(string(indexHtml)))
			})

			It("/index.html serves the static index page", func() {
				Expect(res.Code).To(Equal(http.StatusOK))
				body := res.Body.String()
				Expect(body).To(Equal(string(indexHtml)))
			})
		})
	})

	Describe("GetMeta", func() {
		JustBeforeEach(func() {
			res = httptest.NewRecorder()
			handler.GetMeta(res, req)
		})

		Context("when the repo exists", func() {
			BeforeEach(func() {
				var err error
				locationCache.Add("repo1", fmt.Sprintf("%s/org1/repo1", cfg.GithubAPI))
				req, err = http.NewRequest("GET", "/repo1", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the repo url", func() {
				Expect(res.Code).To(Equal(http.StatusFound))

				headers := res.Header()
				Expect(headers.Get("Location")).To(Equal(fmt.Sprintf("%s/org1/repo1", cfg.GithubAPI)))
			})

			Context("when the user agent is in the NoRedirectAgents list", func() {
				BeforeEach(func() {
					req.Header.Add("User-Agent", "NoRedirect")
				})

				It("returns the second organization in the HTML meta tags,", func() {
					Expect(res.Code).To(Equal(http.StatusOK))

					resBody := res.Body.String()
					Expect(resBody).To(ContainSubstring(fmt.Sprintf(`<meta name="go-import" content="import-prefix/repo1 git %s/org1/repo1">`, cfg.GithubAPI)))
					Expect(resBody).To(ContainSubstring(fmt.Sprintf(`<meta name="go-source" content="import-prefix/repo1 _ %s/org1/repo1">`, cfg.GithubAPI)))
				})
			})

			Context("when go-get=1 is in the query string", func() {
				BeforeEach(func() {
					var err error
					req, err = http.NewRequest("GET", "/repo1?go-get=1", nil)
					Expect(err).NotTo(HaveOccurred())
				})

				It("always returns go-import and go-source meta tags", func() {
					Expect(res.Code).To(Equal(http.StatusOK))

					resBody := res.Body.String()
					Expect(resBody).To(ContainSubstring(fmt.Sprintf(`<meta name="go-import" content="import-prefix/repo1 git %s/org1/repo1">`, cfg.GithubAPI)))
					Expect(resBody).To(ContainSubstring(fmt.Sprintf(`<meta name="go-source" content="import-prefix/repo1 _ %s/org1/repo1">`, cfg.GithubAPI)))
				})

				It("also includes a browser redirect to pkg.go.dev", func() {
					resBody := res.Body.String()
					Expect(resBody).To(ContainSubstring(`<meta http-equiv="refresh" content="0; url=https://pkg.go.dev/import-prefix/repo1">`))
				})

				Context("when the user agent is in the NoRedirectAgents list", func() {
					BeforeEach(func() {
						req.Header.Add("User-Agent", "NoRedirect")
					})

					It("returns go-import and go-source meta tags but no browser redirect", func() {
						Expect(res.Code).To(Equal(http.StatusOK))

						resBody := res.Body.String()
						Expect(resBody).To(ContainSubstring(fmt.Sprintf(`<meta name="go-import" content="import-prefix/repo1 git %s/org1/repo1">`, cfg.GithubAPI)))
						Expect(resBody).To(ContainSubstring(fmt.Sprintf(`<meta name="go-source" content="import-prefix/repo1 _ %s/org1/repo1">`, cfg.GithubAPI)))
						Expect(resBody).NotTo(ContainSubstring(`http-equiv="refresh"`))
					})
				})
			})
		})

		Context("when the repo has a configured module subdirectory", func() {
			BeforeEach(func() {
				var err error
				handler = handlers.NewHandler(ginkgoLogger, cfg, locationCache)

				req, err = http.NewRequest("GET", "/overridden", nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Add("User-Agent", "NoRedirect")
			})

			It("appends the subdirectory as the fourth go-import field", func() {
				Expect(res.Code).To(Equal(http.StatusOK))

				resBody := res.Body.String()
				Expect(resBody).To(ContainSubstring(fmt.Sprintf(
					`<meta name="go-import" content="import-prefix/overridden git %s %s">`,
					cfg.Overrides["overridden"].Repository, cfg.Overrides["overridden"].Path)))
			})

			Context("when go-get=1 is in the query string", func() {
				BeforeEach(func() {
					var err error
					req, err = http.NewRequest("GET", "/overridden?go-get=1", nil)
					Expect(err).NotTo(HaveOccurred())
				})

				It("includes the subdirectory as the fourth go-import field", func() {
					Expect(res.Code).To(Equal(http.StatusOK))

					resBody := res.Body.String()
					Expect(resBody).To(ContainSubstring(fmt.Sprintf(
						`<meta name="go-import" content="import-prefix/overridden git %s %s">`,
						cfg.Overrides["overridden"].Repository, cfg.Overrides["overridden"].Path)))
				})
			})
		})

		Context("when the request includes a subpackage", func() {
			BeforeEach(func() {
				var err error
				locationCache.Add("repo1", fmt.Sprintf("%s/org1/repo1", cfg.GithubAPI))
				req, err = http.NewRequest("GET", "/repo1/subpackage", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("redirects to the base of the repository", func() {
				Expect(res.Code).To(Equal(http.StatusFound))

				headers := res.Header()
				Expect(headers.Get("Location")).To(Equal(fmt.Sprintf("%s/org1/repo1", cfg.GithubAPI)))
			})
		})

		Context("when the repo does not exist", func() {
			BeforeEach(func() {
				var err error
				req, err = http.NewRequest("GET", "/repo3", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a 404 Not Found", func() {
				Expect(res.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("when the repo exists in the override list", func() {
			BeforeEach(func() {
				var err error
				req, err = http.NewRequest("GET", "/overridden", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("redirects to the override value", func() {
				Expect(res.Code).To(Equal(http.StatusFound))

				headers := res.Header()
				Expect(headers.Get("Location")).To(Equal("https://override.example.com/other-org/overridden"))
			})
		})
	})
})
