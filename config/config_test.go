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
		  "GithubAPI": "https://api.github.com/",
		  "GithubAPIKey": "secret-github-api-key-should-be-ignored",
		  "ImportPrefix": "import-prefix-should-come-from-env",
		  "LogLevel": "info",
		  "NoRedirectAgents": [
			"Go-http-client",
			"GoDocBot"
		  ],
		  "OrgList": [
			"cloudfoundry",
			"cloudfoundry-incubator",
			"cloudfoundry-attic"
		  ],
		  "Overrides": {
			"config-server": {
			  "Repository": "https://github.com/cloudfoundry/config-server-release",
			  "Path": "src/config-server"
			},
			"guardian": {
			  "Repository": "https://github.com/cloudfoundry/garden-runc-release",
			  "Path": "src/code.cloudfoundry.org/guardian"
			},
			"stager": {
			  "Repository": "https://github.com/cloudfoundry-incubator/stager"
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

			Expect(parsedConfig.GithubAPI).To(Equal("https://api.github.com/"))
			Expect(parsedConfig.GithubAPIKey).To(Equal(""))
			Expect(parsedConfig.ImportPrefix).To(Equal(""))
			Expect(parsedConfig.LogLevel).To(Equal("info"))
			Expect(parsedConfig.NoRedirectAgents).To(Equal([]string{"Go-http-client", "GoDocBot"}))
			Expect(parsedConfig.OrgList).To(Equal([]string{"cloudfoundry", "cloudfoundry-incubator", "cloudfoundry-attic"}))

			expectedOverrides := map[string]config.Override{
				"config-server": {
					Repository: "https://github.com/cloudfoundry/config-server-release",
					Path:       "src/config-server",
				},
				"guardian": {
					Repository: "https://github.com/cloudfoundry/garden-runc-release",
					Path:       "src/code.cloudfoundry.org/guardian",
				},
				"stager": {
					Repository: "https://github.com/cloudfoundry-incubator/stager",
				},
			}
			Expect(parsedConfig.Overrides).To(Equal(expectedOverrides))
		})
	})
})
