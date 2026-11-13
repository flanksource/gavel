package testrunner

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/testrunner/runners"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var benchPackages = map[string]map[string]string{
	"benchonly": {
		"main_test.go":  "package benchonly\n\nimport \"testing\"\n\nfunc TestMain(m *testing.M) { m.Run() }\n",
		"bench_test.go": "package benchonly\n\nimport \"testing\"\n\nfunc BenchmarkFoo(b *testing.B) {}\n",
	},
	"mixed": {
		"x_test.go": "package mixed\n\nimport \"testing\"\n\nfunc TestX(t *testing.T) {}\nfunc BenchmarkY(b *testing.B) {}\n",
	},
	"empty": {
		"x_test.go": "package empty\n\nimport \"testing\"\n\nfunc TestX(t *testing.T) {}\n",
	},
	// A Ginkgo suite beside a plain-testing benchmark: the suite's bootstrap is
	// the package's runnable test, so benchmarks are opt-in and never displace it.
	"ginkgosuite": {
		"suite_test.go": "package ginkgosuite_test\n\nimport (\n\t\"testing\"\n\n\t. \"github.com/onsi/ginkgo/v2\"\n)\n\nfunc TestSuite(t *testing.T) { RunSpecs(t, \"suite\") }\n",
		"bench_test.go": "package ginkgosuite_test\n\nimport \"testing\"\n\nfunc BenchmarkZ(b *testing.B) {}\n",
	},
}

var _ = Describe("benchOptions", func() {
	var workDir string
	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		for pkg, files := range benchPackages {
			Expect(os.MkdirAll(filepath.Join(workDir, pkg), 0o755)).To(Succeed())
			for name, content := range files {
				Expect(os.WriteFile(filepath.Join(workDir, pkg, name), []byte(content), 0o644)).To(Succeed())
			}
		}
	})

	DescribeTable("decides the benchmark pattern and keeps the empty test pass quiet",
		func(bench, pkg string, given, want runners.TestOptions) {
			o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: workDir, Bench: bench}}
			Expect(o.benchOptions(pkg, given)).To(Equal(want))
		},
		Entry("bench-only package auto-enables bench with -run=^$", "", "./benchonly",
			runners.TestOptions{}, runners.TestOptions{Bench: ".", Focus: runners.FileSelection{Regex: []string{"^$"}}}),
		Entry("--bench=false keeps a bench-only package's benchmarks off", "false", "./benchonly",
			runners.TestOptions{}, runners.TestOptions{}),
		Entry("package with tests only is untouched when --bench is unset", "", "./empty",
			runners.TestOptions{}, runners.TestOptions{}),
		Entry("package without benchmarks is untouched under an explicit pattern", ".", "./empty",
			runners.TestOptions{}, runners.TestOptions{}),
		Entry("mixed package with --bench=true runs benchmarks beside its tests", "true", "./mixed",
			runners.TestOptions{}, runners.TestOptions{Bench: "."}),
		Entry("mixed package with a custom pattern", "BenchmarkY", "./mixed",
			runners.TestOptions{}, runners.TestOptions{Bench: "BenchmarkY"}),
		Entry("Ginkgo suite keeps its specs runnable under an explicit pattern", "^$", "./ginkgosuite",
			runners.TestOptions{}, runners.TestOptions{Bench: "^$"}),
		Entry("Ginkgo suite with benchmarks is not treated as bench-only", "", "./ginkgosuite",
			runners.TestOptions{}, runners.TestOptions{}),
		Entry("Ginkgo suite with --bench=true runs the benchmark beside its specs", "true", "./ginkgosuite",
			runners.TestOptions{}, runners.TestOptions{Bench: "."}),
		Entry("a -bench given after -- wins", "", "./benchonly",
			runners.TestOptions{Bench: "Custom"}, runners.TestOptions{Bench: "Custom"}),
		Entry("a focus given after -- prevents the -run=^$ injection", ".", "./benchonly",
			runners.TestOptions{Focus: runners.FileSelection{Regex: []string{"^TestMain$"}}},
			runners.TestOptions{Bench: ".", Focus: runners.FileSelection{Regex: []string{"^TestMain$"}}}),
	)
})
