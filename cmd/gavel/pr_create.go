package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/github"
	"github.com/spf13/cobra"
)

const (
	prCreateMarkerFile  = ".gavel-pr-create.json"
	prCreateTmpBranchNS = "gavel/pr-create-tmp/"
)

var prCreateScratchSub = filepath.FromSlash(".tmp/pr-create")

var (
	prCreateBase     string
	prCreateDraft    bool
	prCreateMainline int
	prCreateRepo     string
	prCreateModel    string
	prCreateNoCache  bool
)

var prCreateCmd = &cobra.Command{
	Use:          "create <SHA>",
	Short:        "Cherry-pick a commit into a fresh worktree off --base and open a PR",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	RunE:         runPRCreate,
}

func init() {
	prCmd.AddCommand(prCreateCmd)
	prCreateCmd.Flags().StringVar(&prCreateBase, "base", "origin/main", "Base ref to branch from (e.g. origin/main, main)")
	prCreateCmd.Flags().BoolVar(&prCreateDraft, "draft", false, "Open as draft PR")
	prCreateCmd.Flags().IntVar(&prCreateMainline, "mainline", 0, "Mainline parent for cherry-picking a merge commit (1-based)")
	prCreateCmd.Flags().StringVarP(&prCreateRepo, "repo", "R", "", "GitHub repository (owner/repo)")
	prCreateCmd.Flags().StringVar(&prCreateModel, "model", "", "Override the LLM model used for PR title/body/branch")
	prCreateCmd.Flags().BoolVar(&prCreateNoCache, "no-cache", false, "Disable LLM response cache")
}

// prCreateDeps lets tests substitute the GitHub call, browser opener, and AI
// content generator without touching git or the network.
type prCreateDeps struct {
	createPR        func(github.Options, github.CreatePRInput) (*github.CreatePRResult, error)
	openBrowser     func(string)
	generateContent func(ctx context.Context, in commitpkg.PRContentInput) (commitpkg.PRContent, error)
}

func defaultPRCreateDeps() prCreateDeps {
	return prCreateDeps{
		createPR:    github.CreatePR,
		openBrowser: openBrowser,
		generateContent: func(ctx context.Context, in commitpkg.PRContentInput) (commitpkg.PRContent, error) {
			agent, err := commitpkg.BuildAgent(commitpkg.Options{
				Model:   prCreateModel,
				NoCache: prCreateNoCache,
			})
			if err != nil {
				return commitpkg.PRContent{}, err
			}
			return commitpkg.GeneratePRContent(ctx, agent, in)
		},
	}
}

func runPRCreate(cmd *cobra.Command, args []string) error {
	return runPRCreateWithDeps(cmd.Context(), args[0], prCreateOptions{
		Base:     prCreateBase,
		Draft:    prCreateDraft,
		Mainline: prCreateMainline,
		Repo:     prCreateRepo,
	}, defaultPRCreateDeps())
}

type prCreateOptions struct {
	Base     string
	Draft    bool
	Mainline int
	Repo     string
}

func runPRCreateWithDeps(ctx context.Context, sha string, opts prCreateOptions, deps prCreateDeps) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}
	repoRoot, err := gitRepoRoot(workDir)
	if err != nil {
		return err
	}

	pf, err := preflightPRCreate(repoRoot, sha, opts)
	if err != nil {
		return err
	}

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	wtPath := filepath.Join(repoRoot, prCreateScratchSub, pf.shortSHA+"-"+suffix)
	tmpBranch := prCreateTmpBranchNS + pf.shortSHA + "-" + suffix

	if err := os.MkdirAll(filepath.Dir(wtPath), 0o755); err != nil {
		return fmt.Errorf("create scratch dir: %w", err)
	}
	if err := gitWorktreeAdd(repoRoot, wtPath, tmpBranch, opts.Base); err != nil {
		return fmt.Errorf("git worktree add: %w", err)
	}
	if err := writeMarker(wtPath, pf.fullSHA, repoRoot); err != nil {
		_ = removePRCreateWorktree(repoRoot, wtPath)
		return err
	}

	keepWorktree := false
	defer func() {
		if keepWorktree {
			return
		}
		if err := removePRCreateWorktree(repoRoot, wtPath); err != nil {
			logger.Warnf("worktree cleanup: %v", err)
		}
	}()

	if err := gitCherryPick(wtPath, pf.fullSHA, opts.Mainline); err != nil {
		if isNoOpCherryPick(wtPath) {
			baseLocal, _ := splitBaseRef(opts.Base)
			return fmt.Errorf("commit %s is already in %s; nothing to PR", pf.shortSHA, baseLocal)
		}
		keepWorktree = true
		printConflictHint(wtPath, tmpBranch)
		return fmt.Errorf("git cherry-pick %s: %w", pf.shortSHA, err)
	}

	content, err := deps.generateContent(ctx, prContentInputForSHA(wtPath))
	if err != nil {
		return fmt.Errorf("generate PR content: %w", err)
	}
	topic := content.Branch + "-" + suffix
	if err := gitRenameBranch(wtPath, tmpBranch, topic); err != nil {
		return fmt.Errorf("rename branch %s -> %s: %w", tmpBranch, topic, err)
	}

	if err := gitPushTopic(wtPath, topic); err != nil {
		return fmt.Errorf("git push %s: %w", topic, err)
	}

	baseLocal, _ := splitBaseRef(opts.Base)
	ghOpts := github.Options{WorkDir: repoRoot}
	if opts.Repo != "" {
		ghOpts.Repo = opts.Repo
	}
	created, err := deps.createPR(ghOpts, github.CreatePRInput{
		Title: content.Title,
		Body:  content.Body,
		Head:  topic,
		Base:  baseLocal,
		Draft: opts.Draft,
	})
	if err != nil {
		return fmt.Errorf("create PR: %w", err)
	}

	logger.Infof("Opened PR #%d against %s: %s", created.Number, created.Base, created.URL)
	printPRCreateSummary(created, content)
	deps.openBrowser(created.URL)
	return nil
}

