package testrunner

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command output ownership", func() {
	It("retains command stderr once in both streamed and final trees without contaminating test attempts", func() {
		const testCount = 200
		const commandMarker = "package diagnostic: "
		const assertionError = "assertion failed"
		assertions := make([]map[string]any, testCount)
		for i := range assertions {
			assertions[i] = map[string]any{"title": fmt.Sprintf("case-%d", i), "status": "passed"}
		}
		assertions[0]["status"] = "failed"
		assertions[0]["failureMessages"] = []string{assertionError}
		report, err := json.Marshal(map[string]any{"testResults": []any{map[string]any{
			"name": "pkg/example.test.ts", "assertionResults": assertions,
		}}})
		Expect(err).NotTo(HaveOccurred())
		commandOutput := commandMarker + strings.Repeat("diagnostic ", 8192)
		run := &runners.TestRun{Framework: parsers.Vitest, Parser: parsers.NewJestJSON("", parsers.Vitest), Process: exec.NewExec("vitest", "run")}
		results, err := (&TestOrchestrator{}).parseTestResults(run, &exec.ExecResult{Stdout: string(report), Stderr: commandOutput}, "pkg")
		Expect(err).NotTo(HaveOccurred())
		Expect(results[0].Tests[0].Stderr).To(Equal(assertionError))
		Expect(results[0].Tests[1].Stderr).To(BeEmpty())
		stampAttempts(results, attemptStamp{Started: time.Now(), PID: 123, ExitCode: 1})
		updates := make(chan []parsers.Test, 1)
		NewTestStreamer(updates).CompletePackage("pkg", parsers.Vitest, results)
		for _, tree := range []parsers.Tests{results.Tree(), <-updates} {
			payload, err := json.Marshal(tree)
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.Count(string(payload), commandMarker)).To(Equal(1))
			Expect(len(payload)).To(BeNumerically("<", 4*len(commandOutput)))
		}
	})
})
