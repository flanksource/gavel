package linters

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lint glob scopes", func() {
	DescribeTable("keeps matching findings and excludes other files", func(scope string) {
		root := GinkgoT().TempDir()
		selected := filepath.Join(root, "configs", "selected.yaml")
		result := &LinterResult{WorkDir: root, Violations: []models.Violation{{File: selected}, {File: filepath.Join(root, "configs", "other.json")}}}
		Expect(FilterViolationsByUserScope([]*LinterResult{result}, root, []string{scope})).To(Equal(1))
		Expect(result.Violations).To(Equal([]models.Violation{{File: selected}}))
	}, Entry("directory glob", "configs/*.yaml"), Entry("recursive glob", "**/*.yaml"))

	It("treats an existing file with glob characters as a literal path", func() {
		root := GinkgoT().TempDir()
		selected := filepath.Join(root, "data[1].json")
		Expect(os.WriteFile(selected, []byte(`{}`), 0o600)).To(Succeed())
		result := &LinterResult{WorkDir: root, Violations: []models.Violation{{File: selected}, {File: filepath.Join(root, "data1.json")}}}
		Expect(FilterViolationsByUserScope([]*LinterResult{result}, root, []string{selected})).To(Equal(1))
		Expect(result.Violations).To(Equal([]models.Violation{{File: selected}}))
	})
})
