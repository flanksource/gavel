package testrunner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("unified Go runner", func() {
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	})

	It("uses a recursive selector for an unfocused mixed run", func() {
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: ".", Recursive: true}}
		plans, err := o.planGoBatches(map[Framework][]string{
			parsers.GoTest: {"./plain"},
			parsers.Ginkgo: {"./specs"},
		}, nil, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(plans).To(HaveLen(1))
		Expect(plans[0].targets).To(Equal([]string{"./..."}))
		Expect(plans[0].packages).To(ConsistOf("./plain", "./specs"))
	})

	It("uses explicit packages for a focused run", func() {
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: ".", Recursive: true}}
		plans, err := o.planGoBatches(map[Framework][]string{
			parsers.GoTest: {"./plain"},
			parsers.Ginkgo: {"./specs"},
		}, nil, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(plans).To(HaveLen(1))
		Expect(plans[0].targets).To(ConsistOf("./plain", "./specs"))
	})

	It("keeps automatic benchmark-only packages in a focused follow-up", func() {
		workDir, err := filepath.Abs("..")
		Expect(err).NotTo(HaveOccurred())
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: workDir, Recursive: true}}
		plans, err := o.planGoBatches(map[Framework][]string{
			parsers.GoTest: {"./testrunner/runners/testdata/overlayplain", "./testrunner/runners/testdata/overlaybench"},
		}, nil, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(plans).To(HaveLen(2))
		Expect(plans[0].targets).To(Equal([]string{"./..."}))
		Expect(plans[1].targets).To(Equal([]string{"./testrunner/runners/testdata/overlaybench"}))
		Expect(plans[1].args).To(ContainElements("-bench=.", "-run=^$"))
	})

	It("maps packages without discovered tests in a recursive run", func() {
		workDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module example.com/service\n\ngo 1.26.1\n"), 0o600)).To(Succeed())
		paths, err := resolveGoBatchPaths(workDir, []string{"."}, []string{"example.com/service", "example.com/service/empty"}, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(Equal(map[string]string{
			"example.com/service":       ".",
			"example.com/service/empty": "./empty",
		}))
	})

	It("does not map a nested package to a selected package that only shares its suffix", func() {
		workDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module example.com/service\n\ngo 1.26.1\n"), 0o600)).To(Succeed())
		paths, err := resolveGoBatchPaths(workDir, []string{"./har"}, []string{"example.com/service/examples/har", "example.com/service/har"}, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(Equal(map[string]string{
			"example.com/service/har":          "./har",
			"example.com/service/examples/har": "./examples/har",
		}))
	})

	It("maps packages exactly when the workdir is a subdirectory of a module", func() {
		module := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(module, "go.mod"), []byte("module example.com/service\n\ngo 1.26.1\n"), 0o600)).To(Succeed())
		workDir := filepath.Join(module, "sub")
		Expect(os.MkdirAll(workDir, 0o755)).To(Succeed())
		paths, err := resolveGoBatchPaths(workDir, []string{"./har"}, []string{"example.com/service/sub/examples/har", "example.com/service/sub/har"}, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(Equal(map[string]string{
			"example.com/service/sub/har":          "./har",
			"example.com/service/sub/examples/har": "./examples/har",
		}))
	})

	It("limits a Ginkgo-only run of a mixed package to its RunSpecs bootstrap", func() {
		workDir, err := filepath.Abs("..")
		Expect(err).NotTo(HaveOccurred())
		pkg := "./testrunner/runners/testdata/overlayfixture"
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: workDir}}
		plans, err := o.planGoBatches(map[Framework][]string{parsers.Ginkgo: {pkg}}, map[Framework][]string{parsers.Ginkgo: {"--focus=reports metadata"}}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(plans).To(HaveLen(1))
		Expect(plans[0].args).To(Equal([]string{"-ginkgo.focus=reports metadata", "-run=^(?:TestSuite)$"}))
	})

	It("runs a mixed Go and Ginkgo package once and replays both results from cache", func() {
		workDir, err := filepath.Abs("..")
		Expect(err).NotTo(HaveOccurred())
		pkg := "./testrunner/runners/testdata/overlayfixture"
		plainPkg := "./testrunner/runners/testdata/overlayplain"
		for run := 0; run < 2; run++ {
			o := &TestOrchestrator{
				RunOptions: RunOptions{WorkDir: workDir, Recursive: false},
				registry:   DefaultRegistry(workDir),
			}
			results, err := o.detectAndRun([]Framework{parsers.GoTest, parsers.Ginkgo}, []string{pkg, plainPkg}, nil)
			Expect(err).NotTo(HaveOccurred())
			var ordinary, plain, spec *parsers.Test
			commands := map[string]bool{}
			for _, suite := range results {
				commands[suite.Command] = true
				for i := range suite.Tests {
					test := &suite.Tests[i]
					if test.Name == "TestOrdinary" {
						ordinary = test
					}
					if test.Name == "TestPlain" {
						plain = test
					}
					if test.Name == "reports metadata" {
						spec = test
					}
				}
			}
			Expect(ordinary).NotTo(BeNil())
			Expect(plain).NotTo(BeNil())
			Expect(spec).NotTo(BeNil())
			Expect(commands).To(HaveLen(1))
			Expect(spec.PackagePath).To(Equal(pkg))
			if run == 1 {
				Expect(spec.Cached).To(BeTrue())
			}
		}
	})

	It("accepts Ginkgo focus and label filters through go test", func() {
		workDir, err := filepath.Abs("..")
		Expect(err).NotTo(HaveOccurred())
		o := &TestOrchestrator{
			RunOptions: RunOptions{WorkDir: workDir},
			registry:   DefaultRegistry(workDir),
		}
		results, err := o.detectAndRun([]Framework{parsers.Ginkgo}, []string{"./testrunner/runners/testdata/overlayfixture"}, []string{"--focus", "reports metadata", "--label-filter=inherited && leaf"})
		Expect(err).NotTo(HaveOccurred())
		var specs []parsers.Test
		for _, suite := range results {
			if suite.Framework == parsers.Ginkgo {
				specs = append(specs, suite.Tests...)
			}
		}
		// Ginkgo reports the specs its filters leave out as skipped, as its own JSON report does.
		ran := map[string]bool{}
		for _, spec := range specs {
			ran[spec.Name] = !spec.Skipped
			if spec.Name == "reports metadata" {
				Expect(spec.Passed).To(BeTrue())
			}
		}
		Expect(ran).To(Equal(map[string]bool{"reports metadata": true, "passes without report entries": false}), "results: %+v", results)
	})

	It("rejects parallel Ginkgo nodes before execution", func() {
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: ".", Nodes: 2}, registry: DefaultRegistry(".")}
		_, err := o.detectAndRun([]Framework{parsers.Ginkgo}, nil, nil)
		Expect(err).To(MatchError(ContainSubstring("--nodes=2 is unsupported")))
	})

	It("keeps benchmark-only packages in a batch with benchmark flags", func() {
		workDir, err := filepath.Abs("..")
		Expect(err).NotTo(HaveOccurred())
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: workDir}}
		plans, err := o.planGoBatches(map[Framework][]string{
			parsers.GoTest: {"./testrunner/runners/testdata/overlayplain", "./testrunner/runners/testdata/overlaybench"},
		}, nil, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(plans).To(HaveLen(2))
		byPackage := map[string]goBatchPlan{}
		for _, plan := range plans {
			Expect(plan.packages).To(HaveLen(1))
			byPackage[plan.packages[0]] = plan
		}
		Expect(byPackage["./testrunner/runners/testdata/overlayplain"].args).To(BeEmpty())
		Expect(byPackage["./testrunner/runners/testdata/overlaybench"].args).To(ContainElements("-bench=.", "-run=^$"))
		o.registry = DefaultRegistry(workDir)
		results, err := o.detectAndRun([]Framework{parsers.GoTest}, []string{"./testrunner/runners/testdata/overlaybench"}, nil)
		Expect(err).NotTo(HaveOccurred())
		var benchmarks []parsers.Test
		for _, suite := range results {
			for _, test := range suite.Tests {
				if test.Benchmark != nil {
					benchmarks = append(benchmarks, test)
				}
			}
		}
		Expect(benchmarks).To(HaveLen(1), "results: %+v", results)
		Expect(benchmarks[0].Name).To(Equal("BenchmarkOnly"))
	})

	It("maps JSON package names without requiring a go.mod", func() {
		workDir := GinkgoT().TempDir()
		paths, err := resolveGoBatchPaths(workDir, []string{".", "./pkg"}, []string{"_/tmp/example/pkg", "_/tmp/example"}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(Equal(map[string]string{
			"_/tmp/example/pkg": "./pkg",
			"_/tmp/example":     ".",
		}))
	})

	It("maps overlapping package suffixes to their longest path", func() {
		paths, err := resolveGoBatchPaths(GinkgoT().TempDir(), []string{"./pkg", "./nested/pkg"}, []string{"_/tmp/example/nested/pkg", "_/tmp/example/pkg"}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(Equal(map[string]string{
			"_/tmp/example/nested/pkg": "./nested/pkg",
			"_/tmp/example/pkg":        "./pkg",
		}))
	})

	It("keeps distinct Ginkgo runtime arguments on the test binary", func() {
		workDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte("test: {frameworks: {ginkgo: {args: [--focus=selected]}}}\n"), 0o600)).To(Succeed())
		code := `package sample
import (
	"testing"
	. "github.com/onsi/ginkgo/v2"
)
func TestSample(t *testing.T) { RunSpecs(t, "sample") }
`
		Expect(os.WriteFile(filepath.Join(workDir, "sample_test.go"), []byte(code), 0o600)).To(Succeed())
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: workDir}}
		plans, err := o.planGoBatches(map[Framework][]string{parsers.GoTest: {"."}, parsers.Ginkgo: {"."}}, nil, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(plans).To(HaveLen(2))
		var ginkgoPlan goBatchPlan
		for _, plan := range plans {
			if plan.framework == parsers.Ginkgo {
				ginkgoPlan = plan
			}
		}
		process, err := o.buildGoBatchCommand(ginkgoPlan)
		Expect(err).NotTo(HaveOccurred())
		Expect(process.Args).To(ContainElements("-run=^(?:TestSample)$", "-args", "-ginkgo.focus=selected"))
	})

	DescribeTable("preserves explicit Go selectors for Ginkgo batches", func(configArgs string, passThroughArgs, wantGoArgs []string) {
		workDir := GinkgoT().TempDir()
		config := fmt.Sprintf("test: {frameworks: {ginkgo: {args: [%s]}}}\n", configArgs)
		Expect(os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte(config), 0o600)).To(Succeed())
		code, err := os.ReadFile("runners/testdata/overlayfixture/fixture_test.go")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(workDir, "sample_test.go"), code, 0o600)).To(Succeed())
		for _, mixed := range []bool{false, true} {
			packages := map[Framework][]string{parsers.Ginkgo: {"."}}
			if mixed {
				packages[parsers.GoTest] = []string{"."}
			}
			o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: workDir}}
			plans, err := o.planGoBatches(packages, map[Framework][]string{parsers.Ginkgo: passThroughArgs}, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(plans).To(HaveLen(len(packages)))
			plan := plans[len(plans)-1]
			Expect(plan.framework).To(Equal(parsers.Ginkgo))
			Expect(plan.args).To(Equal(passThroughArgs))
			process, err := o.buildGoBatchCommand(plan)
			Expect(err).NotTo(HaveOccurred())
			wantTail := append(append([]string(nil), wantGoArgs...), ".", "-args", "-ginkgo.focus=selected")
			Expect(process.Args[len(process.Args)-len(wantTail):]).To(Equal(wantTail))
		}
	},
		Entry("configured equals form", "--focus=selected, -run=^$", nil, []string{"-run=^$"}),
		Entry("configured separate value", "--focus=selected, -run, ^$", nil, []string{"-run", "^$"}),
		Entry("pass-through equals form", "--focus=selected", []string{"-run=^$"}, []string{"-run=^$"}),
		Entry("pass-through separate value", "--focus=selected", []string{"-run", "^$"}, []string{"-run", "^$"}),
		Entry("pass-through overrides configuration", "--focus=selected, -run=^$", []string{"-run=^TestOrdinary$"}, []string{"-run=^$", "-run=^TestOrdinary$"}),
	)
})
