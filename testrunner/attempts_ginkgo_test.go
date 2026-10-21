package testrunner

import (
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/traces"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// sigquitDump is the shape a Go binary prints on SIGQUIT under -json: a header
// with the runtime's gp/m annotations, then one block per goroutine.
const sigquitDump = `=== RUN   TestHang
SIGQUIT: quit
PC=0x188fe550c m=0 sigcode=0

goroutine 0 gp=0x1049c75e0 m=0 mp=0x1049c8020 [idle]:
runtime.pthread_cond_wait(0x1049c8588, 0x1049c8548)
	/usr/local/go/src/runtime/sys_darwin.go:547 +0x20

goroutine 1 gp=0x5b07cfcb81e0 m=nil [chan receive]:
testing.(*T).Run(0x14000003a00, {0x1009a1b2a, 0x8}, 0x100a3c1e8)
	/usr/local/go/src/testing/testing.go:1751 +0x3ac

goroutine 7 [select]:
example.com/hang.TestHang(0x14000003c00)
	/src/hang/hang_test.go:12 +0x34
`

var _ = Describe("stampAttempts", func() {
	It("counts the goroutines of a timed-out test's dump, keys its process and marks a batch-shared one", func() {
		results := parsers.TestSuiteResults{{Framework: parsers.GoTest, Tests: parsers.Tests{
			{Name: "TestHang", Failed: true, TimedOut: true, Stdout: sigquitDump},
			{Name: "TestOK", Passed: true},
		}}}
		usage := traces.ProcessUsage{Key: "42@1792000000", PeakCPUPercent: 87.5, PeakRSSBytes: 1 << 20}
		stampAttempts(results, attemptStamp{RunKind: "initial", PID: 42, ExitCode: 2, Usage: usage, SharedProcess: true})

		hang, ok := results[0].Tests[0].Attempts[0], results[0].Tests[1].Attempts[0]
		Expect(hang.StackTrace).To(HavePrefix("SIGQUIT: quit"))
		Expect(hang.GoroutineCt).To(Equal(3))
		Expect([]any{hang.ProcessKey, hang.CPUPercent, hang.RSS, hang.SharedProcess}).To(Equal([]any{usage.Key, usage.PeakCPUPercent, usage.PeakRSSBytes, true}))
		Expect(ok.GoroutineCt).To(BeZero(), "a passing test carries no dump")
		Expect(ok.ProcessKey).To(Equal(usage.Key))
	})
})
