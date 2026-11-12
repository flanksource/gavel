package lint

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/git/changeset"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/golangci"
	"github.com/flanksource/gavel/linters/ruff"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	reactPackageJSON = `{"name":"sample","dependencies":{"react":"18.0.0"}}` + "\n"
	reactDoctorFinds = `{"diagnostics":[{"filePath":"selected.tsx","rule":"no-danger","severity":"warning","line":42},{"filePath":"other.tsx","rule":"no-danger","severity":"error","line":7}]}`
)

var _ = Describe("Changed lint execution", func() {
	var root string
	modify := func(files ...string) {
		for _, file := range files {
			content, err := os.ReadFile(filepath.Join(root, file))
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(filepath.Join(root, file), append(content, '\n'), 0o600)).To(Succeed())
		}
	}
	reactDoctor := func(since string) []*linters.LinterResult {
		results, err := Execute(Options{WorkDir: root, Linters: []string{"react-doctor"}, Since: since})
		Expect(err).NotTo(HaveOccurred())
		return results
	}
	BeforeEach(func() {
		root = GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("GAVEL_CHANGED_BASE", "")
		GinkgoT().Setenv("CI", "false")
		Expect(os.Mkdir(filepath.Join(root, "bin"), 0o755)).To(Succeed())
		script, err := os.ReadFile("../linters/reactdoctor/testdata/react-doctor.sh")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(root, "bin", "react-doctor"), script, 0o700)).To(Succeed())
		GinkgoT().Setenv("PATH", filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
		GinkgoT().Setenv("REACT_DOCTOR_TEST_REPORT", reactDoctorFinds)
		GinkgoT().Setenv("REACT_DOCTOR_TEST_EXIT", "0")
		for file, content := range map[string]string{
			"package.json": reactPackageJSON,
			".gavel.yaml":  "lint: {}\n",
			"selected.tsx": "export const selected = 1\n",
			"other.tsx":    "export const other = 1\n",
		} {
			Expect(os.WriteFile(filepath.Join(root, file), []byte(content), 0o600)).To(Succeed())
		}
		for _, args := range [][]string{{"init", "-q"}, {"config", "user.name", "Test"}, {"config", "user.email", "test@example.com"}, {"config", "commit.gpgsign", "false"}, {"add", "--", "package.json", ".gavel.yaml", "selected.tsx", "other.tsx"}, {"commit", "-qm", "initial"}} {
			cmd := exec.Command("git", args...)
			cmd.Dir = root
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(out))
		}
		Expect(os.WriteFile(filepath.Join(root, ".git", "info", "exclude"), []byte("bin/\n"), 0o600)).To(Succeed())
	})
	It("reports a project-wide scanner's findings only in the changed file", func() {
		modify("selected.tsx")
		results := reactDoctor("HEAD")
		Expect(results).To(HaveLen(1))
		Expect(results[0].Violations).To(HaveExactElements(SatisfyAll(HaveField("File", HaveSuffix("selected.tsx")), HaveField("Line", 42))))
		Expect(linters.HasFailures(results)).To(BeFalse())
	})
	It("reruns the project for a changed package.json and .gavel.yaml but reports only changed files", func() {
		modify("package.json", ".gavel.yaml")
		results := reactDoctor("HEAD")
		Expect(results).To(HaveExactElements(SatisfyAll(HaveField("Linter", "react-doctor"), HaveField("Skipped", false), HaveField("Violations", BeEmpty()))))
		Expect(linters.HasFailures(results)).To(BeFalse())
	})
	It("skips an empty selection before resolving linter executables", func() {
		Expect(reactDoctor("HEAD")).To(BeEmpty())
	})
	It("fails an invalid explicit ref instead of running unrestricted", func() {
		_, err := Execute(Options{WorkDir: root, Linters: []string{"react-doctor"}, Since: "missing-reference"})
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
