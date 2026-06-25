package commit

import (
	"context"
	"fmt"
	"strings"

	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/prompting"
)

// groupChangesByAIFunc is the seam tests stub to exercise the --ai-group flow
// without an LLM. It mirrors analyzeCommitMessageWithAIFunc in commit.go.
var groupChangesByAIFunc func(ctx context.Context, opts Options, source stagedSource) ([]commitGroup, error) = groupChangesByAI

// gatherStatusFunc (declared in interactive.go) supplies the repomap-enriched
// status used for scope-first grouping; it is the same seam interactive staging
// uses, so tests can stub it to inject FileMaps without git or repomap.

const (
	choreGroupLabel   = "generated & lock files"
	choreGroupMessage = "chore: update generated and lock files"
)

// aiGroup is one logical commit the LLM proposes.
type aiGroup struct {
	Label string   `json:"label" description:"conventional-commit-style label describing this group's intent, e.g. 'feat: agent check loop'"`
	Files []string `json:"files" description:"repo-relative paths belonging to this commit, exactly as listed in the summary"`
}

// aiGroupingSchema is the structured output handed to the LLM.
type aiGroupingSchema struct {
	Groups []aiGroup `json:"groups" description:"logical groups of related files; each becomes one commit"`
	Ignore []string  `json:"ignore" description:"paths to exclude from real commits (lock files, build artifacts, generated bundles, vendored output); committed separately as a chore commit"`
}

const aiGroupingPrompt = `You are organizing a set of uncommitted changes into clean, logical git commits.

The changed files below are already pre-grouped by code scope (derived from the repository map). Each "[scope: ...]" block is a starting point for one commit. Lines show status, path, and +added/-deleted line counts.

%s

Produce the final commit grouping. Rules:
- Each group MUST represent a single high-level change — one feature, one fix, or one refactor. NEVER put unrelated changes in the same group.
- Treat each scope as the primary boundary: prefer keeping a scope's files together as one commit.
- Split a scope into multiple groups only when it clearly contains unrelated features.
- Merge files from different scopes into one group ONLY when each is a small, related change and there are many such tiny scattered edits — otherwise keep scopes separate.
- Keep a test file in the same group as the feature it tests.
- Every changed file MUST appear in exactly one group's "files" or in "ignore".
- Put lock files, build artifacts, generated bundles, and vendored output in "ignore" — they are committed separately as a chore commit and never analyzed.
- Use repo-relative paths exactly as shown above (for renames, use the new path).
- Give each group a short conventional-commit-style label describing its intent.`

// buildChangeSummary renders a gavel-status-style listing of the staged changes
// for the grouping prompt: one line per change with status, path (rename arrow
// included) and +added/-deleted counts. Diffs are intentionally omitted — the
// grouping decision needs paths and magnitude, not content; per-group diffs are
// sent later during message generation.
func buildChangeSummary(changes []stagedChange) string {
	var b strings.Builder
	for _, c := range changes {
		path := c.Path
		if c.Status == "renamed" && c.PreviousPath != "" {
			path = c.PreviousPath + " -> " + c.Path
		}
		fmt.Fprintf(&b, "%-8s %s (+%d/-%d)\n", c.Status, path, c.Adds, c.Dels)
	}
	return b.String()
}

// buildScopeGroupedSummary renders the staged changes pre-grouped by repomap
// scope, reusing gavel status' bucketing as the first-line grouping. Each scope
// becomes a "[scope: <label>]" block of buildChangeSummary lines. Changes whose
// path is not present in the gathered status (e.g. filtered out) are omitted from
// the summary; assembleGroups' "other" safety net still commits them.
func buildScopeGroupedSummary(workDir string, changes []stagedChange) (string, error) {
	res, err := gatherStatusFunc(workDir)
	if err != nil {
		return "", fmt.Errorf("gather status for scope grouping: %w", err)
	}

	byPath := make(map[string]stagedChange, len(changes))
	for _, c := range changes {
		byPath[c.Path] = c
	}

	var b strings.Builder
	for _, group := range res.ScopeGroups() {
		var bucket []stagedChange
		for _, f := range group.Files {
			if c, ok := byPath[f.Path]; ok {
				bucket = append(bucket, c)
			}
		}
		if len(bucket) == 0 {
			continue
		}
		fmt.Fprintf(&b, "[scope: %s]\n%s\n", group.Label, buildChangeSummary(bucket))
	}

	return b.String(), nil
}

