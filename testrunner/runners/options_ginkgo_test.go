package runners

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// parsedFlags is the part of TestOptions the shared flags fill.
type parsedFlags struct {
	Focus, Skip FileSelection
	FailFast    bool
	ExtraArgs   []string
}

func parseFlags(argv ...string) parsedFlags {
	opts, err := ParseTestArgs(GinkgoT().TempDir(), argv)
	Expect(err).NotTo(HaveOccurred())
	return parsedFlags{Focus: opts.Focus, Skip: opts.Skip, FailFast: opts.FailFast, ExtraArgs: opts.ExtraArgs}
}

// valueFlagEntries builds =value and space-separated entries for each
// spelling, placed between unrelated runner arguments, after the table body.
func valueFlagEntries(body func([]string, parsedFlags), names []string, want func(value string) parsedFlags) []any {
	entries := []any{body}
	for _, name := range names {
		entries = append(entries,
			Entry(name+"=value", []string{"--", "-race", name + "=x", "-count=1"}, want("x")),
			Entry(name+" value", []string{"--", "-race", name, "x", "-count=1"}, want("x")),
		)
	}
	return entries
}

var surrounding = []string{"-race", "-count=1"}

var _ = Describe("ParseTestArgs", func() {
	check := func(argv []string, want parsedFlags) {
		Expect(parseFlags(argv...)).To(Equal(want))
	}

	DescribeTable("focus spellings",
		valueFlagEntries(check,
			[]string{"-run", "--run", "--focus", "-focus", "-ginkgo.focus", "--ginkgo.focus", "-t", "--testNamePattern", "--test-name-pattern"},
			func(v string) parsedFlags {
				return parsedFlags{Focus: FileSelection{Regex: []string{v}}, ExtraArgs: surrounding}
			})...,
	)

	DescribeTable("skip spellings",
		valueFlagEntries(check,
			[]string{"-skip", "--skip", "-ginkgo.skip", "--ginkgo.skip"},
			func(v string) parsedFlags {
				return parsedFlags{Skip: FileSelection{Regex: []string{v}}, ExtraArgs: surrounding}
			})...,
	)

	DescribeTable("fail-fast spellings", check,
		Entry("-failfast", []string{"--", "-race", "-failfast", "-count=1"}, parsedFlags{FailFast: true, ExtraArgs: surrounding}),
		Entry("--fail-fast", []string{"--", "-race", "--fail-fast", "-count=1"}, parsedFlags{FailFast: true, ExtraArgs: surrounding}),
		Entry("-ginkgo.fail-fast", []string{"--", "-race", "-ginkgo.fail-fast", "-count=1"}, parsedFlags{FailFast: true, ExtraArgs: surrounding}),
		Entry("--bail", []string{"--", "-race", "--bail", "-count=1"}, parsedFlags{FailFast: true, ExtraArgs: surrounding}),
		Entry("--failfast=true", []string{"--", "-race", "--failfast=true", "-count=1"}, parsedFlags{FailFast: true, ExtraArgs: surrounding}),
		Entry("--failfast=false", []string{"--", "-race", "--failfast=false", "-count=1"}, parsedFlags{ExtraArgs: surrounding}),
	)

	It("unions repeated focus flags", func() {
		Expect(parseFlags("--", "-run", "A", "--focus=B", "-t", "A").Focus).To(Equal(FileSelection{Regex: []string{"A", "B"}}))
	})

	DescribeTable("rejects malformed flags",
		func(argv []string, message string) {
			_, err := ParseTestArgs(GinkgoT().TempDir(), argv, GinkgoFlags...)
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("missing value at the end", []string{"--", "-run"}, "-run requires a value"),
		Entry("missing value before a second --", []string{"--", "--focus", "--"}, "--focus requires a value"),
		Entry("empty =value", []string{"--", "-run="}, "-run requires a non-empty value"),
		Entry("non-boolean fail-fast", []string{"--", "--bail=maybe"}, "expected true or false"),
		Entry("flag before --", []string{"-run", "x"}, "unexpected flag -run before --"),
		Entry("conflicting parallelism", []string{"--", "-p", "2", "-p=3"}, "conflicting parallelism 2 and 3"),
		Entry("non-positive parallelism", []string{"--", "-p", "0"}, "expected a positive number"),
		Entry("conflicting label filters", []string{"--", "--label-filter", "a", "--label-filter=b"}, `conflicting label filters "a" and "b"`),
		Entry("conflicting benchmarks", []string{"--", "-bench", "A", "-bench", "B"}, `conflicting benchmark patterns "A" and "B"`),
		Entry("several Ginkgo nodes", []string{"--", "--nodes", "2"}, "--nodes=2 is unsupported"),
	)

	Describe("paths", func() {
		var cwd string

		BeforeEach(func() {
			cwd = GinkgoT().TempDir()
			writeTree(cwd, map[string]string{"pkg/a_test.go": "package pkg\n"})
		})

		It("splits positional paths into directories and files", func() {
			opts, err := ParseTestArgs(cwd, []string{"pkg", "pkg/a_test.go"})
			Expect(err).NotTo(HaveOccurred())
			Expect(opts.Paths).To(Equal([]string{"./pkg"}))
			Expect(opts.Files).To(Equal(FileSelection{Exact: []string{"pkg/a_test.go"}}))
			Expect(opts.ExtraArgs).To(BeEmpty())
		})

		It("rejects a missing positional path", func() {
			_, err := ParseTestArgs(cwd, []string{"nope"})
			Expect(err).To(MatchError(ContainSubstring("starting path nope does not exist")))
		})

		It("takes an existing file after -- as a test file and keeps other words as runner arguments", func() {
			opts, err := ParseTestArgs(cwd, []string{"--", "pkg/a_test.go", "word"})
			Expect(err).NotTo(HaveOccurred())
			Expect(opts.Files).To(Equal(FileSelection{Exact: []string{"pkg/a_test.go"}}))
			Expect(opts.ExtraArgs).To(Equal([]string{"word"}))
		})
	})

	Describe("runner flags", func() {
		It("fills Ginkgo's label filter and drops a single node", func() {
			opts, err := ParseTestArgs(GinkgoT().TempDir(), []string{"--", "--label-filter", "fast", "--nodes=1", "-v"}, GinkgoFlags...)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts.LabelFilter).To(Equal("fast"))
			Expect(opts.ExtraArgs).To(Equal([]string{"-v"}))
		})

		It("fills go test's tags, parallelism and benchmark", func() {
			opts, err := ParseTestArgs(GinkgoT().TempDir(), []string{"--", "-tags", "a,b", "-tags=b,c", "-p", "3", "-bench=.", "-v"}, GoFlags...)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts.Tags).To(Equal([]string{"a", "b", "c"}))
			Expect(opts.Parallel).To(Equal(3))
			Expect(opts.Bench).To(Equal("."))
			Expect(opts.ExtraArgs).To(Equal([]string{"-v"}))
		})

		It("keeps another runner's flag and its value as runner arguments", func() {
			Expect(parseFlags("--", "--label-filter", "fast").ExtraArgs).To(Equal([]string{"--label-filter", "fast"}))
		})
	})
})

