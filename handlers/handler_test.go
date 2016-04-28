package handlers_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
  "fmt"
	"strconv"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
  "github.com/cloudfoundry/go-fetcher/config"
  "github.com/cloudfoundry/go-fetcher/util"
	gconf "github.com/onsi/ginkgo/config"
)

var _ = Describe("Import Path Redirect Service", func() {
	var (
		port    string
		absPath string
		session *gexec.Session
		err     error
		c       *config.Config
		req     *http.Request
		res     *http.Response
		client  *http.Client
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
		configFile := os.Getenv("ROOT_DIR") + "/config.json"
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

		err := os.Remove(os.Getenv("ROOT_DIR") + "/config.json")

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
    BeforeEach( func() {
      client = &http.Client{}

      req, err = http.NewRequest("GET",
				"http://:" + port +
				"/something/something-else/test?go-get=1", nil)

      Expect(err).NotTo(HaveOccurred())
    })

    Context("when the user agent is not part of the NoRedirectAgents list", func() {
			It("will redirect", func() {
        req.Header.Set("User-Agent", "Mozilla/5.0")
				res, err = client.Do(req)
			  Expect(err).NotTo(HaveOccurred())
			  defer res.Body.Close()

				var body []byte
        body, err = ioutil.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				expectedMeta := fmt.Sprintf("<meta http-equiv=\"refresh\" content=\"0; url=https://godoc.org/%s/something\">", c.Host)
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
					Expect(body).NotTo(ContainSubstring("<meta http-equiv=\"refresh\" content=\"0; url=https://godoc.org/something/something\">"))
				}
      })
		})
	})
})
