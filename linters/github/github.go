package githublint

import (
	"errors"
	"fmt"
	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

type GitHub struct {
	linters.RunOptions
	AuditOptions github.AuditOptions
}

func New(workDir string) *GitHub                            { return &GitHub{RunOptions: linters.RunOptions{WorkDir: workDir}} }
func (l *GitHub) Name() string                              { return "github" }
func (l *GitHub) InProcess() bool                           { return true }
func (l *GitHub) RequiresExecutable() bool                  { return false }
func (l *GitHub) DefaultIncludes() []string                 { return nil }
func (l *GitHub) DefaultExcludes() []string                 { return nil }
func (l *GitHub) SupportsJSON() bool                        { return true }
func (l *GitHub) JSONArgs() []string                        { return nil }
func (l *GitHub) SupportsFix() bool                         { return false }
func (l *GitHub) FixArgs() []string                         { return nil }
func (l *GitHub) SetOptions(opts linters.RunOptions)        { l.RunOptions = opts }
func (l *GitHub) ValidateConfig(*models.LinterConfig) error { return nil }
func (l *GitHub) DryRunCommand() (string, []string) {
	return "gavel", append([]string{"lint", "github", "--"}, l.ExtraArgs...)
}
func (l *GitHub) Run(ctx commonsContext.Context, t *clicky.Task) ([]models.Violation, error) {
	result, err := l.RunResult(ctx, t)
	return result.Violations, err
}
func (l *GitHub) RunResult(ctx commonsContext.Context, _ *clicky.Task) (*linters.LinterResult, error) {
	result := &linters.LinterResult{}
	if l.Fix {
		return result, fmt.Errorf("github auditing cannot mutate GitHub; --fix is unsupported")
	}
	if len(l.Files) > 0 {
		return result, fmt.Errorf("github audits a repository; use --cwd instead of file scopes")
	}
	opts, selection, err := l.configuration()
	if err != nil {
		return result, err
	}
	if opts.PR != nil {
		result.Violations, err = github.Audit(ctx, opts)
		return result, err
	}
	repos, selectionErr := opts.Client.AuditRepositories(ctx, selection)
	var failures []error
	if selectionErr != nil {
		failures = append(failures, selectionErr)
	}
	for _, repo := range repos {
		if err := ctx.Err(); err != nil {
			failures = append(failures, err)
			break
		}
		current := opts
		current.Client.Repo = repo
		found, err := github.Audit(ctx, current)
		for i := range found {
			found[i].Message = models.StringPtr(repo + ": " + *found[i].Message)
		}
		result.Violations = append(result.Violations, found...)
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", repo, err))
		}
	}
	if len(repos) == 0 && selectionErr == nil {
		result.Skipped = true
		result.Error = "no repositories remain after selection and exclusions"
	}
	result.Sources = map[string]linters.SourceFile{}
	for _, v := range result.Violations {
		if v.File != "" {
			result.Sources[v.File] = linters.SourceFile{}
		}
	}
	return result, errors.Join(failures...)
}
