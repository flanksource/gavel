package lint

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/git/changeset"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/golangci"
	"github.com/flanksource/gavel/linters/ruff"
	"github.com/flanksource/gavel/linters/uir"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Changed lint execution", func() {
	var root string
	BeforeEach(func() {
		root = GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("GAVEL_CHANGED_BASE", "")
		GinkgoT().Setenv("CI", "false")
		Expect(os.Mkdir(filepath.Join(root, "bin"), 0o755)).To(Succeed())
		script, err := os.ReadFile("../linters/uir/testdata/uir.sh")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(root, "bin", "uir"), script, 0o700)).To(Succeed())
		GinkgoT().Setenv("PATH", filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
		for file, content := range map[string]string{".uir.yaml": "version: 1\n", "selected.go": "package sample\n", "other.go": "package sample\n"} {
			Expect(os.WriteFile(filepath.Join(root, file), []byte(content), 0o600)).To(Succeed())
		}
		for _, args := range [][]string{{"init", "-q"}, {"config", "user.name", "Test"}, {"config", "user.email", "test@example.com"}, {"config", "commit.gpgsign", "false"}, {"add", "--", "."}, {"commit", "-qm", "initial"}} {
			cmd := exec.Command("git", args...)
			cmd.Dir = root
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(out))
		}
		GinkgoT().Setenv("UIR_TEST_REPORT", `[{"linter":"uir","success":true,"violations":[{"file":"selected.go","line":42,"severity":"warning"},{"file":"other.go","severity":"error"}]}]`)
		GinkgoT().Setenv("UIR_TEST_EXIT", "1")
	})
	It("retains all findings in a changed file and removes errors in unchanged files", func() {
		Expect(os.WriteFile(filepath.Join(root, "selected.go"), []byte("// changed comment\npackage sample\n"), 0o600)).To(Succeed())
		results, err := Execute(Options{WorkDir: root, Linters: []string{"uir"}, Since: "HEAD"})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Violations).To(HaveExactElements(SatisfyAll(HaveField("File", "selected.go"), HaveField("Line", 42))))
		Expect(linters.HasFailures(results)).To(BeFalse())
	})
	It("analyses the whole project for a changed native configuration but reports only changed files", func() {
		Expect(os.WriteFile(filepath.Join(root, ".uir.yaml"), []byte("version: 1\n# changed rules\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "selected.go"), []byte("// changed comment\npackage sample\n"), 0o600)).To(Succeed())
		results, err := Execute(Options{WorkDir: root, Linters: []string{"uir"}, Since: "HEAD"})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Violations).To(HaveExactElements(HaveField("File", "selected.go")))
		Expect(linters.HasFailures(results)).To(BeFalse())
	})
	It("reports no violations when only the native configuration changed", func() {
		Expect(os.WriteFile(filepath.Join(root, ".uir.yaml"), []byte("version: 1\n# changed rules\n"), 0o600)).To(Succeed())
		results, err := Execute(Options{WorkDir: root, Linters: []string{"uir"}, Since: "HEAD"})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Success).To(BeTrue())
		Expect(results[0].Violations).To(BeEmpty())
	})
	It("widens Go analysis for module changes without widening the report scope", func() {
		opts := Options{WorkDir: root, Files: []string{"go.work.sum", "selected.go"}, Selection: changeset.NewSelection(changeset.Options{Changed: true})}
		for _, linter := range []linters.Linter{golangci.NewGolangciLint(root), uir.New(root)} {
			inv, scopes, err := scopeInvocation(linterInvocation{linter: linter, projectRoot: root, files: opts.Files}, opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(inv.skip).To(BeFalse(), linter.Name())
			Expect(inv.files).To(BeEmpty(), linter.Name())
			Expect(scopes).To(Equal(opts.Files), linter.Name())
		}
	})
	It("skips an empty selection before resolving linter executables", func() {
		results, err := Execute(Options{WorkDir: root, Linters: []string{"uir"}, Since: "HEAD"})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(BeEmpty())
	})
	It("fails an invalid explicit ref instead of running unrestricted", func() {
		_, err := Execute(Options{WorkDir: root, Linters: []string{"uir"}, Since: "missing-reference"})
		Expect(err).To(HaveOccurred())
	})
	It("routes mixed files to their language and keeps Go package context", func() {
		Expect(os.WriteFile(filepath.Join(root, "selected.py"), []byte("value = 1\n"), 0o600)).To(Succeed())
		opts := Options{WorkDir: root, Files: []string{"selected.go", "selected.py"}, Selection: changeset.NewSelection(changeset.Options{Changed: true})}
		python, _, err := scopeInvocation(linterInvocation{linter: ruff.NewRuff(root), projectRoot: root, files: opts.Files}, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(python.files).To(Equal([]string{"selected.py"}))
		goInvocation, _, err := scopeInvocation(linterInvocation{linter: golangci.NewGolangciLint(root), projectRoot: root, files: opts.Files}, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(goInvocation.files).To(BeEmpty())
	})
})

var _ = Describe("Changed project manifest lint execution", func() {
	It("reruns the project for a changed package.json and .gavel.yaml but reports only changed files", func() {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("GAVEL_CHANGED_BASE", "")
		GinkgoT().Setenv("CI", "false")
		Expect(os.Mkdir(filepath.Join(root, "bin"), 0o755)).To(Succeed())
		script, err := os.ReadFile("../linters/reactdoctor/testdata/react-doctor.sh")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(root, "bin", "react-doctor"), script, 0o700)).To(Succeed())
		GinkgoT().Setenv("PATH", filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
		GinkgoT().Setenv("REACT_DOCTOR_TEST_REPORT", `{"diagnostics":[{"filePath":"selected.tsx","rule":"no-danger","severity":"warning","line":42},{"filePath":"other.tsx","rule":"no-danger","severity":"error","line":7}]}`)
		GinkgoT().Setenv("REACT_DOCTOR_TEST_EXIT", "0")
		files := map[string]string{
			"package.json": `{"name":"sample","dependencies":{"react":"18.0.0"}}` + "\n",
			".gavel.yaml":  "lint: {}\n",
			"selected.tsx": "export const selected = 1\n",
			"other.tsx":    "export const other = 1\n",
		}
		for file, content := range files {
			Expect(os.WriteFile(filepath.Join(root, file), []byte(content), 0o600)).To(Succeed())
		}
		for _, args := range [][]string{{"init", "-q"}, {"config", "user.name", "Test"}, {"config", "user.email", "test@example.com"}, {"config", "commit.gpgsign", "false"}, {"add", "--", "package.json", ".gavel.yaml", "selected.tsx", "other.tsx"}, {"commit", "-qm", "initial"}} {
			cmd := exec.Command("git", args...)
			cmd.Dir = root
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(out))
		}
		Expect(os.WriteFile(filepath.Join(root, ".git", "info", "exclude"), []byte("bin/\n"), 0o600)).To(Succeed())
		for _, file := range []string{"package.json", ".gavel.yaml"} {
			Expect(os.WriteFile(filepath.Join(root, file), []byte(files[file]+"\n"), 0o600)).To(Succeed())
		}
		results, err := Execute(Options{WorkDir: root, Linters: []string{"react-doctor"}, Since: "HEAD"})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveExactElements(SatisfyAll(HaveField("Linter", "react-doctor"), HaveField("Skipped", false), HaveField("Violations", BeEmpty()))))
		Expect(linters.HasFailures(results)).To(BeFalse())
	})
})
