package fixtures_test

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func visibleFixtureLines(value string) []string {
	var lines []string
	for _, line := range strings.Split(ansi.Strip(value), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, strings.TrimSpace(line))
		}
	}
	return lines
}

var _ = Describe("fixture result rendering", func() {
	It("renders the native CEL trace directly beneath the failed result", func() {
		trace := "cel: a == b && c > d\n     │    │    │   │\n     │    │    │   └─ int(4)"
		result := fixtures.FixtureResult{
			Status:        task.StatusFAIL,
			Test:          fixtures.FixtureTest{Name: "traced fixture"},
			Error:         "a == b && c > d is false",
			CELExpression: "a == b && c > d",
			CELTrace:      trace,
		}

		lines := visibleFixtureLines(result.Pretty().ANSI())

		Expect(lines).To(ContainElements(
			"cel: a == b && c > d",
			"│    │    │   │",
			"│    │    │   └─ int(4)",
		))
		Expect(strings.Join(lines, "\n")).NotTo(ContainSubstring("cel: cel:"))
	})

	It("puts every requested detail below the header", func() {
		result := fixtures.FixtureResult{
			Status:        task.StatusFAIL,
			Test:          fixtures.FixtureTest{Name: "bounded fixture"},
			Error:         "\x1b[31mfirst error\nsecond error\x1b[0m",
			CELExpression: "actual == expected",
			CELVars:       map[string]any{"actual": 1, "expected": 2},
			Command:       "formula evaluate",
			Stderr:        "stderr one\nstderr two",
			Stdout:        "stdout one\nstdout two",
			Display: &fixtures.DisplayOptions{
				ShowPassed:  true,
				ShowCommand: true,
				ShowCELVars: true,
				ShowStdout:  fixtures.OutputAlways,
				ShowStderr:  fixtures.OutputAlways,
			},
		}

		lines := visibleFixtureLines(result.Pretty().ANSI())

		Expect(lines).To(HaveLen(11), "one header plus every detail line")
		Expect(lines[0]).To(ContainSubstring("bounded fixture"))
		Expect(lines[1]).To(ContainSubstring("error: first error"))
		Expect(lines[2]).To(ContainSubstring("second error"))
		Expect(lines[3]).To(ContainSubstring("cel: actual == expected"))
		Expect(lines[4]).To(ContainSubstring("$ formula evaluate"))
		Expect(lines[10]).To(ContainSubstring("stdout two"))
		Expect(result.Stderr).To(Equal("stderr one\nstderr two"))
		Expect(result.Stdout).To(Equal("stdout one\nstdout two"))
	})

	It("keeps error, command, stderr, and stdout on separate tree lines", func() {
		result := fixtures.FixtureResult{
			Status:  task.StatusFAIL,
			Test:    fixtures.FixtureTest{Name: "tree fixture"},
			Error:   "assertion failed",
			Command: "formula evaluate",
			Stderr:  "diagnostic",
			Stdout:  "calculation",
			Display: &fixtures.DisplayOptions{
				ShowPassed:  true,
				ShowCommand: true,
				ShowStdout:  fixtures.OutputAlways,
				ShowStderr:  fixtures.OutputAlways,
			},
		}
		node := fixtures.FixtureNode{Name: "tree fixture", Type: fixtures.TestNode, Results: &result}

		lines := visibleFixtureLines(api.NewTree(node.Tree()).ANSI())

		Expect(lines).To(HaveLen(5))
		Expect(lines[0]).To(ContainSubstring("tree fixture"))
		Expect(lines[1]).To(ContainSubstring("error: assertion failed"))
		Expect(lines[2]).To(ContainSubstring("$ formula evaluate"))
		Expect(lines[3]).To(ContainSubstring("stderr: diagnostic"))
		Expect(lines[4]).To(ContainSubstring("stdout: calculation"))
	})

	It("keeps SGR in terminal output and converts it for HTML", func() {
		result := fixtures.FixtureResult{
			Status: task.StatusFAIL,
			Test:   fixtures.FixtureTest{Name: "ANSI fixture"},
			Stdout: "\x1b[31mred <&>\x1b[0m\x1b[2J",
			Display: &fixtures.DisplayOptions{
				ShowPassed: true,
				ShowStdout: fixtures.OutputAlways,
				ShowStderr: fixtures.OutputNever,
			},
		}

		pretty := result.Pretty()

		Expect(pretty.ANSI()).To(ContainSubstring("\x1b[31mred <&>\x1b[0m"))
		Expect(pretty.ANSI()).NotTo(ContainSubstring("\x1b[2J"))
		Expect(pretty.HTML()).To(ContainSubstring("color:"))
		Expect(pretty.HTML()).To(ContainSubstring("red &lt;&amp;&gt;"))
		Expect(pretty.HTML()).NotTo(ContainSubstring("\x1b["))
	})

	It("provides a one-line task summary", func() {
		result := fixtures.FixtureResult{
			Status: task.StatusFAIL,
			Test:   fixtures.FixtureTest{Name: "short fixture"},
			Stdout: "first\nsecond",
		}

		lines := visibleFixtureLines(result.PrettyShort().ANSI())

		Expect(lines).To(HaveLen(1))
		Expect(lines[0]).To(ContainSubstring("short fixture"))
		Expect(lines[0]).NotTo(ContainSubstring("first"))
	})

	It("separates failed and error counts", func() {
		Expect(fixtures.Stats{Failed: 1, Error: 6}.Pretty().String()).To(Equal("1/6 errors"))
	})
})
