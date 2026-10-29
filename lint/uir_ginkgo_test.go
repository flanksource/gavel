package lint

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/baseline"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/uir"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UIR native configuration", func() {
	DescribeTable("keeps testdata projects out of automatic discovery and allows explicit targets", func(target string) {
		root := GinkgoT().TempDir()
		fixture := filepath.Join(root, "lint", "testdata", "project")
		for _, directory := range []string{root, fixture} {
			Expect(os.MkdirAll(directory, 0o700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(directory, ".uir.yaml"), []byte("version: 1\n"), 0o600)).To(Succeed())
		}
		Expect(os.WriteFile(filepath.Join(fixture, "sample.go"), []byte("package sample\n"), 0o600)).To(Succeed())
		opts := Options{WorkDir: root}
		expected := root
		switch target {
		case "directory":
			opts.Files = []string{fixture}
			expected = fixture
		case "file":
			opts.Files = []string{filepath.Join(fixture, "sample.go")}
			expected = fixture
		case "working directory":
			opts.WorkDir = fixture
			expected = fixture
		}
		Expect(resolveLinterInvocations(uir.New(root), opts)).To(HaveExactElements(WithTransform(func(inv linterInvocation) string { return inv.projectRoot }, Equal(expected))))
	}, Entry("automatic", "automatic"), Entry("explicit directory", "directory"), Entry("explicit file", "file"), Entry("explicit working directory", "working directory"))

	It("fails when the required UIR binary is missing and still previews a dry run", func() {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("PATH", "")
		_, _, err := resolveLinterExecutable(context.Background(), uir.New(root), root, root, true, false)
		Expect(err).To(MatchError(ContainSubstring("uir executable not found")))
		binary, _, err := resolveLinterExecutable(context.Background(), uir.New(root), root, root, true, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(binary).To(Equal("uir"))
	})
	It("schedules nested configs as independent invocations", func() {
		root := GinkgoT().TempDir()
		nested := filepath.Join(root, "nested")
		Expect(os.MkdirAll(nested, 0o700)).To(Succeed())
		for _, directory := range []string{root, nested} {
			Expect(os.WriteFile(filepath.Join(directory, ".uir.yaml"), []byte("version: 1\n"), 0o600)).To(Succeed())
		}
		invocations := resolveLinterInvocations(uir.New(root), Options{WorkDir: root})
		Expect(invocations).To(HaveLen(2))
		Expect(invocations[0].linter).NotTo(BeIdenticalTo(invocations[1].linter))
		for _, invocation := range invocations {
			Expect(isInvocationCovered(invocation, []string{root})).To(BeFalse())
		}
	})
	It("activates only with configuration and registers the adapter", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		root := GinkgoT().TempDir()
		adapter := uir.New(root)
		selected, _ := shouldSelectLinter(root, verify.GavelConfig{}, adapter, false)
		Expect(selected).To(BeFalse())
		Expect(os.WriteFile(filepath.Join(root, ".uir.yaml"), []byte("version: 1\n"), 0o600)).To(Succeed())
		selected, _ = shouldSelectLinter(root, verify.GavelConfig{}, adapter, false)
		Expect(selected).To(BeTrue())
		Expect(builtinLinterRegistry(root).Has("uir")).To(BeTrue())
	})

	It("schedules the requested project with only global directory configuration", func() {
		home, root := GinkgoT().TempDir(), GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		Expect(os.Mkdir(filepath.Join(home, ".uir.d"), 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(home, ".uir.d", "rules.yaml"), []byte("version: 1\n"), 0o600)).To(Succeed())
		adapter := uir.New(root)
		invocations := resolveLinterInvocations(adapter, Options{WorkDir: root})
		Expect(invocations).To(HaveLen(1))
		Expect(invocations[0].projectRoot).To(Equal(root))
		selected, _ := shouldSelectLinter(root, verify.GavelConfig{}, adapter, false)
		Expect(selected).To(BeTrue())
	})

	It("discovers directory-only project configs and expands fragment scopes", func() {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		directory := filepath.Join(root, ".uir.d")
		Expect(os.Mkdir(directory, 0o700)).To(Succeed())
		path := filepath.Join(directory, "rules.yaml")
		Expect(os.WriteFile(path, []byte("version: 1\n"), 0o600)).To(Succeed())
		invocations := resolveLinterInvocations(uir.New(root), Options{WorkDir: root, Files: []string{path}})
		Expect(invocations).To(HaveLen(1))
		Expect(invocations[0].projectRoot).To(Equal(root))
		result := &linters.LinterResult{Linter: "uir", WorkDir: root}
		Expect(resultScopes(result, root, []string{path})).To(ContainElement(root))
	})

	It("keeps file-scoped runs inside a module beneath globally configured home", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		root := filepath.Join(home, "project")
		Expect(os.Mkdir(root, 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(home, ".uir.yaml"), []byte("version: 1\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.org/project\n"), 0o600)).To(Succeed())
		invocations := resolveLinterInvocations(uir.New(root), Options{WorkDir: root, Files: []string{"sample.go"}})
		Expect(invocations).To(HaveLen(1))
		Expect(invocations[0].projectRoot).To(Equal(root))
	})
	It("expands config changes to the project but filters ordinary file scopes after analysis", func() {
		root := GinkgoT().TempDir()
		config := filepath.Join(root, ".uir.yaml")
		Expect(os.WriteFile(config, []byte("version: 1\n"), 0o600)).To(Succeed())
		makeResult := func() *linters.LinterResult {
			return &linters.LinterResult{Linter: "uir", WorkDir: root, Success: true, Violations: []models.Violation{{File: "a.go"}, {File: "b.go"}, {Severity: models.SeverityWarning}}}
		}
		result := makeResult()
		Expect(applyPostLintFilters([]*linters.LinterResult{result}, root, []string{config})).To(Succeed())
		Expect(result.Violations).To(HaveLen(3))
		result = makeResult()
		Expect(applyPostLintFilters([]*linters.LinterResult{result}, root, []string{"a.go"})).To(Succeed())
		Expect(result.Violations).To(HaveLen(2))
	})
	It("uses existing baseline keys and retains execution failures", func() {
		finding := models.NewViolationBuilder().WithFile("a.go").WithMessage("too large").WithRuleFromLinter("uir", "body").Build()
		finding.Severity = models.SeverityError
		original := []*linters.LinterResult{{Linter: "uir", Success: true, Violations: []models.Violation{finding}}}
		current := []*linters.LinterResult{{Linter: "uir", Success: true, Violations: []models.Violation{finding}}}
		filtered := baseline.FilterNewViolations(current, baseline.ExtractViolationKeys(original))
		Expect(linters.HasFailures(filtered)).To(BeFalse())
		current[0].Success, current[0].Error = false, "execution failed"
		filtered = baseline.FilterNewViolations(current, baseline.ExtractViolationKeys(original))
		Expect(linters.HasFailures(filtered)).To(BeTrue())
	})
})