// groupChangesByAI splits the staged changes into logical commit groups plus an
// ignore list: it pre-groups by repomap scope (gavel status' bucketing), asks
// the LLM to sub-group/merge within that partition, then maps the response back
// onto the staged changes via assembleGroups. It builds its own agent (like
// generateCommitAnalysis) so the grouping seam can be stubbed in tests without
// an LLM.
func groupChangesByAI(ctx context.Context, opts Options, source stagedSource) ([]commitGroup, error) {
	summary, err := buildScopeGroupedSummary(opts.WorkDir, source.Changes)
	if err != nil {
		return nil, err
	}

	agent, err := BuildAgent(opts, opts.groupModel())
	if err != nil {
		return nil, err
	}

	schema := &aiGroupingSchema{}
	prompting.Prepare()
	resp, err := agent.ExecutePrompt(ctx, clickyai.PromptRequest{
		Name:             "commit grouping",
		Prompt:           fmt.Sprintf(aiGroupingPrompt, summary),
		StructuredOutput: schema,
	})
	if err != nil {
		return nil, fmt.Errorf("execute AI grouping prompt: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("AI grouping prompt returned error: %s", resp.Error)
	}

	return assembleGroups(source.Changes, *schema), nil
}

// assembleGroups maps an LLM grouping response back onto the staged changes. It
// is pure (no LLM) so it can be unit-tested directly. Guarantees:
//   - each change lands in exactly one group;
//   - ignore-listed changes form a single trailing chore group with a preset
//     message (no LLM call);
//   - any change the LLM neither grouped nor ignored is committed in a trailing
//     "other" group rather than silently dropped (CW-2: fail loud, never lose);
//   - unknown paths returned by the LLM are warned and skipped.
func assembleGroups(changes []stagedChange, schema aiGroupingSchema) []commitGroup {
	byPath := make(map[string]stagedChange, len(changes))
	for _, c := range changes {
		byPath[c.Path] = c
	}
	assigned := make(map[string]bool, len(changes))

	pick := func(paths []string, kind string) []stagedChange {
		var out []stagedChange
		for _, p := range paths {
			c, ok := byPath[p]
			if !ok {
				logger.Warnf("ai-group: LLM referenced unknown %s file %q; skipping", kind, p)
				continue
			}
			if assigned[c.Path] {
				continue
			}
			assigned[c.Path] = true
			out = append(out, c)
		}
		return out
	}

	var groups []commitGroup
	for _, g := range schema.Groups {
		picked := pick(g.Files, "group")
		if len(picked) == 0 {
			continue
		}
		groups = append(groups, commitGroup{Label: g.Label, Changes: picked})
	}

	chore := pick(schema.Ignore, "ignore")

	var other []stagedChange
	for _, c := range changes {
		if !assigned[c.Path] {
			other = append(other, c)
		}
	}
	if len(other) > 0 {
		paths := make([]string, len(other))
		for i, c := range other {
			paths[i] = c.Path
		}
		logger.Warnf("ai-group: %d file(s) not assigned to any group; committing as 'other': %s",
			len(other), strings.Join(paths, ", "))
		groups = append(groups, commitGroup{Label: "other", Changes: other})
	}

	if len(chore) > 0 {
		groups = append(groups, commitGroup{
			Label:   choreGroupLabel,
			Message: choreGroupMessage,
			Changes: chore,
		})
	}

	return groups
}
