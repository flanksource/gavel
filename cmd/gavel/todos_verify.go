package main

import (
	"context"
	"fmt"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var (
	verifyCmdModel     string
	verifyCmdThreshold int
	verifyCmdStrict    bool
)

var todosVerifyCmd = &cobra.Command{
	Use:          "verify [ids-or-files...]",
	SilenceUsage: true,
	Short:        "AI-verify whether a TODO's commits implement its acceptance criteria",
	Long: `AI code review for TODOs: score whether the committed work actually implements
each TODO's acceptance criteria. This is gavel's AI-review surface (it replaces
the former top-level 'gavel verify' command).

For each TODO, an AI model reviews the diff of its commits against its task-list
criteria and returns a score plus an implemented/not-implemented verdict. A TODO
is promoted to "verified" when it is implemented and scores at/above --threshold
(default 80). --strict exits non-zero if any reviewed TODO is not implemented,
which makes it usable as a CI/quality gate. The model defaults to .gavel.yaml
verify.model; override with --model.

With no arguments it verifies every TODO matching --status; pass ids/paths to
select a subset. The same scoring engine is also available as AI-verification
fixtures (see 'gavel fixtures --help', FORMAT 3).

Examples:
  gavel todos verify                       # score all todos
  gavel todos verify 3f2a1b                 # score one todo
  gavel todos verify --threshold 90         # require a higher score
  gavel todos verify --strict               # non-zero if any unimplemented
  gavel todos verify --model claude-code-sonnet`,
	RunE: runTodosVerify,
}

func init() {
	todosCmd.AddCommand(todosVerifyCmd)
	todosVerifyCmd.Flags().StringVar(&todosDir, "dir", "", "Deprecated; runtime TODOs are stored in PostgreSQL (must be omitted)")
	todosVerifyCmd.Flags().StringVar(&filterStatus, "status", "", "Filter TODOs by status")
	todosVerifyCmd.Flags().StringVar(&verifyCmdModel, "model", "", "Model for verification (default: .gavel.yaml verify.model)")
	todosVerifyCmd.Flags().IntVar(&verifyCmdThreshold, "threshold", 0, "Score at/above which (with implemented) a TODO is promoted to verified (default 80)")
	todosVerifyCmd.Flags().BoolVar(&verifyCmdStrict, "strict", false, "Exit non-zero when any verified TODO is not implemented")
}

func runTodosVerify(_ *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir, todosDir)
	if err != nil {
		return err
	}

	ctx := context.Background()
	todoList, err := resolveVerifyTODOs(ctx, provider, workDir, args)
	if err != nil {
		return err
	}
	if len(todoList) == 0 {
		logger.Infof("No TODOs found")
		return nil
	}

	notImplemented := 0
	for _, todo := range todoList {
		result, err := todos.RunIssueVerification(ctx, provider, todo, todos.VerifyOptions{
			WorkDir:   todoWorkDir(workDir, todo),
			Model:     verifyCmdModel,
			Threshold: verifyCmdThreshold,
		})
		if err != nil {
			logger.Warnf("%s: %v", todo.Filename(), err)
			continue
		}
		fmt.Println(result.Pretty().ANSI())
		if result.Implemented != nil && !*result.Implemented {
			notImplemented++
		}
	}

	if verifyCmdStrict && notImplemented > 0 {
		return fmt.Errorf("%d TODO(s) not implemented", notImplemented)
	}
	return nil
}

// resolveVerifyTODOs loads the TODOs to verify: explicit refs/args or, when none
// are given, every TODO matching the status filter.
func resolveVerifyTODOs(ctx context.Context, provider todos.Provider, _ string, args []string) ([]*types.TODO, error) {
	filters := todos.DiscoveryFilters{}
	if filterStatus != "" {
		filters.IncludeStatuses = []types.Status{types.Status(filterStatus)}
	}
	todoList, err := resolveRequestedTODOs(ctx, provider, args, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to discover TODOs: %w", err)
	}
	return todoList, nil
}
