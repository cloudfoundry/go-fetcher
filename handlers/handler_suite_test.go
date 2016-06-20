package handlers_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoFetch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoFetch Suite")
}
