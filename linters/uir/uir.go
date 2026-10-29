package uir

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

type UIR struct {
	linters.RunOptions
	fileCount, ruleCount int
}

func New(workDir string) *UIR                           { return &UIR{RunOptions: linters.RunOptions{WorkDir: workDir}} }
func (u *UIR) Name() string                             { return "uir" }
func (u *UIR) IndependentProjects() bool                { return true }
func (u *UIR) NewProject(workDir string) linters.Linter { return New(workDir) }
func (u *UIR) SetOptions(options linters.RunOptions)    { u.RunOptions = options }
func (u *UIR) ProjectRootMarkers() []string             { return []string{".uir.yaml", ".uir.d/"} }
func (u *UIR) RequiresExecutable() bool                 { return true }
func (u *UIR) DefaultIncludes() []string {
	return []string{".uir.yaml", "**/.uir.d/*.yaml", "**/.uir.d/*.yml", "**/*.go", "**/go.mod", "**/go.sum", "**/go.work"}
}
func (u *UIR) DefaultExcludes() []string { return []string{"vendor/**", "node_modules/**"} }
func (u *UIR) SupportsJSON() bool        { return true }
func (u *UIR) JSONArgs() []string        { return []string{"--format", "json"} }
func (u *UIR) SupportsFix() bool         { return false }
func (u *UIR) FixArgs() []string         { return nil }
func (u *UIR) TagsArgs(tags []string) []string {
	return []string{"--tags=" + strings.Join(tags, ",")}
}
func (u *UIR) ValidateConfig(config *models.LinterConfig) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}
	return nil
}
func (u *UIR) GetFileCount() int { return u.fileCount }
func (u *UIR) GetRuleCount() int { return u.ruleCount }

func (u *UIR) DryRunCommand() (string, []string) {
	binary := u.Executable
	if binary == "" {
		binary = "uir"
	}
	args := []string{"lint"}
	if u.Config != nil {
		args = append(args, u.Config.Args...)
	}
	args = append(args, u.ExtraArgs...)
	args = append(args, "--format", "json", "--no-progress")
	return binary, args
}

func (u *UIR) Run(ctx commonsContext.Context, _ *clicky.Task) (violations []models.Violation, err error) {
	fragment, err := u.writeFragment(ctx)
	if err != nil {
		return nil, err
	}
	if fragment != "" {
		defer func() { err = errors.Join(err, removeFragment(fragment)) }()
	}
	binary, args := u.DryRunCommand()
	result, err := u.RunReport(ctx, u.Name(), binary, args...)
	return u.found(result, err)
}

func (u *UIR) parse(output []byte, exitCode int) ([]models.Violation, error) {
	return u.found(linters.ParseReport(output, exitCode))
}

// found keeps a uir report's findings and counts, refusing another linter's.
func (u *UIR) found(result *linters.LinterResult, err error) ([]models.Violation, error) {
	if result == nil {
		return nil, err
	}
	if result.Linter != "uir" {
		return nil, fmt.Errorf("expected one uir linter result, got %q", result.Linter)
	}
	u.fileCount, u.ruleCount = result.FileCount, result.RuleCount
	return result.Violations, err
}
