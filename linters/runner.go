package linters

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/task"
	commonsCtx "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/cache"
	"github.com/flanksource/gavel/models"
)

// Runner orchestrates execution of multiple linters with intelligent debouncing
type Runner struct {
	registry       *Registry
	violationCache *cache.ViolationCache
	linterStats    *cache.LinterStats
	config         *models.Config
	workDir        string
	noCache        bool
}

// RunnerOptions configures the runner behavior
type RunnerOptions struct {
	NoCache bool // Disable caching
}

// NewRunner creates a new linter runner with intelligent debouncing
func NewRunner(config *models.Config, workDir string) (*Runner, error) {
	return NewRunnerWithOptions(config, workDir, RunnerOptions{})
}

// NewRunnerWithOptions creates a new linter runner with custom options
func NewRunnerWithOptions(config *models.Config, workDir string, opts RunnerOptions) (*Runner, error) {
	var violationCache *cache.ViolationCache
	var linterStats *cache.LinterStats
	var err error

	if !opts.NoCache {
		violationCache, err = cache.NewViolationCache()
		if err != nil {
			return nil, fmt.Errorf("failed to create violation cache: %w", err)
		}

		linterStats, err = cache.NewLinterStats()
		if err != nil {
			return nil, fmt.Errorf("failed to create linter stats: %w", err)
		}
	}

	return &Runner{
		registry:       DefaultRegistry,
		violationCache: violationCache,
		linterStats:    linterStats,
		config:         config,
		workDir:        workDir,
		noCache:        opts.NoCache,
	}, nil
}

