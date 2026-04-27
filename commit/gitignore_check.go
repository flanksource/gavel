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
	// FolderSiblings is the count of other pending violations under the same
	// directory. Populated by the iterative checker so the prompt can describe
	// how many files a "add folder" choice would batch-unstage. Zero means no
	// siblings (or the violation didn't come from the iterative checker).
	FolderSiblings int
}

type Decision int

const (
	DecisionCancel Decision = iota
	DecisionGitIgnorePattern
	DecisionGitIgnoreFolder
	DecisionGitIgnoreFile
	DecisionAllow
	DecisionAllowFolder
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

	plan, err := iterateDecisions(ctx, p, violations, decider)
	if err != nil {
		return CheckOutcome{}, err
	}
	if plan.cancelled {
		return CheckOutcome{Cancelled: true}, nil
	}
	return applyPlan(p, plan)
}

// iterateDecisions walks the remaining violations one decision at a time and,
// after each decision, recomputes the remaining set against the patterns and
// allow-list accumulated so far. A single "add folder" or "add pattern" choice
// therefore batch-covers every sibling file it implies, instead of re-prompting
// for each one.
func iterateDecisions(ctx context.Context, p CheckParams, violations []Violation, decider Decider) (decisionPlan, error) {
	plan := decisionPlan{
		basePatterns: append([]string(nil), p.Config.GitIgnore...),
		baseAllow:    append([]string(nil), p.Config.Allow...),
	}
	remaining := append([]Violation(nil), violations...)

	for len(remaining) > 0 {
		v := remaining[0]
		v.FolderSiblings = countFolderSiblings(v.File, remaining[1:])

		d, err := decider(ctx, v)
		if err != nil {
			return decisionPlan{}, fmt.Errorf("gitignore prompt: %w", err)
		}
		if d == DecisionCancel {
			return decisionPlan{cancelled: true}, nil
		}

		switch d {
		case DecisionGitIgnorePattern:
			plan.gitIgnoreEntries = appendUnique(plan.gitIgnoreEntries, v.Pattern)
		case DecisionGitIgnoreFolder:
			entry := gitIgnoreFolderEntry(v.File)
			if entry == "" {
				entry = v.File
			}
			plan.gitIgnoreEntries = appendUnique(plan.gitIgnoreEntries, entry)
		case DecisionGitIgnoreFile:
			plan.gitIgnoreEntries = appendUnique(plan.gitIgnoreEntries, v.File)
		case DecisionAllow:
			plan.allowEntries = appendUnique(plan.allowEntries, v.File)
			plan.allowGitIgnoreLines = appendUnique(plan.allowGitIgnoreLines, "!"+v.File)
		case DecisionAllowFolder:
			folder := gitIgnoreFolderEntry(v.File)
			if folder == "" {
				// File is at the repo root — fall back to a per-file allow.
				plan.allowEntries = appendUnique(plan.allowEntries, v.File)
				plan.allowGitIgnoreLines = appendUnique(plan.allowGitIgnoreLines, "!"+v.File)
			} else {
				plan.allowEntries = appendUnique(plan.allowEntries, folder+"**")
				plan.allowGitIgnoreLines = appendUnique(plan.allowGitIgnoreLines, "!"+folder)
			}
		default:
			return decisionPlan{}, fmt.Errorf("unknown gitignore decision %v for %q", d, v.File)
		}

		// Collect the files this decision directly covers.
		covered := coveredFiles(d, v, remaining)
		switch d {
		case DecisionGitIgnorePattern, DecisionGitIgnoreFolder, DecisionGitIgnoreFile:
			plan.unstage = appendUniqueAll(plan.unstage, covered)
		case DecisionAllow, DecisionAllowFolder:
			plan.allowed = appendUniqueAll(plan.allowed, covered)
		}

		// Rebuild the remaining set by re-evaluating every originally-staged
		// file against the accumulated patterns / allow. Files newly covered
		// by this round's decision drop out automatically.
		patterns := append(append([]string(nil), plan.basePatterns...), plan.gitIgnoreEntries...)
		allow := append(append([]string(nil), plan.baseAllow...), plan.allowEntries...)
		next, err := EvaluateGitIgnoreMatches(p.StagedFiles, patterns, allow)
		if err != nil {
			return decisionPlan{}, err
		}
		remaining = filterDecided(next, plan.unstage, plan.allowed)
	}
	return plan, nil
}

// decisionPlan accumulates the outcome of the per-decision loop so apply
// writes disk state exactly once at the end.
type decisionPlan struct {
	basePatterns        []string
	baseAllow           []string
	gitIgnoreEntries    []string
	allowEntries        []string
	allowGitIgnoreLines []string // negation lines (`!path`, `!folder/`) added to .gitignore as overrides
	unstage             []string
	allowed             []string
	cancelled           bool
}

// countFolderSiblings returns how many of the other violations live in the
// same directory as file. Returns 0 for root-level files because a "folder"
// decision would not be offered there.
func countFolderSiblings(file string, others []Violation) int {
	dir := gitIgnoreFolderEntry(file)
	if dir == "" {
		return 0
	}
	n := 0
	for _, o := range others {
		if strings.HasPrefix(filepath.ToSlash(o.File)+"/", dir) {
			n++
		}
	}
	return n
}

