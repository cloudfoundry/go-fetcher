package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestGoFetch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoFetch Suite")
}

var goFetchBinary string

var _ = SynchronizedBeforeSuite(func() []byte {
	goFetchServerPath, err := gexec.Build("github.com/cloudfoundry/go-fetcher/")
	Expect(err).NotTo(HaveOccurred())
	return []byte(goFetchServerPath)
}, func(goFetchServerPath []byte) {
	goFetchBinary = string(goFetchServerPath)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})
