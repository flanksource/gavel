package main

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/workspace"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	fixturesFromStep string
	fixturesFromRun  string
)

// fixtureResumeFlags are --from-step / --from-run over a fixtures command's
// positional files.
type fixtureResumeFlags struct {
	FromStep string
	FromRun  string
	Files    []string
	Dirs     workspace.Dirs
}

// resume builds the resume the flags ask for, the zero fixtureRerun without
// --from-step.
func (f fixtureResumeFlags) resume() (fixtureRerun, error) {
	if f.FromStep == "" {
		if f.FromRun != "" {
			return fixtureRerun{}, errors.New("--from-run needs --from-step")
		}
		return fixtureRerun{}, nil
	}
	if len(f.Files) != 1 {
		return fixtureRerun{}, fmt.Errorf("--from-step resumes one fixture file, got %d", len(f.Files))
	}
	file, err := filepath.Abs(f.Files[0])
	if err != nil {
		return fixtureRerun{}, fmt.Errorf("--from-step: %w", err)
	}
	if info, err := os.Stat(file); err != nil || !info.Mode().IsRegular() {
		return fixtureRerun{}, fmt.Errorf("--from-step resumes one fixture file, and %q is not a file", f.Files[0])
	}
	if f.FromRun != "" {
		snapshot, err := f.namedRun()
		if err != nil {
			return fixtureRerun{}, fmt.Errorf("--from-run: %w", err)
		}
		rerun, found, err := f.resumeRun(snapshot, file)
		if err == nil && !found {
			err = fmt.Errorf("--from-run: run %s has no steps of %s", f.FromRun, file)
		}
		return rerun, err
	}
	return f.resumeNewest(file)
}

// namedRun is the run --from-run names: a saved run's id or a snapshot file.
func (f fixtureResumeFlags) namedRun() (*testui.Snapshot, error) {
	if strings.HasPrefix(f.FromRun, snapshots.PerRunPrefix) && filepath.Ext(f.FromRun) == "" {
		return snapshots.LoadRun(f.Dirs.StateRoot, f.FromRun)
	}
	return snapshots.LoadRunFile(f.FromRun)
}

// resumeNewest resumes the newest saved run that holds steps of file.
func (f fixtureResumeFlags) resumeNewest(file string) (fixtureRerun, error) {
	runs, err := snapshots.ListRuns(f.Dirs.StateRoot, snapshots.ListRunsOptions{})
	if err != nil {
		return fixtureRerun{}, fmt.Errorf("--from-step: %w", err)
	}
	for _, run := range runs {
		if !slices.ContainsFunc(run.FixtureFiles, func(candidate string) bool { return filepath.Base(candidate) == filepath.Base(file) }) {
			continue
		}
		snapshot, err := snapshots.LoadRun(f.Dirs.StateRoot, run.RunID)
		if err != nil {
			return fixtureRerun{}, fmt.Errorf("--from-step: %w", err)
		}
		rerun, found, err := f.resumeRun(snapshot, file)
		if err != nil {
			return fixtureRerun{}, err
		}
		if found {
			logger.Infof("Resuming %s from step %s of %s", file, f.FromStep, run.RunID)
			return rerun, nil
		}
	}
	return fixtureRerun{}, fmt.Errorf("--from-step: no saved run of %s under %s; name one with --from-run", file, filepath.Join(f.Dirs.StateRoot, snapshots.Dir))
}

func (f fixtureResumeFlags) resumeRun(snapshot *testui.Snapshot, file string) (fixtureRerun, bool, error) {
	rerun, err := resumeFixture(snapshot.Tests, fixtureResumeRequest{File: file, FromStep: f.FromStep, WorkDir: runDirectory(snapshot.Metadata)})
	return rerun, len(rerun.Resume.Prior) > 0 || rerun.Params != nil, err
}

// runDirectory is the directory a run's relative fixture files are relative to.
func runDirectory(meta *testui.SnapshotMetadata) string {
	if meta == nil {
		return ""
	}
	if meta.Rerun != nil {
		return meta.Rerun.Cwd
	}
	workDir, _ := meta.Args["work_dir"].(string)
	return workDir
}

// withParams are the resumed run's params answers with the supplied ones over
// them.
func (f fixtureRerun) withParams(supplied map[string]any) map[string]any {
	if f.Params == nil {
		return supplied
	}
	merged := maps.Clone(f.Params)
	maps.Copy(merged, supplied)
	return merged
}

// fixtureCommandOmits are the flags a rerun command leaves out: it names its
// own step and run, and that run holds the params.
var fixtureCommandOmits = []string{"param", "params", "from-step", "from-run"}

// fixtureCommand is the command that runs cmd again with the flags it was
// given, up to the fixture files.
func fixtureCommand(cmd *cobra.Command) (*fixtures.RerunBase, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("resolve the rerun command's directory: %w", err)
	}
	argv := append([]string{os.Args[0]}, strings.Fields(cmd.CommandPath())[1:]...)
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if slices.Contains(fixtureCommandOmits, flag.Name) {
			return
		}
		values := []string{flag.Value.String()}
		if list, ok := flag.Value.(pflag.SliceValue); ok {
			values = list.GetSlice()
		}
		for _, value := range values {
			argv = append(argv, "--"+flag.Name+"="+value)
		}
	})
	return &fixtures.RerunBase{Cwd: cwd, Argv: argv}, nil
}

// saveFixtureRun saves a terminal fixtures run as a per-run snapshot, which
// --from-step resumes. It leaves last.json, which `gavel test` owns.
func saveFixtureRun(run finishedRun, runner *fixtures.Runner, tree *fixtures.FixtureNode) error {
	var err error
	if run.Tests, err = testrunner.FixtureRunTests(runner, tree, run.Opts.WorkDir); err != nil {
		return fmt.Errorf("save the fixture run: %w", err)
	}
	saveFinishedRun(run)
	return nil
}

func init() {
	fixturesCmd.Flags().StringVar(&fixturesFromStep, "from-step", "", "Rerun one fixture file from this step id, reusing what the earlier steps left in --from-run instead of running them")
	fixturesCmd.Flags().StringVar(&fixturesFromRun, "from-run", "", "The run --from-step reuses: a saved run id (run-…) or a snapshot file such as .gavel/running.json (default the newest saved run of the file)")
}
