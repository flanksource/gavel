package commit

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/commons/logger"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
)

var (
	ErrSinceInvalidRef      = errors.New("--since must resolve to a commit (ref, sha, or ~N / HEAD~N)")
	ErrSincePushed          = errors.New("--since range includes commits already on a remote; choose a --since at or after the remote frontier (e.g. origin/main)")
	ErrSinceNoDuplicates    = errors.New("no commits in range share a Gavel-Issue-Id trailer; nothing to merge")
	ErrSinceNeedsConfirm    = errors.New("--since rewrites history and needs confirmation; re-run with --yes or in an interactive terminal")
	ErrSinceWithMessage     = errors.New("--since cannot be combined with --message")
	ErrSinceWithCommitAll   = errors.New("--since cannot be combined with --commit-all")
	ErrSinceWithInteractive = errors.New("--since cannot be combined with --interactive")
)

// issueGroup holds the commits in a --since range that share one
// Gavel-Issue-Id trailer, oldest-first.
type issueGroup struct {
	IssueID string
	Commits []models.Commit
}

// validateSinceOptions rejects flag combinations that conflict with the
// history-only --since dedup mode. --since operates on committed history and
// has no per-commit message / picker / directory-grouping semantics.
func validateSinceOptions(opts Options) error {
	if opts.Since == "" {
		return nil
	}
	if strings.TrimSpace(opts.Message) != "" {
		return ErrSinceWithMessage
	}
	if opts.CommitAll {
		return ErrSinceWithCommitAll
	}
	if opts.Interactive {
		return ErrSinceWithInteractive
	}
	return nil
}

// resolveSinceRef normalizes a --since value into a commit-ish git can use as
// a rebase base: a leading "~" or a bare integer becomes HEAD~N; anything else
// is used literally. The result is validated to peel to a real commit.
func resolveSinceRef(workDir, since string) (string, error) {
	ref := strings.TrimSpace(since)
	if ref == "" {
		return "", ErrSinceInvalidRef
	}
	switch {
	case strings.HasPrefix(ref, "~"):
		ref = "HEAD" + ref
	case isAllDigits(ref):
		ref = "HEAD~" + ref
	}
	if !validRef(workDir, ref+"^{commit}") {
		return "", fmt.Errorf("%w: %q", ErrSinceInvalidRef, since)
	}
	return ref, nil
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// publishedCommits returns the short hashes of commits in sinceRef..HEAD that
// are already reachable from a remote-tracking ref. A non-empty result means
// the merge would rewrite published history and must be refused.
func publishedCommits(workDir, sinceRef string) ([]string, error) {
	spec := sinceRef + "..HEAD"
	inRange, err := gitRevList(workDir, spec)
	if err != nil {
		return nil, err
	}
	// Commits in range not reachable from any remote ref are safe to rewrite.
	unpublished, err := gitRevList(workDir, spec, "--not", "--remotes")
	if err != nil {
		return nil, err
	}
	safe := make(map[string]struct{}, len(unpublished))
	for _, h := range unpublished {
		safe[h] = struct{}{}
	}
	var published []string
	for _, h := range inRange {
		if _, ok := safe[h]; !ok {
			published = append(published, shortHash(h))
		}
	}
	return published, nil
}

func gitRevList(workDir, spec string, extra ...string) ([]string, error) {
	args := append([]string{"rev-list", spec}, extra...)
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git rev-list %s: %w: %s", strings.Join(args[1:], " "), err, strings.TrimSpace(stderr.String()))
	}
	return splitLines(string(out)), nil
}

// readIssueGroups reads sinceRef..HEAD oldest-first and groups commits by their
// Gavel-Issue-Id trailer. It returns the full ordered commit list plus the
// groups that have more than one commit (duplicates), in order of first
// appearance.
func readIssueGroups(workDir, sinceRef string) ([]models.Commit, []issueGroup, error) {
	commits, err := gavelgit.CommitsInRange(workDir, sinceRef+"..HEAD")
	if err != nil {
		return nil, nil, fmt.Errorf("read commits in %s..HEAD: %w", sinceRef, err)
	}
	// CommitsInRange returns newest-first; rebase reasons oldest-first.
	ordered := make([]models.Commit, 0, len(commits))
	for i := len(commits) - 1; i >= 0; i-- {
		ordered = append(ordered, commits[i])
	}

	order := make([]string, 0)
	byID := make(map[string]*issueGroup)
	for _, c := range ordered {
		id := strings.TrimSpace(c.Trailers[trailerIssueID])
		if id == "" {
			continue
		}
		g, ok := byID[id]
		if !ok {
			g = &issueGroup{IssueID: id}
			byID[id] = g
			order = append(order, id)
		}
		g.Commits = append(g.Commits, c)
	}

	var dups []issueGroup
	for _, id := range order {
		if g := byID[id]; len(g.Commits) > 1 {
			dups = append(dups, *g)
		}
	}
	return ordered, dups, nil
}

