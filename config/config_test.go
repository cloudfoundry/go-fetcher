package config_test

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/go-fetcher/config"
)

var _ = Describe("Load Configuration", func() {

	var (
		tmpDir   string
		filePath string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		jsonContent := []byte(fmt.Sprintf(` {
				"importPrefix": "test",
				"orgList": ["test_org"],
				"NoRedirectAgents": ["test_agent"]
		}`))

		err = ioutil.WriteFile(tmpDir+"/config.json", jsonContent, 0644)
		Expect(err).NotTo(HaveOccurred())
		filePath = tmpDir + "/config.json"
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when there is a config file", func() {
		It("returns the parsed configuration", func() {
			c, err := config.Parse(filePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(c.ImportPrefix).To(Equal("test"))
			Expect(c.OrgList).To(Equal([]string{"test_org"}))
			Expect(c.NoRedirectAgents).To(Equal([]string{"test_agent"}))
		})
	})

})
