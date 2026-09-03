package config_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/go-fetcher/config"
)

var _ = Describe("config", func() {

	Describe("Parse", func() {
		var configFile string

		BeforeEach(func() {
			configJson := []byte(`{
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

			Expect(os.WriteFile(configFile, configJson, 0644)).To(Succeed())
		})

		It("returns the parsed configuration", func() {
			cfg, err := config.Parse(configFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(cfg.GithubAPI).To(Equal("https://api.github.com/"))
			Expect(cfg.GithubAPIKey).To(Equal(""))
			Expect(cfg.ImportPrefix).To(Equal(""))
			Expect(cfg.LogLevel).To(Equal("info"))
			Expect(cfg.NoRedirectAgents).To(Equal([]string{"Go-http-client", "GoDocBot"}))
			Expect(cfg.OrgList).To(Equal([]string{"cloudfoundry", "cloudfoundry-incubator", "cloudfoundry-attic"}))

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
			Expect(cfg.Overrides).To(Equal(expectedOverrides))
		})
	})

	Describe("Config.PopulateFromEnv()", func() {
		var cfg *config.Config

		BeforeEach(func() {
			cfg = &config.Config{}
		})

		Context("when NO expected ENV vars are set", func() {
			It("loads the default values", func() {
				cfg.PopulateFromEnv()

				Expect(cfg.GithubAPIKey).To(Equal(""))
				Expect(cfg.ImportPrefix).To(Equal("code.cloudfoundry.org"))
			})
		})

		Context("when expected ENV vars are set", func() {
			var githubApiKey = "secret-github-api-key"
			var importPrefix = "some.import-prefix.tld"

			BeforeEach(func() {
				GinkgoT().Setenv("GITHUB_API_KEY", githubApiKey)
				GinkgoT().Setenv("IMPORT_PREFIX", importPrefix)
			})

			It("loads the ENV values", func() {
				cfg.PopulateFromEnv()

				Expect(cfg.GithubAPIKey).To(Equal(githubApiKey))
				Expect(cfg.ImportPrefix).To(Equal(importPrefix))
			})
		})
	})
})
