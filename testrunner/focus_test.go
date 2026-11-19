package testrunner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

// parsedTestOptions parses o's argv with each framework's flags and merges
// gavel's own flags, as Run's detection and detectAndRun do.
func parsedTestOptions(o *TestOrchestrator, frameworks ...Framework) (map[Framework]*runners.TestOptions, error) {
	base, err := o.baseTestOptions()
	if err != nil {
		return nil, err
	}
	registry := DefaultRegistry(o.WorkDir)
	detected := map[Framework]*runners.TestOptions{}
	for _, fw := range frameworks {
		runner, ok := registry.Get(fw)
		Expect(ok).To(BeTrue(), "runner %s", fw)
		parsed, err := runners.ParseTestArgs(base.Cwd, o.testArgv(), runner.Flags()...)
		if err != nil {
			return nil, err
		}
		merged, err := base.Merge(*parsed)
		if err != nil {
			return nil, err
		}
		detected[fw] = &merged
	}
	return detected, nil
}

// resolveFor resolves passThrough for the frameworks that have a package.
func resolveFor(passThrough []string, packages map[Framework][]string, others ...Framework) (map[Framework]runners.TestOptions, error) {
	o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: GinkgoT().TempDir(), PassThroughArgs: passThrough}}
	detected, err := parsedTestOptions(o, append(lo.Keys(packages), others...)...)
	if err != nil {
		return nil, err
	}
	return o.resolveTestOptions(detected, packages)
}

var goAndGinkgo = map[Framework][]string{parsers.GoTest: {"./plain"}, parsers.Ginkgo: {"./specs"}}

