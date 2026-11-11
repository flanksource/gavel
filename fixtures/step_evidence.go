package fixtures

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"sync"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"

	"github.com/flanksource/gavel/models"
)

// stepEvidence is what one attempt of a step recorded as it ran: its stdout
// and stderr, the structured details a UI renders (a query step's "query"
// record), and the command it ran. Each record is kept here first, then
// published to the live view (the fixture's progress node) and, when the step
// runs as a clicky task of its own, to that task's snapshot — so what a running
// step showed is what its result keeps, even when Execute failed.
//
// Both streams are read for workflow commands (see workflow_commands.go): a
// command line is applied and kept out of the output, everything else is kept
// and streamed as it is written.
type stepEvidence struct {
	started  time.Time
	live     io.Writer
	task     *task.Task
	progress func(done, total int) error

	mu             sync.Mutex
	stdout, stderr bytes.Buffer
	details        map[string]any
	command, cwd   string

	lines    [2]evidenceLine
	filter   workflowFilter
	groups   []string
	findings []models.Violation
	failures []error
}

// evidenceLine is where one stream is within its current line: pending holds
// a line start that may be a workflow command until its newline; plain marks
// a line already known not to be one, streamed as it arrives.
type evidenceLine struct {
	pending []byte
	plain   bool
}

// newStepEvidence starts an attempt's record. ctx carries the live output
// stream and the step's own task, when the runner gave it either.
func newStepEvidence(ctx context.Context, progress func(done, total int) error) *stepEvidence {
	e := &stepEvidence{started: time.Now(), live: fixtureOutput(ctx), task: fixtureTask(ctx), progress: progress, details: map[string]any{}}
	if e.task != nil {
		e.task.SetOutputProvider(e.output)
		e.task.SetDetailsProvider(e.snapshot)
	}
	return e
}

// recorder is the attempt's record. A StepContext the runner did not start —
// a kind's unit test — records into one of its own with no live view.
func (s *StepContext) recorder() *stepEvidence {
	if s.evidence == nil {
		s.evidence = newStepEvidence(context.Background(), s.Opts.Progress)
	}
	return s.evidence
}

// Stdout is where the step writes its standard output (a process's stdout, a
// SQL program's PRINT messages): kept as the result's Stdout and streamed to
// the live view as it is written.
func (s *StepContext) Stdout() io.Writer { return evidenceStream{s.recorder(), false} }

// Stderr is Stdout for the step's error output.
func (s *StepContext) Stderr() io.Writer { return evidenceStream{s.recorder(), true} }

// SetDetail records key in the result's Metadata (copied to plain values, so
// the caller may go on mutating value) and publishes it to the live view.
func (s *StepContext) SetDetail(key string, value any) error {
	plain, err := PlainValue(value)
	if err != nil {
		return fmt.Errorf("step detail %q: %w", key, err)
	}
	s.recorder().setDetail(key, plain)
	return nil
}

func (e *stepEvidence) setDetail(key string, plain any) {
	e.mu.Lock()
	e.details[key] = plain
	e.mu.Unlock()
	e.publish(key, plain)
}

// publish sets key on the live view only.
func (e *stepEvidence) publish(key string, plain any) {
	if stream, ok := e.live.(*outputStream); ok {
		stream.setMetadata(key, plain)
	}
}

// SetCommand records the command the step runs and where, as the result's
// Command and CWD.
func (s *StepContext) SetCommand(command, cwd string) {
	e := s.recorder()
	e.mu.Lock()
	defer e.mu.Unlock()
	e.command, e.cwd = command, cwd
}

// Progress reports done of total units of the step's work (tests, linters) to
// the live view and the step's task.
func (s *StepContext) Progress(done, total int) error {
	return s.recorder().reportProgress(done, total)
}

func (e *stepEvidence) reportProgress(done, total int) error {
	if e.task != nil {
		e.task.SetProgress(done, total)
	}
	if e.progress == nil {
		return nil
	}
	return e.progress(done, total)
}

