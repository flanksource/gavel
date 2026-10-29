// Package command runs an external linter a .gavel.yaml declares under
// lint.linters.<name>.command, speaking gavel's report protocol.
package command

import (
	"errors"

	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
)

// Linter runs the declared command with --format json and the files to lint
// appended, from the work dir, and reads its report (linters.ParseReport).
// Declaring it activates it; enabled: false turns it off.
type Linter struct {
	linters.RunOptions
	name                 string
	command, includes    []string
	fileCount, ruleCount int
}

// New is the linter lint.linters.<name> declares; its Command must not be empty.
func New(workDir, name string, cfg verify.LintLinterConfig) *Linter {
	return &Linter{RunOptions: linters.RunOptions{WorkDir: workDir}, name: name, command: cfg.Command, includes: cfg.Includes}
}

// Declared lists the linters cfg declares by command, by name.
func Declared(workDir string, cfg verify.LintConfig) []*Linter {
	var out []*Linter
	for name, linter := range cfg.Linters {
		if len(linter.Command) > 0 {
			out = append(out, New(workDir, name, linter))
		}
	}
	return out
}

func (l *Linter) Name() string                          { return l.name }
func (l *Linter) SetOptions(options linters.RunOptions) { l.RunOptions = options }
func (l *Linter) RequiresExecutable() bool              { return false }
func (l *Linter) HasDefaultActivation(string) bool      { return true }
func (l *Linter) DefaultIncludes() []string {
	if len(l.includes) == 0 {
		return []string{"**/*"}
	}
	return l.includes
}
func (l *Linter) DefaultExcludes() []string { return nil }
func (l *Linter) SupportsJSON() bool        { return true }
func (l *Linter) JSONArgs() []string        { return []string{"--format", "json"} }
func (l *Linter) SupportsFix() bool         { return false }
func (l *Linter) FixArgs() []string         { return nil }
func (l *Linter) GetFileCount() int         { return l.fileCount }
func (l *Linter) GetRuleCount() int         { return l.ruleCount }
func (l *Linter) ValidateConfig(config *models.LinterConfig) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}
	return nil
}

func (l *Linter) DryRunCommand() (string, []string) {
	args := append(append(append([]string{}, l.command[1:]...), l.ExtraArgs...), l.JSONArgs()...)
	return l.command[0], append(args, l.Files...)
}

func (l *Linter) Run(ctx commonsContext.Context, _ *clicky.Task) ([]models.Violation, error) {
	binary, args := l.DryRunCommand()
	result, err := l.RunReport(ctx, l.name, binary, args...)
	if result == nil {
		return nil, err
	}
	l.fileCount, l.ruleCount = result.FileCount, result.RuleCount
	return result.Violations, err
}
