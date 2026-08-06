package runners

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("build tag mappings", func() {
	It("maps tags to go test", func() {
		Expect(NewGoTest(GinkgoT().TempDir()).BuildTagsArgs([]string{"integration", "postgres"})).To(
			Equal([]string{"-tags=integration,postgres"}),
		)
	})

	It("maps tags to ginkgo", func() {
		Expect(NewGinkgo(GinkgoT().TempDir()).BuildTagsArgs([]string{"integration", "postgres"})).To(
			Equal([]string{"--tags=integration,postgres"}),
		)
	})

	It("excludes packages whose tests are disabled by build constraints", func() {
		workDir := GinkgoT().TempDir()
		packageDir := filepath.Join(workDir, "e2e")
		Expect(os.MkdirAll(packageDir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(packageDir, "e2e_test.go"), []byte(`//go:build e2e

package e2e

import "testing"

func TestE2E(t *testing.T) {}
`), 0o644)).To(Succeed())

		runner := NewGoTest(workDir)
		packages, err := runner.DiscoverPackages(workDir, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(packages).To(BeEmpty())

		runner.SetBuildTags([]string{"e2e"})
		packages, err = runner.DiscoverPackages(workDir, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(packages).To(Equal([]string{"./e2e"}))
	})

	It("applies build constraints before classifying Ginkgo packages", func() {
		workDir := GinkgoT().TempDir()
		packageDir := filepath.Join(workDir, "integration")
		Expect(os.MkdirAll(packageDir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(packageDir, "suite_test.go"), []byte(`//go:build integration

package integration

import . "github.com/onsi/ginkgo/v2"
`), 0o644)).To(Succeed())

		runner := NewGinkgo(workDir)
		packages, err := runner.DiscoverPackages(workDir, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(packages).To(BeEmpty())

		runner.SetBuildTags([]string{"integration"})
		packages, err = runner.DiscoverPackages(workDir, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(packages).To(Equal([]string{"./integration"}))
	})
})