// fill folds the record into result wherever the kind left the field empty:
// what a kind set itself wins. The findings the step's workflow commands
// reported are added as its children, and only ever worsen its status.
func (e *stepEvidence) fill(result FixtureResult) FixtureResult {
	e.consume(false, nil, true)
	e.consume(true, nil, true)
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, field := range []struct {
		target *string
		value  string
	}{{&result.Stdout, e.stdout.String()}, {&result.Stderr, e.stderr.String()}, {&result.Command, e.command}, {&result.CWD, e.cwd}} {
		if *field.target == "" {
			*field.target = field.value
		}
	}
	if len(e.details) > 0 && result.Metadata == nil {
		result.Metadata = map[string]any{}
	}
	for key, value := range e.details {
		if _, ok := result.Metadata[key]; !ok {
			result.Metadata[key] = value
		}
	}
	if result.Duration == 0 {
		result.Duration = time.Since(e.started)
	}
	return e.applyFindingsLocked(result)
}

func (e *stepEvidence) applyFindingsLocked(result FixtureResult) FixtureResult {
	var errorCount, warnCount int
	var firstError string
	for _, v := range e.findings {
		switch v.Severity {
		case models.SeverityError:
			if errorCount == 0 {
				firstError = v.Rule.Method + ": " + *v.Message
			}
			errorCount++
		case models.SeverityWarning:
			warnCount++
		}
		result.Children = append(result.Children, ViolationNode(v))
	}
	if len(e.failures) > 0 {
		return result.Errorf(errors.Join(e.failures...), "workflow commands")
	}
	completed := isPassingStatus(result.Status) || result.Status == task.StatusWarning
	switch status := FindingsStatus(errorCount, warnCount); {
	case status == task.StatusFAIL && completed:
		result.Status = status
		result.Error = fmt.Sprintf("%d error finding(s), %d other: %s", errorCount, len(e.findings)-errorCount, firstError)
	case status == task.StatusWarning && isPassingStatus(result.Status):
		result.Status = status
		result.Error = fmt.Sprintf("%d warning finding(s)", warnCount)
	}
	return result
}

func (e *stepEvidence) output() task.OutputSnapshot {
	e.mu.Lock()
	defer e.mu.Unlock()
	return task.OutputSnapshot{Stdout: e.stdout.String(), Stderr: e.stderr.String()}
}

func (e *stepEvidence) snapshot() any {
	e.mu.Lock()
	defer e.mu.Unlock()
	details := maps.Clone(e.details)
	if e.command != "" {
		details["command"] = e.command
	}
	if e.cwd != "" {
		details["cwd"] = e.cwd
	}
	return details
}

// evidenceStream is Stdout or Stderr: it keeps each write, then streams it to
// the live view. A write never fails — it may be teed from a running process's
// pipe, and failing it would cut the process's own capture short.
type evidenceStream struct {
	e      *stepEvidence
	stderr bool
}

func (w evidenceStream) Write(p []byte) (int, error) {
	w.e.consume(w.stderr, p, false)
	return len(p), nil
}

// consume reads p into the stream's record; final ends its last line, as when
// the step finished without a trailing newline. The commands it read are
// applied after the lock is released: they publish, and a publish may call
// back into the record.
func (e *stepEvidence) consume(stderr bool, p []byte, final bool) {
	e.mu.Lock()
	kept, commands, found := e.scanLocked(stderr, p, final)
	var findings any
	var err error
	if found {
		findings, err = PlainValue(e.findings)
	}
	e.mu.Unlock()
	if len(kept) > 0 {
		_, _ = e.live.Write(kept)
	}
	if err != nil {
		e.fail(fmt.Errorf("publish workflow findings: %w", err))
	} else if found {
		e.publish("findings", findings)
	}
	for _, cmd := range commands {
		e.apply(cmd)
	}
}