var _ = Describe("runner arguments after --", func() {
	DescribeTable("read a focus in any position and spelling for go test and Ginkgo alike",
		func(passThrough []string) {
			resolved, err := resolveFor(passThrough, goAndGinkgo)
			Expect(err).NotTo(HaveOccurred())
			focus := runners.FileSelection{Regex: []string{"TestFoo"}}
			Expect(resolved).To(HaveLen(2))
			for fw, opts := range resolved {
				Expect(opts.Focus).To(Equal(focus), "%s", fw)
				Expect(opts.ExtraArgs).To(Equal([]string{"-count=1"}), "%s", fw)
			}
		},
		Entry("go spaced, before an unknown flag", []string{"-run", "TestFoo", "-count=1"}),
		Entry("go equals, after an unknown flag", []string{"-count=1", "-run=TestFoo"}),
		Entry("ginkgo spaced", []string{"--focus", "TestFoo", "-count=1"}),
		Entry("ginkgo equals", []string{"-count=1", "--focus=TestFoo"}),
		Entry("native -ginkgo.focus", []string{"-ginkgo.focus=TestFoo", "-count=1"}),
		Entry("native --ginkgo.focus spaced", []string{"-count=1", "--ginkgo.focus", "TestFoo"}),
		Entry("vitest -t", []string{"-t", "TestFoo", "-count=1"}),
		Entry("vitest --testNamePattern", []string{"-count=1", "--testNamePattern=TestFoo"}),
	)

	DescribeTable("reject a focus flag without a pattern",
		func(passThrough []string, message string) {
			_, err := resolveFor(passThrough, goAndGinkgo)
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("missing go pattern", []string{"-run"}, "-run requires a value"),
		Entry("empty ginkgo pattern", []string{"--focus="}, "requires a non-empty value"),
	)

	It("selects tests in jest and vitest from a go test focus", func() {
		resolved, err := resolveFor([]string{"-run=TestFoo"}, map[Framework][]string{parsers.Jest: {"./web"}, parsers.Vitest: {"./app"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(lo.Keys(resolved)).To(ConsistOf(parsers.Jest, parsers.Vitest))
		for fw, opts := range resolved {
			Expect(opts.Focus).To(Equal(runners.FileSelection{Regex: []string{"TestFoo"}}), "%s", fw)
			Expect(opts.ExtraArgs).To(BeEmpty(), "%s", fw)
		}
	})

	It("treats go test and Ginkgo as one family: a Ginkgo-only flag is not unknown, and narrows to Ginkgo", func() {
		resolved, err := resolveFor([]string{"--label-filter", "smoke"}, goAndGinkgo)
		Expect(err).NotTo(HaveOccurred())
		Expect(lo.Keys(resolved)).To(Equal([]Framework{parsers.Ginkgo}))
		Expect(resolved[parsers.Ginkgo].LabelFilter).To(Equal("smoke"))
		Expect(resolved[parsers.Ginkgo].ExtraArgs).To(BeEmpty())
	})

	It("passes unknown arguments to go test and Ginkgo together", func() {
		resolved, err := resolveFor([]string{"--trace"}, goAndGinkgo)
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved[parsers.GoTest].ExtraArgs).To(Equal([]string{"--trace"}))
		Expect(resolved[parsers.Ginkgo].ExtraArgs).To(Equal([]string{"--trace"}))
	})

	It("rejects unknown arguments across framework families, naming only frameworks with packages", func() {
		_, err := resolveFor([]string{"--trace"},
			map[Framework][]string{parsers.GoTest: {"./plain"}, parsers.Vitest: {"./app"}},
			parsers.Ginkgo, parsers.Jest)
		Expect(err).To(MatchError(And(
			ContainSubstring("(--trace) need one framework family"),
			ContainSubstring("got 2 (go test, vitest) for"),
		)))
	})
})

var _ = Describe("build tags", func() {
	It("carries tags to the Go frameworks without removing unsupported frameworks", func() {
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: GinkgoT().TempDir(), Tags: []string{"integration", "postgres"}}}
		packages := map[Framework][]string{parsers.GoTest: {"./plain"}, parsers.Ginkgo: {"./specs"}, parsers.Jest: {"./web"}}
		detected, err := parsedTestOptions(o, lo.Keys(packages)...)
		Expect(err).NotTo(HaveOccurred())
		resolved, err := o.resolveTestOptions(detected, packages)
		Expect(err).NotTo(HaveOccurred())
		Expect(lo.Keys(resolved)).To(ConsistOf(parsers.GoTest, parsers.Ginkgo, parsers.Jest))
		Expect(resolved[parsers.Ginkgo].Tags).To(Equal([]string{"integration", "postgres"}))
		process, err := runners.GoTestCommand(resolved[parsers.GoTest].ForPackages("./plain"), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(process.Args).To(ContainElement("-tags=integration,postgres"))
	})

	It("rejects tags when no selected framework supports them", func() {
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: GinkgoT().TempDir(), Tags: []string{"integration"}}}
		packages := map[Framework][]string{parsers.Jest: {"./web"}, parsers.Vitest: {"./app"}}
		detected, err := parsedTestOptions(o, lo.Keys(packages)...)
		Expect(err).NotTo(HaveOccurred())
		_, err = o.resolveTestOptions(detected, packages)
		Expect(err).To(MatchError(ContainSubstring("--tags is not supported")))
	})

	It("salts run-cache fingerprints with tags", func() {
		workDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module example.com/tags\n\ngo 1.24\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "tags.go"), []byte("package tags\n"), 0o644)).To(Succeed())
		plain, err := newSelectorContext(workDir, nil)
		Expect(err).NotTo(HaveOccurred())
		tagged, err := newSelectorContext(workDir, []string{"integration"})
		Expect(err).NotTo(HaveOccurred())
		plainFingerprint, err := plain.hasher.Effective("example.com/tags")
		Expect(err).NotTo(HaveOccurred())
		taggedFingerprint, err := tagged.hasher.Effective("example.com/tags")
		Expect(err).NotTo(HaveOccurred())
		Expect(taggedFingerprint.Hex).NotTo(Equal(plainFingerprint.Hex))
	})
})

