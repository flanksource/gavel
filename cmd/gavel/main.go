package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/formatters"
	"github.com/flanksource/clicky/mcp"
	"github.com/flanksource/clicky/shutdown"
	"github.com/flanksource/commons/logger"
	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	commit     = "unknown"
	date       = "unknown"
	workingDir string
	exitCode   int
)

var rootCmd = &cobra.Command{
	Use:   "gavel",
	Short: "Gavel CLI",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		clicky.Flags.UseFlags()
	},
}

func getWorkingDir() (string, error) {
	if workingDir != "" {
		if workingDir == "~" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get home directory: %w", err)
			}
			return home, nil
		}
		absPath, err := filepath.Abs(workingDir)
		if err != nil {
			return "", fmt.Errorf("failed to resolve working directory: %w", err)
		}
		info, err := os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("working directory does not exist: %w", err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("working directory is not a directory: %s", absPath)
		}
		return absPath, nil
	}
	return os.Getwd()
}

func init() {
	clicky.BindAllFlags(rootCmd.PersistentFlags(), "format", "tasks")
	logger.Configure(logger.Flags{LogToStderr: true, Color: true})
	rootCmd.PersistentFlags().StringVar(&workingDir, "cwd", "", "Working directory")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gavel %s (commit: %s, built: %s, go: %s)\n",
				version, commit, date, runtime.Version())
		},
	})

	rootCmd.AddCommand(newGavelMCPCommand())
}

func newGavelMCPCommand() *cobra.Command {
	cfg := mcp.DefaultConfig()
	cfg.Name = "gavel"
	cfg.Description = "Gavel CLI as an MCP server - lint, test, commit, pr, fixtures, status, and repository workflows"
	cfg.Version = version
	cfg.Security.TimeoutSeconds = 600
	cfg.Security.RequireConfirmation = false

	return mcp.NewMcpServer(rootCmd).
		WithConfig(cfg).
		AutoExpose().
		WithExclude(gavelMCPExcludedCommands()...).
		IgnoreParams("*", gavelMCPGlobalIgnoredParams()...).
		IgnoreParams("test", gavelMCPTestIgnoredParams()...).
		IgnoreParams("commit", gavelMCPCommitIgnoredParams()...).
		IgnoreParams("pr create", "base").
		IgnoreParams("pr list", gavelMCPPRListIgnoredParams()...).
		IgnoreParams("pr status", gavelMCPPRStatusIgnoredParams()...).
		IgnoreParams("repomap get", "args").
		IgnoreParams("repomap view", "args").
		WithFormat(formatters.FormatOptions{Markdown: true, NoColor: true}).
		Command()
}

// gavelMCPExcludedCommands returns regexp patterns matched against the
// space-joined cobra command path, e.g. "git amend-commits". Anchor with ^...$
// so "^git$" does not match "git summary".
func gavelMCPExcludedCommands() []string {
	return []string{
		// long-running servers that would block the MCP request indefinitely
		"^ssh$", "^ssh ",
		"^ui$", "^ui ",
		// top-level lint/test are enough; framework/linter aliases duplicate the same options
		"^lint ",
		"^test ",
		// OS-level daemon/service installers
		"^system$", "^system ",
		// expensive AI, history-rewriting, or large-analysis workflows
		"^verify$",
		"^todos$", "^todos ",
		"^bench$", "^bench ",
		"^pr fix$",
		"^git amend-commits$",
		"^git analyze$",
		"^git history$",
		"^git init-config$",
		"^status$",
	}
}

func gavelMCPGlobalIgnoredParams() []string {
	return []string{
		"--format",
		"--no-color",
		"--no-progress",
		"--timestamps",
		"--loglevel",
		"--ui",
		"--triage",
		"--sync-todos",
		"--todo-template",
		"--todos-dir",
		"--baseline",
		"--work-dir",
	}
}

func gavelMCPTestIgnoredParams() []string {
	return []string{
		"addr",
		"bench",
		"detach",
		"diagnostics",
		"fixture-files",
		"fixtures",
		"idle-timeout",
		"lint",
		"lint-timeout",
		"nodes",
		"recursive",
		"skip-hooks",
		"auto-stop",
		"cache",
		"framework",
	}
}

func gavelMCPCommitIgnoredParams() []string {
	return []string{
		"force",
		"interactive",
		"lint",
		"lint-secrets",
		"max-files",
		"max-lines",
		"model",
		"precommit",
		"push",
		"tree",
	}
}

func gavelMCPPRListIgnoredParams() []string {
	return []string{
		"all",
		"interval",
		"menu-bar",
		"persist-port",
		"port",
	}
}

func gavelMCPPRStatusIgnoredParams() []string {
	return append(gavelMCPFormatIgnoredParams(), "interval")
}

func gavelMCPFormatIgnoredParams() []string {
	return []string{
		"csv",
		"dump-schema",
		"filter",
		"html",
		"json",
		"markdown",
		"output",
		"pdf",
		"pretty",
		"slack",
		"table",
		"tree",
		"yaml",
	}
}

func main() {
	defer shutdown.RecoverAndShutdown()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
