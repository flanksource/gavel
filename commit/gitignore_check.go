package commit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/verify"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"golang.org/x/term"
)

const (
	IgnoreCheckModePrompt = CheckModePrompt
	IgnoreCheckModeFail   = CheckModeFail
	IgnoreCheckModeSkip   = CheckModeSkip
	IgnoreCheckModeFalse  = CheckModeFalse
)

var ErrGitIgnoreCancelled = errors.New("commit cancelled: staged file matched a .gavel.yaml gitignore pattern")

type Violation struct {
	File    string
	Pattern string
}

type Decision int

const (
	DecisionCancel Decision = iota
	DecisionGitIgnorePattern
	DecisionGitIgnoreFolder
	DecisionGitIgnoreFile
	DecisionAllow
)

const DecisionGitIgnore = DecisionGitIgnoreFile

type Decider func(ctx context.Context, v Violation) (Decision, error)

type CheckParams struct {
	WorkDir     string
	GitRoot     string
	StagedFiles []string
	Config      verify.CommitConfig
	Decider     Decider
	SaveDir     string
	Mode        string
}

type CheckOutcome struct {
	Cancelled  bool
	Unstaged   []string
	GitIgnored []string
	Allowed    []string
}

type gitIgnoreChoice struct {
	Text     string
	Decision Decision
}

var (
	interactiveDeciderFunc = runPromptDecider
	stdinIsTerminal        = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }
)

// EvaluateGitIgnoreMatches returns the staged files matched by any pattern in
// patterns, unless shadowed by an allow pattern. Uses go-git's .gitignore
// semantics (anchored patterns, globstar, trailing slash for directories).
// Returns an error when any pattern or allow entry is whitespace-only, so the
// caller sees a loud failure instead of a silently-discarded rule.
func EvaluateGitIgnoreMatches(stagedFiles, patterns, allow []string) ([]Violation, error) {
	if len(stagedFiles) == 0 || len(patterns) == 0 {
		return nil, nil
	}

	blockers, err := parsePatterns(patterns, "commit.gitignore")
	if err != nil {
		return nil, err
	}
	allowMatchers, err := parsePatterns(allow, "commit.allow")
	if err != nil {
		return nil, err
	}
	allowMatcher := gitignore.NewMatcher(allowMatchers)

	var violations []Violation
	for _, file := range stagedFiles {
		parts := splitGitPath(file)
		if allowMatcher.Match(parts, false) {
			continue
		}
		for i, p := range blockers {
			if p == nil {
				continue
			}
			if gitignore.NewMatcher([]gitignore.Pattern{p}).Match(parts, false) {
				violations = append(violations, Violation{
					File:    file,
					Pattern: patterns[i],
				})
				break
			}
		}
	}
	return violations, nil
}

func parsePatterns(raw []string, field string) ([]gitignore.Pattern, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	parsed := make([]gitignore.Pattern, len(raw))
	for i, p := range raw {
		if strings.TrimSpace(p) == "" {
			return nil, fmt.Errorf("%s: pattern #%d is empty", field, i+1)
		}
		parsed[i] = gitignore.ParsePattern(p, nil)
	}
	return parsed, nil
}

func splitGitPath(p string) []string {
	return strings.Split(filepath.ToSlash(p), "/")
}

