package prompt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/flanksource/captain/pkg/api"
	"path/filepath"
	"strings"
	"sync"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// Captain's async prompt-run seam, swapped out by specs.
var (
	launchPrompt   = captaincli.LaunchPrompt
	awaitPromptRun = captaincli.AwaitPromptRun
	stopPromptRun  = captaincli.StopPromptRun
)

type reviewsInput struct {
	Reviews     []review
	WorkDir     string
	File        string
	Stdin       []byte
	Concurrency int
	Publisher   Publisher
	OnEvent     func(int, api.Event)
}

// runReviews launches every review's Captain run into one clicky group, which
// runs at most Concurrency of them at once, then awaits them all. Every review
// is published as pending before the first launch; results keep review order.
func runReviews(ctx context.Context, input reviewsInput) []*linters.LinterResult {
	for _, review := range input.Reviews {
		input.publish(&linters.LinterResult{Linter: review.Name, WorkDir: input.WorkDir, Status: linters.LinterStatusPending, Violations: []models.Violation{}})
	}
	group := clicky.StartGroup[captaincli.PromptRunSummary](testui.PromptLintTaskGroupName, clickytask.WithConcurrency(max(input.Concurrency, 1)))
	awaits := make([]func() []*linters.LinterResult, len(input.Reviews))
	for i, review := range input.Reviews {
		awaits[i] = input.launch(ctx, review, group.Group)
	}
	perReview := make([][]*linters.LinterResult, len(input.Reviews))
	var wg sync.WaitGroup
	for i, await := range awaits {
		wg.Go(func() { perReview[i] = await() })
	}
	wg.Wait()
	group.WaitFor()
	var results []*linters.LinterResult
	for _, got := range perReview {
		results = append(results, got...)
	}
	return results
}

func (input reviewsInput) publish(result *linters.LinterResult) {
	if input.Publisher != nil {
		input.Publisher.UpsertLintResult(result)
	}
}

// launch starts one review's run as a task of group and returns the await for
// its results: the run's own, then the linters its agent ran. A multi-model
// prompt launches one run per model, each reported on its own.
func (input reviewsInput) launch(ctx context.Context, review review, group *clickytask.Group) func() []*linters.LinterResult {
	run := &reviewRun{WorkDir: input.WorkDir, Name: review.Name, Publisher: input.Publisher}
	start := time.Now()
	launched, err := launchPrompt(ctx, invocation(invocationInput{
		File: input.File, Prompt: review.Prompt, Stdin: stdinReader(input.Stdin), Tools: run.tools(),
		Group: group, Name: review.Name, OnEvent: input.OnEvent,
	}))
	if err != nil {
		result := run.result(summaryInput{Name: review.Name, Err: err, Elapsed: time.Since(start)})
		input.finish(result)
		return func() []*linters.LinterResult { return []*linters.LinterResult{result} }
	}
	run.setRun(launched.RunID, launched.SessionID)
	if len(launched.Runs) == 0 {
		running := func(violations []models.Violation) {
			input.publish(&linters.LinterResult{Linter: review.Name, WorkDir: input.WorkDir, Status: linters.LinterStatusRunning, RunID: launched.RunID, SessionID: launched.SessionID, Violations: violations})
		}
		running(run.Reported())
		run.setOnReport(running)
		return func() []*linters.LinterResult {
			result := input.await(ctx, run, review.Name, launched.RunID, start)
			return append([]*linters.LinterResult{result}, run.linted(result.SessionID)...)
		}
	}
	if input.Publisher != nil {
		input.Publisher.RemoveLintResult(review.Name, input.WorkDir)
	}
	for _, variant := range launched.Runs {
		input.publish(&linters.LinterResult{Linter: review.Name + ":" + variant.Selector, WorkDir: input.WorkDir, Status: linters.LinterStatusRunning, RunID: variant.RunID, SessionID: variant.SessionID, Violations: []models.Violation{}})
	}
	return func() []*linters.LinterResult {
		results := make([]*linters.LinterResult, len(launched.Runs))
		var wg sync.WaitGroup
		for i, variant := range launched.Runs {
			wg.Go(func() { results[i] = input.await(ctx, run, review.Name+":"+variant.Selector, variant.RunID, start) })
		}
		wg.Wait()
		return append(results, run.linted("")...)
	}
}

