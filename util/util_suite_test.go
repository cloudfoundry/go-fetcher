package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"os"
	"path/filepath"
)

var err  error

func TestUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Util Suite")
}

var _ = BeforeSuite(func(){
	os.Setenv("APP_NAME", "code-acceptance")
	os.Setenv("DOMAIN", "cfapps.io")

	absPath, err := filepath.Abs("..")
	Expect(err).NotTo(HaveOccurred())
	os.Setenv("ROOT_DIR", absPath)
})

var _ = AfterSuite(func(){
	err = os.Remove(os.Getenv("ROOT_DIR") + "/manifest.yml")
	Expect(err).NotTo(HaveOccurred())
	err = os.Remove(os.Getenv("ROOT_DIR") + "/config.json")
	Expect(err).NotTo(HaveOccurred())

	os.Unsetenv("APP_NAME")
	os.Unsetenv("DOMAIN")
	os.Unsetenv("ROOT_DIR")
})
