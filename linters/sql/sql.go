// Package sql adapts UIR migration rules to Gavel's lint runner.
package sql

import (
	"flag"
	"fmt"
	"io"

	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	uirlint "github.com/flanksource/uir/lint"
)

type Linter struct {
	linters.RunOptions
	Stdin io.Reader
}

func New(workDir string) *Linter                          { return &Linter{RunOptions: linters.RunOptions{WorkDir: workDir}} }
func (*Linter) Name() string                              { return "sql" }
func (*Linter) InProcess() bool                           { return true }
func (*Linter) RequiresExecutable() bool                  { return false }
func (*Linter) DefaultIncludes() []string                 { return []string{"**/*.sql", "**/*.xml"} }
func (*Linter) DefaultExcludes() []string                 { return nil }
func (*Linter) SupportsJSON() bool                        { return true }
func (*Linter) JSONArgs() []string                        { return nil }
func (*Linter) SupportsFix() bool                         { return false }
func (*Linter) FixArgs() []string                         { return nil }
func (l *Linter) SetOptions(options linters.RunOptions)   { l.RunOptions = options }
func (*Linter) ValidateConfig(*models.LinterConfig) error { return nil }

func (l *Linter) DryRunCommand() (string, []string) {
	args := append([]string{"lint", "sql"}, l.Files...)
	if len(l.ExtraArgs) > 0 {
		args = append(append(args, "--"), l.ExtraArgs...)
	}
	return "gavel", args
}

func (l *Linter) Run(ctx commonsContext.Context, task *clicky.Task) ([]models.Violation, error) {
	result, err := l.RunResult(ctx, task)
	if result == nil {
		return nil, err
	}
	return result.Violations, err
}

func (l *Linter) RunResult(ctx commonsContext.Context, _ *clicky.Task) (*linters.LinterResult, error) {
	options := uirlint.SQLCommandOptions{WorkDir: l.WorkDir, Files: l.Files, Stdin: l.Stdin}
	flags := flag.NewFlagSet(l.Name(), flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.Dialect, "dialect", "", "SQL dialect")
	flags.IntVar(&options.CompatibilityLevel, "compatibility-level", 0, "SQL Server compatibility level")
	if err := flags.Parse(l.ExtraArgs); err != nil {
		return nil, err
	}
	if flags.NArg() != 0 {
		return nil, fmt.Errorf("unexpected SQL lint arguments %v", flags.Args())
	}
	return uirlint.RunSQL(ctx, options)
}
