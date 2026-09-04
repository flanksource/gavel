package verifier

import (
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
)

const (
	// maxFeedbackLines bounds the failure digest fed back into the next turn's
	// prompt. A red run can hold thousands of leaves; the agent needs the first
	// handful and a count, not the whole tree pasted into its context.
	maxFeedbackLines = 50
	// maxOutputTailBytes bounds one node's captured output in the digest.
	maxOutputTailBytes = 2000
)

// Feedback renders the failing part of a report as the text the agent reads on
// its next turn. It is the only thing the agent sees of the verification, so it
// names the node, why it failed, and the tail of what it printed.
//
// One cap covers the whole digest — failing nodes and unmet criteria alike — and
// the note counting what was cut comes last, after everything it summarises.
func Feedback(nodes []api.VerifyNode, checklist []api.VerifyChecklistItem) string {
	var lines []string
	truncated := 0
	add := func(line string) {
		if len(lines) >= maxFeedbackLines {
			truncated++
			return
		}
		lines = append(lines, line)
	}
	var walk func(prefix string, nodes []api.VerifyNode)
	walk = func(prefix string, nodes []api.VerifyNode) {
		for i := range nodes {
			node := &nodes[i]
			if len(node.Children) > 0 {
				walk(joinPath(prefix, node.Name), node.Children)
				continue
			}
			if node.Passed || node.Skipped || node.Pending {
				continue
			}
			add(failureLine(joinPath(prefix, node.Name), node))
		}
	}
	walk("", nodes)

	for _, item := range checklist {
		if item.Passed != nil && *item.Passed {
			continue
		}
		line := "- criterion not met: " + item.Item
		if item.Message != "" {
			line += ": " + item.Message
		}
		add(line)
	}
	if truncated > 0 {
		lines = append(lines, fmt.Sprintf("… (%d more failures truncated)", truncated))
	}
	return strings.Join(lines, "\n")
}

func failureLine(name string, node *api.VerifyNode) string {
	line := "- " + name
	if node.TimedOut {
		line += " (timed out)"
	}
	if node.Message != "" {
		line += ": " + node.Message
	}
	if output := tail(strings.TrimSpace(node.Stderr + "\n" + node.Stdout)); output != "" {
		line += "\n  " + strings.ReplaceAll(output, "\n", "\n  ")
	}
	return line
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + " › " + name
}

// tail keeps the end of a node's output: a failing command explains itself on
// its last lines, and its first ones are usually the banner.
func tail(output string) string {
	if len(output) <= maxOutputTailBytes {
		return output
	}
	return "…" + output[len(output)-maxOutputTailBytes:]
}
