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
)

var _ = Describe("Handler", func() {
	var server *ghttp.Server
	var handler *handlers.Handler
	var req *http.Request
	var res *httptest.ResponseRecorder

	BeforeEach(func() {
		server = ghttp.NewServer()
		server.RouteToHandler("HEAD", "/org1/repo1", ghttp.RespondWith(http.StatusOK, ""))
		server.RouteToHandler("HEAD", "/org2/repo1", ghttp.RespondWith(http.StatusOK, ""))
		server.RouteToHandler("HEAD", "/org2/repo2", ghttp.RespondWith(http.StatusOK, ""))
		server.RouteToHandler("HEAD", "/org2/overridden", ghttp.RespondWith(http.StatusOK, ""))

		server.AllowUnhandledRequests = true
		server.UnhandledRequestStatusCode = http.StatusNotFound

		handler = handlers.NewHandler(config.Config{
			OrgList: []string{
				fmt.Sprintf("%s/org1/", server.URL()),
				fmt.Sprintf("%s/org2/", server.URL())},
			ImportPrefix:     "import-prefix",
			NoRedirectAgents: []string{"NoRedirect"},
			Overrides: map[string]string{
				"overridden": fmt.Sprintf("%s/other-org/overridden", server.URL())},
		})
	})

	AfterEach(func() {
		if server.HTTPTestServer != nil {
			server.Close()
		}
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
