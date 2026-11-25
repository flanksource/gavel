package testrunner

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const overlayFixture = "./testrunner/runners/testdata/overlayfixture"

var _ = Describe("unified Go runner", func() {
	It("reports missing metadata as a package failure while keeping other packages", func() {
		workDir, err := filepath.Abs("..")
		Expect(err).NotTo(HaveOccurred())
		const missing = "./testrunner/runners/testdata/overlaymissing"
		o := &TestOrchestrator{
			RunOptions: RunOptions{WorkDir: workDir, Recursive: false, Frameworks: []string{"go", "ginkgo"}, StartingPaths: []string{
				missing,
				"./testrunner/runners/testdata/overlayplain",
				"./testrunner/runners/testdata/overlayfixture",
			}},
			registry: DefaultRegistry(workDir),
		}
		results, err := o.Run()
		Expect(err).NotTo(HaveOccurred())
		Expect(results.All()).To(ContainElement(And(HaveField("Name", "TestPlain"), HaveField("Passed", true))))
		Expect(results.All()).To(ContainElement(And(HaveField("Name", "reports metadata"), HaveField("Passed", true))))
		failures := results.Failed()
		Expect(failures).To(HaveLen(1))
		Expect(failures[0].PackagePath).To(Equal(missing))
		Expect(failures[0].Framework).To(Equal(parsers.Ginkgo))
		Expect(failures[0].Message).To(ContainSubstring("without Ginkgo metadata"))
		Expect(failures[0].Attempts).To(HaveLen(1))
		Expect(failures[0].Attempts[0].ExitCode).To(HaveValue(Equal(1)))
	})

	It("uses a recursive selector for an unfocused mixed run", func() {
		root := GinkgoT().TempDir()
		Expect(os.Mkdir(filepath.Join(root, "plain"), 0o755)).To(Succeed())
		Expect(os.Mkdir(filepath.Join(root, "specs"), 0o755)).To(Succeed())
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: root, Recursive: true}}
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
		root := GinkgoT().TempDir()
		Expect(os.Mkdir(filepath.Join(root, "plain"), 0o755)).To(Succeed())
		Expect(os.Mkdir(filepath.Join(root, "specs"), 0o755)).To(Succeed())
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: root, Recursive: true}}
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
		Expect(plans[1].opts.Bench).To(Equal("."))
		Expect(plans[1].opts.Focus).To(Equal(runners.FileSelection{Regex: []string{"^$"}}))
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

	It("runs a mixed Go and Ginkgo package once and replays both results from cache", func() {
		workDir, err := filepath.Abs("..")
		Expect(err).NotTo(HaveOccurred())
		pkg := "./testrunner/runners/testdata/overlayfixture"
		plainPkg := "./testrunner/runners/testdata/overlayplain"
		for run := 0; run < 2; run++ {
			o := &TestOrchestrator{
				RunOptions: RunOptions{WorkDir: workDir, Recursive: false, Frameworks: []string{"go", "ginkgo"}, StartingPaths: []string{pkg, plainPkg}},
				registry:   DefaultRegistry(workDir),
			}
			results, err := o.Run()
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

	DescribeTable("accepts Ginkgo focus and label filters through go test",
		func(passThrough []string) {
			workDir, err := filepath.Abs("..")
			Expect(err).NotTo(HaveOccurred())
			o := &TestOrchestrator{
				RunOptions: RunOptions{WorkDir: workDir, Frameworks: []string{"ginkgo"}, StartingPaths: []string{overlayFixture}, PassThroughArgs: passThrough},
				registry:   DefaultRegistry(workDir),
			}
			results, err := o.Run()
			Expect(err).NotTo(HaveOccurred())
			var specs []parsers.Test
			for _, suite := range results {
				if suite.Framework == parsers.Ginkgo {
					specs = append(specs, suite.Tests...)
				}
			}
			// Ginkgo reports the specs its filters leave out as skipped, as its own JSON report does.
			Expect(specs).To(ConsistOf(
				And(HaveField("Name", "reports metadata"), HaveField("Passed", true)),
				And(HaveField("Name", "passes without report entries"), HaveField("Skipped", true)),
			), "results: %+v", results)
			// A flag before the packages ends go test's package list, so it tests the module root instead.
			Expect(results.All()).NotTo(ContainElement(HaveField("Package", "github.com/flanksource/gavel")), "results: %+v", results)
		},
		Entry("--focus with a label filter", []string{"--focus", "reports metadata", "--label-filter=inherited && leaf"}),
		Entry("native --ginkgo.focus=", []string{"--ginkgo.focus=reports metadata"}),
		Entry("space-separated -ginkgo.focus", []string{"-ginkgo.focus", "reports metadata"}),
		Entry("-race with a Ginkgo focus", []string{"-race", "--ginkgo.focus=reports metadata"}),
		Entry("vitest-style -t", []string{"-t", "reports metadata"}),
	)

	It("puts the packages before the user's flags in the go test argv", func() {
		workDir, err := filepath.Abs("..")
		Expect(err).NotTo(HaveOccurred())
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: workDir, StartingPaths: []string{overlayFixture},
			PassThroughArgs: []string{"-race", "-ginkgo.focus", "reports metadata"}}}
		packages := map[Framework][]string{parsers.Ginkgo: {overlayFixture}}
		detected, err := parsedTestOptions(o, parsers.Ginkgo)
		Expect(err).NotTo(HaveOccurred())
		resolved, err := o.resolveTestOptions(detected, packages)
		Expect(err).NotTo(HaveOccurred())
		plans, err := o.planGoBatches(packages, resolved, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(plans).To(HaveLen(1))
		process, err := o.buildGoBatchCommand(plans[0], 0)
		Expect(err).NotTo(HaveOccurred())
		args := process.Args
		Expect(args).NotTo(ContainElement("-args"))
		pkg := slices.Index(args, overlayFixture)
		Expect(pkg).To(BeNumerically(">", 0), "argv: %v", args)
		for _, flag := range []string{"-race", "-ginkgo.focus=reports metadata"} {
			Expect(slices.Index(args, flag)).To(BeNumerically(">", pkg), "%s follows the packages in %v", flag, args)
		}
	})

	It("rejects parallel Ginkgo nodes before execution", func() {
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: ".", Nodes: 2}, registry: DefaultRegistry(".")}
		_, err := o.detectAndRun(map[Framework]*runners.TestOptions{parsers.Ginkgo: {}})
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
		plain := byPackage["./testrunner/runners/testdata/overlayplain"].opts
		Expect(plain.Bench).To(BeEmpty())
		Expect(plain.Focus.Empty()).To(BeTrue())
		bench := byPackage["./testrunner/runners/testdata/overlaybench"].opts
		Expect(bench.Bench).To(Equal("."))
		Expect(bench.Focus).To(Equal(runners.FileSelection{Regex: []string{"^$"}}))
		o.registry = DefaultRegistry(workDir)
		o.Frameworks = []string{"go"}
		o.StartingPaths = []string{"./testrunner/runners/testdata/overlaybench"}
		results, err := o.Run()
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
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		Expect(os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte("test: {frameworks: {ginkgo: {args: [--focus=selected]}}}\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "sample_test.go"), []byte("package sample\nimport \"github.com/onsi/ginkgo/v2\"\n"), 0o600)).To(Succeed())
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
		process, err := o.buildGoBatchCommand(ginkgoPlan, 0)
		Expect(err).NotTo(HaveOccurred())
		Expect(process.Args).To(ContainElement("-ginkgo.focus=selected"))
		Expect(process.Args).NotTo(ContainElement("-args"))
	})
})
