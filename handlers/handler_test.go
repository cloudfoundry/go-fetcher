package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry/go-fetcher/cache"
	"github.com/cloudfoundry/go-fetcher/config"
	"github.com/cloudfoundry/go-fetcher/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Handler", func() {
	var (
		handler       *handlers.Handler
		req           *http.Request
		res           *httptest.ResponseRecorder
		logger        *lagertest.TestLogger
		locationCache *cache.LocationCache
		cfg           config.Config
	)

	BeforeEach(func() {
		cfg = config.Config{
			LogLevel:         "info",
			OrgList:          []string{"org1", "org2"},
			ImportPrefix:     "import-prefix",
			NoRedirectAgents: []string{"NoRedirect"},
			Overrides: map[string]string{
				"overridden": "http://override.org/other-org/overridden"},
			GithubURL:    "http://example.com",
			GithubAPIKey: "somekey-somekey",
		}

		logger = lagertest.NewTestLogger("test")
		clock := clock.NewClock()
		locationCache = cache.NewLocationCache(clock)
		handler = handlers.NewHandler(logger, cfg, locationCache)
	})

	Describe("GetMeta", func() {
		JustBeforeEach(func() {
			res = httptest.NewRecorder()
			handler.GetMeta(res, req)
		})

		Context("when the repo exists", func() {
			BeforeEach(func() {
				var err error
				locationCache.Add("repo1", fmt.Sprintf("%s/org1/repo1", cfg.GithubURL))
				req, err = http.NewRequest("GET", "/repo1", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the repo url", func() {
				Expect(res.Code).To(Equal(http.StatusFound))

				headers := res.Header()
				Expect(headers.Get("Location")).To(Equal(fmt.Sprintf("%s/org1/repo1", cfg.GithubURL)))
			})

			Context("when the user agent is in the NoRedirectAgents list", func() {
				BeforeEach(func() {
					req.Header.Add("User-Agent", "NoRedirect")
				})

				It("returns the second organiztion in the HTML meta tags,", func() {
					Expect(res.Code).To(Equal(http.StatusOK))

					resBody := res.Body.String()
					Expect(resBody).To(ContainSubstring(fmt.Sprintf("<meta name=\"go-import\" content=\"import-prefix/repo1 git %s/org1/repo1\">", cfg.GithubURL)))
					Expect(resBody).To(ContainSubstring(fmt.Sprintf("<meta name=\"go-source\" content=\"import-prefix/repo1 _ %s/org1/repo1\">", cfg.GithubURL)))
				})
			})
		})

		Context("when the request includes a subpackage", func() {
			BeforeEach(func() {
				var err error
				locationCache.Add("repo1", fmt.Sprintf("%s/org1/repo1", cfg.GithubURL))
				req, err = http.NewRequest("GET", "/repo1/subpackage", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("redirects to the base of the repository", func() {
				Expect(res.Code).To(Equal(http.StatusFound))

				headers := res.Header()
				Expect(headers.Get("Location")).To(Equal(fmt.Sprintf("%s/org1/repo1", cfg.GithubURL)))
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
				Expect(headers.Get("Location")).To(Equal("http://override.org/other-org/overridden"))
			})
		})
	})
})