var _ = Describe("TestOptions.Merge", func() {
	It("treats a token as extra only when neither parse consumed it", func() {
		cwd := GinkgoT().TempDir()
		argv := []string{"--", "--label-filter", "fast", "-run", "A", "-race"}
		goOpts, err := ParseTestArgs(cwd, argv, GoFlags...)
		Expect(err).NotTo(HaveOccurred())
		Expect(goOpts.ExtraArgs).To(Equal([]string{"--label-filter", "fast", "-race"}))
		ginkgoOpts, err := ParseTestArgs(cwd, argv, GinkgoFlags...)
		Expect(err).NotTo(HaveOccurred())

		merged, err := goOpts.Merge(*ginkgoOpts)
		Expect(err).NotTo(HaveOccurred())
		Expect(merged.ExtraArgs).To(Equal([]string{"-race"}))
		Expect(merged.LabelFilter).To(Equal("fast"))
		Expect(merged.Focus).To(Equal(FileSelection{Regex: []string{"A"}}))
	})

	DescribeTable("rejects incompatible options",
		func(a, b TestOptions, message string) {
			_, err := a.Merge(b)
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("scalar conflict", TestOptions{Parallel: 2}, TestOptions{Parallel: 3}, "conflicting parallelism 2 and 3"),
		Entry("different working directories", TestOptions{Cwd: "/a"}, TestOptions{Cwd: "/b"}, "cannot merge test options for /a and /b"),
	)

	It("adds up options parsed from different arguments", func() {
		cwd := GinkgoT().TempDir()
		cli, err := ParseTestArgs(cwd, []string{"--", "-run", "A", "-race"}, GoFlags...)
		Expect(err).NotTo(HaveOccurred())
		config, err := ParseTestArgs(cwd, []string{"--", "-count=1", "-skip", "B"}, GoFlags...)
		Expect(err).NotTo(HaveOccurred())

		merged, err := cli.Merge(*config)
		Expect(err).NotTo(HaveOccurred())
		Expect(merged.ExtraArgs).To(Equal([]string{"-race", "-count=1"}))
		Expect(merged.Focus).To(Equal(FileSelection{Regex: []string{"A"}}))
		Expect(merged.Skip).To(Equal(FileSelection{Regex: []string{"B"}}))
	})

	It("unions lists and keeps the set side of each scalar", func() {
		merged, err := TestOptions{Paths: []string{"./a"}, Tags: []string{"x"}, Parallel: 2}.
			Merge(TestOptions{Paths: []string{"./a", "./b"}, Tags: []string{"y"}, Bench: ".", FailFast: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Join(merged.Paths, " ")).To(Equal("./a ./b"))
		Expect(merged.Tags).To(Equal([]string{"x", "y"}))
		Expect(merged.Parallel).To(Equal(2))
		Expect(merged.Bench).To(Equal("."))
		Expect(merged.FailFast).To(BeTrue())
	})
})
