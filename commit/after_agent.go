package commit

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/flanksource/clicky/prompt"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
)

// AgentRunMetadata identifies the todo run an after-agent commit belongs to, so
// the generated commit message can be stamped with Gavel-Issue-Id /
// Claude-Session-Id trailers without relying on the GAVEL_ISSUE_ID /
// GAVEL_SESSION_ID env vars (those cover the case where the agent runs
// `gavel commit` itself).
type AgentRunMetadata struct {
	IssueID   string
	SessionID string
}

// AgentRun describes one commit an agent run wants cut through gavel's pipeline.
// The zero value beyond WorkDir/Cwd/Meta is the plain end-of-run auto-commit;
// captain's phase hooks fill in the rest to cut a fixup chain instead.
type AgentRun struct {
	// WorkDir is the run's source directory and Cwd the TODO's, joined and
	// resolved to a git root before anything is staged.
	WorkDir string
	Cwd     string
	Meta    AgentRunMetadata

	// Fixup, when set, commits against that hash with `fixup!` rather than
	// generating a message — so a per-turn commit costs no LLM call.
	Fixup string
	// Message overrides the generated subject. Empty leaves message generation
	// to the pipeline, which is the reason to route through it at all.
	Message string
	// DryRun reports what would be committed without writing.
	DryRun bool
	// SkipGates bypasses the pre-commit pipeline (hooks, lint, gitignore,
	// file-size, linked-deps, tidy). Callers that run their own cheap gates set
	// it; `gates: full` is precisely the caller that does not.
	SkipGates bool
}

// RunAfterAgent stages and commits what an agent changed, driving the same
// pipeline as `gavel commit` (Stage=all) in the git root of the agent's working
// directory. It is shared by the CLI (`todos run --commit`), the dashboard's
// auto-commit, and captain's commit hook via its Do callback. A run that staged
// nothing is a no-op (nil result), not an error. The returned Result carries the
// commit hashes so callers can hand them to issue verification.
func RunAfterAgent(ctx context.Context, run AgentRun) (*Result, error) {
	meta := run.Meta
	commitDir := resolveAgentCommitDir(run.WorkDir, run.Cwd)
	if root := repomap.FindGitRoot(commitDir); root != "" {
		commitDir = root
	}

	cfg, err := verify.LoadGavelConfig(commitDir)
	if err != nil {
		logger.Warnf("Failed to load .gavel.yaml: %v", err)
	}

	// Scope any prompt this commit raises (gitignore / linked-deps / file-size /
	// compatibility) to the todo and session, so the dashboard can surface it on the
	// todo detail page and the session view. When no UI sink is installed this is
	// inert and the commit keeps its terminal/non-TTY behavior.
	if meta.IssueID != "" || meta.SessionID != "" {
		scope := prompt.Scope{Owner: meta.IssueID, Kind: "commit"}
		if meta.SessionID != "" {
			scope.Labels = map[string]string{"session": meta.SessionID}
		}
		ctx = prompt.WithScope(ctx, scope)
	}

	// Scope the commit to the files the agent's session actually edited. Without
	// a session id (e.g. a codex run with no on-disk Claude log) fall back to
	// staging the whole change set, logging the reason rather than failing.
	stage := StageAll
	if meta.SessionID != "" {
		stage = meta.SessionID
	} else {
		logger.Infof("commit: no agent session id; staging all changes")
	}

	result, err := Run(ctx, Options{
		WorkDir:     commitDir,
		Stage:       stage,
		AddMetadata: true,
		IssueID:     meta.IssueID,
		SessionID:   meta.SessionID,
		Config:      cfg.Commit,
		AI:          cfg.AI,
		PR:          cfg.PR,
		Fixup:       run.Fixup,
		Message:     run.Message,
		DryRun:      run.DryRun,
		Force:       run.SkipGates,
		// Autosquash stays off here even for fixups: the caller cutting a chain
		// collapses it once, at the end of the run, rather than rebasing after
		// every turn.
		Autosquash: false,
	})
	if err != nil {
		if errors.Is(err, ErrNothingStaged) {
			logger.Infof("commit: no changes to commit")
			return nil, nil
		}
		return nil, err
	}
	for _, c := range result.Commits {
		logger.Infof("Committed %s: %s", c.Hash, firstLine(c.Message))
	}
	return result, nil
}

// resolveAgentCommitDir resolves the directory the agent worked in, mirroring how
// the executors derive their working directory from the TODO's cwd.
func resolveAgentCommitDir(workDir, cwd string) string {
	if cwd != "" {
		if filepath.IsAbs(cwd) {
			return filepath.Clean(cwd)
		}
		if workDir != "" {
			return filepath.Clean(filepath.Join(workDir, cwd))
		}
		return filepath.Clean(cwd)
	}
	return workDir
}
