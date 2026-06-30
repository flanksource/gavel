package commit

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flanksource/captain/pkg/ai/history"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
)

// resolveEnvSessionID returns the agent session id exported into the commit
// process, preferring gavel's own GAVEL_SESSION_ID, then Claude Code's
// CLAUDE_SESSION_ID, then Codex's CODEX_SESSION_ID. Empty when none is set —
// `--stage=session` then falls back to committing the already-staged set.
func resolveEnvSessionID() string {
	return firstNonEmpty(os.Getenv(EnvSessionID), os.Getenv(EnvClaudeSessionID), os.Getenv(EnvCodexSessionID))
}

// stageSessionFiles stages exactly the files an agent wrote to during the
// session identified by sessionID — for Claude the file_path of every
// Edit/Write/MultiEdit/NotebookEdit tool call, for Codex the targets of every
// apply_patch — filtered by .gitignore and the repo's .gavel.yaml
// commit.gitignore. It scopes a commit to what the agent changed rather than the
// whole working tree, and backs `gavel commit --stage=session|<session-id>` and
// the todo runner's auto-commit. Files edited outside workDir, no longer present,
// or matching an ignore rule are skipped (each logged).
func stageSessionFiles(workDir, sessionID string, cfg verify.CommitConfig) error {
	sessionFile, err := findSessionFile(sessionID)
	if err != nil {
		return fmt.Errorf("stage session %q: %w", sessionID, err)
	}
	logger.Infof("commit: staging files from session %s (%s)", sessionID, sessionFile)

	modified, err := sessionModifiedFiles(sessionFile)
	if err != nil {
		return fmt.Errorf("stage session %q: read session log %s: %w", sessionID, sessionFile, err)
	}
	logger.Infof("commit: session %s edited %d file(s)", sessionID, len(modified))
	for _, p := range modified {
		logger.V(1).Infof("commit:   edited %s", p)
	}

	absWork, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve work dir %q: %w", workDir, err)
	}

	candidates := make([]string, 0, len(modified))
	for _, p := range modified {
		abs := p
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(absWork, abs)
		}
		rel, err := filepath.Rel(absWork, abs)
		if err != nil || strings.HasPrefix(rel, "..") {
			logger.Infof("commit: skipping %s (edited outside %s)", p, absWork)
			continue
		}
		if _, statErr := os.Stat(abs); statErr != nil {
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

// findSessionFile locates the on-disk session log for sessionID, trying Claude
// project logs first (keyed by id) and falling back to Codex rollouts (whose
// filenames embed the session uuid).
func findSessionFile(sessionID string) (string, error) {
	f, claudeErr := history.FindSessionFile(sessionID)
	if claudeErr == nil {
		return f, nil
	}
	cf, codexErr := findCodexSessionFile(sessionID)
	if codexErr == nil {
		return cf, nil
	}
	return "", fmt.Errorf("no claude or codex session log found (%v; %v)", claudeErr, codexErr)
}

// findCodexSessionFile finds the Codex rollout whose filename carries sessionID.
// Codex names every rollout rollout-<timestamp>-<session-uuid>.jsonl, so the id
// alone identifies the file under ~/.codex/sessions.
func findCodexSessionFile(sessionID string) (string, error) {
	files, err := history.FindCodexSessionFiles()
	if err != nil {
		return "", fmt.Errorf("list codex sessions: %w", err)
	}
	for _, f := range files {
		if strings.Contains(filepath.Base(f), sessionID) {
			return f, nil
		}
	}
	return "", fmt.Errorf("no codex rollout for session %s", sessionID)
}

// sessionModifiedFiles returns the files an agent wrote to during a session,
// dispatching on the log format: Codex rollouts go through apply_patch parsing,
// Claude logs through the Edit/Write tool_use file_path.
func sessionModifiedFiles(sessionFile string) ([]string, error) {
	if history.IsCodexSession(sessionFile) {
		toolUses, err := history.ExtractCodexToolUses(sessionFile)
		if err != nil {
			return nil, fmt.Errorf("extract codex tool uses from %s: %w", sessionFile, err)
		}
		return codexModifiedFiles(toolUses), nil
	}
	return history.SessionModifiedFiles(sessionFile)
}

// applyPatchFileRE / applyPatchMoveRE match the file targets of a Codex
// apply_patch envelope. The path runs to the next newline (real or JSON-escaped
// \n), quote, or backslash, so the same expression works whether the command is
// a clean shell string or the raw JSON arguments of a function call.
var (
	applyPatchFileRE = regexp.MustCompile(`\*\*\* (?:Add|Update|Delete) File: ([^\n"\\]+)`)
	applyPatchMoveRE = regexp.MustCompile(`\*\*\* Move to: ([^\n"\\]+)`)
)

// codexModifiedFiles extracts the files a Codex session edited by scanning the
// apply_patch envelopes in its shell tool calls. Only Bash tool uses are read so
// patch text quoted inside agent/user messages is never mistaken for an edit.
func codexModifiedFiles(toolUses []history.ToolUse) []string {
	var files []string
	seen := make(map[string]struct{}, len(toolUses))
	for _, tu := range toolUses {
		if tu.Tool != "Bash" {
			continue
		}
		command, _ := tu.Input["command"].(string)
		if command == "" {
			continue
		}
		for _, p := range parseApplyPatchPaths(command) {
			if _, dup := seen[p]; dup {
				continue
			}
			seen[p] = struct{}{}
			files = append(files, p)
		}
	}
	return files
}

// parseApplyPatchPaths returns the distinct file paths named by *** Add/Update/
// Delete File and *** Move to markers in a Codex apply_patch command.
func parseApplyPatchPaths(command string) []string {
	var paths []string
	for _, re := range []*regexp.Regexp{applyPatchFileRE, applyPatchMoveRE} {
		for _, m := range re.FindAllStringSubmatch(command, -1) {
			if p := strings.TrimSpace(m[1]); p != "" {
				paths = append(paths, p)
			}
		}
	}
	return paths
}

// filterIgnoredPaths drops paths the repo's .gitignore (naming one to `git add`
// would error) or .gavel.yaml commit.gitignore excludes, logging each skip.
func filterIgnoredPaths(workDir string, candidates []string, cfg verify.CommitConfig) ([]string, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	ignored := make(map[string]struct{})

	absToRel := make(map[string]string, len(candidates))
	abs := make([]string, 0, len(candidates))
	for _, c := range candidates {
		p := filepath.Join(workDir, c)
		absToRel[p] = c
		abs = append(abs, p)
	}
	_, gitIgnored := utils.PartitionGitIgnored(abs, workDir)
	for _, p := range gitIgnored {
		rel := absToRel[p]
		logger.Infof("commit: skipping %s (matches .gitignore)", rel)
		ignored[rel] = struct{}{}
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
