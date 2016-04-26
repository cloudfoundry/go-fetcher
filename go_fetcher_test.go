package main_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Import Path Redirect Service", func() {
	var (
		port    string
		domain  string
		session *gexec.Session
		err     error
	)

	BeforeEach(func() {
		Expect(err).NotTo(HaveOccurred())
		port = strconv.Itoa(8182 + config.GinkgoConfig.ParallelNode)
		domain = "localhost"
		os.Setenv("PORT", port)

		args := []string{
			"-orgList", "https://github.com/cloudfoundry/,https://github.com/cloudfoundry-incubator",
			"-domain", domain,
		}

		session, err = gexec.Start(exec.Command(goFetchBinary, args...), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gbytes.Say("go-fetch-server.ready"))
	})

	AfterEach(func() {
		session.Kill().Wait()
	})

	Context("when a request specifies go-get=1", func() {
		It("responds appropriately", func() {
			res, err := http.Get("http://:" + port + "/something/something-else?go-get=1")
			Expect(err).NotTo(HaveOccurred())
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("<meta name=\"go-import\" content=\"" + domain + "/something git https://github.com/cloudfoundry/something\">"))
		})
	})
})