// buildRebaseTodo renders the interactive-rebase todo that collapses each
// duplicate group into its earliest commit. For a group's first commit it
// emits `pick`, then `fixup` for every later member (folding their changes in),
// then `exec git commit --amend -F <msgFile>` to set the AI-simplified message.
// Commits that are later members of a group are skipped where they originally
// sit; every other commit is preserved as a plain `pick`.
func buildRebaseTodo(ordered []models.Commit, dups []issueGroup, msgFiles map[string]string) string {
	consumed := make(map[string]struct{})
	firstByHash := make(map[string]issueGroup)
	for _, g := range dups {
		firstByHash[g.Commits[0].Hash] = g
		for _, c := range g.Commits[1:] {
			consumed[c.Hash] = struct{}{}
		}
	}

	var b strings.Builder
	for _, c := range ordered {
		if _, skip := consumed[c.Hash]; skip {
			continue
		}
		if g, ok := firstByHash[c.Hash]; ok {
			fmt.Fprintf(&b, "pick %s %s\n", c.Hash, c.Subject)
			for _, m := range g.Commits[1:] {
				fmt.Fprintf(&b, "fixup %s %s\n", m.Hash, m.Subject)
			}
			fmt.Fprintf(&b, "exec git commit --amend -F %s\n", msgFiles[g.IssueID])
			continue
		}
		fmt.Fprintf(&b, "pick %s %s\n", c.Hash, c.Subject)
	}
	return b.String()
}

// simplifyGroupMessage asks the LLM to fold several commit messages that all
// belong to the same Gavel-Issue-Id into one conventional commit message.
// Under the test stub it returns a deterministic message instead of calling AI.
func simplifyGroupMessage(ctx context.Context, opts Options, msgs []string) (string, error) {
	if os.Getenv(testEnvVar) == "1" {
		logger.V(1).Infof("%s=1, returning stub squash message", testEnvVar)
		return stubMessage, nil
	}
	agent, err := BuildAgent(opts, opts.messageModel())
	if err != nil {
		return "", err
	}
	prompt := "The following git commit messages all belong to the same logical change " +
		"and will be merged into a single commit. Write ONE clean Conventional Commits " +
		"message (a `type(scope): subject` line, optionally followed by a blank line and a " +
		"concise body) that summarizes them. Do not invent a Gavel-Issue-Id or other " +
		"trailers. Return only the commit message.\n\n---\n" +
		strings.Join(msgs, "\n\n---\n")
	resp, err := agent.ExecutePrompt(ctx, clickyai.PromptRequest{
		Name:   "fixup-squash-message",
		Prompt: prompt,
	})
	if err != nil {
		return "", fmt.Errorf("simplify squash message: %w", err)
	}
	msg := strings.TrimSpace(resp.Result)
	if msg == "" {
		return "", fmt.Errorf("simplify squash message: empty response")
	}
	return msg, nil
}

// ensureIssueTrailer appends `Gavel-Issue-Id: <id>` to msg unless it is already
// present, keeping the merged commit attributable to its issue.
func ensureIssueTrailer(msg, issueID string) string {
	if issueID == "" || hasTrailer(msg, trailerIssueID) {
		return msg
	}
	return strings.TrimRight(msg, "\n") + "\n\n" + trailerIssueID + ": " + issueID
}

// confirmSquash prints the merge plan and asks the user to confirm. --yes
// auto-confirms; a non-interactive terminal without --yes is refused so we
// never silently rewrite history. Returns (false, nil) when the user declines.
func confirmSquash(opts Options, dups []issueGroup, merged map[string]string) (bool, error) {
	rewritten := 0
	for _, g := range dups {
		rewritten += len(g.Commits)
		fmt.Fprintf(os.Stderr, "\nGavel-Issue-Id %s — merging %d commits:\n", g.IssueID, len(g.Commits))
		for _, c := range g.Commits {
			fmt.Fprintf(os.Stderr, "  %s %s\n", shortHash(c.Hash), c.Subject)
		}
		fmt.Fprintf(os.Stderr, "  → %s\n", firstLine(merged[g.IssueID]))
	}
	fmt.Fprintf(os.Stderr, "\nRewrite %d commit(s) into %d? [y/N]: ", rewritten, len(dups))

	if opts.AssumeYes {
		fmt.Fprintln(os.Stderr, "y (--yes)")
		return true, nil
	}
	if !stdinIsTerminal() {
		fmt.Fprintln(os.Stderr)
		return false, ErrSinceNeedsConfirm
	}
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

// runSequencedRebase runs `git rebase -i <sinceRef>` non-interactively, feeding
// git our precomputed todo via GIT_SEQUENCE_EDITOR (a `cp` that overwrites the
// generated todo) and a no-op GIT_EDITOR. On conflict it aborts and surfaces a
// clear error, mirroring runAutosquash.
func runSequencedRebase(workDir, sinceRef, todoPath string) error {
	cmd := exec.Command("git", "rebase", "-i", sinceRef)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(),
		"GIT_SEQUENCE_EDITOR=cp '"+todoPath+"'",
		"GIT_EDITOR=:",
	)
	cmd.Stdout = os.Stdout
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	stderrStr := stderr.String()
	if stderrStr != "" {
		os.Stderr.WriteString(stderrStr)
	}
	if err == nil {
		return nil
	}
	if strings.Contains(stderrStr, "CONFLICT") || strings.Contains(stderrStr, "could not apply") {
		_ = runGitRebaseAbort(workDir)
		return fmt.Errorf("rebase onto %s conflicted while merging Gavel-Issue-Id commits; aborted. Resolve the overlap manually", sinceRef)
	}
	return fmt.Errorf("git rebase -i %s: %w", sinceRef, err)
}

