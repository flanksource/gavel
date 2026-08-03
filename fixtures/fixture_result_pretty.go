package fixtures

import (
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/task"
)

const maxFixtureDetailLines = 5

func (f FixtureResult) PrettyShort() api.Textable {
	if f.Display != nil && !f.Display.ShowPassed && isPassingStatus(f.Status) {
		return api.Text{}
	}
	return f.fixtureHeader()
}

func (f FixtureResult) Pretty() api.Text {
	if f.Display != nil && !f.Display.ShowPassed && isPassingStatus(f.Status) {
		return api.Text{}
	}

	text := f.fixtureHeader()
	details := f.fixtureDetailLines()
	if len(details) > maxFixtureDetailLines {
		omitted := len(details) - maxFixtureDetailLines + 1
		details = append(details[:maxFixtureDetailLines-1], clicky.Text(fmt.Sprintf("… %d more lines", omitted), "text-muted"))
	}
	for _, detail := range details {
		text = text.NewLine().Append("  ").Add(detail)
	}
	return text
}

func (f FixtureResult) fixtureHeader() api.Text {
	text := f.Status.Pretty().Append(" ")
	if f.Test.Name != "" {
		text = text.Add(f.Test.Pretty())
	} else {
		text = text.Append(f.Name, "italic text-orange-500")
	}
	if f.Duration > 0 {
		text = text.Space().Append(fmt.Sprintf("(%s)", f.Duration), "text-muted")
	}
	return text
}

func (f FixtureResult) fixtureDetailLines() []api.Text {
	failed := f.Status == task.StatusFAIL || f.Status == task.StatusERR || f.Status == task.StatusFailed
	var lines []api.Text
	lines = appendFixtureDetail(lines, "error: ", f.Error, "text-red-600")
	lines = appendFixtureDetail(lines, "cel: ", f.CELExpression, "text-red-500 font-mono")
	if f.Command != "" && f.showCommand() {
		command := f.Command
		if f.CWD != "" {
			command += " (cwd: " + relativePath(f.CWD) + ")"
		}
		lines = appendFixtureDetail(lines, "$ ", command, "text-gray-500 font-mono")
	}
	if len(f.CELVars) > 0 && f.showCELVars() {
		keys := make([]string, 0, len(f.CELVars))
		for key := range f.CELVars {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			lines = appendFixtureDetail(lines, "var: ", fmt.Sprintf("%s=%v", key, f.CELVars[key]), "text-gray-500 font-mono")
		}
	}
	if f.Stderr != "" && f.showStderr(failed) {
		lines = appendFixtureDetail(lines, "stderr: ", f.Stderr, "text-red-500 font-mono text-xs")
	}
	if f.Stdout != "" && f.showStdout(failed) {
		lines = appendFixtureDetail(lines, "stdout: ", fixtureStdoutContent(f.Stdout), "font-mono text-xs")
	}
	return lines
}

func appendFixtureDetail(lines []api.Text, label, content, style string) []api.Text {
	visible := (api.ANSIText{Content: content}).VisibleLines()
	for i, line := range visible {
		prefix := strings.Repeat(" ", len(label))
		if i == 0 {
			prefix = label
		}
		lines = append(lines, api.Text{Style: style}.Append(prefix).Add(line))
	}
	return lines
}
