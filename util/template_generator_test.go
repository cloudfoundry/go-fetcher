package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry/go-fetcher/util"

	"io/ioutil"
)

var _ = Describe("Generate Application Templates", func(){

  var (
		err error
	)

	Context("When the environment variables are present", func(){
		It("should generate the application manifest", func() {
			err = util.GenerateManifest("manifest.yml.template")
			Expect(err).NotTo(HaveOccurred())
			Expect("../manifest.yml").To(BeAnExistingFile())

			var content []byte
			content, err = ioutil.ReadFile("../manifest.yml")
			Expect(string(content)).To(ContainSubstring("code-acceptance\n"))
		})

		It("should generate the json configuration", func(){
			err := util.GenerateConfig("config.json.template")
			Expect(err).NotTo(HaveOccurred())
			Expect("../config.json").To(BeAnExistingFile())

			var content []byte
			content, err = ioutil.ReadFile("../config.json")
			Expect(string(content)).To(ContainSubstring("code-acceptance.cfapps.io"))
		})
	})

	Context("When environment variables are missing", func(){
		It("should generate the application manifest", func() {
			err = util.GenerateManifest("manifest.yml.template")
			Expect(err).To(HaveOccurred())
		})

		It("should generate the application manifest", func() {
			err = util.GenerateConfig("config.json.template")
			Expect(err).To(HaveOccurred())
		})
	})
})
