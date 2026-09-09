package commit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/verify"
)

// resolveEnvSessionID returns the agent session id exported into the commit
// process, preferring gavel's own GAVEL_SESSION_ID before Captain's provider
// marker detection. Empty when no marker supplies an ID — `--stage=session`
// then falls back to committing the already-staged set.
func resolveEnvSessionID() string {
	if sessionID := strings.TrimSpace(os.Getenv(EnvSessionID)); sessionID != "" {
		return sessionID
	}
	current := captaincli.CurrentEnvironmentSession()
	if current == nil {
		return ""
	}
	return current.SessionID
}

// stageSessionFiles stages exactly the files an agent wrote to during the
// session identified by sessionID or its prefix. Captain owns Claude/Codex
// history discovery and changed-file extraction; Gavel applies its commit-specific
// repository and ignore filters before staging. It scopes a commit to what the
// agent changed rather than the whole working tree, and backs
// `gavel commit --stage=session|<session-id>` and the todo runner's auto-commit.
// Files edited outside workDir, no longer present, or matching an ignore rule
// are skipped (each logged).
func stageSessionFiles(workDir, sessionID string, cfg verify.CommitConfig) error {
	value, err := captaincli.RunChanges(captaincli.ChangesOptions{
		SessionID: sessionID,
		All:       true,
		Agents:    true,
		Plans:     true,
		Ignored:   true,
	})
	if err != nil {
		return fmt.Errorf("stage session %q: %w", sessionID, err)
	}
	changes, ok := value.(captaincli.ChangesResult)
	if !ok {
		return fmt.Errorf("stage session %q: captain changes returned %T, expected cli.ChangesResult", sessionID, value)
	}
	// Captain returns absolute paths: a session's working directory moves as the
	// agent works, so only an absolute path is meaningful to this process.
	modified := make([]string, 0, len(changes.Files))
	for _, file := range changes.Files {
		modified = append(modified, string(file.Path))
	}
	logger.Infof("commit: session %s (%s) edited %d file(s)", changes.SessionID, changes.Source, len(modified))
	for _, p := range modified {
		logger.V(1).Infof("commit:   edited %s", p)
	}

	absWork, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve work dir %q: %w", workDir, err)
	}

	candidates := make([]string, 0, len(modified))
	for _, p := range modified {
		if !filepath.IsAbs(p) {
			// Captain anchors every path to the cwd of the tool use that wrote it,
			// so anything still relative is a fragment it could not resolve (an
			// unexpanded shell variable, say). Joining it onto workDir here is a
			// guess, and guessing is what used to manufacture paths that pointed
			// nowhere and were then reported as "no longer present".
			logger.Infof("commit: skipping %s (path could not be resolved)", p)
			continue
		}
		rel, err := filepath.Rel(absWork, p)
		if err != nil || strings.HasPrefix(rel, "..") {
			// Report the absolute path: "outside the repo" is precisely the case
			// where a repo-relative rendering would be meaningless.
			logger.Infof("commit: skipping %s (edited outside %s)", p, absWork)
			continue
		}
		if _, statErr := os.Stat(p); statErr != nil {
			logger.Infof("commit: skipping %s (no longer present)", rel)
			continue
		}
		candidates = append(candidates, rel)
	}

	keep, err := filterIgnoredPaths(absWork, candidates, cfg)
	if err != nil {
		return err
	}
	logger.Infof("commit: staging %d of %d session file(s)", len(keep), len(modified))
	if len(keep) == 0 {
		return fmt.Errorf("stage session %q (%d edited, all skipped): %w", sessionID, len(modified), ErrSessionNoFiles)
	}
	return addFiles(workDir, keep)
}

// filterIgnoredPaths drops paths the repo's .gitignore (naming one to `git add`
// would error) or .gavel.yaml commit.gitignore excludes, logging each skip.
func filterIgnoredPaths(workDir string, candidates []string, cfg verify.CommitConfig) ([]string, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	ignored := make(map[string]struct{})

	gitIgnored, err := gitCheckIgnore(workDir, candidates)
	if err != nil {
		return nil, fmt.Errorf("check .gitignore: %w", err)
	}
	for c := range gitIgnored {
		logger.Infof("commit: skipping %s (matches .gitignore)", c)
		ignored[c] = struct{}{}
	}

	violations, err := EvaluateGitIgnoreMatches(candidates, cfg.GitIgnore, cfg.Allow)
	if err != nil {
		return nil, fmt.Errorf("evaluate .gavel.yaml commit.gitignore: %w", err)
	}
	for _, v := range violations {
		logger.Infof("commit: skipping %s (matches .gavel.yaml commit.gitignore %q)", v.File, v.Pattern)
		ignored[v.File] = struct{}{}
	}

	keep := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if _, skip := ignored[c]; !skip {
			keep = append(keep, c)
		}
	}
	return keep, nil
}
