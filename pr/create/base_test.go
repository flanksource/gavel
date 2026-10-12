package prcreate

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResolveBase", func() {
	const configuredBase = "origin/develop"
	var dir string

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir = GinkgoT().TempDir()
	})

	writeConfig := func(base string) {
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("pr:\n  base: "+base+"\n"), 0o600)).To(Succeed())
	}

	It("prefers an explicit override over the configured base", func() {
		writeConfig(configuredBase)
		Expect(ResolveBase(dir, "release")).To(Equal("release"))
	})

	It("uses the configured pr.base when no override is given", func() {
		writeConfig(configuredBase)
		Expect(ResolveBase(dir, "")).To(Equal(configuredBase))
	})

	It("falls back to the default base when nothing is configured", func() {
		Expect(ResolveBase(dir, "")).To(Equal(DefaultBase))
	})

	It("fails on an unreadable configuration", func() {
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("pr: [unterminated\n"), 0o600)).To(Succeed())
		_, err := ResolveBase(dir, "")
		Expect(err).To(MatchError(ContainSubstring("load PR configuration")))
	})
})
