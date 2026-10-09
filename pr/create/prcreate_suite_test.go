package prcreate

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPRCreateSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PR Create Suite")
}
