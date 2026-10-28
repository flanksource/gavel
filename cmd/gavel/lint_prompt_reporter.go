package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/formatters"
	"github.com/flanksource/gavel/lint"
	lintprompt "github.com/flanksource/gavel/lint/prompt"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"golang.org/x/term"
)

type promptLintReportOptions struct {
	Options   LintOptions
	Writer    io.Writer
	Publisher lintprompt.Publisher
	Cancel    context.CancelCauseFunc
}

type promptFindingKey struct {
	Source, WorkDir, File, RulePackage, Rule, Message string
	Line                                              int
}

type promptLintReporter struct {
	mu         sync.Mutex
	options    promptLintReportOptions
	exclusions linters.RuleExclusions
	snapshot   testui.Snapshot
	path       string
	seen       map[promptFindingKey]bool
	errors     map[string]string
	err        error
	finished   bool
}

func newPromptLintReporter(options promptLintReportOptions) (*promptLintReporter, error) {
	if options.Writer == nil || options.Cancel == nil || options.Options.Dirs.StateRoot == "" {
		return nil, errors.New("prompt lint reporting requires a writer, cancellation callback, and state root")
	}
	exclusions, err := lint.ParseRuleExclusions(options.Options.Exclude)
	if err != nil {
		return nil, err
	}
	reporter := &promptLintReporter{
		options: options, exclusions: exclusions,
		seen: map[promptFindingKey]bool{}, errors: map[string]string{},
		snapshot: testui.Snapshot{
			Metadata: &testui.SnapshotMetadata{Version: version, Started: time.Now().UTC(), Kind: "lint"},
			Git:      snapshotGitInfo(options.Options.WorkDir),
			Status:   testui.SnapshotStatus{Running: true, LintRun: true},
			Lint:     []*linters.LinterResult{},
		},
	}
	reporter.path, err = snapshots.SavePerRun(options.Options.Dirs.StateRoot, &reporter.snapshot, reporter.snapshot.Metadata.Started, "")
	if err != nil {
		return nil, fmt.Errorf("create prompt lint snapshot: %w", err)
	}
	return reporter, nil
}

func (r *promptLintReporter) UpsertLintResult(result *linters.LinterResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.finished || r.err != nil {
		return
	}
	result = r.filtered(result)
	index := slices.IndexFunc(r.snapshot.Lint, func(old *linters.LinterResult) bool {
		return old.Linter == result.Linter && old.WorkDir == result.WorkDir
	})
	if index < 0 {
		r.snapshot.Lint = append(r.snapshot.Lint, result)
	} else {
		previous := r.snapshot.Lint[index]
		if result.InFlight() && !previous.InFlight() {
			return
		}
		if result.InFlight() {
			result.Violations = append(slices.Clone(previous.Violations), result.Violations...)
			result = r.filtered(result)
		}
		r.snapshot.Lint[index] = result
	}
	err := snapshots.WriteSnapshot(r.path, &r.snapshot)
	if err == nil && r.options.Publisher != nil {
		r.options.Publisher.UpsertLintResult(result)
	}
	if err == nil {
		err = r.emit(result)
	}
	if err != nil {
		r.err = fmt.Errorf("publish prompt lint result: %w", err)
		r.options.Cancel(r.err)
	}
}

func (r *promptLintReporter) RemoveLintResult(linter, workDir string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.finished || r.err != nil {
		return
	}
	r.snapshot.Lint = slices.DeleteFunc(r.snapshot.Lint, func(result *linters.LinterResult) bool {
		return result.Linter == linter && result.WorkDir == workDir
	})
	if err := snapshots.WriteSnapshot(r.path, &r.snapshot); err != nil {
		r.err = fmt.Errorf("remove prompt lint result: %w", err)
		r.options.Cancel(r.err)
		return
	}
	if r.options.Publisher != nil {
		r.options.Publisher.RemoveLintResult(linter, workDir)
	}
}

func (r *promptLintReporter) filtered(result *linters.LinterResult) *linters.LinterResult {
	copy := *result
	copy.Violations = []models.Violation{}
	seen := map[promptFindingKey]bool{}
	for _, violation := range r.exclusions.Filter(result.Linter, slices.Clone(result.Violations)) {
		key := promptFindingIdentity(result, violation)
		if !seen[key] {
			copy.Violations = append(copy.Violations, violation)
			seen[key] = true
		}
	}
	return &copy
}

func promptFindingIdentity(result *linters.LinterResult, violation models.Violation) promptFindingKey {
	key := promptFindingKey{Source: violation.Source, WorkDir: result.WorkDir, File: violation.File, Line: violation.Line}
	if key.Source == "" {
		key.Source = result.Linter
	}
	if violation.Rule != nil {
		key.RulePackage, key.Rule = violation.Rule.Package, violation.Rule.Method
	}
	if violation.Message != nil {
		key.Message = *violation.Message
	}
	return key
}

func (r *promptLintReporter) emit(result *linters.LinterResult) error {
	delta := *result
	delta.Violations = nil
	for _, violation := range result.Violations {
		key := promptFindingIdentity(result, violation)
		if !r.seen[key] {
			delta.Violations = append(delta.Violations, violation)
			r.seen[key] = true
		}
	}
	errorKey := result.Linter + "\x00" + result.WorkDir
	if result.Error == r.errors[errorKey] {
		delta.Error = ""
	}
	if len(delta.Violations) == 0 && delta.Error == "" {
		return nil
	}
	r.errors[errorKey] = result.Error
	if delta.InFlight() {
		delta.Success = true
	}
	if r.options.Options.Summary {
		return r.write(linters.NewSummaryView([]*linters.LinterResult{&delta}, r.options.Options.SummaryLimit))
	}
	return r.write(&delta)
}

func (r *promptLintReporter) write(value any) error {
	noColor := true
	if file, ok := r.options.Writer.(*os.File); ok {
		noColor = !term.IsTerminal(int(file.Fd()))
	}
	output, err := clicky.Formatter.FormatWithOptions(clicky.FormatOptions{Format: "pretty", NoColor: noColor || formatters.IsNoColor()}, value)
	if err != nil {
		return err
	}
	output = strings.TrimSuffix(output, "\n") + "\n"
	written, err := io.WriteString(r.options.Writer, output)
	if err == nil && written != len(output) {
		err = io.ErrShortWrite
	}
	return err
}

func (r *promptLintReporter) Finish(results []*linters.LinterResult, runErr error) (*testui.Snapshot, error) {
	for _, result := range results {
		r.UpsertLintResult(result)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if results != nil {
		r.snapshot.Lint = make([]*linters.LinterResult, len(results))
		for index, result := range results {
			r.snapshot.Lint[index] = r.filtered(result)
		}
	}
	r.finished = true
	r.snapshot.Status.Running = false
	r.snapshot.Status.Stopped = errors.Is(runErr, context.Canceled)
	r.snapshot.Metadata.Ended = time.Now().UTC()
	r.snapshot.Metadata.Changes = r.options.Options.Selection.Snapshot()
	if cause := errors.Join(runErr, r.err); cause != nil {
		r.snapshot.Error = cause.Error()
	}
	writeErr := snapshots.WriteSnapshot(r.path, &r.snapshot)
	if writeErr == nil && r.snapshot.Git != nil && r.snapshot.Git.SHA != "" {
		_, writeErr = snapshots.Save(r.options.Options.Dirs.StateRoot, &r.snapshot)
	}
	if r.err == nil && writeErr == nil {
		writeErr = r.write(r.snapshot.Pretty())
	}
	return &r.snapshot, errors.Join(r.err, writeErr)
}
