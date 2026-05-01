package commit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/status"
)

// Indirection points so tests don't touch git or the terminal.
var (
	gatherStatusFunc = func(workDir string) (*status.Result, error) {
		return status.GatherBase(workDir, status.Options{NoRepomap: false})
	}
	addFilesFunc      = addFiles
	resetAllStagedFn  = resetAllStaged
	runTreePickerFunc = runTreePicker
	interactiveStdout = os.Stdout
)

func validateInteractiveOptions(opts Options) error {
	if !opts.Interactive {
		if opts.Summary {
			logger.Warnf("--summary has no effect without --interactive; ignoring")
		}
		return nil
	}
	if opts.CommitAll {
		return ErrInteractiveWithCommitAll
	}
	if strings.TrimSpace(opts.Message) != "" {
		return ErrInteractiveWithMessage
	}
	if !stdinIsTerminal() {
		return ErrInteractiveNonTTY
	}
	return nil
}

// runInteractiveStaging gathers all changed files (staged, unstaged, untracked)
// in workDir, asks the user to select a subset via the tree picker, then
// resets the index and stages exactly the chosen paths. Returns the list of
// selected (now staged) file paths so the caller can verify and proceed with
// the standard commit pipeline.
func runInteractiveStaging(_ context.Context, opts Options) ([]string, error) {
	statusResult, err := gatherStatusFunc(opts.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("gather candidate files: %w", err)
	}

	candidates, skipped := filterCandidates(statusResult.Files)
	for _, c := range skipped {
		logger.Warnf("skipping %s (conflict — resolve manually before committing)", c.Path)
	}
	if len(candidates) == 0 {
		return nil, ErrNothingStaged
	}

	if opts.Summary {
		printCandidateSummary(statusResult, candidates)
	}

	selected, err := runTreePickerFunc(candidates)
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, ErrInteractiveEmpty
	}

	if err := resetAllStagedFn(opts.WorkDir); err != nil {
		return nil, fmt.Errorf("reset index before staging selection: %w", err)
	}
	if err := addFilesFunc(opts.WorkDir, selected); err != nil {
		return nil, fmt.Errorf("stage selected files: %w", err)
	}

	fmt.Fprintf(interactiveStdout, "selected %d of %d files; continuing with normal commit pipeline\n",
		len(selected), len(candidates))
	return selected, nil
}

// filterCandidates drops conflict files and returns the remainder along with
// the list of files that were skipped for the caller to surface.
func filterCandidates(files []status.FileStatus) (kept, skipped []status.FileStatus) {
	for _, f := range files {
		if f.State == status.StateConflict {
			skipped = append(skipped, f)
			continue
		}
		kept = append(kept, f)
	}
	return
}

func printCandidateSummary(result *status.Result, candidates []status.FileStatus) {
	if result == nil {
		return
	}
	view := *result
	view.Files = candidates
	fmt.Fprintln(interactiveStdout, view.Pretty().ANSI())
}

// resetAllStaged unstages everything currently in the index. Used to clear
// state before re-staging just the user's selection. We use `git reset HEAD --`
// (no paths after `--`) which does NOT touch the working tree.
func resetAllStaged(workDir string) error {
	cmd := exec.Command("git", "reset", "--mixed")
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		// `git reset --mixed` with an empty index can fail benignly when there
		// are no commits yet; tolerate that one case so the first commit on a
		// fresh repo still works.
		if strings.Contains(string(out), "ambiguous argument 'HEAD'") ||
			strings.Contains(string(out), "unknown revision") {
			return nil
		}
		return fmt.Errorf("git reset: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
