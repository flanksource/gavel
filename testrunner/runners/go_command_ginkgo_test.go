package runners

import (
	"path/filepath"
	"regexp"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	ginkgoSuiteSource = `package suite_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
)

func TestSuite(t *testing.T) { RunSpecs(t, "suite") }
`
	ginkgoSpecsSource = `package suite_test

import . "github.com/onsi/ginkgo/v2"

var _ = Describe("specs", func() {})
`
	plainTestsSource = `package plain

import "testing"

func TestA(t *testing.T) {}
func TestB(t *testing.T) {}
func helper(t *testing.T) {}
`
)

// goModule writes a module with one Ginkgo suite (./suite) and one plain
// go test package (./plain) and returns its root.
func goModule() string {
	root := GinkgoT().TempDir()
	writeTree(root, map[string]string{
		"go.mod":              "module example.com/m\n",
		"suite/suite_test.go": ginkgoSuiteSource,
		"suite/specs_test.go": ginkgoSpecsSource,
		"plain/a_test.go":     plainTestsSource,
		"plain/c_test.go":     "package plain\n\nimport \"testing\"\n\nfunc TestC(t *testing.T) {}\n",
		"plain/impl.go":       "package plain\n",
	})
	return root
}

// goTestArgs parses argv for the runner and returns the go test argv for
// packages, with the generated overlay path replaced by "-overlay=<overlay>".
func goTestArgs(root string, flags []FlagSpec, ginkgo []string, packages []string, argv ...string) ([]string, error) {
	opts, err := ParseTestArgs(root, argv, flags...)
	Expect(err).NotTo(HaveOccurred())
	process, err := GoTestCommand(opts.ForPackages(packages...), ginkgo)
	if err != nil {
		return nil, err
	}
	overlay := regexp.MustCompile("^-overlay=" + regexp.QuoteMeta(filepath.Join(root, ".tmp", "gavel-ginkgo-overlay")) + `/[0-9a-f]{16}\.json$`)
	args := process.Args
	for i, arg := range args {
		if overlay.MatchString(arg) {
			args[i] = "-overlay=<overlay>"
		}
	}
	return args, nil
}

var _ = Describe("GoTestCommand", func() {
	var root string

	BeforeEach(func() {
		root = goModule()
	})

	DescribeTable("plain go test packages",
		func(argv []string, want []string) {
			Expect(goTestArgs(root, GoFlags, nil, []string{"./plain", "./other"}, argv...)).To(Equal(want))
		},
		Entry("focus renders as -run after the packages", []string{"--", "-v", "--focus", "TestA", "-count=1"},
			[]string{"test", "-json", "-v", "./plain", "./other", "-run=TestA", "-v", "-count=1"}),
		Entry("skip, fail-fast and benchmark", []string{"--", "-skip", "TestB", "-failfast", "-bench", "."},
			[]string{"test", "-json", "-v", "./plain", "./other", "-skip=TestB", "-failfast", "-bench=."}),
		Entry("tags and parallelism precede the packages", []string{"--", "-tags", "a,b", "-p", "2"},
			[]string{"test", "-json", "-v", "-tags=a,b", "-p=2", "./plain", "./other"}),
		Entry("Ginkgo runner arguments are dropped", []string{"--", "-ginkgo.v", "-count=1"},
			[]string{"test", "-json", "-v", "./plain", "./other", "-count=1"}),
	)

	It("arms the timeout before the packages and keeps a user override after them", func() {
		opts, err := ParseTestArgs(root, []string{"--", "-run", "TestFoo", "-timeout=1s"}, GoFlags...)
		Expect(err).NotTo(HaveOccurred())
		opts.Timeout = 4*time.Minute + 30*time.Second
		run, err := NewGoTest(root).BuildCommand(opts.ForPackages("./pkg"))
		Expect(err).NotTo(HaveOccurred())
		Expect(run.Process.Args).To(Equal([]string{"test", "-json", "-v", "-timeout=4m30s", "./pkg", "-run=TestFoo", "-timeout=1s"}))
	})

	It("rejects a Ginkgo label filter for plain packages", func() {
		_, err := goTestArgs(root, GinkgoFlags, nil, []string{"./plain"}, "--", "--label-filter", "fast")
		Expect(err).To(MatchError(UnsupportedOptionError{Framework: "go test", Option: "label filters"}))
	})

	It("renders names, label filter and fail-fast as Ginkgo filters after the package and runner arguments", func() {
		args, err := goTestArgs(root, GinkgoFlags, []string{"./suite"}, []string{"./suite"},
			"--", "--focus", "a b", "--skip=slow", "--label-filter", "fast", "--fail-fast", "-ginkgo.v", "--nodes=1")
		Expect(err).NotTo(HaveOccurred())
		Expect(args).To(Equal([]string{"test", "-json", "-v", "-overlay=<overlay>", "./suite", "-failfast", "-ginkgo.v",
			"-ginkgo.focus=a b", "-ginkgo.skip=slow", "-ginkgo.label-filter=fast", "-ginkgo.fail-fast"}))
	})

	It("builds a Ginkgo run without a file-based report", func() {
		opts, err := ParseTestArgs(root, []string{"--", "--focus", "x"}, GinkgoFlags...)
		Expect(err).NotTo(HaveOccurred())
		run, err := NewGinkgo(root).BuildCommand(opts.ForPackages("./suite"))
		Expect(err).NotTo(HaveOccurred())
		Expect(run.Framework).To(Equal(parsers.Ginkgo))
		Expect(run.ReportPath).To(BeEmpty())
		Expect(run.Process.Args[len(run.Process.Args)-1]).To(Equal("-ginkgo.focus=x"))
	})

	It("refuses to filter a run that mixes Ginkgo suites with plain packages", func() {
		_, err := goTestArgs(root, GinkgoFlags, []string{"./suite"}, []string{"./suite", "./plain"}, "--", "--focus", "x")
		Expect(err).To(MatchError(ContainSubstring("mixes Ginkgo suites [./suite] with plain go test packages")))
	})

	DescribeTable("test file selections",
		func(ginkgo []string, pkg string, argv []string, want []string) {
			Expect(goTestArgs(root, GinkgoFlags, ginkgo, []string{pkg}, argv...)).To(Equal(want))
		},
		Entry("a plain test file runs exactly its Test functions", nil, "./plain", []string{"plain/a_test.go"},
			[]string{"test", "-json", "-v", "./plain", "-run=(?:^TestA$)|(?:^TestB$)"}),
		Entry("the focus narrows the file's Test functions", nil, "./plain", []string{"plain/a_test.go", "--", "-run", "B"},
			[]string{"test", "-json", "-v", "./plain", "-run=^TestB$"}),
		Entry("a non-test source file runs the whole package", nil, "./plain", []string{"plain/impl.go"},
			[]string{"test", "-json", "-v", "./plain"}),
		Entry("a Ginkgo spec file runs the suite bootstrap focused on the file", []string{"./suite"}, "./suite", []string{"suite/specs_test.go"},
			[]string{"test", "-json", "-v", "-overlay=<overlay>", "./suite", "-run=^TestSuite$", `-ginkgo.focus-file=(^|/)specs_test\.go$`}),
	)

	It("fails when the focus selects none of the file's Test functions", func() {
		_, err := goTestArgs(root, GoFlags, nil, []string{"./plain"}, "plain/a_test.go", "--", "-run", "TestC")
		Expect(err).To(MatchError(ContainSubstring("no test in [a_test.go] matches focus regex:TestC")))
	})
})