// finish publishes a run's final result and warns when it failed.
func (input reviewsInput) finish(result *linters.LinterResult) {
	if !result.Success {
		logger.Warnf("%s: %s", result.Linter, result.Error)
	}
	input.publish(result)
}

func (input reviewsInput) await(ctx context.Context, run *reviewRun, name, runID string, start time.Time) *linters.LinterResult {
	summary, err := awaitPromptRun(ctx, runID)
	run.setOnReport(nil)
	if ctx.Err() != nil {
		// Awaiting stops with the caller, the run does not: stop it too.
		if stopErr := stopPromptRun(runID); stopErr != nil && !errors.Is(stopErr, captaincli.ErrPromptRunTerminal) {
			err = errors.Join(err, stopErr)
		}
	}
	result := run.result(summaryInput{Name: name, RunID: runID, Summary: summary, Err: err, Elapsed: time.Since(start)})
	input.finish(result)
	return result
}

type summaryInput struct {
	Name, RunID string
	Summary     captaincli.PromptRunSummary
	Err         error
	Elapsed     time.Duration
}

// result is the run's lint result: its structured violations unioned with the
// ones reported through report_violation, which survive even a failed run.
func (r *reviewRun) result(input summaryInput) *linters.LinterResult {
	output := runOutput{Name: input.Name, WorkDir: r.WorkDir, Data: input.Summary.StructuredOutput, Duration: input.Summary.Duration, Elapsed: input.Elapsed}
	if input.Err != nil {
		output.Error = input.Err.Error()
	} else if !input.Summary.Success {
		output.Error = firstNonEmpty(input.Summary.Error, "prompt run did not succeed")
	}
	result := output.result()
	result.RunID = input.RunID
	result.SessionID = firstNonEmpty(input.Summary.SessionID, r.identity().SessionID)
	result.Violations = mergeViolations(result.Violations, r.Reported())
	if result.Violations == nil {
		result.Violations = []models.Violation{}
	}
	return result
}

type runOutput struct {
	Name, WorkDir, Duration, Error string
	Data                           map[string]any
	Elapsed                        time.Duration
}

func (output runOutput) result() *linters.LinterResult {
	result := &linters.LinterResult{Linter: output.Name, WorkDir: output.WorkDir, Duration: output.Elapsed, Error: output.Error}
	if output.Duration != "" {
		var err error
		result.Duration, err = time.ParseDuration(output.Duration)
		if err != nil {
			result.Error = fmt.Sprintf("invalid prompt duration %q: %v", output.Duration, err)
		}
	}
	if result.Error != "" {
		return result
	}
	violations, err := decodeViolations(output.Data, output.WorkDir)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Success, result.Violations = true, violations
	return result
}

func decodeViolations(data map[string]any, dir string) ([]models.Violation, error) {
	var output struct {
		Violations []models.Violation `json:"violations"`
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("encode prompt lint output: %w", err)
	}
	if err := json.Unmarshal(raw, &output); err != nil {
		return nil, fmt.Errorf("decode prompt lint output: %w", err)
	}
	if output.Violations == nil {
		return nil, fmt.Errorf("prompt lint output requires a violations array")
	}
	for i := range output.Violations {
		err := locateCode(dir, &output.Violations[i])
		if err == nil {
			err = validateViolation(&output.Violations[i], dir)
		}
		if err != nil {
			return nil, fmt.Errorf("prompt lint violation %d: %w", i+1, err)
		}
	}
	return output.Violations, nil
}

// validateViolation checks a violation the prompt reported, resolves its file
// against dir and defaults its source.
func validateViolation(v *models.Violation, dir string) error {
	if v.Message == nil || strings.TrimSpace(*v.Message) == "" {
		return fmt.Errorf("violation requires a nonempty message")
	}
	if v.Severity != "" && v.Severity != models.SeverityError && v.Severity != models.SeverityWarning && v.Severity != models.SeverityInfo {
		return fmt.Errorf("violation has invalid severity %q", v.Severity)
	}
	if v.Line < 0 || v.Column < 0 || v.EndLine < 0 || v.EndColumn < 0 {
		return fmt.Errorf("violation has a negative location")
	}
	if v.File != "" && !filepath.IsAbs(v.File) {
		v.File = filepath.Join(dir, v.File)
	}
	if v.Source == "" {
		v.Source = "prompt"
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
