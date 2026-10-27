package prompt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	clickytask "github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
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

// runReviews runs each review in a clicky group, at most Concurrency at once.
// Every review is published as pending before the group starts.
func runReviews(ctx context.Context, input reviewsInput) []*linters.LinterResult {
	for _, review := range input.Reviews {
		input.publish(&linters.LinterResult{Linter: review.Name, WorkDir: input.WorkDir, Status: linters.LinterStatusPending, Violations: []models.Violation{}})
	}
	group := clicky.StartGroup[[]*linters.LinterResult]("lint prompt", clickytask.WithConcurrency(max(input.Concurrency, 1)))
	tasks := make([]clickytask.TypedTask[[]*linters.LinterResult], 0, len(input.Reviews))
	for _, review := range input.Reviews {
		tasks = append(tasks, group.Add(review.Name, func(_ flanksourceContext.Context, t *clickytask.Task) ([]*linters.LinterResult, error) {
			results := input.run(ctx, review)
			for _, result := range results {
				if !result.Success {
					t.Warnf("%s: %s", result.Linter, result.Error)
				}
			}
			return results, nil
		}))
	}
	group.WaitFor()
	var results []*linters.LinterResult
	for i, task := range tasks {
		got, err := task.GetResult()
		if err != nil {
			got = []*linters.LinterResult{{Linter: input.Reviews[i].Name, WorkDir: input.WorkDir, Error: err.Error()}}
		}
		results = append(results, got...)
	}
	return results
}

func (input reviewsInput) publish(result *linters.LinterResult) {
	if input.Publisher != nil {
		input.Publisher.UpsertLintResult(result)
	}
}

// run launches one review and awaits it. A multi-model prompt launches one run
// per model, each reported on its own.
func (input reviewsInput) run(ctx context.Context, review review) []*linters.LinterResult {
	run := &reviewRun{WorkDir: input.WorkDir}
	start := time.Now()
	inv := invocation(input.File, review.Prompt, stdinReader(input.Stdin), run.tools())
	inv.OnEvent = input.OnEvent
	launched, err := launchPrompt(ctx, inv)
	if err != nil {
		result := run.result(summaryInput{Name: review.Name, Err: err, Elapsed: time.Since(start)})
		input.publish(result)
		return []*linters.LinterResult{result}
	}
	if len(launched.Runs) == 0 {
		running := func(violations []models.Violation) {
			input.publish(&linters.LinterResult{Linter: review.Name, WorkDir: input.WorkDir, Status: linters.LinterStatusRunning, RunID: launched.RunID, Violations: violations})
		}
		running(run.Reported())
		run.setOnReport(running)
		return []*linters.LinterResult{input.await(ctx, run, review.Name, launched.RunID, start)}
	}
	if input.Publisher != nil {
		input.Publisher.RemoveLintResult(review.Name, input.WorkDir)
	}
	results := make([]*linters.LinterResult, len(launched.Runs))
	done := make(chan struct{})
	for i, variant := range launched.Runs {
		input.publish(&linters.LinterResult{Linter: review.Name + ":" + variant.Selector, WorkDir: input.WorkDir, Status: linters.LinterStatusRunning, RunID: variant.RunID, Violations: []models.Violation{}})
		go func() {
			defer func() { done <- struct{}{} }()
			results[i] = input.await(ctx, run, review.Name+":"+variant.Selector, variant.RunID, start)
		}()
	}
	for range launched.Runs {
		<-done
	}
	return results
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
	input.publish(result)
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
