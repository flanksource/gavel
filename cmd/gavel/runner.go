// `gavel runner exec`, run inside a sandbox job: fetches the run's spec and source from gavel
// serve with the job's run token, runs it, posts the results and exits with the job's exit code.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/sandbox/runclient"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/cobra"
)

func init() {
	runnerCmd := &cobra.Command{Use: "runner", Short: "Runs inside gavel's sandbox jobs", Hidden: true}
	var run, workspace, source string
	execCmd := &cobra.Command{
		Use:    "exec",
		Short:  "Run a sandbox job: fetch its spec and source, run it, post its results",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			code, err := runnerExec(cmd.Context(), runnerOptions{
				Run: run, Workspace: workspace, Source: source, Getenv: os.Getenv, Command: runCommandIn, Output: cmd.OutOrStdout(),
			})
			if err != nil {
				return err
			}
			if code != 0 {
				os.Exit(code)
			}
			return nil
		},
	}
	execCmd.Flags().StringVar(&run, "run", "", "The sandbox run's ID")
	execCmd.Flags().StringVar(&workspace, "workspace", "/workspace", "Scratch directory for the source")
	execCmd.Flags().StringVar(&source, "source", "", "A checkout of the job's commit to clone the source from, instead of downloading it")
	runnerCmd.AddCommand(execCmd)
	var taskDir string
	runTaskCmd := &cobra.Command{
		Use:    "run-task",
		Short:  "Run a step's whole run in a substrate environment, relaying it on stdout",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Stdout is the run's channel to its caller, so everything else goes to stderr, the
			// environment's log.
			channel := os.Stdout
			os.Stdout = os.Stderr
			return runnerRunTask(cmd.Context(), runTaskOptions{Dir: taskDir}, channel, os.Stdin)
		},
	}
	runTaskCmd.Flags().StringVar(&taskDir, "dir", ".", "The run's directory, holding its task and workspace bundle")
	runnerCmd.AddCommand(runTaskCmd)
	rootCmd.AddCommand(runnerCmd)
}

type runnerOptions struct {
	Run       string
	Workspace string
	// Source is a checkout of the run's commit to clone the source from; empty downloads it.
	Source string
	Getenv func(string) string
	// Command runs name with args in dir, streaming its output, and returns its exit code.
	Command func(ctx context.Context, dir, name string, args []string, output io.Writer) (int, error)
	Output  io.Writer
}