// coveredFiles reports which of the current remaining violations this one
// decision unstages/allows. DecisionGitIgnoreFile and DecisionAllow cover only
// the chosen file; folder / pattern decisions cover every matching sibling.
func coveredFiles(d Decision, v Violation, remaining []Violation) []string {
	switch d {
	case DecisionGitIgnoreFile, DecisionAllow:
		return []string{v.File}
	case DecisionGitIgnoreFolder, DecisionAllowFolder:
		dir := gitIgnoreFolderEntry(v.File)
		if dir == "" {
			return []string{v.File}
		}
		out := []string{v.File}
		for _, o := range remaining[1:] {
			if strings.HasPrefix(filepath.ToSlash(o.File)+"/", dir) {
				out = append(out, o.File)
			}
		}
		return out
	case DecisionGitIgnorePattern:
		out := []string{v.File}
		for _, o := range remaining[1:] {
			if o.Pattern == v.Pattern {
				out = append(out, o.File)
			}
		}
		return out
	default:
		return nil
	}
}

func filterDecided(violations []Violation, unstaged, allowed []string) []Violation {
	decided := make(map[string]struct{}, len(unstaged)+len(allowed))
	for _, f := range unstaged {
		decided[f] = struct{}{}
	}
	for _, f := range allowed {
		decided[f] = struct{}{}
	}
	out := violations[:0]
	for _, v := range violations {
		if _, ok := decided[v.File]; ok {
			continue
		}
		out = append(out, v)
	}
	return out
}

func appendUnique(dst []string, v string) []string {
	for _, x := range dst {
		if x == v {
			return dst
		}
	}
	return append(dst, v)
}

func appendUniqueAll(dst, src []string) []string {
	for _, v := range src {
		dst = appendUnique(dst, v)
	}
	return dst
}

func formatViolationsError(violations []Violation) error {
	lines := make([]string, 0, len(violations)+1)
	lines = append(lines, "staged files match commit.gitignore patterns:")
	for _, v := range violations {
		lines = append(lines, fmt.Sprintf("  - %s (pattern: %s)", v.File, v.Pattern))
	}
	return errors.New(strings.Join(lines, "\n"))
}

func applyPlan(p CheckParams, plan decisionPlan) (CheckOutcome, error) {
	var outcome CheckOutcome

	gitRoot := p.GitRoot
	if gitRoot == "" {
		gitRoot = p.WorkDir
	}

	if len(plan.unstage) > 0 {
		appended, err := appendGitIgnore(gitRoot, plan.gitIgnoreEntries)
		if err != nil {
			return CheckOutcome{}, fmt.Errorf("append .gitignore: %w", err)
		}
		if err := resetFiles(p.WorkDir, plan.unstage); err != nil {
			return CheckOutcome{}, fmt.Errorf("unstage: %w", err)
		}
		outcome.Unstaged = plan.unstage
		outcome.GitIgnored = appended
	}

	if len(plan.allowed) > 0 {
		saveDir := p.SaveDir
		if saveDir == "" {
			saveDir = gitRoot
		}
		added, err := appendAllow(saveDir, plan.allowEntries)
		if err != nil {
			return CheckOutcome{}, fmt.Errorf("update .gavel.yaml: %w", err)
		}
		outcome.Allowed = added

		negated, err := appendGitIgnoreAllow(gitRoot, plan.allowGitIgnoreLines)
		if err != nil {
			return CheckOutcome{}, fmt.Errorf("append .gitignore allow override: %w", err)
		}
		outcome.GitIgnored = appendUniqueAll(outcome.GitIgnored, negated)
	}

	return outcome, nil
}

// appendGitIgnore appends lines to {gitRoot}/.gitignore, preserving any
// existing content. Lines already present (exact match) are skipped. Returns
// the subset of entries actually written so callers can report them.
func appendGitIgnore(gitRoot string, entries []string) ([]string, error) {
	return appendGitIgnoreLines(gitRoot, entries)
}

// appendGitIgnoreAllow appends explicit-allow negation lines (`!entry`) to
// {gitRoot}/.gitignore. Entries already prefixed with "!" pass through as-is;
// otherwise the prefix is added. The negations live in the same .gitignore so
// the override is self-documenting and survives a future broad ignore.
func appendGitIgnoreAllow(gitRoot string, entries []string) ([]string, error) {
	negated := make([]string, 0, len(entries))
	for _, e := range entries {
		if e == "" {
			continue
		}
		if strings.HasPrefix(e, "!") {
			negated = append(negated, e)
			continue
		}
		negated = append(negated, "!"+e)
	}
	return appendGitIgnoreLines(gitRoot, negated)
}

func appendGitIgnoreLines(gitRoot string, entries []string) ([]string, error) {
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
		text := fmt.Sprintf("Add folder %q to .gitignore (unstage now)", dir)
		if v.FolderSiblings > 0 {
			text = fmt.Sprintf("Add folder %q to .gitignore (unstages %d files)", dir, v.FolderSiblings+1)
		}
		choices = append(choices, gitIgnoreChoice{
			Text:     text,
			Decision: DecisionGitIgnoreFolder,
		})
	}
	choices = append(choices,
		gitIgnoreChoice{
			Text:     fmt.Sprintf("Add file %q to .gitignore (unstage now)", v.File),
			Decision: DecisionGitIgnoreFile,
		},
		gitIgnoreChoice{
			Text:     fmt.Sprintf("Allow this file (%q in .gavel.yaml + !%s in .gitignore)", v.File, v.File),
			Decision: DecisionAllow,
		},
	)
	if folder := gitIgnoreFolderEntry(v.File); folder != "" {
		choices = append(choices, gitIgnoreChoice{
			Text:     fmt.Sprintf("Allow folder %q (and everything under it: !%s in .gitignore + %s** in .gavel.yaml)", folder, folder, folder),
			Decision: DecisionAllowFolder,
		})
	}
	choices = append(choices,
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
