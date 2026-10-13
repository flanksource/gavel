package fixtures_test

import (
	"context"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/processruntime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == fixtures.WorkerCommand {
		output := os.NewFile(3, "fixture-worker-results")
		if err := fixtures.RunWorker(os.Stdin, output); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

var _ = Describe("Fixture runtime isolation", func() {
	It("refuses capped node execution without a worker executable", func() {
		memory := "1GiB"
		_, _, err := fixtures.RunNodes(context.Background(), nil, fixtures.RunOptions{ProcessRuntime: processruntime.Options{Mem: &memory}})
		Expect(err).To(MatchError(ContainSubstring("worker-capable executable")))
	})
	It("keeps node execution in process after explicit cap resets", func() {
		memory, timeout, cpu := "", "", 0.0
		_, snapshot, err := fixtures.RunNodes(context.Background(), nil, fixtures.RunOptions{ProcessRuntime: processruntime.Options{Mem: &memory, Timeout: &timeout, CPU: &cpu}})
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot).NotTo(BeNil())
	})
	It("preserves isolated node results, progress, and child environment", func() {
		dir := GinkgoT().TempDir()
		tree, err := fixtures.ParseMarkdownDocument("runtime", "| Name | Command | CEL |\n| --- | --- | --- |\n| environment | printenv GAVEL_RUNTIME_TEST | stdout == 'child\\n' |\n", dir)
		Expect(err).NotTo(HaveOccurred())
		executable, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		timeout := "20s"
		var progress []fixtures.ExecutionSnapshot
		ctx := fixtures.WithProgressSink(context.Background(), func(_ context.Context, snapshot fixtures.ExecutionSnapshot) error {
			progress = append(progress, snapshot)
			return nil
		})
		results, snapshot, err := fixtures.RunNodes(ctx, tree.Children, fixtures.RunOptions{WorkDir: dir, ExecutablePath: executable, ProcessRuntime: processruntime.Options{Timeout: &timeout, Env: map[string]string{"GAVEL_RUNTIME_TEST": "child"}}})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].IsOK()).To(BeTrue(), results[0].Error)
		Expect(results[0].Test.Name).To(Equal("environment"))
		Expect(snapshot).NotTo(BeNil())
		Expect(progress).NotTo(BeEmpty())
		Expect(os.Getenv("GAVEL_RUNTIME_TEST")).To(BeEmpty())
	})
	It("terminates capped whole-file fixtures and returns a failure", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "runtime.md")
		Expect(os.WriteFile(path, []byte("| Name | Command |\n| --- | --- |\n| delayed | sleep 30 |\n"), 0o600)).To(Succeed())
		executable, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		timeout := "300ms"
		runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: dir, ExecutablePath: executable, ProcessRuntime: processruntime.Options{Timeout: &timeout}})
		Expect(err).NotTo(HaveOccurred())
		_, err = runner.Run()
		Expect(err).To(MatchError(ContainSubstring("timed out")))
	})
	It("preserves setup, build, and saved benchmark reports in a worker", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "benchmark.md")
		document := "---\nsetup:\n  envVars:\n    - name: GAVEL_SETUP_TEST\n      value: setup\nbuild: printenv GAVEL_RUNTIME_TEST > build.env\n---\n\n| Name | Command | CEL |\n| --- | --- | --- |\n| build | cat build.env | stdout == 'child\\n' |\n| setup | printenv GAVEL_SETUP_TEST | stdout == 'setup\\n' |\n"
		Expect(os.WriteFile(path, []byte(document), 0o600)).To(Succeed())
		executable, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		timeout := "20s"
		runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: dir, ExecutablePath: executable, Benchmark: true, ProcessRuntime: processruntime.Options{Timeout: &timeout, Env: map[string]string{"GAVEL_RUNTIME_TEST": "child"}}})
		Expect(err).NotTo(HaveOccurred())
		tree, err := runner.Run()
		Expect(err).NotTo(HaveOccurred())
		Expect(tree.Stats.Passed).To(Equal(2))
		report := runner.BenchmarkReport()
		Expect(report).NotTo(BeNil())
		Expect(report.Path).To(BeAnExistingFile())
		Expect(report.Fixtures).To(HaveLen(2))
		Expect(report.Phases).To(ContainElement(HaveField("Kind", "cleanup")))
	})
})
