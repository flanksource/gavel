package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const verifyReportDocument = `---
codeBlocks: [bash]
---

# Definition of done

### command: prints ok

` + "```bash\necho ok\n```" + `

- cel: exitCode == 0
`

const verifyReportRedDocument = `---
codeBlocks: [bash]
---

# Definition of done

### command: prints trouble

` + "```bash\necho trouble\n```" + `

- cel: stdout.contains('ok')
`

// ndjsonLine mirrors captain's ExternalVerifier stdout protocol.
type ndjsonLine struct {
	Progress *api.VerifyReport `json:"progress,omitempty"`
	Report   *api.VerifyReport `json:"report,omitempty"`
}

func runVerifyReportCommand(document, cwd string) (lines []ndjsonLine, err error) {
	GinkgoHelper()
	previousFormat, previousCwd := clicky.Flags.Format, workingDir
	clicky.Flags.Format, workingDir = VerifyReportFormat, cwd
	DeferCleanup(func() { clicky.Flags.Format, workingDir = previousFormat, previousCwd })

	var stdout bytes.Buffer
	fixturesRunCmd.SetIn(strings.NewReader(document))
	fixturesRunCmd.SetOut(&stdout)
	fixturesRunCmd.SetErr(GinkgoWriter)
	fixturesRunCmd.SetContext(context.Background())
	DeferCleanup(func() { fixturesRunCmd.SetIn(nil); fixturesRunCmd.SetOut(nil); fixturesRunCmd.SetErr(nil) })

	err = runFixturesVerifyReport(fixturesRunCmd, nil)
	for _, raw := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		var line ndjsonLine
		Expect(json.Unmarshal([]byte(raw), &line)).To(Succeed(), raw)
		lines = append(lines, line)
	}
	return lines, err
}

var _ = Describe("gavel fixtures run --format captain-verify-report", func() {
	It("answers a passing document with exactly one valid report line", func() {
		lines, err := runVerifyReportCommand(verifyReportDocument, GinkgoT().TempDir())

		Expect(err).ToNot(HaveOccurred())
		reports := reportLines(lines)
		Expect(reports).To(HaveLen(1))
		Expect(reports[0].Validate()).To(Succeed())
		Expect(reports[0].Kind).To(Equal("fixture"))
		Expect(reports[0].Passed).To(BeTrue())
		Expect(reports[0].Summary.Passed).To(Equal(1))
	})

	It("reports a failing document as a verdict, not a runner error", func() {
		lines, err := runVerifyReportCommand(verifyReportRedDocument, GinkgoT().TempDir())

		Expect(err).ToNot(HaveOccurred(), "a red definition of done exits 0 with a report")
		reports := reportLines(lines)
		Expect(reports).To(HaveLen(1))
		Expect(reports[0].Validate()).To(Succeed())
		Expect(reports[0].Passed).To(BeFalse())
		Expect(reports[0].State).To(Equal(api.VerifyStateFailed))
		Expect(reports[0].Feedback).To(ContainSubstring("prints trouble"))
	})

	It("streams progress lines before the report", func() {
		lines, err := runVerifyReportCommand(verifyReportDocument, GinkgoT().TempDir())

		Expect(err).ToNot(HaveOccurred())
		Expect(len(lines)).To(BeNumerically(">", 1), "expected at least one progress line")
		for _, line := range lines[:len(lines)-1] {
			Expect(line.Progress).ToNot(BeNil())
			Expect(line.Report).To(BeNil())
			Expect(line.Progress.Validate()).To(Succeed())
		}
		Expect(lines[len(lines)-1].Report).ToNot(BeNil())
	})

	It("refuses a format it does not implement", func() {
		_, err := runVerifyReportCommandWithFormat(verifyReportDocument, GinkgoT().TempDir(), "json")
		Expect(err).To(MatchError(ContainSubstring("implements only --format captain-verify-report")))
	})

	It("refuses an empty document rather than passing vacuously", func() {
		_, err := runVerifyReportCommand("   \n", GinkgoT().TempDir())
		Expect(err).To(MatchError(ContainSubstring("no fixture document on stdin")))
	})
})

// The engine publishes progress from its own goroutine while the command
// writes the verdict; a reader parses the stream line by line, so two writers
// interleaving inside a line would corrupt both.
var _ = Describe("ndjsonWriter", func() {
	It("keeps concurrent progress and report lines whole", func() {
		var stdout safeBuffer
		out := &ndjsonWriter{out: &stdout}
		progress := api.VerifyReport{Kind: "fixture", Name: strings.Repeat("progress-", 200), State: api.VerifyStateRunning}
		report := api.VerifyReport{Kind: "fixture", Name: strings.Repeat("verdict-", 200), State: api.VerifyStatePassed, Passed: true}

		const writers = 16
		var wg sync.WaitGroup
		start := make(chan struct{})
		for i := 0; i < writers; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				<-start
				for n := 0; n < 50; n++ {
					if i == 0 {
						Expect(out.write("report", report)).To(Succeed())
						continue
					}
					Expect(out.write("progress", progress)).To(Succeed())
				}
			}(i)
		}
		close(start)
		wg.Wait()

		lines := strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n")
		Expect(lines).To(HaveLen(writers * 50))
		for _, raw := range lines {
			var line ndjsonLine
			Expect(json.Unmarshal([]byte(raw), &line)).To(Succeed(), raw)
			Expect((line.Progress != nil) != (line.Report != nil)).To(BeTrue(), "each line carries exactly one of progress/report")
		}
	})
})

// safeBuffer is a bytes.Buffer whose writes are serialised, so the assertion
// is on the writer under test rather than on the buffer's own races.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func runVerifyReportCommandWithFormat(document, cwd, format string) ([]ndjsonLine, error) {
	GinkgoHelper()
	previous := clicky.Flags.Format
	clicky.Flags.Format = format
	DeferCleanup(func() { clicky.Flags.Format = previous })
	fixturesRunCmd.SetIn(strings.NewReader(document))
	fixturesRunCmd.SetOut(GinkgoWriter)
	fixturesRunCmd.SetContext(context.Background())
	DeferCleanup(func() { fixturesRunCmd.SetIn(nil); fixturesRunCmd.SetOut(nil) })
	previousCwd := workingDir
	workingDir = cwd
	DeferCleanup(func() { workingDir = previousCwd })
	return nil, runFixturesVerifyReport(fixturesRunCmd, nil)
}

func reportLines(lines []ndjsonLine) []*api.VerifyReport {
	var reports []*api.VerifyReport
	for _, line := range lines {
		if line.Report != nil {
			reports = append(reports, line.Report)
		}
	}
	return reports
}