type preflightResult struct {
	fullSHA  string
	shortSHA string
}

func preflightPRCreate(repoRoot, sha string, opts prCreateOptions) (preflightResult, error) {
	if err := runGitQuiet(repoRoot, "rev-parse", "--git-dir"); err != nil {
		return preflightResult{}, fmt.Errorf("not a git repository: %s", repoRoot)
	}
	if err := runGitQuiet(repoRoot, "cat-file", "-e", sha+"^{commit}"); err != nil {
		return preflightResult{}, fmt.Errorf("commit %q not found in %s", sha, repoRoot)
	}
	if err := runGitQuiet(repoRoot, "rev-parse", "--verify", opts.Base); err != nil {
		return preflightResult{}, fmt.Errorf("base ref %q not resolvable", opts.Base)
	}
	if err := refuseInProgressOps(repoRoot); err != nil {
		return preflightResult{}, err
	}
	if dirty, _ := gitWorkingTreeDirty(repoRoot); dirty {
		logger.Warnf("source repo has uncommitted changes; the new worktree is isolated, but check you're not mid-something")
	}

	full, err := captureGit(repoRoot, "rev-parse", sha)
	if err != nil {
		return preflightResult{}, fmt.Errorf("resolve SHA %q: %w", sha, err)
	}
	full = strings.TrimSpace(full)
	if len(full) < 8 {
		return preflightResult{}, fmt.Errorf("git rev-parse returned unexpectedly short SHA %q", full)
	}

	parents, err := captureGit(repoRoot, "rev-list", "--parents", "-n", "1", full)
	if err != nil {
		return preflightResult{}, fmt.Errorf("rev-list parents: %w", err)
	}
	parts := strings.Fields(strings.TrimSpace(parents))
	if len(parts) > 2 && opts.Mainline == 0 {
		return preflightResult{}, fmt.Errorf("commit %s is a merge commit; pass --mainline 1 (or 2) to choose the parent", full[:8])
	}

	return preflightResult{fullSHA: full, shortSHA: full[:8]}, nil
}

func refuseInProgressOps(repoRoot string) error {
	gitDir, err := captureGit(repoRoot, "rev-parse", "--git-dir")
	if err != nil {
		return fmt.Errorf("locate .git: %w", err)
	}
	gitDir = strings.TrimSpace(gitDir)
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(repoRoot, gitDir)
	}
	for _, m := range []string{"REBASE_HEAD", "CHERRY_PICK_HEAD", "MERGE_HEAD"} {
		if _, err := os.Stat(filepath.Join(gitDir, m)); err == nil {
			return fmt.Errorf("source repo has %s in progress; finish or abort it before running pr create", m)
		}
	}
	return nil
}

// splitBaseRef returns (branchForGitHub, refForGit). origin/main -> ("main", "origin/main").
func splitBaseRef(base string) (string, string) {
	branch := strings.TrimPrefix(base, "refs/heads/")
	if i := strings.IndexByte(branch, '/'); i >= 0 {
		// strip a single remote prefix; "origin/main" -> "main", but
		// "feature/foo" stays "feature/foo" if no remote of that name exists.
		// Heuristic: only strip "origin/" since gavel pushes there exclusively.
		branch = strings.TrimPrefix(branch, "origin/")
	}
	return branch, base
}

func prContentInputForSHA(wtPath string) commitpkg.PRContentInput {
	msg, _ := captureGit(wtPath, "show", "-s", "--format=%B", "HEAD")
	rawFiles, _ := captureGit(wtPath, "show", "--name-only", "--format=", "HEAD")
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(rawFiles), "\n") {
		if s := strings.TrimSpace(line); s != "" {
			files = append(files, s)
		}
	}
	return commitpkg.PRContentInput{
		Commits: []commitpkg.PRCommitInput{{Message: strings.TrimSpace(msg), Files: files}},
	}
}

