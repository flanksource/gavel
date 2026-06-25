package commit

import (
	"context"
	"errors"
	"path/filepath"

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

// RunAfterAgent stages and commits everything an agent changed after a TODO run,
// driving the same pipeline as `gavel commit` (Stage=all) in the git root of the
// agent's working directory (workDir joined with the TODO's cwd). It is shared by
// the CLI (`todos run --commit`) and the dashboard's auto-commit. A run that
// staged nothing is a no-op, not an error.
func RunAfterAgent(ctx context.Context, workDir, cwd string, meta AgentRunMetadata) error {
	commitDir := resolveAgentCommitDir(workDir, cwd)
	if root := repomap.FindGitRoot(commitDir); root != "" {
		commitDir = root
	}

	cfg, err := verify.LoadGavelConfig(commitDir)
	if err != nil {
		logger.Warnf("Failed to load .gavel.yaml: %v", err)
	}

	result, err := Run(ctx, Options{
		WorkDir:     commitDir,
		Stage:       StageAll,
		AddMetadata: true,
		IssueID:     meta.IssueID,
		SessionID:   meta.SessionID,
		Config:      cfg.Commit,
	})
	if err != nil {
		if errors.Is(err, ErrNothingStaged) {
			logger.Infof("commit: no changes to commit")
			return nil
		}
		return err
	}
	for _, c := range result.Commits {
		logger.Infof("Committed %s: %s", c.Hash, firstLine(c.Message))
	}
	return nil
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