var _ = Describe("--failed", func() {
	writeSnapshot := func(tests ...parsers.Test) string {
		data, err := json.Marshal(testui.Snapshot{Tests: tests})
		Expect(err).NotTo(HaveOccurred())
		path := filepath.Join(GinkgoT().TempDir(), "failed.json")
		Expect(os.WriteFile(path, data, 0o644)).To(Succeed())
		return path
	}

	It("narrows to failed packages and focuses each framework on its failed tests", func() {
		path := writeSnapshot(
			parsers.Test{Name: "TestFoo", PackagePath: "./go", Framework: parsers.GoTest, Failed: true},
			parsers.Test{Name: "does (work)", PackagePath: "./ginkgo", Framework: parsers.Ginkgo, Failed: true},
			parsers.Test{Name: "renders [a] map", PackagePath: "./web", Framework: parsers.Vitest, Failed: true},
			parsers.Test{Name: "opens page", PackagePath: "./jest", Framework: parsers.Jest, Failed: true},
			parsers.Test{Name: "logs in", PackagePath: "./e2e", Framework: parsers.Playwright, Failed: true},
		)
		filtered, names, err := applyFailedFilter(map[Framework][]string{
			parsers.GoTest:     {"./go", "./passing"},
			parsers.Ginkgo:     {"./ginkgo"},
			parsers.Vitest:     {"./web"},
			parsers.Jest:       {"./jest"},
			parsers.Playwright: {"./e2e"},
		}, "", path)
		Expect(err).NotTo(HaveOccurred())
		Expect(filtered).To(Equal(map[Framework][]string{
			parsers.GoTest: {"./go"}, parsers.Ginkgo: {"./ginkgo"}, parsers.Vitest: {"./web"}, parsers.Jest: {"./jest"}, parsers.Playwright: {"./e2e"},
		}))
		Expect(failedTestOptions(names)).To(Equal(map[Framework]runners.FileSelection{
			parsers.GoTest:     {Exact: []string{"TestFoo"}},
			parsers.Ginkgo:     {Regex: []string{`does \(work\)`}},
			parsers.Vitest:     {Regex: []string{`renders \[a\] map$`}},
			parsers.Jest:       {Regex: []string{`opens page$`}},
			parsers.Playwright: {Regex: []string{`logs in$`}},
		}))
	})

	It("rejects failed packages missing from discovery", func() {
		path := writeSnapshot(parsers.Test{Name: "fails in CI", PackagePath: "./ci-only", Framework: parsers.Ginkgo, Failed: true})
		_, _, err := applyFailedFilter(map[Framework][]string{parsers.Ginkgo: {"./local"}}, "", path)
		Expect(err).To(MatchError(ContainSubstring("did not match any detected packages")))
	})

	It("unions the failures of every run started within a duration window", func() {
		stateRoot := GinkgoT().TempDir()
		now := time.Now().UTC()
		record := func(age time.Duration, tests ...parsers.Test) {
			_, err := snapshots.SavePerRun(stateRoot, &testui.Snapshot{Tests: tests}, now.Add(-age), "")
			Expect(err).NotTo(HaveOccurred())
		}
		web := func(name string, failed bool) parsers.Test {
			return parsers.Test{Name: name, PackagePath: "./web", Framework: parsers.Vitest, Failed: failed, Passed: !failed}
		}
		record(48*time.Hour, web("failed before the window", true))
		record(3*time.Hour, web("times out under load", true), web("renders map", true))
		record(time.Hour, web("renders map", true), web("times out under load", false))

		filtered, names, err := applyFailedFilter(map[Framework][]string{parsers.Vitest: {"./web", "./other"}}, stateRoot, "1d")
		Expect(err).NotTo(HaveOccurred())
		Expect(filtered).To(Equal(map[Framework][]string{parsers.Vitest: {"./web"}}))
		Expect(names).To(Equal(map[Framework][]string{parsers.Vitest: {"renders map", "times out under load"}}))
	})

	It("rejects a duration window with no failed runs", func() {
		_, _, err := applyFailedFilter(map[Framework][]string{parsers.Vitest: {"./web"}}, GinkgoT().TempDir(), "1d")
		Expect(err).To(MatchError(ContainSubstring("no failed tests found in runs started within 1d")))
	})
})
