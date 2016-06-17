package util_test

import (
	"github.com/cloudfoundry/go-fetcher/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var _ = Describe("Generate Application Templates", func() {

	var (
		err                                      error
		manifestTemplateFile, manifestTargetFile string
		configTemplateFile, configTargetFile     string
	)

	AfterEach(func() {
		err = os.Remove(manifestTargetFile)
		Expect(err).NotTo(HaveOccurred())
		err = os.Remove(configTargetFile)
		Expect(err).NotTo(HaveOccurred())

		os.Unsetenv("APP_NAME")
		os.Unsetenv("DOMAIN")
		os.Unsetenv("ROOT_DIR")
		os.Unsetenv("SERVICE_NAME")
		os.Unsetenv("GITHUB_APIKEY")
	})

	BeforeEach(func() {
		os.Setenv("APP_NAME", "code-acceptance")
		os.Setenv("DOMAIN", "cfapps.io")
		os.Setenv("SERVICE_NAME", "code-acceptance-papertrail")
		os.Setenv("GITHUB_APIKEY", "some-key-key")

		absPath, err := filepath.Abs("..")
		Expect(err).NotTo(HaveOccurred())
		os.Setenv("ROOT_DIR", absPath)

		manifestTemplateFile = "manifest.yml.template"
		manifestTargetFile = fmt.Sprintf("../manifest-%d.yml", GinkgoParallelNode())
		err = util.GenerateManifest(manifestTemplateFile, manifestTargetFile)
		Expect(err).NotTo(HaveOccurred())

		configTemplateFile = "config.json.template"
		configTargetFile = fmt.Sprintf("../config-%d.json", GinkgoParallelNode())
		err = util.GenerateConfig(configTemplateFile, configTargetFile)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("When the environment variables are present", func() {

		It("should generate the application manifest", func() {
			Expect(manifestTargetFile).To(BeAnExistingFile())

			var content []byte
			content, err = ioutil.ReadFile(manifestTargetFile)
			Expect(string(content)).To(ContainSubstring("code-acceptance\n"))
		})

		It("should generate the json configuration", func() {
			Expect(configTargetFile).To(BeAnExistingFile())

			var content []byte
			content, err = ioutil.ReadFile(configTargetFile)
			Expect(string(content)).To(ContainSubstring("code-acceptance.cfapps.io"))
		})
	})

	Context("When environment variables are missing", func() {

		JustBeforeEach(func() {
			os.Unsetenv("APP_NAME")
			os.Unsetenv("DOMAIN")
			os.Unsetenv("ROOT_DIR")
		})

		It("should generate the application manifest", func() {
			err = util.GenerateManifest(manifestTemplateFile, manifestTargetFile)
			Expect(err).To(HaveOccurred())
		})

		It("should generate the application manifest", func() {
			err = util.GenerateConfig(configTemplateFile, configTargetFile)
			Expect(err).To(HaveOccurred())
		})
	})
})
