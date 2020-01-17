package main_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/cloudfoundry/go-fetcher/config"
	. "github.com/onsi/ginkgo"
	gconf "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Import Path Redirect Service", func() {
	var (
		port             string
		configFile       string
		session          *gexec.Session
		conf             *config.Config
		fakeGithubServer *ghttp.Server
	)

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

		port = strconv.Itoa(8182 + gconf.GinkgoConfig.ParallelNode)

		os.Setenv("PORT", port)

		configFile = fmt.Sprintf("config-%d.json", GinkgoParallelNode())
		conf = &config.Config{
			LogLevel:         "debug",
			ImportPrefix:     "the.canonical.import.path",
			GithubURL:        fakeGithubServer.URL(),
			OrgList:          []string{"cloudfoundry", "cloudfoundry-incubator", "cloudfoundry-attic"},
			NoRedirectAgents: []string{"some-agent", "some-other-agent"},
		}

		bytes, err := json.Marshal(conf)
		Expect(err).NotTo(HaveOccurred())
		err = ioutil.WriteFile(configFile, bytes, 0644)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("CONFIG", configFile)

		session, err = gexec.Start(exec.Command(goFetchBinary), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gbytes.Say("go-fetcher.started"))
	})

	AfterEach(func() {
		session.Kill().Wait()
		fakeGithubServer.Close()

		err := os.Remove(configFile)

		os.Unsetenv("APP_NAME")
		os.Unsetenv("DOMAIN")
		os.Unsetenv("PORT")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the user agent is part of the NoRedirectAgents list", func() {
		It("responds appropriately", func() {
			client := &http.Client{}
			req, err := http.NewRequest("GET", "http://:"+port+"/repository-1/something-else/test", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("User-Agent", conf.NoRedirectAgents[0])

			res, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring(fmt.Sprintf(
				`<meta name="go-import" content="%s/repository-1 git %s/cloudfoundry/repository-1">`,
				conf.ImportPrefix,
				fakeGithubServer.URL())))

			Expect(body).To(ContainSubstring(fmt.Sprintf(
				`<meta name="go-source" content="%s/repository-1 _ %s/cloudfoundry/repository-1">`,
				conf.ImportPrefix,
				fakeGithubServer.URL())))
		})
	})

	Describe("Redirects", func() {
		Context("when go-get is not set", func() {
			var redirectCount int
			var client *http.Client

			BeforeEach(func() {
				redirectCount = 0

				client = &http.Client{
					CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
						redirectCount++
						return errors.New("don't follow redirect in test")
					},
				}
			})

			Context("when the repo is in cloudfoundry", func() {
				It("will redirect to the true cloudfoundry source via HTTP redirects", func() {
					req, err := http.NewRequest("GET", "http://:"+port+"/repository-2", nil)
					Expect(err).NotTo(HaveOccurred())

					res, err := client.Do(req)
					Expect(res).NotTo(BeNil())
					Expect(res.StatusCode).To(Equal(http.StatusFound))
					Expect(res.Header.Get("Location")).To(Equal(fmt.Sprintf("%s/cloudfoundry/repository-2", fakeGithubServer.URL())))
					Expect(err).To(MatchError(ContainSubstring("don't follow redirect in test")))

					Expect(redirectCount).To(Equal(1))
				})
			})

			Context("when the repo is in cloudfoundry-incubator", func() {
				It("will redirect to the true cloudfoundry-incubator source via HTTP redirects", func() {
					req, err := http.NewRequest("GET", "http://:"+port+"/repo-in-incubator", nil)
					Expect(err).NotTo(HaveOccurred())

					res, err := client.Do(req)
					Expect(res).NotTo(BeNil())
					Expect(res.StatusCode).To(Equal(http.StatusFound))
					Expect(res.Header.Get("Location")).To(Equal(fmt.Sprintf("%s/cloudfoundry-incubator/repo-in-incubator", fakeGithubServer.URL())))
					Expect(err).To(MatchError(ContainSubstring("don't follow redirect in test")))

					Expect(redirectCount).To(Equal(1))
				})
			})

			Context("when the repo is in cloudfoundry-attic", func() {
				It("will redirect to the true cloudfoundry-attic source via HTTP redirects", func() {
					req, err := http.NewRequest("GET", "http://:"+port+"/repo-in-attic", nil)
					Expect(err).NotTo(HaveOccurred())

					res, err := client.Do(req)
					Expect(res).NotTo(BeNil())
					Expect(res.StatusCode).To(Equal(http.StatusFound))
					Expect(res.Header.Get("Location")).To(Equal(fmt.Sprintf("%s/cloudfoundry-attic/repo-in-attic", fakeGithubServer.URL())))
					Expect(err).To(MatchError(ContainSubstring("don't follow redirect in test")))

					Expect(redirectCount).To(Equal(1))
				})
			})
		})

		Context("when go-get is set", func() {
			It("will redirect to godoc.org with an HTML meta tag redirect", func() {
				client := &http.Client{}

				req, err := http.NewRequest("GET", "http://:"+port+"/repository-1/test?go-get=1", nil)
				Expect(err).NotTo(HaveOccurred())

				res, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer res.Body.Close()

				var body []byte
				body, err = ioutil.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				expectedMeta := fmt.Sprintf("<meta http-equiv=\"refresh\" content=\"0; url=https://godoc.org/%s/repository-1/test\">", conf.ImportPrefix)
				Expect(body).To(ContainSubstring(expectedMeta))
			})
		})
	})
})
