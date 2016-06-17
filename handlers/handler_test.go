package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry/go-fetcher/config"
	"github.com/cloudfoundry/go-fetcher/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Handler", func() {
	var server *ghttp.Server
	var handler *handlers.Handler
	var req *http.Request
	var res *httptest.ResponseRecorder
	var logger *lagertest.TestLogger

	BeforeEach(func() {
		server = ghttp.NewServer()

		logger = lagertest.NewTestLogger("test")
		handler = handlers.NewHandler(config.Config{
			LogLevel: "info",
			OrgList: []string{
				fmt.Sprintf("%s/org1/", server.URL()),
				fmt.Sprintf("%s/org2/", server.URL())},
			ImportPrefix:     "import-prefix",
			NoRedirectAgents: []string{"NoRedirect"},
			Overrides: map[string]string{
				"overridden": fmt.Sprintf("%s/other-org/overridden", server.URL())},
			GithubAPIKey:         "somekey-somekey",
			GithubStatusEndpoint: fmt.Sprintf("%s/", server.URL()),
		}, logger)
	})

	AfterEach(func() {
		if server.HTTPTestServer != nil {
			server.Close()
		}
	})

	Context("for Status handler function", func() {
		BeforeEach(func() {
			var err error
			req, err = http.NewRequest("GET", "/status", nil)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			res = httptest.NewRecorder()
			handler.Status(res, req)
			server.AllowUnhandledRequests = true
			server.UnhandledRequestStatusCode = http.StatusNotFound
		})

		Context("successful response", func() {
			BeforeEach(func() {
				server.RouteToHandler("GET", "/somekey-somekey", ghttp.RespondWith(http.StatusOK, "", http.Header{"X-RateLimit-Remaining": []string{"20000"}}))
			})

			It("should return remainig ratelimit value", func() {
				resBody := res.Body.String()
				Expect(resBody).To(ContainSubstring(fmt.Sprintf("remaining: 20000")))
				Expect(res.Code).To(Equal(http.StatusOK))
			})
		})

		Context("failed response > 400", func() {
			BeforeEach(func() {
				server.RouteToHandler("GET", "/somekey-somekey", ghttp.RespondWith(403, "Failed"))
			})

			It("should return error", func() {
				resBody := res.Body.String()
				Expect(resBody).To(ContainSubstring(fmt.Sprintf("error: Failed")))
				Expect(res.Code).To(Equal(403))
			})
		})
		Context("Get request errored", func() {
			BeforeEach(func() {
				server.RouteToHandler("GET", "/somekey-somekey", ghttp.RespondWith(302, "Redirect requires location in Header"))
			})

			It("should return error", func() {
				Expect(res.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Context("for GetMeta handler function", func() {
		BeforeEach(func() {
			server.RouteToHandler("HEAD", "/org1/repo1", ghttp.RespondWith(http.StatusOK, ""))
			server.RouteToHandler("HEAD", "/org2/repo1", ghttp.RespondWith(http.StatusOK, ""))
			server.RouteToHandler("HEAD", "/org2/repo2", ghttp.RespondWith(http.StatusOK, ""))
			server.RouteToHandler("HEAD", "/org2/overridden", ghttp.RespondWith(http.StatusOK, ""))

			server.AllowUnhandledRequests = true
			server.UnhandledRequestStatusCode = http.StatusNotFound
		})

		JustBeforeEach(func() {
			res = httptest.NewRecorder()
			handler.GetMeta(res, req)
		})

		Context("when the repo exists in the first and second organizations", func() {
			BeforeEach(func() {
				var err error
				req, err = http.NewRequest("GET", "/repo1", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the repo in the first organization url", func() {
				Expect(res.Code).To(Equal(http.StatusFound))

				headers := res.Header()
				Expect(headers.Get("Location")).To(Equal(fmt.Sprintf("%s/org1/repo1", server.URL())))
			})

			Context("when the user agent is in the NoRedirectAgents list", func() {
				BeforeEach(func() {
					req.Header.Add("User-Agent", "NoRedirect")
				})

				It("returns the second organiztion in the HTML meta tags,", func() {
					Expect(res.Code).To(Equal(http.StatusOK))

					resBody := res.Body.String()
					Expect(resBody).To(ContainSubstring(fmt.Sprintf("<meta name=\"go-import\" content=\"import-prefix/repo1 git %s/org1/repo1\">", server.URL())))
					Expect(resBody).To(ContainSubstring(fmt.Sprintf("<meta name=\"go-source\" content=\"import-prefix/repo1 _ %s/org1/repo1\">", server.URL())))
				})
			})
		})

		Context("when the repo exists only in the second organization", func() {
			BeforeEach(func() {
				var err error
				req, err = http.NewRequest("GET", "/repo2", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the repo in the second organization url", func() {
				Expect(res.Code).To(Equal(http.StatusFound))

				headers := res.Header()
				Expect(headers.Get("Location")).To(Equal(fmt.Sprintf("%s/org2/repo2", server.URL())))
			})

			Context("when the user agent is in the NoRedirectAgents list", func() {
				BeforeEach(func() {
					req.Header.Add("User-Agent", "NoRedirect")
				})

				It("returns the second organiztion in the HTML meta tags,", func() {
					Expect(res.Code).To(Equal(http.StatusOK))

					resBody := res.Body.String()
					Expect(resBody).To(ContainSubstring(fmt.Sprintf("<meta name=\"go-import\" content=\"import-prefix/repo2 git %s/org2/repo2\">", server.URL())))
					Expect(resBody).To(ContainSubstring(fmt.Sprintf("<meta name=\"go-source\" content=\"import-prefix/repo2 _ %s/org2/repo2\">", server.URL())))
				})
			})
		})

		Context("when the request includes a subpackage", func() {
			BeforeEach(func() {
				var err error
				req, err = http.NewRequest("GET", "/repo1/subpackage", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("redirects to the base of the repository", func() {
				Expect(res.Code).To(Equal(http.StatusFound))

				headers := res.Header()
				Expect(headers.Get("Location")).To(Equal(fmt.Sprintf("%s/org1/repo1", server.URL())))
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
				Expect(headers.Get("Location")).To(Equal(fmt.Sprintf("%s/other-org/overridden", server.URL())))
			})
		})

		Context("when there is an error communicated with the backing server", func() {
			BeforeEach(func() {
				server.Close()

				var err error
				req, err = http.NewRequest("GET", "/repo2", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns status code 502", func() {
				Expect(res.Code).To(Equal(http.StatusBadGateway))
				Expect(res.Body.String()).To(ContainSubstring("connection refused"))
			})
		})
	})
})
