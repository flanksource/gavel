// `gavel runner exec`, run inside a sandbox job: fetches the run's spec and source from gavel
// serve with the job's run token, runs it, posts the results and exits with the job's exit code.
package main

import (
	"bytes"
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

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/sandbox/runclient"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/todos/remote"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/cobra"
)

func init() {
	runnerCmd := &cobra.Command{Use: "runner", Short: "Runs inside gavel's sandbox jobs", Hidden: true}
	var run, workspace string
	execCmd := &cobra.Command{
		Use:    "exec",
		Short:  "Run a sandbox job: fetch its spec and source, run it, post its results",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			code, err := runnerExec(cmd.Context(), runnerOptions{
				Run: run, Workspace: workspace, Getenv: os.Getenv, Command: runCommandIn, Output: cmd.OutOrStdout(),
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
	runnerCmd.AddCommand(execCmd)
	rootCmd.AddCommand(runnerCmd)
}

type runnerOptions struct {
	Run       string
	Workspace string
	Getenv    func(string) string
	// Command runs name with args in dir, streaming its output, and returns its exit code.
	Command func(ctx context.Context, dir, name string, args []string, output io.Writer) (int, error)
	Output  io.Writer
	// Agent is the provider an agent run's turns go to; nil builds it from the job's model.
	Agent captainai.Provider
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
	src, err := checkout(ctx, api, spec, opts.Workspace)
	if err != nil {
		return 0, err
	}
	if spec.Kind == "agent" {
		return 0, runAgentJob(ctx, api, spec, src, opts.Agent, opts.Output)
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

// checkout fetches the run's source bundle and checks out its commit in <workspace>/src.
func checkout(ctx context.Context, api runAPI, spec sandbox.RunSpec, workspace string) (string, error) {
	bundle := filepath.Join(workspace, "src.bundle")
	if err := api.download(ctx, "source", bundle); err != nil {
		return "", err
	}
	src := filepath.Join(workspace, "src")
	if out, err := exec.CommandContext(ctx, "git", "clone", "--quiet", bundle, src).CombinedOutput(); err != nil {
		return "", fmt.Errorf("clone the source bundle: %w: %s", err, strings.TrimSpace(string(out)))
	}
	target := spec.SHA
	if target == "" {
		target = spec.Ref
	}
	if target != "" {
		if out, err := exec.CommandContext(ctx, "git", "-C", src, "checkout", "--quiet", "--detach", target).CombinedOutput(); err != nil {
			return "", fmt.Errorf("check out %s: %w: %s", target, err, strings.TrimSpace(string(out)))
		}
	}
	return src, nil
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

// runAgentJob runs an agent run's step in src and reports it to serve as it goes: its messages,
// its agent's transcripts, and once it ends, its commits and then its result.
func runAgentJob(ctx context.Context, api runAPI, spec sandbox.RunSpec, src string, agent captainai.Provider, output io.Writer) error {
	var job remote.Job
	if err := json.Unmarshal(spec.Agent, &job); err != nil {
		return fmt.Errorf("decode the agent job of run %s: %w", api.Run, err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find the home directory the agent writes its transcripts under: %w", err)
	}
	transcripts := &transcriptUploader{api: api, home: home}
	uploading, stopUploading := context.WithCancel(ctx)
	uploaderDone := make(chan struct{})
	go func() {
		defer close(uploaderDone)
		transcripts.every(uploading, transcriptInterval, output)
	}()

	send := func(message remote.Message) error { return api.postMessage(ctx, message) }
	finished, err := remote.Execute(ctx, job, src, agent, send)
	stopUploading()
	<-uploaderDone
	if err != nil {
		return err
	}
	if err := transcripts.sync(ctx); err != nil {
		fmt.Fprintf(output, "gavel runner: %s\n", err)
	}
	if err := pushRunCommits(ctx, api, finished.Result); err != nil {
		finished.Err = errors.Join(finished.Err, err)
	}
	return send(remote.ResultMessage(finished.Result, finished.Err))
}

// pushRunCommits uploads the commits the run made on its worktree's branch, if it made any, and
// points the result's worktree at the branch serve put them on.
func pushRunCommits(ctx context.Context, api runAPI, result promptrun.Result) error {
	if result.Response == nil || result.Response.Workspace == nil || result.Response.Workspace.Worktree == nil {
		return nil
	}
	worktree := result.Response.Workspace.Worktree
	if worktree.BranchDeleted || worktree.Head == "" || worktree.Head == worktree.Setup {
		return nil
	}
	file, err := os.CreateTemp("", "gavel-run-*.bundle")
	if err != nil {
		return fmt.Errorf("bundle the run's commits: %w", err)
	}
	_ = file.Close()
	defer os.Remove(file.Name())
	if out, err := exec.CommandContext(ctx, "git", "-C", worktree.Repo, "bundle", "create", "--quiet", file.Name(),
		"refs/heads/"+worktree.Branch, "^"+worktree.Setup).CombinedOutput(); err != nil {
		return fmt.Errorf("bundle the run's commits on %s: %w: %s", worktree.Branch, err, strings.TrimSpace(string(out)))
	}
	data, err := os.Open(file.Name())
	if err != nil {
		return fmt.Errorf("read the bundle of the run's commits: %w", err)
	}
	defer data.Close()
	resp, err := api.Do(ctx, http.MethodPost, "commits", data)
	if err != nil {
		return fmt.Errorf("push the commits of run %s: %w", api.Run, err)
	}
	defer resp.Body.Close()
	var pushed struct{ Branch, SHA string }
	if err := json.NewDecoder(resp.Body).Decode(&pushed); err != nil {
		return fmt.Errorf("read where run %s's commits were pushed: %w", api.Run, err)
	}
	worktree.Branch, worktree.Head = pushed.Branch, pushed.SHA
	return nil
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

func (a runAPI) postMessage(ctx context.Context, message remote.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("encode a %s message of run %s: %w", message.Kind, a.Run, err)
	}
	resp, err := a.Do(ctx, http.MethodPost, "messages", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("send a %s message of run %s: %w", message.Kind, a.Run, err)
	}
	return resp.Body.Close()
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
