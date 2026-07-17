package commit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/status"
)

// Indirection points so tests don't touch git or the terminal.
var (
	gatherStatusFunc = func(workDir string) (*status.Result, error) {
		return status.GatherBase(workDir, status.Options{NoRepomap: false})
	}
	addFilesFunc                = addFiles
	resetAllStagedFn            = resetAllStaged
	gitRmCachedFunc             = gitRmCached
	runTreePickerFunc           = runTreePicker
	startCandidateSummariesFunc = startCandidateSummaries
	interactiveStdout           = os.Stdout
)

// candidateSummarySession owns the per-file AI summary stream used by one
// picker invocation. Files contains the candidate slice after its AI states
// have been initialized. Stop cancels in-flight requests, closes the agent,
// waits for the stream to settle, and restores Clicky's task renderer.
type candidateSummarySession struct {
	Files   []status.FileStatus
	Updates <-chan status.AISummaryUpdate
	Stop    func()
}

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
func runInteractiveStaging(ctx context.Context, opts Options) ([]string, error) {
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

	var summaryUpdates <-chan status.AISummaryUpdate
	if opts.Summary {
		summaries, err := startCandidateSummariesFunc(ctx, opts, candidates)
		if err != nil {
			return nil, fmt.Errorf("start candidate summaries: %w", err)
		}
		defer summaries.Stop()
		candidates = summaries.Files
		summaryUpdates = summaries.Updates
	}

	picked, err := runTreePickerFunc(candidates, opts.WorkDir, summaryUpdates)
	if err != nil {
		return nil, err
	}
	if len(picked.Selected) == 0 {
		return nil, ErrInteractiveEmpty
	}

	if len(picked.RmCached) > 0 {
		if err := gitRmCachedFunc(opts.WorkDir, picked.RmCached); err != nil {
			return nil, fmt.Errorf("git rm --cached for newly-ignored files: %w", err)
		}
	}
	if err := resetAllStagedFn(opts.WorkDir); err != nil {
		return nil, fmt.Errorf("reset index before staging selection: %w", err)
	}
	if err := addFilesFunc(opts.WorkDir, picked.Selected); err != nil {
		return nil, fmt.Errorf("stage selected files: %w", err)
	}

	fmt.Fprintf(interactiveStdout, "selected %d of %d files; continuing with normal commit pipeline\n",
		len(picked.Selected), len(candidates))
	return picked.Selected, nil
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

// startCandidateSummaries starts the same per-file AI summary pipeline used by
// `gavel status --ai`, but feeds its updates to the Bubble Tea picker instead
// of installing Clicky's status renderer. The task renderer is disabled while
// Bubble Tea owns the alternate screen so the two renderers cannot corrupt one
// another's frames.
func startCandidateSummaries(ctx context.Context, opts Options, candidates []status.FileStatus) (*candidateSummarySession, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	model, err := opts.messageModel()
	if err != nil {
		return nil, err
	}
	agent, err := BuildAgent(opts, model)
	if err != nil {
		return nil, err
	}

	summaryPrompt, err := status.ResolveSummaryPrompt(opts.WorkDir)
	if err != nil {
		_ = agent.Close()
		return nil, err
	}

	prompting.Prepare()
	result := &status.Result{Files: append([]status.FileStatus(nil), candidates...)}
	result.PrepareAISummaries()

	streamCtx, cancel := context.WithCancel(ctx)
	previousNoRender := clickytask.IsNoRender()
	clickytask.SetNoRender(true)
	rawUpdates := status.StreamAISummaries(
		streamCtx,
		opts.WorkDir,
		agent,
		result.Files,
		clickyai.DefaultConfig().MaxConcurrent,
		summaryPrompt,
	)

	updates := make(chan status.AISummaryUpdate, len(result.Files)*2+1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer close(updates)
		for update := range rawUpdates {
			updates <- update
		}
	}()

	var stopOnce sync.Once
	stop := func() {
		stopOnce.Do(func() {
			cancel()
			_ = agent.Close()
			<-done
			clickytask.SetNoRender(previousNoRender)
		})
	}

	return &candidateSummarySession{
		Files:   result.Files,
		Updates: updates,
		Stop:    stop,
	}, nil
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
