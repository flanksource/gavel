package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGavelCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gavel CLI Suite")
}

// Specs here load .gavel.yaml, which layers ~/.gavel.yaml under the workspace's,
// so without an isolated HOME they assert against the developer's own config.
var _ = BeforeEach(func() {
	GinkgoT().Setenv("HOME", GinkgoT().TempDir())
})
