package main

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("config runtime resolution", func() {
	It("reports invalid saved runtime defaults without affecting structural config output", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		Expect(os.WriteFile(filepath.Join(home, ".captain.yaml"), []byte("ai: [broken"), 0o600)).To(Succeed())
		target := GinkgoT().TempDir()
		_, err := runConfig(ConfigOptions{Args: []string{target}})
		Expect(err).NotTo(HaveOccurred())
		_, err = runConfig(ConfigOptions{Args: []string{target}, Resolve: true})
		Expect(err).To(MatchError(ContainSubstring("load saved AI defaults")))
	})
})
