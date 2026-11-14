package testrunner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/flanksource/gavel/slots"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test runner slots", func() {
	var root string
	BeforeEach(func() {
		root = GinkgoT().TempDir()
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		GinkgoT().Setenv("XDG_CACHE_HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("GOWORK", "off")
		Expect(os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("slots: {global: 1}\n"), 0o600)).To(Succeed())
	})
	It("batches an unfocused run into one ./... go test with slots enabled", func() {
		Expect(os.Mkdir(filepath.Join(root, "plain"), 0o700)).To(Succeed())
		Expect(os.Mkdir(filepath.Join(root, "specs"), 0o700)).To(Succeed())
		Expect(os.Mkdir(filepath.Join(root, "more"), 0o700)).To(Succeed())
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: root, Recursive: true}}
		plans, err := o.planGoBatches(map[Framework][]string{parsers.GoTest: {"./plain", "./more"}, parsers.Ginkgo: {"./specs"}}, nil, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(plans).To(HaveLen(1))
		Expect(plans[0].targets).To(Equal([]string{"./..."}))
		Expect(plans[0].ginkgo).To(Equal([]string{"./specs"}))
		Expect(plans[0].runtime.Slots).To(Equal(slots.Limits{Global: 1}))
		Expect(goBatchParallelism(plans[0])).To(Equal(runtime.GOMAXPROCS(0)), "one slot per parallel test binary, not per package")
	})
	DescribeTable("runs a batch with -p set to the granted slots",
		func(args []string, wantParallel, granted int, expectedP string) {
			plan := goBatchPlan{framework: parsers.GoTest, packages: []string{"./a", "./b", "./c"}, targets: []string{"./..."}, args: args}
			Expect(goBatchParallelism(plan)).To(Equal(wantParallel))
			o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: root}}
			process, err := o.buildGoBatchCommand(plan, granted)
			Expect(err).NotTo(HaveOccurred())
			var parallelism []string
			for i, arg := range process.Args {
				if arg == "-p" || len(arg) > 3 && arg[:3] == "-p=" {
					parallelism = append(parallelism, process.Args[i])
				}
			}
			Expect(parallelism).To(Equal([]string{expectedP}))
			Expect(process.Args[len(process.Args)-1]).To(Equal("./..."))
		},
		Entry("go's default -p", nil, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0), fmt.Sprintf("-p=%d", runtime.GOMAXPROCS(0))),
		Entry("downgraded to free slots", nil, runtime.GOMAXPROCS(0), 1, "-p=1"),
		Entry("user -p=N replaced", []string{"-p=8"}, 8, 2, "-p=2"),
		Entry("user -p N replaced", []string{"-p", "6", "-race"}, 6, 4, "-p=4"),
	)
	It("rejects a non-numeric -p", func() {
		_, err := goBatchParallelism(goBatchPlan{args: []string{"-p=many"}})
		Expect(err).To(MatchError(ContainSubstring(`got "many"`)))
	})
	It("waits before arming the Go package's timeout", func() {
		Expect(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/slots\n\ngo 1.22\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "sample_test.go"), []byte("package sample\nimport \"testing\"\nfunc TestAdmitted(t *testing.T) {}\n"), 0o600)).To(Succeed())
		warm := exec.Command("go", "test", "-run=^$", ".")
		warm.Dir = root
		output, err := warm.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		lease, err := slots.Acquire(context.Background(), slots.Request{Limits: slots.Limits{Global: 1}, WorkDir: root, Command: "blocker"})
		Expect(err).NotTo(HaveOccurred())
		defer lease.Close()
		released := make(chan error, 1)
		time.AfterFunc(2300*time.Millisecond, func() { released <- lease.Close() })
		started := time.Now()
		result, err := Run(RunOptions{WorkDir: root, StartingPaths: []string{"."}, Frameworks: []string{"go"}, TestTimeout: 2 * time.Second})
		Expect(err).NotTo(HaveOccurred())
		Expect(<-released).To(Succeed())
		Expect(time.Since(started)).To(BeNumerically(">=", 2300*time.Millisecond))
		Expect(parsers.Tests(result.([]parsers.Test)).Sum().Failed).To(BeZero())
	})
})
