package config_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
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
		tmpDir, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		jsonContent := []byte(` {
		  "importPrefix": "test",
		  "orgList": ["test_org"],
		  "NoRedirectAgents": ["test_agent"],
		  "Overrides": {
			"test_repo": {
			  "Repository": "https://golang.example.com/test_repo",
			  "Path": "src/test_repo"
			}
		  },
		  "IndexPath": "some_relative/path"
		}`)

		err = os.WriteFile(tmpDir+"/config.json", jsonContent, 0644)
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
			Expect(c.Overrides["test_repo"]).To(Equal(config.Override{Repository: "https://golang.example.com/test_repo", Path: "src/test_repo"}))
			Expect(c.IndexPath).To(Equal("some_relative/path"))
		})
	})
})