// RunGitIgnoreCheck evaluates the staged files, prompts for each match, and
// applies the user's chosen action (unstage + append a pattern / folder / file
// entry to .gitignore, or append to commit.allow in the repo's .gavel.yaml).
// Cancel on any match aborts without applying any change.
func RunGitIgnoreCheck(ctx context.Context, p CheckParams) (CheckOutcome, error) {
	mode, err := normalizeCheckMode(p.Mode, "--precommit")
	if err != nil {
		return CheckOutcome{}, err
	}
	if mode == CheckModeSkip {
		return CheckOutcome{}, nil
	}

	violations, err := EvaluateGitIgnoreMatches(p.StagedFiles, p.Config.GitIgnore, p.Config.Allow)
	if err != nil {
		return CheckOutcome{}, err
	}
	if len(violations) == 0 {
		return CheckOutcome{}, nil
	}

	// Only escalate on non-TTY if we'd otherwise fall through to the real
	// interactive prompt. Callers that inject a Decider (tests, future
	// non-interactive flows) have already decided how to answer.
	if mode == IgnoreCheckModePrompt && p.Decider == nil && !stdinIsTerminal() {
		logger.Warnf("gitignore check: stdin is not a terminal; escalating to --precommit=fail")
		mode = IgnoreCheckModeFail
	}

	switch mode {
	case IgnoreCheckModeFail:
		return CheckOutcome{}, formatViolationsError(violations)
	case IgnoreCheckModePrompt:
	default:
		return CheckOutcome{}, fmt.Errorf("unknown --precommit mode: %q", mode)
	}

	decider := p.Decider
	if decider == nil {
		decider = interactiveDeciderFunc
	}

	decisions := make([]Decision, len(violations))
	for i, v := range violations {
		d, err := decider(ctx, v)
		if err != nil {
			return CheckOutcome{}, fmt.Errorf("gitignore prompt: %w", err)
		}
		if d == DecisionCancel {
			return CheckOutcome{Cancelled: true}, nil
		}
		decisions[i] = d
	}

	return applyDecisions(p, violations, decisions)
}

func formatViolationsError(violations []Violation) error {
	lines := make([]string, 0, len(violations)+1)
	lines = append(lines, "staged files match commit.gitignore patterns:")
	for _, v := range violations {
		lines = append(lines, fmt.Sprintf("  - %s (pattern: %s)", v.File, v.Pattern))
	}
	return errors.New(strings.Join(lines, "\n"))
}

func applyDecisions(p CheckParams, violations []Violation, decisions []Decision) (CheckOutcome, error) {
	var (
		toUnstage        []string
		toAllow          []string
		gitIgnoreEntries []string
		outcome          CheckOutcome
	)
	for i, v := range violations {
		switch decisions[i] {
		case DecisionGitIgnorePattern, DecisionGitIgnoreFolder, DecisionGitIgnoreFile:
			toUnstage = append(toUnstage, v.File)
			gitIgnoreEntries = append(gitIgnoreEntries, gitIgnoreEntry(v, decisions[i]))
		case DecisionAllow:
			toAllow = append(toAllow, v.File)
		}
	}

	gitRoot := p.GitRoot
	if gitRoot == "" {
		gitRoot = p.WorkDir
	}

	if len(toUnstage) > 0 {
		appended, err := appendGitIgnore(gitRoot, gitIgnoreEntries)
		if err != nil {
			return CheckOutcome{}, fmt.Errorf("append .gitignore: %w", err)
		}
		if err := resetFiles(p.WorkDir, toUnstage); err != nil {
			return CheckOutcome{}, fmt.Errorf("unstage: %w", err)
		}
		outcome.Unstaged = toUnstage
		outcome.GitIgnored = appended
	}

	if len(toAllow) > 0 {
		saveDir := p.SaveDir
		if saveDir == "" {
			saveDir = gitRoot
		}
		added, err := appendAllow(saveDir, toAllow)
		if err != nil {
			return CheckOutcome{}, fmt.Errorf("update .gavel.yaml: %w", err)
		}
		outcome.Allowed = added
	}

	return outcome, nil
}

// appendGitIgnore appends lines to {gitRoot}/.gitignore, preserving any
// existing content. Lines already present (exact match) are skipped. Returns
// the subset of entries actually written so callers can report them.
func appendGitIgnore(gitRoot string, entries []string) ([]string, error) {
	path := filepath.Join(gitRoot, ".gitignore")
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	present := make(map[string]struct{})
	for line := range strings.SplitSeq(string(existing), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		present[trimmed] = struct{}{}
	}

	var buf strings.Builder
	buf.Write(existing)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		buf.WriteByte('\n')
	}

	var written []string
	for _, e := range entries {
		if _, ok := present[e]; ok {
			continue
		}
		buf.WriteString(e)
		buf.WriteByte('\n')
		present[e] = struct{}{}
		written = append(written, e)
	}

	if len(written) == 0 {
		return nil, nil
	}
	if err := os.WriteFile(path, []byte(buf.String()), 0o644); err != nil {
		return nil, err
	}
	return written, nil
}

