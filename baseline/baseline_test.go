package baseline

import (
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBaseline(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Baseline Suite")
}

var _ = Describe("ExtractFailedTestKeys", func() {
	It("extracts keys from flat failed tests", func() {
		tests := []parsers.Test{
			{Name: "TestA", PackagePath: "./pkg/a", Framework: parsers.GoTest, Failed: true},
			{Name: "TestB", PackagePath: "./pkg/a", Framework: parsers.GoTest, Passed: true},
			{Name: "TestC", PackagePath: "./pkg/b", Framework: parsers.GoTest, Failed: true},
		}
		keys := ExtractFailedTestKeys(tests)
		Expect(keys).To(HaveLen(2))
		Expect(keys).To(HaveKey(TestKey{Framework: "go test", PackagePath: "./pkg/a", FullName: "TestA"}))
		Expect(keys).To(HaveKey(TestKey{Framework: "go test", PackagePath: "./pkg/b", FullName: "TestC"}))
	})

	It("extracts keys from nested children", func() {
		tests := []parsers.Test{
			{
				Name:        "./pkg/a",
				PackagePath: "./pkg/a",
				Children: parsers.Tests{
					{Name: "TestA", PackagePath: "./pkg/a", Framework: parsers.GoTest, Failed: true},
					{Name: "TestB", PackagePath: "./pkg/a", Framework: parsers.GoTest, Passed: true},
				},
			},
		}
		keys := ExtractFailedTestKeys(tests)
		Expect(keys).To(HaveLen(1))
		Expect(keys).To(HaveKey(TestKey{Framework: "go test", PackagePath: "./pkg/a", FullName: "TestA"}))
	})

	It("handles suite paths in FullName", func() {
		tests := []parsers.Test{
			{Name: "should work", Suite: []string{"Outer", "Inner"}, PackagePath: "./pkg", Framework: parsers.Ginkgo, Failed: true},
		}
		keys := ExtractFailedTestKeys(tests)
		Expect(keys).To(HaveKey(TestKey{Framework: "ginkgo", PackagePath: "./pkg", FullName: "Outer > Inner > should work"}))
	})

	It("returns empty map for no failures", func() {
		tests := []parsers.Test{
			{Name: "TestA", PackagePath: "./pkg", Framework: parsers.GoTest, Passed: true},
		}
		Expect(ExtractFailedTestKeys(tests)).To(BeEmpty())
	})
})

var _ = Describe("ExtractViolationKeys", func() {
	It("extracts keys with rule method", func() {
		results := []*linters.LinterResult{
			{
				Linter: "eslint",
				Violations: []models.Violation{
					{File: "src/app.ts", Rule: &models.Rule{Method: "no-unused-vars"}},
					{File: "src/index.ts", Rule: &models.Rule{Method: "no-console"}},
				},
			},
		}
		keys := ExtractViolationKeys(results)
		Expect(keys).To(HaveLen(2))
		Expect(keys).To(HaveKey(ViolationKey{Linter: "eslint", File: "src/app.ts", Rule: "no-unused-vars"}))
		Expect(keys).To(HaveKey(ViolationKey{Linter: "eslint", File: "src/index.ts", Rule: "no-console"}))
	})

	It("handles nil rule", func() {
		results := []*linters.LinterResult{
			{Linter: "ruff", Violations: []models.Violation{{File: "main.py"}}},
		}
		keys := ExtractViolationKeys(results)
		Expect(keys).To(HaveKey(ViolationKey{Linter: "ruff", File: "main.py", Rule: ""}))
	})

	It("skips skipped linters", func() {
		results := []*linters.LinterResult{
			{Linter: "eslint", Skipped: true, Violations: []models.Violation{{File: "x.ts"}}},
		}
		Expect(ExtractViolationKeys(results)).To(BeEmpty())
	})
})

var _ = Describe("ExtractFailedTestPackages", func() {
	It("groups by framework and package", func() {
		tests := []parsers.Test{
			{Name: "TestA", PackagePath: "./pkg/a", Framework: parsers.GoTest, Failed: true},
			{Name: "TestB", PackagePath: "./pkg/a", Framework: parsers.GoTest, Failed: true},
			{Name: "TestC", PackagePath: "./pkg/b", Framework: parsers.GoTest, Failed: true},
			{Name: "it works", PackagePath: "./pkg/c", Framework: parsers.Jest, Failed: true},
		}
		pkgs := ExtractFailedTestPackages(tests)
		Expect(pkgs).To(HaveKey(parsers.GoTest))
		Expect(pkgs[parsers.GoTest]).To(HaveLen(2))
		Expect(pkgs[parsers.GoTest]["./pkg/a"]).To(ConsistOf("TestA", "TestB"))
		Expect(pkgs[parsers.GoTest]["./pkg/b"]).To(ConsistOf("TestC"))
		Expect(pkgs[parsers.Jest]["./pkg/c"]).To(ConsistOf("it works"))
	})

	It("skips tests with empty framework", func() {
		tests := []parsers.Test{
			{Name: "TestA", PackagePath: "./pkg", Failed: true},
		}
		Expect(ExtractFailedTestPackages(tests)).To(BeEmpty())
	})
})

var _ = Describe("ExtractFailedLintTargets", func() {
	It("returns unique linters and files", func() {
		results := []*linters.LinterResult{
			{
				Linter: "eslint",
				Violations: []models.Violation{
					{File: "src/a.ts"},
					{File: "src/b.ts"},
					{File: "src/a.ts"}, // duplicate
				},
			},
			{
				Linter: "ruff",
				Violations: []models.Violation{
					{File: "main.py"},
				},
			},
			{Linter: "golangci-lint", Skipped: true},
		}
		names, files := ExtractFailedLintTargets(results)
		Expect(names).To(ConsistOf("eslint", "ruff"))
		Expect(files).To(ConsistOf("src/a.ts", "src/b.ts", "main.py"))
	})

	It("returns nil for no violations", func() {
		results := []*linters.LinterResult{
			{Linter: "eslint", Violations: nil},
		}
		names, files := ExtractFailedLintTargets(results)
		Expect(names).To(BeNil())
		Expect(files).To(BeNil())
	})
})

var _ = Describe("FilterNewTestFailures", func() {
	It("flips baseline-known failures to passed", func() {
		baseline := map[TestKey]bool{
			{Framework: "go test", PackagePath: "./pkg/a", FullName: "TestA"}: true,
		}
		tests := []parsers.Test{
			{Name: "TestA", PackagePath: "./pkg/a", Framework: parsers.GoTest, Failed: true},
			{Name: "TestB", PackagePath: "./pkg/a", Framework: parsers.GoTest, Failed: true},
		}
		filtered := FilterNewTestFailures(tests, baseline)
		Expect(filtered[0].Failed).To(BeFalse())
		Expect(filtered[0].Passed).To(BeTrue())
		Expect(filtered[1].Failed).To(BeTrue())
	})

	It("handles nested children", func() {
		baseline := map[TestKey]bool{
			{Framework: "go test", PackagePath: "./pkg", FullName: "TestChild"}: true,
		}
		tests := []parsers.Test{
			{
				Name: "./pkg",
				Children: parsers.Tests{
					{Name: "TestChild", PackagePath: "./pkg", Framework: parsers.GoTest, Failed: true},
					{Name: "TestNew", PackagePath: "./pkg", Framework: parsers.GoTest, Failed: true},
				},
			},
		}
		filtered := FilterNewTestFailures(tests, baseline)
		Expect(filtered[0].Children[0].Failed).To(BeFalse())
		Expect(filtered[0].Children[0].Passed).To(BeTrue())
		Expect(filtered[0].Children[1].Failed).To(BeTrue())
	})

	It("clears cached summary so it recomputes", func() {
		baseline := map[TestKey]bool{
			{Framework: "go test", PackagePath: "./pkg", FullName: "TestA"}: true,
		}
		summary := parsers.TestSummary{Failed: 1, Total: 1}
		tests := []parsers.Test{
			{Name: "TestA", PackagePath: "./pkg", Framework: parsers.GoTest, Failed: true, Summary: &summary},
		}
		filtered := FilterNewTestFailures(tests, baseline)
		Expect(filtered[0].Summary).To(BeNil())
	})

	It("does nothing with empty baseline", func() {
		tests := []parsers.Test{
			{Name: "TestA", PackagePath: "./pkg", Framework: parsers.GoTest, Failed: true},
		}
		filtered := FilterNewTestFailures(tests, map[TestKey]bool{})
		Expect(filtered[0].Failed).To(BeTrue())
	})
})

var _ = Describe("FilterNewViolations", func() {
	It("removes baseline-known violations", func() {
		baseline := map[ViolationKey]bool{
			{Linter: "eslint", File: "src/a.ts", Rule: "no-console"}: true,
		}
		results := []*linters.LinterResult{
			{
				Linter: "eslint",
				Violations: []models.Violation{
					{File: "src/a.ts", Rule: &models.Rule{Method: "no-console"}},
					{File: "src/b.ts", Rule: &models.Rule{Method: "no-unused-vars"}},
				},
			},
		}
		filtered := FilterNewViolations(results, baseline)
		Expect(filtered[0].Violations).To(HaveLen(1))
		Expect(filtered[0].Violations[0].File).To(Equal("src/b.ts"))
	})

	It("keeps all violations with empty baseline", func() {
		results := []*linters.LinterResult{
			{
				Linter:     "ruff",
				Violations: []models.Violation{{File: "a.py"}, {File: "b.py"}},
			},
		}
		filtered := FilterNewViolations(results, map[ViolationKey]bool{})
		Expect(filtered[0].Violations).To(HaveLen(2))
	})

	It("skips nil and skipped results", func() {
		baseline := map[ViolationKey]bool{
			{Linter: "eslint", File: "x.ts", Rule: ""}: true,
		}
		results := []*linters.LinterResult{
			nil,
			{Linter: "eslint", Skipped: true, Violations: []models.Violation{{File: "x.ts"}}},
		}
		filtered := FilterNewViolations(results, baseline)
		Expect(filtered[0]).To(BeNil())
		Expect(filtered[1].Violations).To(HaveLen(1)) // skipped = untouched
	})
})
