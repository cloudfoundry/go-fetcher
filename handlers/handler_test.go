package handlers_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/cloudfoundry/go-fetcher/config"
	"github.com/cloudfoundry/go-fetcher/util"
	. "github.com/onsi/ginkgo"
	gconf "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Import Path Redirect Service", func() {
	var (
		port       string
		absPath    string
		configFile string
		session    *gexec.Session
		err        error
		c          *config.Config
		req        *http.Request
		res        *http.Response
		client     *http.Client
	)

	BeforeEach(func() {
		Expect(err).NotTo(HaveOccurred())
		port = strconv.Itoa(8182 + gconf.GinkgoConfig.ParallelNode)
		os.Setenv("PORT", port)
		os.Setenv("APP_NAME", "code-acceptance")
		os.Setenv("DOMAIN", "cfapps.io")

		absPath, err = filepath.Abs("..")
		Expect(err).NotTo(HaveOccurred())
		os.Setenv("ROOT_DIR", absPath)

		templateFile := os.Getenv("ROOT_DIR") + "/util/config.json.template"
		configFile = fmt.Sprintf(os.Getenv("ROOT_DIR")+"/config-%d.json", GinkgoParallelNode())
		util.GenerateConfig(templateFile, configFile)
		os.Setenv("CONFIG", configFile)
		c, err = config.Parse(configFile)
		Expect(err).NotTo(HaveOccurred())

		session, err = gexec.Start(exec.Command(goFetchBinary), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gbytes.Say("go-fetch-server.ready"))
	})

	AfterEach(func() {
		session.Kill().Wait()

		err := os.Remove(configFile)

		os.Unsetenv("APP_NAME")
		os.Unsetenv("DOMAIN")
		os.Unsetenv("PORT")
		os.Unsetenv("ROOT_DIR")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when a request specifies go-get=1", func() {
		It("responds appropriately", func() {
			res, err := http.Get("http://:" + port + "/something/something-else/test?go-get=1")
			Expect(err).NotTo(HaveOccurred())
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring(
				"<meta name=\"go-import\" content=\"" +
					c.Host +
					"/something git https://github.com/cloudfoundry/something\">"))

			Expect(body).To(ContainSubstring(
				"<meta name=\"go-source\" content=\"" +
					c.Host +
					"/something _ https://github.com/cloudfoundry/something\">"))
		})
	})

	Context("when attempting to deal with redirects", func() {

		Context("when go-get is set", func() {
			BeforeEach(func() {
				client = &http.Client{}

				req, err = http.NewRequest("GET", "http://:"+port+"/something/something-else/test?go-get=1", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the user agent is not part of the NoRedirectAgents list", func() {
				It("will redirect to godoc.org", func() {
					req.Header.Set("User-Agent", "Mozilla/5.0")
					res, err = client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					defer res.Body.Close()

					var body []byte
					body, err = ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					expectedMeta := fmt.Sprintf("<meta http-equiv=\"refresh\" content=\"0; url=https://godoc.org/%s/something/something-else/test\">", c.Host)
					Expect(body).To(ContainSubstring(expectedMeta))
				})
			})

			Context("when the user agent is part of the NoRedirectAgents list", func() {
				It("will not redirect", func() {
					for _, agent := range c.NoRedirectAgents {
						req.Header.Set("User-Agent", agent)
						res, err = client.Do(req)
						Expect(err).NotTo(HaveOccurred())
						defer res.Body.Close()

						var body []byte
						body, err = ioutil.ReadAll(res.Body)
						Expect(err).NotTo(HaveOccurred())
						expectedMeta := fmt.Sprintf("<meta http-equiv=\"refresh\" content=\"0; url=https://godoc.org/%s/something/something-else/test\">", c.Host)
						Expect(body).NotTo(ContainSubstring(expectedMeta))
					}
				})
			})
		})

		Context("when go-get is not set", func() {
			BeforeEach(func() {
				client = &http.Client{}

				req, err = http.NewRequest("GET", "http://:"+port+"/something", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the user agent is not part of the NoRedirectAgents list", func() {
				It("will redirect github.com", func() {
					req.Header.Set("User-Agent", "Mozilla/5.0")
					res, err = client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					defer res.Body.Close()

					var body []byte
					body, err = ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					expectedMeta := fmt.Sprintf("<meta http-equiv=\"refresh\" content=\"0; url=%s\">", c.OrgList[0]+"something")
					Expect(body).To(ContainSubstring(expectedMeta))
				})
			})

			Context("when the user agent is part of the NoRedirectAgents list", func() {
				It("will not redirect", func() {
					for _, agent := range c.NoRedirectAgents {
						req.Header.Set("User-Agent", agent)
						res, err = client.Do(req)
						Expect(err).NotTo(HaveOccurred())
						defer res.Body.Close()

						var body []byte
						body, err = ioutil.ReadAll(res.Body)
						Expect(err).NotTo(HaveOccurred())
						expectedMeta := fmt.Sprintf("<meta http-equiv=\"refresh\" content=\"0; url=%s\">", c.OrgList[0]+"something")
						Expect(body).NotTo(ContainSubstring(expectedMeta))
					}
				})
			})
		})
	})
})