// scanLocked appends p to the stream's line and reads off every complete
// line: a line that cannot be a workflow command is kept as it arrives, one
// that may be is held until its newline. kept is what it added to the stream,
// commands the ones to apply, found whether it recorded a finding.
func (e *stepEvidence) scanLocked(stderr bool, p []byte, final bool) (kept []byte, commands []workflowCommand, found bool) {
	line, buffer := &e.lines[0], &e.stdout
	if stderr {
		line, buffer = &e.lines[1], &e.stderr
	}
	keep := func(b []byte) {
		buffer.Write(b)
		kept = append(kept, b...)
	}
	line.pending = append(line.pending, p...)
	for len(line.pending) > 0 {
		newline := bytes.IndexByte(line.pending, '\n')
		if line.plain || !mayBeWorkflowCommand(line.pending) {
			if newline < 0 {
				keep(line.pending)
				line.pending, line.plain = line.pending[:0], true
				break
			}
			keep(line.pending[:newline+1])
			line.pending, line.plain = line.pending[newline+1:], false
			continue
		}
		if newline < 0 && !final {
			break
		}
		end := newline
		if end < 0 {
			end = len(line.pending)
		}
		parsed := e.filter.line(string(line.pending[:end]))
		if !parsed.Drop {
			keep([]byte(parsed.Text))
			if newline >= 0 {
				keep([]byte{'\n'})
			}
		}
		if parsed.Err != nil {
			e.findings, found = append(e.findings, invalidWorkflowCommand(parsed.Err)), true
		} else if parsed.Command != nil {
			if cmd, recorded := e.recordLocked(*parsed.Command); recorded {
				found = true
			} else if cmd != nil {
				commands = append(commands, *cmd)
			}
		}
		line.pending = line.pending[end:]
		if newline >= 0 {
			line.pending = line.pending[1:]
		}
	}
	return kept, commands, found
}

// recordLocked records what a command changes in the record itself (the open
// groups, the findings) and returns any command still to apply.
func (e *stepEvidence) recordLocked(cmd workflowCommand) (*workflowCommand, bool) {
	switch cmd.Name {
	case "group":
		e.groups = append(e.groups, cmd.Message)
	case "endgroup":
		if len(e.groups) > 0 {
			e.groups = e.groups[:len(e.groups)-1]
		}
	case "error", "warning", "notice":
		var group string
		if len(e.groups) > 0 {
			group = e.groups[len(e.groups)-1]
		}
		v, err := workflowFinding(cmd, group)
		if err != nil {
			v = invalidWorkflowCommand(err)
		}
		e.findings = append(e.findings, v)
		return nil, true
	case "progress":
		var err error
		if cmd.Done, cmd.Total, err = workflowProgress(cmd); err != nil {
			e.findings = append(e.findings, invalidWorkflowCommand(err))
			return nil, true
		}
		return &cmd, false
	default:
		return &cmd, false
	}
	return nil, false
}

// apply publishes a command that reports on the running step.
func (e *stepEvidence) apply(cmd workflowCommand) {
	switch cmd.Name {
	case "progress":
		if err := e.reportProgress(cmd.Done, cmd.Total); err != nil {
			e.fail(fmt.Errorf("::progress %d/%d: %w", cmd.Done, cmd.Total, err))
		}
	case "description":
		if e.task != nil {
			e.task.SetDescription(cmd.Message)
		}
		e.setDetail("description", cmd.Message)
	case "status":
		if e.task != nil {
			e.task.Infof("%s", cmd.Message)
		}
		e.setDetail("status", cmd.Message)
	case "debug":
		if e.task != nil {
			e.task.Debugf("%s", cmd.Message)
		} else {
			logger.Debugf("%s", cmd.Message)
		}
	}
}

func (e *stepEvidence) fail(err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.failures = append(e.failures, err)
}

// mayBeWorkflowCommand reports whether a line starting with b could still be
// a workflow command: every command line starts with "::".
func mayBeWorkflowCommand(b []byte) bool {
	switch len(b) {
	case 0:
		return true
	case 1:
		return b[0] == ':'
	}
	return b[0] == ':' && b[1] == ':'
}

// invalidWorkflowCommand is a command line that did not parse, reported as
// an error finding rather than dropped.
func invalidWorkflowCommand(err error) models.Violation {
	return models.Violation{Severity: models.SeverityError, Message: models.StringPtr(err.Error()),
		Source: WorkflowCommandSource, Rule: &models.Rule{Method: "::invalid"}}
}
