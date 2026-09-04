package main

import (
	"context"
	"fmt"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/githubpush"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/cobra"
)

var (
	todoPushBaseURL string
	todoPushRepo    string
	todoPushForce   bool
	todoPushUpdate  bool
	todoPushIssue   string
	todoPushLabels  bool
	todoPushPlan    bool
)

var todosPushCmd = &cobra.Command{
	Use:          "push <ref...>",
	SilenceUsage: true,
	Short:        "Push TODOs to GitHub issues",
	Long: `Open a GitHub issue for each TODO — or rewrite the one it is linked to — and
record the link on it.

The issue body is the TODO's portable Markdown without the YAML front matter:
its body, then the current plan under '# Plan' (--plan=false to leave it out),
then the fixture under '# Verification'. The link is stored as a
'owner/repo#number' alias, so the TODO afterwards resolves by that reference
too.

A TODO that already carries a GitHub link is refused: --update rewrites that
issue's title and body, and --force opens a second issue instead. --issue
targets one specific issue, including one gavel did not open, and links it.

This is one-way. Nothing is read back from GitHub, so an --update overwrites
whatever the issue body says now.

Attachments are stored as links relative to the gavel dashboard, which GitHub
cannot fetch. A TODO whose body references one therefore needs an absolute
origin from --base-url or todos.baseUrl in .gavel.yaml.`,
	Example: `  gavel todos push 3f2a1b
  gavel todos push 3f2a1b --repo flanksource/gavel
  gavel todos push 3f2a1b --update
  gavel todos push 3f2a1b --issue flanksource/gavel#412
  gavel todos push 3f2a1b --base-url https://gavel.example.com`,
	Args: cobra.MinimumNArgs(1),
	RunE: runTodosPush,
}

func init() {
	todosCmd.AddCommand(todosPushCmd)
	todosPushCmd.Flags().StringVar(&todoPushBaseURL, "base-url", "",
		"Absolute origin attachment links resolve against (default: .gavel.yaml todos.baseUrl)")
	todosPushCmd.Flags().StringVar(&todoPushRepo, "repo", "",
		"Target owner/repo (default: the workspace's origin remote)")
	todosPushCmd.Flags().BoolVar(&todoPushForce, "force", false,
		"Open a second issue for a TODO that is already linked")
	todosPushCmd.Flags().BoolVar(&todoPushUpdate, "update", false,
		"Rewrite the linked issue instead of refusing")
	todosPushCmd.Flags().StringVar(&todoPushIssue, "issue", "",
		"Rewrite this issue and link it: 123, owner/repo#123, or its URL")
	todosPushCmd.Flags().BoolVar(&todoPushLabels, "labels", true,
		"Copy the TODO's labels onto the issue (use --labels=false to skip)")
	todosPushCmd.Flags().BoolVar(&todoPushPlan, "plan", true,
		"Include the TODO's plan in the issue body (use --plan=false to skip)")
}

func runTodosPush(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	todoList, err := resolveRequestedTODOs(ctx, provider, args, todos.DiscoveryFilters{})
	if err != nil {
		return err
	}
	if len(todoList) == 0 {
		return fmt.Errorf("no TODOs matched %v", args)
	}

	baseURL, err := resolveTodoPushBaseURL(workDir)
	if err != nil {
		return err
	}

	if todoPushIssue != "" && len(todoList) > 1 {
		return fmt.Errorf("--issue names a single issue but %d TODOs matched %v", len(todoList), args)
	}

	opts := githubpush.Options{
		GitHub:  github.Options{WorkDir: workDir, Repo: todoPushRepo},
		BaseURL: baseURL,
		Force:   todoPushForce,
		Update:  todoPushUpdate,
		Issue:   todoPushIssue,
		Labels:  todoPushLabels,
		Plan:    todoPushPlan,
	}
	for _, todo := range todoList {
		result, err := githubpush.Push(ctx, provider, todo.ID, opts)
		if err != nil {
			return err
		}
		verb := "opened"
		if result.Updated {
			verb = "updated"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s → %s (%s)\n", todo.DisplayID(), result.URL, verb)
	}
	return nil
}

// resolveTodoPushBaseURL prefers the flag over the project config. Resolving it
// before any TODO is pushed keeps a malformed origin from opening some issues
// and then failing.
func resolveTodoPushBaseURL(workDir string) (string, error) {
	config, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return "", fmt.Errorf("load .gavel.yaml: %w", err)
	}
	baseURL, err := githubpush.ResolveBaseURL(todoPushBaseURL, config.Todos.BaseURL)
	if err != nil {
		return "", err
	}
	if githubpush.IsLoopback(baseURL) {
		logger.Warnf("base URL %s is loopback: attachment images will only render for viewers on this machine", baseURL)
	}
	return baseURL, nil
}
