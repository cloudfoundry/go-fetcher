package main_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

// NOTE: os.Setenv/os.Unsetenv errors are intentionally ignored in these tests;
// this is removed once the suite moves to Ginkgo v2's GinkgoT().Setenv helper.
//
//nolint:errcheck
var _ = Describe("Import Path Redirect Service", func() {
	var (
		port             string
		importPrefix     = "the.canonical.import.path"
		someAgent        = "some-agent"
		session          *gexec.Session
		fakeGithubServer *ghttp.Server
		httpClient       *http.Client
	)

	BeforeEach(func() {
		port = strconv.Itoa(8182 + GinkgoParallelProcess())
		GinkgoT().Setenv("PORT", port)
	})

	Describe("when config can be parsed", func() {
		BeforeEach(func() {
			fakeGithubServer = ghttp.NewServer()
			fakeGithubServer.RouteToHandler("GET", "/orgs/cloudfoundry/repos", ghttp.RespondWithJSONEncoded(http.StatusOK, []map[string]interface{}{
				{
					"id":       1,
					"name":     "repository-1",
					"html_url": fmt.Sprintf("%s/cloudfoundry/repository-1", fakeGithubServer.URL()),
				},
				{
					"id":       2,
					"name":     "repository-2",
					"html_url": fmt.Sprintf("%s/cloudfoundry/repository-2", fakeGithubServer.URL()),
				},
			}))
			fakeGithubServer.RouteToHandler("GET", "/orgs/cloudfoundry-incubator/repos", ghttp.RespondWithJSONEncoded(http.StatusOK, []map[string]interface{}{
				{
					"id":       3,
					"name":     "repo-in-incubator",
					"html_url": fmt.Sprintf("%s/cloudfoundry-incubator/repo-in-incubator", fakeGithubServer.URL()),
				},
			}))
			fakeGithubServer.RouteToHandler("GET", "/orgs/cloudfoundry-attic/repos", ghttp.RespondWithJSONEncoded(http.StatusOK, []map[string]interface{}{
				{
					"id":       4,
					"name":     "repo-in-attic",
					"html_url": fmt.Sprintf("%s/cloudfoundry-attic/repo-in-attic", fakeGithubServer.URL()),
				},
			}))

			fakeGithubServer.AllowUnhandledRequests = true
			fakeGithubServer.UnhandledRequestStatusCode = http.StatusNotFound

			port = strconv.Itoa(8182 + GinkgoParallelProcess())

			configJson := []byte(fmt.Sprintf(`{
		  "GithubAPI":        "%s",
		  "LogLevel":         "debug",
		  "OrgList":          [
			"cloudfoundry",
			"cloudfoundry-incubator", 
			"cloudfoundry-attic"
		  ],
		  "NoRedirectAgents": [
			"%s", 
			"some-other-agent"
		  ]
		}`, fakeGithubServer.URL(), someAgent))

			configFile := filepath.Join(GinkgoT().TempDir(), fmt.Sprintf("config-%d.json", GinkgoParallelProcess()))
			err := os.WriteFile(configFile, configJson, 0644)
			Expect(err).NotTo(HaveOccurred())

			GinkgoT().Setenv("CONFIG", configFile)
			GinkgoT().Setenv("IMPORT_PREFIX", importPrefix)

			session, err = gexec.Start(exec.Command(goFetchBinary), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			// Startup includes loading the location cache from GitHub; under -race
			// and parallel package runs this can exceed the 1s default, so allow more time.
			Eventually(session, 20*time.Second).Should(gbytes.Say(`"msg":"started"`))
		})

		AfterEach(func() {
			session.Kill().Wait()
			fakeGithubServer.Close()
		})

		Context("when the user agent is part of the NoRedirectAgents list", func() {
			It("responds appropriately", func() {
				httpClient = &http.Client{}
				req, err := http.NewRequest("GET", fmt.Sprintf("http://:%s/repository-1/something-else/test", port), nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("User-Agent", someAgent)

				res, err := httpClient.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer func() { _ = res.Body.Close() }()

				body, err := io.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(body).To(ContainSubstring(fmt.Sprintf(
					`<meta name="go-import" content="%s/repository-1 git %s/cloudfoundry/repository-1">`,
					importPrefix,
					fakeGithubServer.URL())))

				Expect(body).To(ContainSubstring(fmt.Sprintf(
					`<meta name="go-source" content="%s/repository-1 _ %s/cloudfoundry/repository-1">`,
					importPrefix,
					fakeGithubServer.URL())))
			})
		})

		Describe("Redirects", func() {
			Context("when go-get is not set", func() {
				var redirectCount int

				BeforeEach(func() {
					redirectCount = 0

					httpClient = &http.Client{
						CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
							redirectCount++
							return errors.New("don't follow redirect in test")
						},
					}
				})

				Context("when the repo is in cloudfoundry", func() {
					It("will redirect to the true cloudfoundry source via HTTP redirects", func() {
						req, err := http.NewRequest("GET", fmt.Sprintf("http://:%s/repository-2", port), nil)
						Expect(err).NotTo(HaveOccurred())

						res, err := httpClient.Do(req)
						Expect(res).NotTo(BeNil())
						Expect(res.StatusCode).To(Equal(http.StatusFound))
						Expect(res.Header.Get("Location")).To(Equal(fmt.Sprintf("%s/cloudfoundry/repository-2", fakeGithubServer.URL())))
						Expect(err).To(MatchError(ContainSubstring("don't follow redirect in test")))

						Expect(redirectCount).To(Equal(1))
					})
				})

				Context("when the repo is in cloudfoundry-incubator", func() {
					It("will redirect to the true cloudfoundry-incubator source via HTTP redirects", func() {
						req, err := http.NewRequest("GET", fmt.Sprintf("http://:%s/repo-in-incubator", port), nil)
						Expect(err).NotTo(HaveOccurred())

						res, err := httpClient.Do(req)
						Expect(res).NotTo(BeNil())
						Expect(res.StatusCode).To(Equal(http.StatusFound))
						Expect(res.Header.Get("Location")).To(Equal(fmt.Sprintf("%s/cloudfoundry-incubator/repo-in-incubator", fakeGithubServer.URL())))
						Expect(err).To(MatchError(ContainSubstring("don't follow redirect in test")))

						Expect(redirectCount).To(Equal(1))
					})
				})

				Context("when the repo is in cloudfoundry-attic", func() {
					It("will redirect to the true cloudfoundry-attic source via HTTP redirects", func() {
						req, err := http.NewRequest("GET", fmt.Sprintf("http://:%s/repo-in-attic", port), nil)
						Expect(err).NotTo(HaveOccurred())

						res, err := httpClient.Do(req)
						Expect(res).NotTo(BeNil())
						Expect(res.StatusCode).To(Equal(http.StatusFound))
						Expect(res.Header.Get("Location")).To(Equal(fmt.Sprintf("%s/cloudfoundry-attic/repo-in-attic", fakeGithubServer.URL())))
						Expect(err).To(MatchError(ContainSubstring("don't follow redirect in test")))

						Expect(redirectCount).To(Equal(1))
					})
				})
			})

			Context("when go-get is set", func() {
				It("returns go-import and go-source meta tags and a pkg.go.dev browser redirect", func() {
					httpClient = &http.Client{}

					req, err := http.NewRequest("GET", fmt.Sprintf("http://:%s/repository-1/test?go-get=1", port), nil)
					Expect(err).NotTo(HaveOccurred())

					res, err := httpClient.Do(req)
					Expect(err).NotTo(HaveOccurred())
					defer func() { _ = res.Body.Close() }()

					var body []byte
					body, err = io.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())

					expectedImport := fmt.Sprintf(`<meta name="go-import" content="%s/repository-1 git `, importPrefix)
					Expect(body).To(ContainSubstring(expectedImport))

					expectedRedirect := fmt.Sprintf(`<meta http-equiv="refresh" content="0; url=https://pkg.go.dev/%s/repository-1/test">`, importPrefix)
					Expect(body).To(ContainSubstring(expectedRedirect))
				})
			})
		})
	})

	Describe("when config can NOT be parsed", func() {
		BeforeEach(func() {
			configJson := []byte("bad-json")

			configFile := filepath.Join(GinkgoT().TempDir(), fmt.Sprintf("config-%d.json", GinkgoParallelProcess()))
			err := os.WriteFile(configFile, configJson, 0644)
			Expect(err).NotTo(HaveOccurred())

			GinkgoT().Setenv("CONFIG", configFile)
		})

		It("exits 1", func() {
			var err error
			session, err = gexec.Start(exec.Command(goFetchBinary), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
		})
	})
})
