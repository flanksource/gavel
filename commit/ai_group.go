package commit

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	dotprompt "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/internal/prompting"
)

//go:embed commit-grouping.prompt
var groupingPromptTemplate string

// groupingPromptFile is reported as PromptRequest.Source so the AI logging
// middleware identifies the template under -v.
const groupingPromptFile = "commit-grouping.prompt"

// groupChangesByAIFunc is the seam tests stub to exercise the -A grouping flow
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

// renderGroupingPrompt renders the grouping template around the status table and
// returns the user prompt, the JSON output schema (with the runtime MaxCommits cap
// baked in as maxItems on the groups array), and the SchemaStrictness policy — all
// declared in the .prompt frontmatter. maxCommits drives the Handlebars
// conditionals in both the body and the frontmatter schema.
func renderGroupingPrompt(template, table string, maxCommits int) (string, json.RawMessage, api.SchemaStrictness, error) {
	data := map[string]any{
		"table":      table,
		"maxCommits": maxCommits,
	}
	req, _, err := dotprompt.Load(template).Render(data, nil)
	if err != nil {
		return "", nil, "", fmt.Errorf("render grouping prompt: %w", err)
	}
	return req.Prompt.User, req.Prompt.SchemaJSON, req.Prompt.SchemaStrictness, nil
}

// buildStatusTable renders the staged changes as a markdown table (gavel status'
// columns: scope, file, status, +added/-deleted) for the grouping prompt, reusing
// gavel status' repomap scope labelling. Diffs are intentionally omitted — the
// grouping decision needs paths, scope and magnitude, not content; per-group diffs
// are sent later during message generation. Every change gets a row (changes absent
// from the gathered status fall back to the "general" scope) so the LLM can assign
// each path.
func buildStatusTable(workDir string, changes []stagedChange) (string, error) {
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
	sortGroupingRows(rows)

	table, err := clicky.Format(rows, clicky.FormatOptions{Markdown: true, NoColor: true})
	if err != nil {
		return "", fmt.Errorf("render grouping status table: %w", err)
	}
	return table, nil
}

// sortGroupingRows orders the table rows by file for deterministic output.
func sortGroupingRows(rows []groupingRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].File < rows[j].File
	})
}

// groupChangesByAI splits the staged changes into logical commit groups plus an
// ignore list: it renders a gavel-status table, asks the LLM to group it, then maps
// the response back onto the staged changes via assembleGroups. The MaxCommits cap
// is declared as maxItems on the groups array in the .prompt output schema and
// enforced by captain's schemaStrictness=retry policy (bounded fix-up re-asks,
// then a hard error). It builds its own agent (like generateCommitAnalysis) so the
// grouping seam can be stubbed in tests without an LLM.
func groupChangesByAI(ctx context.Context, opts Options, source stagedSource) ([]commitGroup, error) {
	table, err := buildStatusTable(opts.WorkDir, source.Changes)
	if err != nil {
		return nil, err
	}

	agent, err := BuildAgent(opts, opts.groupModel())
	if err != nil {
		return nil, err
	}

	template, err := opts.Config.Grouping.TemplateSource(opts.WorkDir, groupingPromptTemplate)
	if err != nil {
		return nil, fmt.Errorf("resolve commit.grouping prompt override: %w", err)
	}

	prompting.Prepare()

	promptText, schemaJSON, strictness, err := renderGroupingPrompt(template, table, opts.MaxCommits)
	if err != nil {
		return nil, err
	}
	resp, err := agent.ExecutePrompt(ctx, clickyai.PromptRequest{
		Name:             "commit grouping",
		Source:           groupingPromptFile,
		Prompt:           promptText,
		SchemaJSON:       schemaJSON,
		SchemaStrictness: strictness,
	})
	if err != nil {
		return nil, fmt.Errorf("execute AI grouping prompt: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("AI grouping prompt returned error: %s", resp.Error)
	}
	var schema aiGroupingSchema
	if err := clickyai.DecodeStructured(resp, &schema); err != nil {
		return nil, fmt.Errorf("decode AI grouping response: %w", err)
	}

	return assembleGroups(source.Changes, schema), nil
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
