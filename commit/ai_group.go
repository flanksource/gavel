package commit

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/clicky"
	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/prompting"
)

// groupChangesByAIFunc is the seam tests stub to exercise the --ai-group flow
// without an LLM. It mirrors analyzeCommitMessageWithAIFunc in commit.go.
var groupChangesByAIFunc func(ctx context.Context, opts Options, source stagedSource) ([]commitGroup, error) = groupChangesByAI

// gatherStatusFunc (declared in interactive.go) supplies the repomap-enriched
// status used for scope labelling; it is the same seam interactive staging uses,
// so tests can stub it to inject FileMaps without git or repomap.

const (
	choreGroupLabel   = "generated & lock files"
	choreGroupMessage = "chore: update generated and lock files"

	// scopeGeneralFallback labels changes absent from the gathered status (e.g.
	// filtered out) so every path still appears in the grouping table.
	scopeGeneralFallback = "general"

	// maxGroupingAttempts bounds the grouping prompt calls: one base attempt plus
	// up to two consolidation feedback rounds when the LLM exceeds MaxCommits.
	maxGroupingAttempts = 3
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

// groupingRow is one changed file rendered into the markdown status table fed to
// the grouping prompt. clicky.Format turns a []groupingRow into a markdown table.
type groupingRow struct {
	Scope  string `json:"scope" pretty:"label=Scope"`
	File   string `json:"file" pretty:"label=File"`
	Status string `json:"status" pretty:"label=Status"`
	Adds   int    `json:"adds" pretty:"label=+Adds"`
	Dels   int    `json:"dels" pretty:"label=-Dels"`
}

// buildGroupingPrompt assembles the grouping instructions around the rendered
// status table. The scope guidance and commit cap are parameterized: GroupByScope
// promotes scope to the primary commit boundary, otherwise scope is just a hint
// column; maxCommits is stated as a soft limit the LLM should consolidate toward.
func buildGroupingPrompt(table string, maxCommits int, groupByScope bool) string {
	var b strings.Builder
	b.WriteString("You are organizing a set of uncommitted changes into clean, logical git commits.\n\n")
	b.WriteString("The changed files are listed below with their code scope (derived from the repository map), git status, and +added/-deleted line counts.\n\n")
	b.WriteString(table)
	b.WriteString("\nProduce the final commit grouping. Rules:\n")
	b.WriteString("- Each group MUST represent a single high-level change — one feature, one fix, or one refactor. NEVER put unrelated changes in the same group.\n")
	if groupByScope {
		b.WriteString("- Treat each scope as the primary boundary: prefer keeping a scope's files together as one commit, splitting a scope only when it clearly contains unrelated features.\n")
	} else {
		b.WriteString("- Group by logical change, one feature/fix/refactor per commit. The Scope column is a hint, not a hard boundary — split or merge across scopes as the change intent dictates.\n")
	}
	if maxCommits > 0 {
		fmt.Fprintf(&b, "- Produce at most %d commits (the chore commit for lock/generated files does NOT count toward this limit). If there are more logical groups, merge the smallest related ones until you are within the limit.\n", maxCommits)
	}
	b.WriteString("- Keep a test file in the same group as the feature it tests.\n")
	b.WriteString("- Every changed file MUST appear in exactly one group's \"files\" or in \"ignore\".\n")
	b.WriteString("- Put lock files, build artifacts, generated bundles, and vendored output in \"ignore\" — they are committed separately as a chore commit and never analyzed.\n")
	b.WriteString("- Use repo-relative paths exactly as shown above (for renames, use the new path).\n")
	b.WriteString("- Give each group a short conventional-commit-style label describing its intent.")
	return b.String()
}

// buildStatusTable renders the staged changes as a markdown table (gavel status'
// columns: scope, file, status, +added/-deleted) for the grouping prompt, reusing
// gavel status' repomap scope labelling. Diffs are intentionally omitted — the
// grouping decision needs paths, scope and magnitude, not content; per-group diffs
// are sent later during message generation. Every change gets a row (changes absent
// from the gathered status fall back to the "general" scope) so the LLM can assign
// each path.
func buildStatusTable(workDir string, changes []stagedChange, groupByScope bool) (string, error) {
	res, err := gatherStatusFunc(workDir)
	if err != nil {
		return "", fmt.Errorf("gather status for grouping: %w", err)
	}

	scopeByPath := make(map[string]string)
	for _, group := range res.ScopeGroups() {
		for _, f := range group.Files {
			scopeByPath[f.Path] = group.Label
		}
	}

	rows := make([]groupingRow, 0, len(changes))
	for _, c := range changes {
		scope := scopeByPath[c.Path]
		if scope == "" {
			scope = scopeGeneralFallback
		}
		file := c.Path
		if c.Status == "renamed" && c.PreviousPath != "" {
			file = c.PreviousPath + " → " + c.Path
		}
		rows = append(rows, groupingRow{
			Scope:  scope,
			File:   file,
			Status: c.Status,
			Adds:   c.Adds,
			Dels:   c.Dels,
		})
	}
	sortGroupingRows(rows, groupByScope)

	table, err := clicky.Format(rows, clicky.FormatOptions{Markdown: true, NoColor: true})
	if err != nil {
		return "", fmt.Errorf("render grouping status table: %w", err)
	}
	return table, nil
}

// sortGroupingRows orders the table rows: by scope then file when grouping by
// scope (so a scope's files read as a block), else by file alone.
func sortGroupingRows(rows []groupingRow, groupByScope bool) {
	sort.SliceStable(rows, func(i, j int) bool {
		if groupByScope && rows[i].Scope != rows[j].Scope {
			return rows[i].Scope < rows[j].Scope
		}
		return rows[i].File < rows[j].File
	})
}

// groupChangesByAI splits the staged changes into logical commit groups plus an
// ignore list: it renders a gavel-status table, asks the LLM to group it, enforces
// the MaxCommits cap via a consolidation feedback loop, then maps the response back
// onto the staged changes via assembleGroups. It builds its own agent (like
// generateCommitAnalysis) so the grouping seam can be stubbed in tests without an LLM.
func groupChangesByAI(ctx context.Context, opts Options, source stagedSource) ([]commitGroup, error) {
	table, err := buildStatusTable(opts.WorkDir, source.Changes, opts.GroupByScope)
	if err != nil {
		return nil, err
	}

	agent, err := BuildAgent(opts, opts.groupModel())
	if err != nil {
		return nil, err
	}

	basePrompt := buildGroupingPrompt(table, opts.MaxCommits, opts.GroupByScope)
	prompting.Prepare()

	exec := func(feedback string) (aiGroupingSchema, error) {
		prompt := basePrompt
		if feedback != "" {
			prompt = basePrompt + "\n\n" + feedback
		}
		schema := &aiGroupingSchema{}
		resp, err := agent.ExecutePrompt(ctx, clickyai.PromptRequest{
			Name:             "commit grouping",
			Prompt:           prompt,
			StructuredOutput: schema,
		})
		if err != nil {
			return aiGroupingSchema{}, fmt.Errorf("execute AI grouping prompt: %w", err)
		}
		if resp.Error != "" {
			return aiGroupingSchema{}, fmt.Errorf("AI grouping prompt returned error: %s", resp.Error)
		}
		return *schema, nil
	}

	schema, err := groupWithMaxCommits(opts.MaxCommits, exec)
	if err != nil {
		return nil, err
	}

	return assembleGroups(source.Changes, schema), nil
}

// groupWithMaxCommits runs the grouping prompt then re-runs it with a consolidation
// feedback prompt while the (non-chore) group count exceeds maxCommits, up to
// maxGroupingAttempts. The chore/ignore commit does not count toward the limit. If
// the LLM still overshoots after the bounded retries the last grouping is returned
// and a warning is logged — never silently merged or dropped (CW-2).
func groupWithMaxCommits(maxCommits int, exec func(feedback string) (aiGroupingSchema, error)) (aiGroupingSchema, error) {
	schema, err := exec("")
	if err != nil {
		return aiGroupingSchema{}, err
	}

	for attempt := 1; attempt < maxGroupingAttempts && maxCommits > 0 && len(schema.Groups) > maxCommits; attempt++ {
		next, err := exec(buildConsolidationFeedback(schema, maxCommits))
		if err != nil {
			return aiGroupingSchema{}, err
		}
		schema = next
	}

	if maxCommits > 0 && len(schema.Groups) > maxCommits {
		logger.Warnf("ai-group: LLM returned %d commits after %d attempts; limit is %d. Committing as-is.",
			len(schema.Groups), maxGroupingAttempts, maxCommits)
	}
	return schema, nil
}

// buildConsolidationFeedback asks the LLM to merge its previous grouping down to
// the commit limit, echoing the prior groups so it can decide what to combine.
func buildConsolidationFeedback(schema aiGroupingSchema, maxCommits int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Your previous grouping produced %d commits, but the limit is %d. ", len(schema.Groups), maxCommits)
	fmt.Fprintf(&b, "Consolidate by merging the smallest, most related groups so there are at most %d commits. Keep every file assigned to exactly one group and keep the ignore list unchanged.\n\n", maxCommits)
	b.WriteString("Your previous groups were:\n")
	for _, g := range schema.Groups {
		fmt.Fprintf(&b, "- %s: %s\n", g.Label, strings.Join(g.Files, ", "))
	}
	return b.String()
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