// runIssueIdSquash is the entry point for `gavel commit --since=<ref>`: it
// reviews sinceRef..HEAD, finds commits sharing a Gavel-Issue-Id trailer, and
// merges each such group into a single commit with an AI-simplified message.
func runIssueIdSquash(ctx context.Context, opts Options) (*Result, error) {
	sinceRef, err := resolveSinceRef(opts.WorkDir, opts.Since)
	if err != nil {
		return nil, err
	}

	published, err := publishedCommits(opts.WorkDir, sinceRef)
	if err != nil {
		return nil, err
	}
	if len(published) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrSincePushed, strings.Join(published, " "))
	}

	ordered, dups, err := readIssueGroups(opts.WorkDir, sinceRef)
	if err != nil {
		return nil, err
	}
	if len(dups) == 0 {
		return &Result{}, ErrSinceNoDuplicates
	}

	scratch := filepath.Join(opts.WorkDir, ".tmp", "gavel-fixup")
	if err := os.MkdirAll(scratch, 0o755); err != nil {
		return nil, fmt.Errorf("create scratch dir: %w", err)
	}

	merged := make(map[string]string, len(dups))
	msgFiles := make(map[string]string, len(dups))
	result := &Result{DryRun: opts.DryRun}
	for i, g := range dups {
		rawMsgs := make([]string, 0, len(g.Commits))
		for _, c := range g.Commits {
			m, mErr := commitMessage(opts.WorkDir, c.Hash)
			if mErr != nil {
				return nil, fmt.Errorf("read message for %s: %w", shortHash(c.Hash), mErr)
			}
			rawMsgs = append(rawMsgs, m)
		}
		msg, mErr := simplifyGroupMessage(ctx, opts, rawMsgs)
		if mErr != nil {
			return nil, mErr
		}
		msg = ensureIssueTrailer(msg, g.IssueID)
		merged[g.IssueID] = msg

		path := filepath.Join(scratch, fmt.Sprintf("msg-%d.txt", i))
		if err := os.WriteFile(path, []byte(msg+"\n"), 0o644); err != nil {
			return nil, fmt.Errorf("write merged message: %w", err)
		}
		msgFiles[g.IssueID] = path
		result.Commits = append(result.Commits, CommitResult{Message: msg})
	}

	if opts.DryRun {
		printSinceDryRun(dups, merged)
		return result, nil
	}

	ok, err := confirmSquash(opts, dups, merged)
	if err != nil {
		return nil, err
	}
	if !ok {
		logger.Infof("aborted: history not rewritten")
		return &Result{}, nil
	}

	todoPath := filepath.Join(scratch, "todo")
	if err := os.WriteFile(todoPath, []byte(buildRebaseTodo(ordered, dups, msgFiles)), 0o644); err != nil {
		return nil, fmt.Errorf("write rebase todo: %w", err)
	}
	if err := runSequencedRebase(opts.WorkDir, sinceRef, todoPath); err != nil {
		return result, err
	}
	if head, herr := headHash(opts.WorkDir); herr == nil {
		result.Hash = head
	}
	logger.Infof("Merged %d Gavel-Issue-Id group(s) into single commits", len(dups))
	return result, nil
}

func printSinceDryRun(dups []issueGroup, merged map[string]string) {
	fmt.Fprintln(dryRunOutput, "DRY RUN: --since Gavel-Issue-Id merge")
	for _, g := range dups {
		fmt.Fprintf(dryRunOutput, "  issue %s: merge %d commits\n", g.IssueID, len(g.Commits))
		for _, c := range g.Commits {
			fmt.Fprintf(dryRunOutput, "      %s %s\n", shortHash(c.Hash), c.Subject)
		}
		fmt.Fprintf(dryRunOutput, "    -> %s\n", firstLine(merged[g.IssueID]))
	}
}