// --- worktree creation, push, cleanup ---------------------------------------

func gitRepoRoot(workDir string) (string, error) {
	out, err := captureGit(workDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("locate repo root: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func gitWorktreeAdd(repoRoot, path, branch, base string) error {
	cmd := exec.Command("git", "-C", repoRoot, "worktree", "add", "-b", branch, path, base)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitCherryPick(wtPath, fullSHA string, mainline int) error {
	args := []string{"-C", wtPath, "cherry-pick"}
	if mainline > 0 {
		args = append(args, "-m", strconv.Itoa(mainline))
	}
	args = append(args, fullSHA)
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func isNoOpCherryPick(wtPath string) bool {
	out, err := captureGit(wtPath, "status", "--porcelain")
	if err != nil {
		return false
	}
	if strings.TrimSpace(out) != "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(wtPath, ".git")); err == nil {
		return true
	}
	return false
}

func gitRenameBranch(wtPath, from, to string) error {
	return runGitQuiet(wtPath, "branch", "-m", from, to)
}

func gitPushTopic(wtPath, topic string) error {
	cmd := exec.Command("git", "-C", wtPath, "push", "-u", "origin", "HEAD:refs/heads/"+topic)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitWorkingTreeDirty(repoRoot string) (bool, error) {
	out, err := captureGit(repoRoot, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// --- marker + safe cleanup --------------------------------------------------

type prCreateMarker struct {
	CreatedAt  time.Time `json:"createdAt"`
	SourceSHA  string    `json:"sourceSHA"`
	ParentRepo string    `json:"parentRepo"`
	PID        int       `json:"pid"`
}

func writeMarker(wtPath, fullSHA, repoRoot string) error {
	m := prCreateMarker{
		CreatedAt:  time.Now().UTC(),
		SourceSHA:  fullSHA,
		ParentRepo: repoRoot,
		PID:        os.Getpid(),
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal marker: %w", err)
	}
	return os.WriteFile(filepath.Join(wtPath, prCreateMarkerFile), data, 0o644)
}

// removePRCreateWorktree deletes the worktree at path only after three
// independent safety checks pass. Any failure leaves the directory in place.
func removePRCreateWorktree(repoRoot, path string) error {
	if err := assertSafeWorktreePath(repoRoot, path); err != nil {
		return err
	}
	cmd := exec.Command("git", "-C", repoRoot, "worktree", "remove", "--force", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git worktree remove %s: %w (leaving for manual cleanup)", path, err)
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("os.RemoveAll %s: %w", path, err)
	}
	return nil
}

func assertSafeWorktreePath(repoRoot, path string) error {
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("abs repoRoot: %w", err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("abs path: %w", err)
	}
	scratchPrefix := filepath.Join(absRoot, prCreateScratchSub) + string(filepath.Separator)
	if !strings.HasPrefix(absPath, scratchPrefix) {
		return fmt.Errorf("refusing to remove %s: outside %s", absPath, scratchPrefix)
	}
	markerPath := filepath.Join(absPath, prCreateMarkerFile)
	data, err := os.ReadFile(markerPath)
	if err != nil {
		return fmt.Errorf("refusing to remove %s: marker missing (%w)", absPath, err)
	}
	var m prCreateMarker
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("refusing to remove %s: marker unparseable (%w)", absPath, err)
	}
	if !worktreeRegistered(repoRoot, absPath) {
		return fmt.Errorf("refusing to remove %s: not in git worktree list", absPath)
	}
	return nil
}

func worktreeRegistered(repoRoot, absPath string) bool {
	out, err := captureGit(repoRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			if strings.TrimPrefix(line, "worktree ") == absPath {
				return true
			}
		}
	}
	return false
}

// --- small helpers ----------------------------------------------------------

func runGitQuiet(workDir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", workDir}, args...)...)
	return cmd.Run()
}

func captureGit(workDir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", workDir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}

func printConflictHint(wtPath, tmpBranch string) {
	t := clicky.Text("cherry-pick conflict; resolve in ", "text-yellow-600").
		Append(wtPath, "font-bold").NewLine().
		Append("  git -C "+wtPath+" cherry-pick --continue", "text-muted").NewLine().
		Append("  git -C "+wtPath+" push -u origin "+tmpBranch, "text-muted").NewLine()
	fmt.Println(t.ANSI())
}

func printPRCreateSummary(created *github.CreatePRResult, content commitpkg.PRContent) {
	if created == nil {
		return
	}
	t := clicky.Text(fmt.Sprintf("PR #%d", created.Number), "font-bold text-green-600").
		Space().Append(content.Title, "font-bold").NewLine().
		Append(created.URL, "text-muted").NewLine()
	if body := strings.TrimSpace(content.Body); body != "" {
		t = t.NewLine().Append(body, "")
	}
	fmt.Println(t.ANSI())
}
