package config_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/go-fetcher/config"
)

var _ = Describe("Load Configuration", func() {
	var configFile string

	BeforeEach(func() {
		jsonContent := []byte(`{
		  "importPrefix": "test",
		  "orgList": ["test_org"],
		  "NoRedirectAgents": ["test_agent"],
		  "Overrides": {
			"test_repo": {
			  "Repository": "https://golang.example.com/test_repo",
			  "Path": "src/test_repo"
			}
		  }
		}`)

		configFile = filepath.Join(GinkgoT().TempDir(), "config.json")

		Expect(os.WriteFile(configFile, jsonContent, 0644)).To(Succeed())
	})

	Context("when there is a config file", func() {
		It("returns the parsed configuration", func() {
			parsedConfig, err := config.Parse(configFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsedConfig.ImportPrefix).To(Equal("test"))
			Expect(parsedConfig.OrgList).To(Equal([]string{"test_org"}))
			Expect(parsedConfig.NoRedirectAgents).To(Equal([]string{"test_agent"}))
			Expect(parsedConfig.Overrides["test_repo"]).To(Equal(
				config.Override{
					Repository: "https://golang.example.com/test_repo",
					Path:       "src/test_repo",
				},
			))
		})
	})
})
