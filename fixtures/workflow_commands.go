package fixtures

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/models"
)

// WorkflowCommandSource is the Violation.Source of a finding a step reported
// with a `::error`, `::warning` or `::notice` workflow command.
const WorkflowCommandSource = "workflow"

// workflowCommand is one GitHub Actions workflow command a step printed on a
// line of its own: `::name key=value,key=value::message`. Done and Total are
// a `::progress` command's, once read.
type workflowCommand struct {
	Name        string
	Properties  map[string]string
	Message     string
	Done, Total int
}

// workflowCommands are the commands a step's output is parsed for: GitHub's
// group, annotation and debug commands, and gavel's progress, description and
// status. A line naming any other command is ordinary output, as on GitHub.
var workflowCommands = map[string]bool{
	"group": true, "endgroup": true, "notice": true, "warning": true, "error": true, "debug": true,
	"progress": true, "description": true, "status": true, "stop-commands": true,
}

var (
	unescapeMessage  = strings.NewReplacer("%25", "%", "%0D", "\r", "%0A", "\n")
	unescapeProperty = strings.NewReplacer("%25", "%", "%0D", "\r", "%0A", "\n", "%3A", ":", "%2C", ",")
)

// parseWorkflowCommand reads line as a workflow command. ok is false for a
// line that is not one; err is set for a known command whose properties do
// not parse.
func parseWorkflowCommand(line string) (cmd workflowCommand, ok bool, err error) {
	line = strings.TrimSuffix(line, "\r")
	if !strings.HasPrefix(line, "::") {
		return cmd, false, nil
	}
	head, message, found := strings.Cut(line[2:], "::")
	if !found {
		return cmd, false, nil
	}
	name, properties, _ := strings.Cut(head, " ")
	if !workflowCommands[name] {
		return cmd, false, nil
	}
	cmd = workflowCommand{Name: name, Message: unescapeMessage.Replace(message)}
	if properties = strings.TrimSpace(properties); properties == "" {
		return cmd, true, nil
	}
	cmd.Properties = map[string]string{}
	for _, property := range strings.Split(properties, ",") {
		key, value, found := strings.Cut(property, "=")
		if key = strings.TrimSpace(key); !found || key == "" {
			return cmd, true, fmt.Errorf("::%s property %q is not key=value", name, property)
		}
		cmd.Properties[key] = unescapeProperty.Replace(value)
	}
	return cmd, true, nil
}

// workflowLine is what one line of output is once its workflow command, if
// any, is read: Text is what the output keeps of it (a group marker, written
// back in its canonical form for the renderer to draw a section), Drop that
// it keeps nothing, Command the command to apply.
type workflowLine struct {
	Text    string
	Drop    bool
	Command *workflowCommand
	Err     error
}

// workflowFilter reads a stream's lines in order: `::stop-commands::token`
// turns command parsing off until a `::token::` line turns it back on.
type workflowFilter struct {
	stopToken string
}

func (f *workflowFilter) line(line string) workflowLine {
	trimmed := strings.TrimSuffix(line, "\r")
	if f.stopToken != "" {
		if trimmed == "::"+f.stopToken+"::" {
			f.stopToken = ""
			return workflowLine{Drop: true}
		}
		return workflowLine{Text: line}
	}
	cmd, ok, err := parseWorkflowCommand(trimmed)
	switch {
	case !ok:
		return workflowLine{Text: line}
	case err != nil:
		return workflowLine{Drop: true, Err: err}
	case cmd.Name == "stop-commands":
		if cmd.Message == "" {
			return workflowLine{Drop: true, Err: errors.New("::stop-commands needs a token: ::stop-commands::<token>")}
		}
		f.stopToken = cmd.Message
		return workflowLine{Drop: true}
	case cmd.Name == "group":
		return workflowLine{Text: "::group::" + cmd.Message, Command: &cmd}
	case cmd.Name == "endgroup":
		return workflowLine{Text: "::endgroup::", Command: &cmd}
	}
	return workflowLine{Drop: true, Command: &cmd}
}

// StripWorkflowCommands is raw output without its workflow command lines,
// keeping group markers (in canonical form) for the renderer to draw sections:
// what a step's stored stdout and its CEL `stdout` are.
func StripWorkflowCommands(raw string) string {
	if !strings.Contains(raw, "::") {
		return raw
	}
	var filter workflowFilter
	var out strings.Builder
	for _, line := range strings.SplitAfter(raw, "\n") {
		content, newline := strings.CutSuffix(line, "\n")
		parsed := filter.line(content)
		if parsed.Drop {
			continue
		}
		out.WriteString(parsed.Text)
		if newline {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// workflowFinding is the finding an annotation command reports, in the group
// it was printed in.
func workflowFinding(cmd workflowCommand, group string) (models.Violation, error) {
	severity := map[string]models.ViolationSeverity{"error": models.SeverityError, "warning": models.SeverityWarning, "notice": models.SeverityInfo}[cmd.Name]
	rule := cmd.Properties["title"]
	if rule == "" {
		rule = "::" + cmd.Name
	}
	v := models.Violation{Severity: severity, Message: models.StringPtr(cmd.Message), Source: WorkflowCommandSource,
		Rule: &models.Rule{Method: rule}, File: cmd.Properties["file"]}
	if group != "" {
		v.Code = models.StringPtr(group)
	}
	for _, field := range []struct {
		key    string
		target *int
	}{{"line", &v.Line}, {"col", &v.Column}, {"endLine", &v.EndLine}, {"endColumn", &v.EndColumn}} {
		value, ok := cmd.Properties[field.key]
		if !ok {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || n < 0 {
			return models.Violation{}, fmt.Errorf("::%s %s=%q is not a line or column number", cmd.Name, field.key, value)
		}
		*field.target = n
	}
	return v, nil
}

// workflowProgress is the done and total a `::progress` command reports:
// `::progress done=3,total=10::` or `::progress::3/10`.
func workflowProgress(cmd workflowCommand) (done, total int, err error) {
	doneText, totalText := cmd.Properties["done"], cmd.Properties["total"]
	if doneText == "" && totalText == "" {
		var found bool
		if doneText, totalText, found = strings.Cut(strings.TrimSpace(cmd.Message), "/"); !found {
			return 0, 0, fmt.Errorf("::progress %q: want a <done>/<total> message or done and total properties", cmd.Message)
		}
	}
	if done, err = strconv.Atoi(strings.TrimSpace(doneText)); err != nil {
		return 0, 0, fmt.Errorf("::progress done %q is not a number", doneText)
	}
	if total, err = strconv.Atoi(strings.TrimSpace(totalText)); err != nil {
		return 0, 0, fmt.Errorf("::progress total %q is not a number", totalText)
	}
	if done < 0 || total < 0 {
		return 0, 0, fmt.Errorf("::progress %d/%d is negative", done, total)
	}
	return done, total, nil
}
