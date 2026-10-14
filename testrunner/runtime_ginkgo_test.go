package testrunner

import (
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
	"time"
)

var _ = Describe("Shared test runtime", func() {
	It("retains invocation configuration after a positional path rebases execution to the repository root", func() {
		root := GinkgoT().TempDir()
		pkg := filepath.Join(root, "pkg")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("GOWORK", "off")
		Expect(os.MkdirAll(filepath.Join(root, ".git"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(pkg, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/runtime\n\ngo 1.22\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(pkg, ".gavel.yaml"), []byte("runtime: {env: {GAVEL_INVOCATION_TEST: child}}\n"), 0o600)).To(Succeed())
		code := "package sample\nimport (\"os\"; \"testing\")\nfunc TestEnv(t *testing.T) { if os.Getenv(\"GAVEL_INVOCATION_TEST\") != \"child\" { t.Fatal(\"missing invocation environment\") } }\n"
		Expect(os.WriteFile(filepath.Join(pkg, "sample_test.go"), []byte(code), 0o600)).To(Succeed())
		result, err := Run(RunOptions{WorkDir: pkg, StartingPaths: []string{"."}, Frameworks: []string{"go"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(parsers.Tests(result.([]parsers.Test)).Sum().Failed).To(BeZero())
	})
	DescribeTable("previews fixtures without executing build or commands", func(direct bool) {
		dir := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		path := filepath.Join(dir, "preview.fixture.md")
		Expect(os.WriteFile(path, []byte("---\nbuild: touch build.marker\n---\n\n| Name | Command |\n| --- | --- |\n| marker | touch exec.marker |\n"), 0o600)).To(Succeed())
		opts := RunOptions{WorkDir: dir, DryRun: true, FixturesOnly: true, Fixtures: true, FixtureFiles: []string{path}}
		if direct {
			opts.FixtureRunner = &fixtures.RunnerOptions{Paths: []string{path}, WorkDir: dir}
		}
		_, err := Run(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(filepath.Join(dir, "build.marker")).NotTo(BeAnExistingFile())
		Expect(filepath.Join(dir, "exec.marker")).NotTo(BeAnExistingFile())
	}, Entry("direct", true), Entry("discovered", false))
	It("fails dry-run inspection when caller runtime is invalid", func() {
		dir := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		invalidCPU := -1.0
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: dir, Runtime: processruntime.Options{CPU: &invalidCPU}}, registry: DefaultRegistry(dir)}
		_, err := o.displayDryRun(map[Framework][]string{parsers.GoTest: {"."}}, nil)
		Expect(err).To(MatchError(ContainSubstring("runtime cpu")))
	})
	DescribeTable("applies runtime after every adapter builds its command", func(framework parsers.Framework) {
		workDir := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		Expect(os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte("runtime: {env: {RUNTIME_KEY: inherited}, nice: 10, timeout: 7s}\n"), 0o600)).To(Succeed())
		registry := DefaultRegistry(workDir)
		runner, ok := registry.Get(framework)
		Expect(ok).To(BeTrue())
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: workDir}}
		run, err := o.buildCommand(runner, ".", nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(run.Process.Env["RUNTIME_KEY"]).To(Equal("inherited"))
		Expect(run.Process.Timeout).To(Equal(7 * time.Second))
	},
		Entry("Go", parsers.GoTest), Entry("Ginkgo", parsers.Ginkgo), Entry("Jest", parsers.Jest), Entry("Vitest", parsers.Vitest), Entry("Playwright", parsers.Playwright),
	)
})
