package uir

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
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
	command := u.CommandContext(ctx, u.Name(), binary, args...)
	var stdout, stderr bytes.Buffer
	command.Stdout = u.WrapWriter(&stdout)
	command.Stderr = u.WrapWriter(&stderr)
	runErr := command.Run()
	exitCode := 0
	if runErr != nil {
		var exited *exec.ExitError
		if !errors.As(runErr, &exited) {
			return nil, fmt.Errorf("uir execution: %w: %s", runErr, stderr.String())
		}
		exitCode = exited.ExitCode()
	}
	found, err := u.parse(stdout.Bytes(), exitCode)
	if err != nil {
		return found, fmt.Errorf("uir: %w; stderr: %s", err, stderr.String())
	}
	return found, nil
}

func (u *UIR) parse(output []byte, exitCode int) ([]models.Violation, error) {
	var results []*linters.LinterResult
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("decode lint report: %w", err)
	}
	if len(results) != 1 || results[0] == nil || results[0].Linter != "uir" {
		return nil, errors.New("expected one uir linter result")
	}
	result := results[0]
	u.fileCount, u.ruleCount = result.FileCount, result.RuleCount
	if exitCode != 0 && exitCode != 1 {
		return result.Violations, fmt.Errorf("unexpected exit status %d", exitCode)
	}
	if result.Skipped || result.TimedOut || !result.Success || result.Error != "" {
		return result.Violations, fmt.Errorf("lint execution failed: %s", result.Error)
	}
	if (exitCode == 1) != linters.HasFailures(results) {
		return result.Violations, errors.New("lint report disagrees with exit status")
	}
	return result.Violations, nil
}