// appendAllow reads {saveDir}/.gavel.yaml (if present) in isolation,
// appends entries to commit.allow, and writes it back. Dedups against
// existing entries.
func appendAllow(saveDir string, entries []string) ([]string, error) {
	path := filepath.Join(saveDir, ".gavel.yaml")
	cfg, err := verify.LoadSingleGavelConfig(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	present := make(map[string]struct{}, len(cfg.Commit.Allow))
	for _, a := range cfg.Commit.Allow {
		present[a] = struct{}{}
	}

	var added []string
	for _, e := range entries {
		if _, ok := present[e]; ok {
			continue
		}
		cfg.Commit.Allow = append(cfg.Commit.Allow, e)
		present[e] = struct{}{}
		added = append(added, e)
	}

	if len(added) == 0 {
		return nil, nil
	}
	if err := verify.SaveGavelConfig(saveDir, cfg); err != nil {
		return nil, err
	}
	return added, nil
}

func gitIgnoreEntry(v Violation, d Decision) string {
	switch d {
	case DecisionGitIgnorePattern:
		return v.Pattern
	case DecisionGitIgnoreFolder:
		if dir := gitIgnoreFolderEntry(v.File); dir != "" {
			return dir
		}
		return v.File
	case DecisionGitIgnoreFile:
		return v.File
	default:
		return ""
	}
}

func gitIgnoreFolderEntry(file string) string {
	dir := filepath.Dir(filepath.ToSlash(file))
	if dir == "." || dir == "" {
		return ""
	}
	return dir + "/"
}

func gitIgnoreChoices(v Violation) []gitIgnoreChoice {
	choices := []gitIgnoreChoice{
		{
			Text:     fmt.Sprintf("Add matched pattern %q to .gitignore (unstage now)", v.Pattern),
			Decision: DecisionGitIgnorePattern,
		},
	}
	if dir := gitIgnoreFolderEntry(v.File); dir != "" {
		choices = append(choices, gitIgnoreChoice{
			Text:     fmt.Sprintf("Add folder %q to .gitignore (unstage now)", dir),
			Decision: DecisionGitIgnoreFolder,
		})
	}
	choices = append(choices,
		gitIgnoreChoice{
			Text:     fmt.Sprintf("Add file %q to .gitignore (unstage now)", v.File),
			Decision: DecisionGitIgnoreFile,
		},
		gitIgnoreChoice{
			Text:     "Allow this file in ./.gavel.yaml (commit.allow)",
			Decision: DecisionAllow,
		},
		gitIgnoreChoice{
			Text:     "Cancel commit",
			Decision: DecisionCancel,
		},
	)
	return choices
}

// applyGitIgnoreCheck runs the check and, if any file was unstaged, re-reads
// the staged source so the caller sees the updated file list. Returns
// ErrGitIgnoreCancelled when the user cancels.
func applyGitIgnoreCheck(ctx context.Context, opts Options, source stagedSource) (stagedSource, error) {
	if !shouldRunPrecommitChecks(opts.PrecommitMode) {
		return source, nil
	}

	outcome, err := RunGitIgnoreCheck(ctx, CheckParams{
		WorkDir:     opts.WorkDir,
		GitRoot:     opts.WorkDir,
		StagedFiles: source.Files,
		Config:      opts.Config,
		SaveDir:     opts.WorkDir,
		Mode:        opts.PrecommitMode,
	})
	if err != nil {
		return source, err
	}
	if outcome.Cancelled {
		return source, ErrGitIgnoreCancelled
	}
	if len(outcome.Unstaged) == 0 {
		return source, nil
	}
	refreshed, err := readStagedSource(opts.WorkDir)
	if err != nil {
		return source, fmt.Errorf("re-read staged source after gitignore unstage: %w", err)
	}
	return refreshed, nil
}

func runPromptDecider(_ context.Context, v Violation) (Decision, error) {
	header := fmt.Sprintf("Staged %q matches commit.gitignore pattern %q", v.File, v.Pattern)
	choices := gitIgnoreChoices(v)
	items := make([]string, len(choices))
	for i, choice := range choices {
		items[i] = choice.Text
	}
	idx, ok := promptSelectIndex(header, items)
	if !ok {
		return DecisionCancel, nil
	}
	return choices[idx].Decision, nil
}
