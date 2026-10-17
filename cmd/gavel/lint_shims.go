package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	cexec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/lint/shims"
	"github.com/flanksource/gavel/processruntime"
	"github.com/spf13/cobra"
)

type LintInstallShimsOptions struct {
	Timeout  string   `flag:"timeout" default:"30s" help:"Kill the linter after this long (0 disables)"`
	CPULimit float64  `flag:"cpu-limit" default:"200" help:"CPU budget in percent (200 = two cores); caps parallelism and kills a linter that stays above it (0 disables)"`
	MemLimit string   `flag:"mem-limit" default:"2GB" help:"Memory budget; sets GOMEMLIMIT / --max-old-space-size and kills a linter whose process tree exceeds it (empty disables)"`
	Nice     int      `flag:"nice" default:"10" help:"Niceness added to the linter process"`
	Dir      string   `flag:"dir" help:"Directory to write the shims into (default ~/.gavel/shims)"`
	Tools    []string `flag:"tools" help:"Linters to shim (default: eslint, golangci-lint, tsc)"`
	DryRun   bool     `flag:"dry-run" help:"Report what would be written without writing anything"`
	Force    bool     `flag:"force" help:"Overwrite existing files in --dir that are not gavel shims"`
}

func (LintInstallShimsOptions) Help() api.Textable {
	return clicky.Text(`Install PATH shims that make heavy linters cheap when run directly by
editors, agents or scripts.

Each shim runs the real linter through gavel with:
  golangci-lint  run gets --fast-only and --concurrency=<cores>; GOMAXPROCS and
                 GOMEMLIMIT follow --cpu-limit / --mem-limit
  eslint         --cache with the cache kept in the user cache dir;
                 UV_THREADPOOL_SIZE and --max-old-space-size follow the limits
  tsc            UV_THREADPOOL_SIZE and --max-old-space-size follow the limits

Every linter is niced, killed after --timeout (exit 124), and killed when its
process tree stays above --cpu-limit or exceeds --mem-limit (exit 137).
Flags you pass explicitly (e.g. --fast-only=false, --concurrency, --no-cache)
are left alone.

Put the shim dir first on PATH; install-shims prints the export line when it
is not. Set GAVEL_SHIMS=off to run the real linter unchanged. gavel lint
itself always resolves the real binaries and ignores the shims.`)
}

func registerLintShimCommands(lintCmd *cobra.Command) {
	clicky.AddNamedCommand("install-shims", lintCmd, LintInstallShimsOptions{}, runLintInstallShims)
	lintCmd.AddCommand(&cobra.Command{
		Use:                "shim-exec <tool> [args...]",
		Short:              "Run a linter with the limits passed by a gavel shim",
		Hidden:             true,
		DisableFlagParsing: true,
		Args:               cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			os.Exit(runShimExec(cmd.Context(), args[0], args[1:], os.Stdin, os.Stdout, os.Stderr))
		},
	})
}

func runLintInstallShims(opts LintInstallShimsOptions) (any, error) {
	dir := opts.Dir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve home dir for --dir: %w", err)
		}
		dir = filepath.Join(home, ".gavel", "shims")
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve --dir: %w", err)
	}
	gavelPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve gavel executable: %w", err)
	}
	if gavelPath, err = filepath.EvalSymlinks(gavelPath); err != nil {
		return nil, fmt.Errorf("resolve gavel executable: %w", err)
	}
	if len(opts.Tools) == 0 {
		opts.Tools = shims.ToolNames()
	}
	return shims.Install(shims.InstallOptions{
		Dir:       dir,
		GavelPath: gavelPath,
		Tools:     opts.Tools,
		Settings:  shims.Settings{Timeout: opts.Timeout, CPU: opts.CPULimit, Mem: opts.MemLimit, Nice: opts.Nice},
		DryRun:    opts.DryRun,
		Force:     opts.Force,
	})
}

// runShimExec runs tool with the limits a shim passed through its environment
// and returns the exit code the shim should exit with.
func runShimExec(ctx context.Context, tool string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fail := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "gavel shim: "+format+"\n", a...)
		return 126
	}
	real, err := shims.LookPath(tool)
	if err != nil {
		return fail("%v", err)
	}
	if os.Getenv(shims.EnvBypass) == "off" {
		return passthrough(ctx, real, args, stdin, stdout, stderr)
	}
	profile, ok := shims.Profiles[tool]
	if !ok {
		return fail("no shim profile for %q", tool)
	}
	settings, err := shims.SettingsFromEnv(os.Getenv)
	if err != nil {
		return fail("%v", err)
	}
	runtime, limits, err := settings.Runtime()
	if err != nil {
		return fail("%v", err)
	}
	if args, err = profile.Inject(args, limits); err != nil {
		return fail("%v", err)
	}
	runtime.Env = profile.Env(limits, map[string]string{"GOMEMLIMIT": os.Getenv("GOMEMLIMIT"), "NODE_OPTIONS": os.Getenv("NODE_OPTIONS")})

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	if timeout := settings.TimeoutDuration(); timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail("getwd: %v", err)
	}
	cmd := processruntime.NewCommand(ctx, runtime, real, args...)
	cmd.Dir, cmd.Stdin, cmd.Stdout, cmd.Stderr = cwd, stdin, stdout, stderr
	return shimExitCode(ctx, cmd.Run(), tool, settings, stderr)
}

func shimExitCode(ctx context.Context, err error, tool string, settings shims.Settings, stderr io.Writer) int {
	var exitErr *exec.ExitError
	var limitErr *cexec.ResourceLimitError
	switch {
	case err == nil:
		return 0
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		fmt.Fprintf(stderr, "gavel shim: %s exceeded the %s timeout\n", tool, settings.Timeout)
		return 124
	case errors.As(err, &limitErr):
		fmt.Fprintf(stderr, "gavel shim: %s killed: %s (cpu-limit %v%%, mem-limit %s)\n", tool, limitErr.Reason, settings.CPU, settings.Mem)
		return 137
	case errors.Is(ctx.Err(), context.Canceled):
		return 130
	case errors.As(err, &exitErr):
		return exitErr.ExitCode()
	default:
		fmt.Fprintf(stderr, "gavel shim: %s: %v\n", tool, err)
		return 126
	}
}

func passthrough(ctx context.Context, real string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, real, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr
	err := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &exitErr):
		return exitErr.ExitCode()
	default:
		fmt.Fprintf(stderr, "gavel shim: %v\n", err)
		return 126
	}
}
