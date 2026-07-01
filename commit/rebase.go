package commit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/flanksource/commons/logger"
)

func runGitFetch(workDir, ref string) error {
	cmd := exec.Command("git", "fetch", "origin", ref)
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runGitRebase runs `git rebase [-Xours|-Xtheirs] <upstream>` and returns
// (clean, err). clean == false with err == nil means the rebase started but
// hit a conflict (caller is responsible for aborting).
func runGitRebase(workDir, upstream, strategyOpt string) (bool, error) {
	args := []string{"rebase", "--autostash"}
	if strategyOpt != "" {
		args = append(args, "-X"+strategyOpt)
	}
	args = append(args, upstream)
	cmd := exec.Command("git", args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	var stderr strings.Builder
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	stderrStr := stderr.String()
	os.Stderr.WriteString(stderrStr)
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if strings.Contains(stderrStr, "CONFLICT") || strings.Contains(stderrStr, "could not apply") {
			return false, nil
		}
	}
	return false, fmt.Errorf("git rebase %s: %w", upstream, err)
}

func runGitRebaseAbort(workDir string) error {
	cmd := exec.Command("git", "rebase", "--abort")
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git rebase --abort: %w", err)
	}
	return nil
}

// rebaseOnto fetches origin/<upstreamBranch> then attempts to rebase HEAD
// onto it. On conflict it aborts the rebase and prompts the user to retry
// with -Xours, -Xtheirs, or cancel the push entirely.
func rebaseOnto(workDir, upstreamBranch string) error {
	if upstreamBranch == "" {
		return nil
	}
	if err := runGitFetch(workDir, upstreamBranch); err != nil {
		return fmt.Errorf("git fetch origin %s: %w", upstreamBranch, err)
	}
	upstream := "origin/" + upstreamBranch

	clean, err := runGitRebase(workDir, upstream, "")
	if err != nil {
		return err
	}
	if clean {
		return nil
	}

	if abortErr := runGitRebaseAbort(workDir); abortErr != nil {
		return fmt.Errorf("rebase conflict and abort failed: %w", abortErr)
	}
	logger.Warnf("Rebase onto %s hit conflicts and was aborted.", upstream)

	choice, ok := promptSelectIndex(
		context.Background(),
		fmt.Sprintf("Rebase onto %s conflicted. Retry with which strategy?", upstream),
		[]string{
			"Cancel push",
			"Retry with -Xours (prefer our changes)",
			"Retry with -Xtheirs (prefer remote changes)",
		},
	)
	if !ok || choice == 0 {
		return fmt.Errorf("push cancelled: rebase onto %s conflicted", upstream)
	}
	strategy := "ours"
	if choice == 2 {
		strategy = "theirs"
	}
	logger.Infof("Retrying rebase onto %s with -X%s", upstream, strategy)
	clean, err = runGitRebase(workDir, upstream, strategy)
	if err != nil {
		return err
	}
	if !clean {
		_ = runGitRebaseAbort(workDir)
		return fmt.Errorf("rebase onto %s still conflicted with -X%s; resolve manually", upstream, strategy)
	}
	return nil
}
