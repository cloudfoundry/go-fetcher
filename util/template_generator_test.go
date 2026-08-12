package util_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/go-fetcher/util"
)

var _ = Describe("Generate Application Templates", func() {
	var testDir string
	var manifestTargetFile string
	var configTargetFile string

	BeforeEach(func() {
		testDir = GinkgoT().TempDir()
	})

	Context("with environment overrides", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("APP_NAME", "golang-redirector")
			GinkgoT().Setenv("DOMAIN", "example.com")
		})

		Describe("manifest generation", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("INSTANCES", "10")
				GinkgoT().Setenv("MEMORY", "20M")
				GinkgoT().Setenv("DISK_QUOTA", "30M")

				manifestTargetFile = filepath.Join(testDir, fmt.Sprintf("manifest-%d.yml", GinkgoParallelProcess()))
				Expect(util.GenerateManifest(manifestTargetFile)).To(Succeed())
			})

			It("should generate the application manifest", func() {
				Expect(manifestTargetFile).To(BeAnExistingFile())

				content, err := os.ReadFile(manifestTargetFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal(`---
applications:
  - name: golang-redirector
    memory: 20M
    instances: 10
    disk_quota: 30M
    routes:
    - route: golang-redirector.example.com
    env:
      CONFIG: config.json
`))
			})
		})

		Describe("config generation", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("GITHUB_APIKEY", "FAKE_GIT_HUB_API_KEY")

				configTargetFile = filepath.Join(testDir, fmt.Sprintf("config-%d.json", GinkgoParallelProcess()))
				Expect(util.GenerateConfig(configTargetFile)).To(Succeed())
			})

			It("should generate the json configuration", func() {
				Expect(configTargetFile).To(BeAnExistingFile())

				content, err := os.ReadFile(configTargetFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal(`{
  "LogLevel": "info",
  "ImportPrefix": "golang-redirector.example.com",
  "OrgList": [
    "cloudfoundry",
    "cloudfoundry-incubator",
    "cloudfoundry-attic"
  ],
  "NoRedirectAgents": [
    "Go-http-client",
    "GoDocBot"
  ],
  "Overrides": {
    "config-server": {
      "Repository": "https://github.com/cloudfoundry/config-server-release",
      "Path": "src/config-server"
    },
    "stager": {
      "Repository": "https://github.com/cloudfoundry-incubator/stager"
    }
  },
  "GithubURL": "https://api.github.com",
  "GithubAPIKey": "FAKE_GIT_HUB_API_KEY",
  "IndexPath": "public/index.html"
}
`))

			})
		})
	})

	Context("without environment overrides", func() {
		Describe("manifest generation", func() {
			BeforeEach(func() {
				manifestTargetFile = filepath.Join(testDir, fmt.Sprintf("manifest-%d.yml", GinkgoParallelProcess()))
				Expect(util.GenerateManifest(manifestTargetFile)).To(Succeed())
			})

			It("should generate the application manifest", func() {
				Expect(manifestTargetFile).To(BeAnExistingFile())

				content, err := os.ReadFile(manifestTargetFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal(`---
applications:
  - name: code
    memory: 128M
    instances: 2
    disk_quota: 128M
    routes:
    - route: code.cloudfoundry.org
    env:
      CONFIG: config.json
`))
			})
		})

		Describe("config generation", func() {
			BeforeEach(func() {
				configTargetFile = filepath.Join(testDir, fmt.Sprintf("config-%d.json", GinkgoParallelProcess()))
				Expect(util.GenerateConfig(configTargetFile)).To(Succeed())
			})

			It("should generate the application manifest", func() {
				Expect(configTargetFile).To(BeAnExistingFile())

				content, err := os.ReadFile(configTargetFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal(`{
  "LogLevel": "info",
  "ImportPrefix": "code.cloudfoundry.org",
  "OrgList": [
    "cloudfoundry",
    "cloudfoundry-incubator",
    "cloudfoundry-attic"
  ],
  "NoRedirectAgents": [
    "Go-http-client",
    "GoDocBot"
  ],
  "Overrides": {
    "config-server": {
      "Repository": "https://github.com/cloudfoundry/config-server-release",
      "Path": "src/config-server"
    },
    "stager": {
      "Repository": "https://github.com/cloudfoundry-incubator/stager"
    }
  },
  "GithubURL": "https://api.github.com",
  "IndexPath": "public/index.html"
}
`))
			})
		})
	})
})