// runnerExec runs the sandbox job and returns its exit code.
func runnerExec(ctx context.Context, opts runnerOptions) (int, error) {
	if opts.Getenv("GAVEL_DB") != "off" {
		return 0, errors.New("gavel runner exec needs GAVEL_DB=off: a sandbox job never reaches a database")
	}
	client, err := runclient.FromEnv(opts.Getenv)
	if err != nil {
		return 0, fmt.Errorf("%w; gavel runner exec runs inside a sandbox job", err)
	}
	if opts.Run != client.Run {
		return 0, fmt.Errorf("--run %s doesn't match GAVEL_RUN_ID %s", opts.Run, client.Run)
	}
	api := runAPI{client}

	var spec sandbox.RunSpec
	if err := api.getJSON(ctx, "spec", &spec); err != nil {
		return 0, err
	}
	src, err := checkout(ctx, api, spec, opts.Workspace, opts.Source)
	if err != nil {
		return 0, err
	}
	if spec.Kind == "agent" {
		return 0, errors.New("an agent run runs whole on substrate, through `gavel runner run-task`, not as a sandbox job")
	}
	name, args, err := jobCommand(spec, src)
	if err != nil {
		return 0, err
	}
	before := runSnapshots(src)
	code, err := opts.Command(ctx, src, name, args, opts.Output)
	if err != nil {
		return 0, fmt.Errorf("run %s: %w", name, err)
	}
	after := runSnapshots(src)
	paths := make([]string, 0, len(after))
	for path := range after {
		if !before[path] {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := api.postResults(ctx, path); err != nil {
			return code, err
		}
	}
	return code, nil
}

// checkout checks out the run's commit in <workspace>/src: cloned from source, a checkout of it,
// when there is one, or else from the run's source bundle, fetched first. The job runs in a clone
// of its own, so what it writes stays out of source.
func checkout(ctx context.Context, api runAPI, spec sandbox.RunSpec, workspace, source string) (string, error) {
	src := filepath.Join(workspace, "src")
	if source != "" {
		// A clone starts at source's head, the run's commit.
		if out, err := exec.CommandContext(ctx, "git", "clone", "--quiet", source, src).CombinedOutput(); err != nil {
			return "", fmt.Errorf("clone the source from %s: %w: %s", source, err, strings.TrimSpace(string(out)))
		}
		return src, detach(ctx, src, spec.SHA, spec.SHA)
	}
	bundle := filepath.Join(workspace, "src.bundle")
	if err := api.download(ctx, "source", bundle); err != nil {
		return "", err
	}
	if out, err := exec.CommandContext(ctx, "git", "clone", "--quiet", bundle, src).CombinedOutput(); err != nil {
		return "", fmt.Errorf("clone the source bundle: %w: %s", err, strings.TrimSpace(string(out)))
	}
	target, name := spec.SHA, spec.SHA
	if target == "" && spec.Ref != "" {
		// The clone has a local branch only for the default branch name git guesses, so the
		// ref is fetched from the bundle by its own name.
		if out, err := exec.CommandContext(ctx, "git", "-C", src, "fetch", "--quiet", bundle, spec.Ref).CombinedOutput(); err != nil {
			return "", fmt.Errorf("fetch %s from the source bundle: %w: %s", spec.Ref, err, strings.TrimSpace(string(out)))
		}
		target, name = "FETCH_HEAD", spec.Ref
	}
	return src, detach(ctx, src, target, name)
}

// detach checks target, which name names, out in src, detached; an empty target leaves src as
// it is.
func detach(ctx context.Context, src, target, name string) error {
	if target == "" {
		return nil
	}
	if out, err := exec.CommandContext(ctx, "git", "-C", src, "checkout", "--quiet", "--detach", target).CombinedOutput(); err != nil {
		return fmt.Errorf("check out %s: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// jobCommand is what the run's kind runs in src.
func jobCommand(spec sandbox.RunSpec, src string) (string, []string, error) {
	gavel, err := os.Executable()
	if err != nil {
		return "", nil, fmt.Errorf("find the gavel executable: %w", err)
	}
	switch spec.Kind {
	case "test", "lint":
		return gavel, append([]string{spec.Kind, "--work-dir", src}, spec.Args...), nil
	case "push-check":
		config, err := verify.LoadSingleGavelConfig(filepath.Join(src, verify.GavelConfigFileName))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", nil, fmt.Errorf("read the repository's %s: %w", verify.GavelConfigFileName, err)
		}
		if command := strings.TrimSpace(config.SSH.Cmd); command != "" {
			return "sh", []string{"-c", command}, nil
		}
		return gavel, []string{"test", "--lint", "--work-dir", src}, nil
	default:
		return "", nil, fmt.Errorf("run kind %q is not one of test, lint, push-check, agent", spec.Kind)
	}
}

// runSnapshots lists the per-run snapshots in src's .gavel directory.
func runSnapshots(src string) map[string]bool {
	matches, _ := filepath.Glob(filepath.Join(src, snapshots.Dir, snapshots.PerRunPrefix+"*.json"))
	found := map[string]bool{}
	for _, match := range matches {
		found[match] = true
	}
	return found
}

func runCommandIn(ctx context.Context, dir, name string, args []string, output io.Writer) (int, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = output, output
	err := cmd.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), nil
	}
	if err != nil {
		return 0, err
	}
	return 0, nil
}

// runAPI calls gavel serve's run API with the job's run token.
type runAPI struct{ *runclient.Client }

func (a runAPI) getJSON(ctx context.Context, path string, out any) error {
	resp, err := a.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return fmt.Errorf("get the %s of run %s: %w", path, a.Run, err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode the %s of run %s: %w", path, a.Run, err)
	}
	return nil
}

func (a runAPI) download(ctx context.Context, path, to string) error {
	resp, err := a.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return fmt.Errorf("get the %s of run %s: %w", path, a.Run, err)
	}
	defer resp.Body.Close()
	file, err := os.Create(to)
	if err != nil {
		return fmt.Errorf("save the %s of run %s: %w", path, a.Run, err)
	}
	if _, err := io.Copy(file, resp.Body); err != nil {
		_ = file.Close()
		return fmt.Errorf("save the %s of run %s: %w", path, a.Run, err)
	}
	return file.Close()
}

func (a runAPI) postResults(ctx context.Context, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("read results %s: %w", path, err)
	}
	defer file.Close()
	resp, err := a.Do(ctx, http.MethodPost, "results", file)
	if err != nil {
		return fmt.Errorf("post results %s of run %s: %w", filepath.Base(path), a.Run, err)
	}
	return resp.Body.Close()
}