// Close closes any resources held by the runner
func (r *Runner) Close() error {
	var errs []error

	if r.violationCache != nil {
		if err := r.violationCache.Close(); err != nil {
			errs = append(errs, fmt.Errorf("violation cache close error: %w", err))
		}
	}

	if r.linterStats != nil {
		if err := r.linterStats.Close(); err != nil {
			errs = append(errs, fmt.Errorf("linter stats close error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	return nil
}

// RunEnabledLinters runs all enabled linters with intelligent debouncing
func (r *Runner) RunEnabledLinters() ([]LinterResult, error) {
	return r.RunEnabledLintersOnFiles(nil, false)
}

// RunEnabledLintersOnFiles runs enabled linters on specific files
func (r *Runner) RunEnabledLintersOnFiles(specificFiles []string, fix bool) ([]LinterResult, error) {
	var results []LinterResult

	enabledLinters := r.config.GetEnabledLinters()
	logger.Infof("Running %d enabled linters: %v", len(enabledLinters), enabledLinters)

	ctx := context.Background()

	for _, linterName := range enabledLinters {
		result, err := r.RunWithIntelligentDebounce(ctx, linterName, specificFiles, fix)
		if err != nil {
			logger.Warnf("Failed to run linter %s: %v", linterName, err)
			results = append(results, LinterResult{
				Linter:  linterName,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}
		logger.Infof(result.Pretty().ANSI())

		results = append(results, *result)
	}

	return results, nil
}

// RunWithIntelligentDebounce executes a linter with intelligent debouncing
func (r *Runner) RunWithIntelligentDebounce(ctx context.Context, linterName string, files []string, fix bool) (*LinterResult, error) {

	// Get linter from registry
	linter, ok := r.registry.Get(linterName)
	if !ok {
		return nil, fmt.Errorf("unknown linter: %s", linterName)
	}

	// Get configuration
	config := r.config.GetLinterConfig(linterName, r.workDir)

	// Apply per-linter timeout
	timeout := config.GetTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Check intelligent debounce (only if cache is enabled)
	if r.linterStats != nil {
		shouldSkip, actualDebounce, err := r.linterStats.ShouldSkipLinter(linterName, r.workDir, config.Debounce)
		if err != nil {
			logger.Warnf("Failed to check debounce for %s: %v", linterName, err)
		} else if shouldSkip {
			logger.Debugf("Skipping %s due to intelligent debounce (%v)", linterName, actualDebounce)
			// Load cached violations and return
			return r.loadCachedResult(linterName, actualDebounce)
		}
	}

	opts := RunOptions{
		WorkDir:    r.workDir,
		Files:      files,
		Config:     config,
		ArchConfig: r.config, // Pass full config for all_language_excludes macro
		ForceJSON:  config.OutputFormat == "json" || config.OutputFormat == "",
		Fix:        fix,
		NoCache:    r.noCache,
	}

	if mixin, ok := linter.(OptionsMixin); ok {
		mixin.SetOptions(opts)
	}

	task := clicky.StartTask[[]models.Violation](fmt.Sprintf("Running %s", linterName), func(fCtx commonsCtx.Context, t *task.Task) ([]models.Violation, error) {
		return linter.Run(fCtx, t)
	})
	violations, err := task.GetResult()

	// Record execution stats
	if r.linterStats != nil {
		if statsErr := r.linterStats.RecordExecution(linterName, r.workDir, task.Duration(), len(violations), task.Error() == nil); statsErr != nil {
			logger.Warnf("Failed to record execution stats for %s: %v", linterName, statsErr)
		}
	}

	// Cache violations if successful
	if task.Error() == nil && len(violations) > 0 && r.violationCache != nil {
		r.cacheViolations(linterName, violations)
	}

	r.updateTaskStatus(task.Task, linterName, task.IsOk(), violations, task.Error())

	timedOut := ctx.Err() == context.DeadlineExceeded
	result := &LinterResult{
		Linter:     linterName,
		WorkDir:    r.workDir,
		Success:    task.IsOk() && !timedOut,
		TimedOut:   timedOut,
		Duration:   task.Duration(),
		Violations: violations,
		Error:      r.formatError(err),
	}

	// If the linter provides metadata, include it in the result
	if metadata, ok := linter.(MetadataProvider); ok {
		result.FileCount = metadata.GetFileCount()
		result.RuleCount = metadata.GetRuleCount()
	}

	return result, nil
}

// loadCachedResult loads cached violations for debounced linters
func (r *Runner) loadCachedResult(linterName string, debounce time.Duration) (*LinterResult, error) {

	task := clicky.StartTask[*LinterResult](fmt.Sprintf("Skipping %s (debounced for %v)", linterName, debounce), func(ctx commonsCtx.Context, t *clicky.Task) (
		*LinterResult, error) {
		var violations []models.Violation
		if r.violationCache != nil {
			cachedViolations, err := r.violationCache.GetViolationsBySource(linterName)
			if err != nil {
				logger.Warnf("Failed to load cached violations for %s: %v", linterName, err)
			} else {
				violations = cachedViolations
				logger.Debugf("Loaded %d cached violations for %s", len(violations), linterName)
			}
		}

		result := &LinterResult{
			Linter:       linterName,
			WorkDir:      r.workDir,
			Success:      true,
			Duration:     t.Duration(),
			Violations:   violations,
			Debounced:    true,
			DebounceUsed: debounce,
		}

		// For cached results, we can't provide file/rule counts since we didn't actually analyze
		// TODO: Consider caching metadata along with violations

		return result, nil

	})

	return task.GetResult()
}

// cacheViolations stores violations in the cache
func (r *Runner) cacheViolations(linterName string, violations []models.Violation) {
	if r.violationCache == nil {
		return
	}

	// Group violations by file
	fileViolations := make(map[string][]models.Violation)
	for _, v := range violations {
		fileViolations[v.File] = append(fileViolations[v.File], v)
	}

	// Store each file's violations
	for file, vList := range fileViolations {
		if err := r.violationCache.StoreViolations(file, vList); err != nil {
			logger.Debugf("Failed to cache linter violations for %s: %v", file, err)
		}
	}
}

// updateTaskStatus updates the task manager status. The task name is the
// sole output line for this linter in the CI step log, so it carries
// status and (on problems) a short reason snippet within the 80-char
// budget. Errors are NOT passed to task.Errorf — that writes buffered
// child lines that render as indented sub-lines under the task, which
// the step-log compact format does not want.
func ApplyLinterTaskStatus(task *clicky.Task, linterName string, success bool, violations []models.Violation, err error) {
	if success {
		if len(violations) > 0 {
			task.SetName(buildLinterLabel(linterName, len(violations), true, firstViolationSnippet(violations), ""))
			task.Warning()
		} else {
			task.SetName(linterName)
			task.Success()
		}
	} else {
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		task.SetName(buildLinterLabel(linterName, len(violations), false, firstViolationSnippet(violations), errMsg))
		task.Failed()
	}
}

func (r *Runner) updateTaskStatus(task *clicky.Task, linterName string, success bool, violations []models.Violation, err error) {
	ApplyLinterTaskStatus(task, linterName, success, violations, err)
}

// firstViolationSnippet returns a compact "file:line rule: message" for
// the first violation in the slice, or "" if there are none.
func firstViolationSnippet(violations []models.Violation) string {
	if len(violations) == 0 {
		return ""
	}
	v := violations[0]
	loc := v.File
	if v.Line > 0 {
		loc = fmt.Sprintf("%s:%d", v.File, v.Line)
	}
	rule := v.Source
	msg := ""
	if v.Message != nil {
		msg = *v.Message
	}
	switch {
	case loc != "" && rule != "" && msg != "":
		return fmt.Sprintf("%s %s: %s", loc, rule, msg)
	case loc != "" && msg != "":
		return fmt.Sprintf("%s: %s", loc, msg)
	case loc != "":
		return loc
	case msg != "":
		return msg
	default:
		return ""
	}
}

// buildLinterLabel assembles the compact one-line task name for a linter
// result. Total length is kept within labelBudget characters; the
// trailing details portion is right-truncated with an ellipsis.
func buildLinterLabel(linterName string, violationCount int, success bool, firstViolationLine, errorMessage string) string {
	var prefix, details string
	switch {
	case !success:
		prefix = fmt.Sprintf("%s (failed)", linterName)
		if errorMessage != "" {
			details = "err: " + firstNonBlankLineLinter(errorMessage)
		} else if firstViolationLine != "" {
			details = firstViolationLine
		}
	case violationCount > 0:
		prefix = fmt.Sprintf("%s (%d)", linterName, violationCount)
		details = firstViolationLine
	default:
		prefix = linterName
	}
	return joinLabelLinter(prefix, details, labelBudget)
}

const labelBudget = 68

// joinLabelLinter joins prefix + details with two spaces, right-
// truncating details with `…` so the total length stays within budget.
// If the prefix alone exceeds the budget, it is right-truncated too.
func joinLabelLinter(prefix, details string, budget int) string {
	prefix = stripANSILinter(prefix)
	details = stripANSILinter(strings.TrimSpace(details))
	if runeLenLinter(prefix) > budget {
		return truncateRunesLinter(prefix, budget)
	}
	if details == "" {
		return prefix
	}
	const sep = "  "
	remaining := budget - runeLenLinter(prefix) - len(sep)
	if remaining <= 1 {
		return prefix
	}
	return prefix + sep + truncateRunesLinter(details, remaining)
}

var ansiReLinter = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSILinter(s string) string {
	return ansiReLinter.ReplaceAllString(s, "")
}

func firstNonBlankLineLinter(s string) string {
	s = stripANSILinter(s)
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

func truncateRunesLinter(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if runeLenLinter(s) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	runes := []rune(s)
	return string(runes[:n-1]) + "…"
}

func runeLenLinter(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

// hasArg checks if args contains a specific argument or argument prefix
// hasPathArg checks if the args already contain a path argument
// formatError formats an error for display
func (r *Runner) formatError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// GetIntelligentDebounceForLinter returns the recommended debounce for a linter
func (r *Runner) GetIntelligentDebounceForLinter(linterName string) (time.Duration, error) {
	if r.linterStats == nil {
		return 30 * time.Second, nil
	}
	return r.linterStats.GetIntelligentDebounce(linterName, r.workDir)
}

// GetStatsForLinter returns execution statistics for a linter
func (r *Runner) GetStatsForLinter(linterName string) (*cache.ExecutionStats, error) {
	if r.linterStats == nil {
		return nil, fmt.Errorf("linter stats not available")
	}
	return r.linterStats.GetStats(linterName, r.workDir)
}
