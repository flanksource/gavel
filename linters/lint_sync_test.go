package linters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/todos"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLintSync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lint Sync Suite")
}

var _ = Describe("SyncLintTodos", func() {
	var todosDir string

	BeforeEach(func() {
		todosDir = GinkgoT().TempDir()
	})

	makeResults := func(violations ...models.Violation) []*LinterResult {
		return []*LinterResult{{
			Linter:     "golangci-lint",
			Success:    true,
			Violations: violations,
		}}
	}

	violation := func(file string, line int, msg, rule string) models.Violation {
		v := models.Violation{
			File:     file,
			Line:     line,
			Message:  models.StringPtr(msg),
			Source:   "golangci-lint",
			Severity: models.SeverityWarning,
		}
		if rule != "" {
			v.Rule = &models.Rule{Method: rule}
		}
		return v
	}

	It("creates TODO files grouped by file", func() {
		results := makeResults(
			violation("pkg/foo.go", 10, "unused var", "unused"),
			violation("pkg/foo.go", 20, "error not checked", "errcheck"),
			violation("pkg/bar.go", 5, "shadowed var", "shadow"),
		)

		syncResult, err := SyncLintTodos(results, SyncOptions{
			TodosDir: todosDir,
			GroupBy:  GroupByFile,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(syncResult.Created).To(HaveLen(2))
		Expect(syncResult.Updated).To(BeEmpty())
		Expect(syncResult.Completed).To(BeEmpty())

		// Verify files exist
		entries, _ := os.ReadDir(todosDir)
		Expect(entries).To(HaveLen(2))

		// Verify content of one file
		fooPath := filepath.Join(todosDir, "pkg-foo-go.md")
		parsed, err := todos.ParseFrontmatterFromFile(fooPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(parsed.Frontmatter.Status).To(BeEquivalentTo("pending"))
		Expect(parsed.Frontmatter.Priority).To(BeEquivalentTo("medium"))
		Expect(parsed.Frontmatter.Path).To(ContainElement("pkg/foo.go"))
	})

	It("creates TODO files grouped by package", func() {
		results := makeResults(
			violation("pkg/a/foo.go", 10, "unused", "unused"),
			violation("pkg/a/bar.go", 5, "shadow", "shadow"),
			violation("pkg/b/baz.go", 1, "error", "errcheck"),
		)

		syncResult, err := SyncLintTodos(results, SyncOptions{
			TodosDir: todosDir,
			GroupBy:  GroupByPackage,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(syncResult.Created).To(HaveLen(2)) // pkg/a and pkg/b
	})

	It("creates TODO files grouped by message", func() {
		results := makeResults(
			violation("pkg/foo.go", 10, "unused var", "unused"),
			violation("pkg/bar.go", 5, "unused var", "unused"),
			violation("pkg/baz.go", 1, "error not checked", "errcheck"),
		)

		syncResult, err := SyncLintTodos(results, SyncOptions{
			TodosDir: todosDir,
			GroupBy:  GroupByMessage,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(syncResult.Created).To(HaveLen(2)) // unused and errcheck
	})

	It("updates existing TODOs on re-sync", func() {
		results := makeResults(
			violation("pkg/foo.go", 10, "unused var", "unused"),
		)
		_, err := SyncLintTodos(results, SyncOptions{
			TodosDir: todosDir,
			GroupBy:  GroupByFile,
		})
		Expect(err).NotTo(HaveOccurred())

		// Re-sync with updated violations
		results = makeResults(
			violation("pkg/foo.go", 10, "unused var", "unused"),
			violation("pkg/foo.go", 30, "new issue", "govet"),
		)
		syncResult, err := SyncLintTodos(results, SyncOptions{
			TodosDir: todosDir,
			GroupBy:  GroupByFile,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(syncResult.Created).To(BeEmpty())
		Expect(syncResult.Updated).To(HaveLen(1))
	})

	It("completes TODOs when violations are resolved", func() {
		results := makeResults(
			violation("pkg/foo.go", 10, "unused", "unused"),
			violation("pkg/bar.go", 5, "shadow", "shadow"),
		)
		_, err := SyncLintTodos(results, SyncOptions{
			TodosDir: todosDir,
			GroupBy:  GroupByFile,
		})
		Expect(err).NotTo(HaveOccurred())

		// Re-sync with only foo.go violations - bar.go should complete
		results = makeResults(
			violation("pkg/foo.go", 10, "unused", "unused"),
		)
		syncResult, err := SyncLintTodos(results, SyncOptions{
			TodosDir: todosDir,
			GroupBy:  GroupByFile,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(syncResult.Completed).To(HaveLen(1))

		barPath := filepath.Join(todosDir, "pkg-bar-go.md")
		parsed, err := todos.ParseFrontmatterFromFile(barPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(parsed.Frontmatter.Status).To(BeEquivalentTo("completed"))
	})

	It("reopens completed TODOs when violations reappear", func() {
		results := makeResults(
			violation("pkg/foo.go", 10, "unused", "unused"),
		)
		_, err := SyncLintTodos(results, SyncOptions{
			TodosDir: todosDir,
			GroupBy:  GroupByFile,
		})
		Expect(err).NotTo(HaveOccurred())

		// Complete it
		_, err = SyncLintTodos([]*LinterResult{{
			Linter:  "golangci-lint",
			Success: true,
		}}, SyncOptions{
			TodosDir: todosDir,
			GroupBy:  GroupByFile,
		})
		Expect(err).NotTo(HaveOccurred())

		// Re-sync with violations - should reopen
		syncResult, err := SyncLintTodos(results, SyncOptions{
			TodosDir: todosDir,
			GroupBy:  GroupByFile,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(syncResult.Updated).To(HaveLen(1))

		fooPath := filepath.Join(todosDir, "pkg-foo-go.md")
		parsed, err := todos.ParseFrontmatterFromFile(fooPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(parsed.Frontmatter.Status).To(BeEquivalentTo("pending"))
	})

	It("maps severity to priority correctly", func() {
		results := []*LinterResult{{
			Linter:  "golangci-lint",
			Success: true,
			Violations: []models.Violation{
				{File: "err.go", Line: 1, Severity: models.SeverityError, Source: "golangci-lint", Message: models.StringPtr("err")},
			},
		}}
		_, err := SyncLintTodos(results, SyncOptions{
			TodosDir: todosDir,
			GroupBy:  GroupByFile,
		})
		Expect(err).NotTo(HaveOccurred())

		parsed, err := todos.ParseFrontmatterFromFile(filepath.Join(todosDir, "err-go.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(parsed.Frontmatter.Priority).To(BeEquivalentTo("high"))
	})

	It("skips failed linter results", func() {
		results := []*LinterResult{{
			Linter:  "golangci-lint",
			Success: false,
			Violations: []models.Violation{
				violation("pkg/foo.go", 10, "unused", "unused"),
			},
		}}
		syncResult, err := SyncLintTodos(results, SyncOptions{
			TodosDir: todosDir,
			GroupBy:  GroupByFile,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(syncResult.Created).To(BeEmpty())
	})
})
