package handlers_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
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

		util.GenerateConfig(os.Getenv("ROOT_DIR") + "/util/config.json.template")
		configFile := os.Getenv("ROOT_DIR") + "/config.json"
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
			res, err := http.Get("http://:" + port + "/something/something-else?go-get=1")
			Expect(err).NotTo(HaveOccurred())
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("<meta name=\"go-import\" content=\"" + c.Host + "/something git https://github.com/cloudfoundry/something\">"))
		})
	})
})
