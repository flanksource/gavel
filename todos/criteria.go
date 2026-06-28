package todos

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

const acceptanceCriteriaSection = "Acceptance Criteria"

// criteriaInsertBefore keeps the criteria section above the operational sections
// when it is first added to a body that already has them.
var criteriaInsertBefore = []string{
	"Steps to Reproduce", "Implementation", "Verification",
	"Custom Validations", "Latest Failure", "Verification Result",
	"Attempts", "Failure History",
}

// criteriaSchema is the structured output for criteria generation: the static
// catalog checks the model judged applicable, plus any custom criteria.
type criteriaSchema struct {
	ApplicableChecks flexStringList `json:"applicable_checks" description:"IDs from the provided catalog that are relevant to verifying THIS issue is done. Omit any that do not apply."`
	CustomCriteria   flexStringList `json:"custom_criteria" description:"Additional functionality-specific acceptance criteria not covered by the catalog. Each a single, testable true/false assertion. Leave empty when the catalog already covers the issue."`
}

// flexStringList is a []string that also tolerates array items a model returns
// as objects (e.g. {"criterion": "..."}) or a wrapping non-array, so a model
// ignoring the string item type doesn't fail the whole response. It mirrors the
// lenient parsing in git.parseStringArrayResult.
type flexStringList []string

func (f *flexStringList) UnmarshalJSON(data []byte) error {
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		// Not an array (bare string, object, or null): yield nothing rather than
		// failing — the caller treats an empty list as "no criteria".
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s := flexString(item); s != "" {
			out = append(out, s)
		}
	}
	*f = out
	return nil
}

// flexString extracts a string from a JSON value that is either a bare string or
// an object carrying the text under a common key.
func flexString(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return strings.TrimSpace(s)
	}
	var obj map[string]any
	if json.Unmarshal(raw, &obj) == nil {
		for _, key := range []string{"criterion", "text", "description", "name", "title", "value", "id"} {
			if v, ok := obj[key].(string); ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
	}
	return ""
}

// Generate drafts acceptance criteria for an issue: the model is seeded with the
// static verify.AllChecks catalog, selects the applicable ones, and only adds
// custom criteria when the catalog doesn't cover the issue's specific behavior.
func Generate(ctx context.Context, agent clickyai.Agent, title, body string) ([]types.AcceptanceCriterion, error) {
	schema := &criteriaSchema{}
	prompting.Prepare()
	resp, err := agent.ExecutePrompt(ctx, clickyai.PromptRequest{
		Name:             "acceptance criteria: " + title,
		Prompt:           buildCriteriaPrompt(title, body),
		StructuredOutput: schema,
	})
	if err != nil {
		return nil, fmt.Errorf("execute acceptance-criteria prompt: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("acceptance-criteria prompt returned error: %s", resp.Error)
	}
	return criteriaFromSchema(schema), nil
}

func buildCriteriaPrompt(title, body string) string {
	var b strings.Builder
	b.WriteString("Define the acceptance criteria for the following issue — the concrete, testable\n")
	b.WriteString("conditions that must hold for it to be considered done.\n\n")
	b.WriteString("First, from the standard catalog below, choose only the checks that genuinely\n")
	b.WriteString("apply to THIS issue (return their ids in `applicable_checks`). Then add custom\n")
	b.WriteString("criteria in `custom_criteria` only for functionality the catalog does not cover.\n")
	b.WriteString("Keep custom criteria specific and verifiable; do not restate catalog checks.\n\n")
	b.WriteString("## Catalog\n")
	byCategory := verify.ChecksByCategory(verify.AllChecks)
	for _, cat := range verify.AllCategories {
		checks := byCategory[cat]
		if len(checks) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n### %s\n", cat)
		for _, c := range checks {
			fmt.Fprintf(&b, "- %s: %s\n", c.ID, c.Description)
		}
	}
	fmt.Fprintf(&b, "\n## Issue\n\n**Title:** %s\n\n%s\n", title, strings.TrimSpace(body))
	return b.String()
}

func criteriaFromSchema(schema *criteriaSchema) []types.AcceptanceCriterion {
	descriptions := checkDescriptions()
	var out []types.AcceptanceCriterion
	for _, id := range schema.ApplicableChecks {
		id = strings.TrimSpace(id)
		desc, ok := descriptions[id]
		if !ok {
			continue
		}
		out = append(out, types.AcceptanceCriterion{CheckID: id, Text: desc})
	}
	for _, c := range schema.CustomCriteria {
		c = strings.TrimSpace(c)
		if c != "" {
			out = append(out, types.AcceptanceCriterion{Text: c})
		}
	}
	return out
}

// ParseAcceptanceCriteria reads the "## Acceptance Criteria" checklist from a
// markdown body. A leading "<id>:" matching a verify.AllChecks id marks a
// selected static check; anything else is a custom criterion.
func ParseAcceptanceCriteria(body string) []types.AcceptanceCriterion {
	valid := checkDescriptions()
	var out []types.AcceptanceCriterion
	for _, line := range sectionLines(body, acceptanceCriteriaSection) {
		text, done, ok := parseChecklistItem(line)
		if !ok {
			continue
		}
		c := types.AcceptanceCriterion{Text: text, Done: done}
		if id, rest, found := strings.Cut(text, ":"); found {
			id = strings.TrimSpace(id)
			if _, known := valid[id]; known {
				c.CheckID = id
				c.Text = strings.TrimSpace(rest)
			}
		}
		out = append(out, c)
	}
	return out
}

// RenderCriteriaSection renders the criteria as a "## Acceptance Criteria"
// checklist, re-prefixing selected static checks with their id so they
// round-trip through ParseAcceptanceCriteria.
func RenderCriteriaSection(criteria []types.AcceptanceCriterion) string {
	var b strings.Builder
	b.WriteString("## " + acceptanceCriteriaSection + "\n\n")
	for _, c := range criteria {
		box := "[ ]"
		if c.Done {
			box = "[x]"
		}
		text := c.Text
		if c.CheckID != "" {
			text = c.CheckID + ": " + c.Text
		}
		fmt.Fprintf(&b, "- %s %s\n", box, text)
	}
	return b.String()
}

// UpsertCriteriaSection replaces (or inserts) the criteria section in body.
func UpsertCriteriaSection(body string, criteria []types.AcceptanceCriterion) string {
	return ReplaceOrAppendSection(body, acceptanceCriteriaSection, RenderCriteriaSection(criteria), criteriaInsertBefore...)
}

func checkDescriptions() map[string]string {
	out := make(map[string]string, len(verify.AllChecks))
	for _, c := range verify.AllChecks {
		out[c.ID] = c.Description
	}
	return out
}

// sectionLines returns the lines inside the "## <header>" section (excluding the
// heading), stopping at the next "## " heading or end of body.
func sectionLines(body, header string) []string {
	want := "## " + header
	var out []string
	inSection := false
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == want || strings.HasPrefix(trimmed, want+" ") {
			inSection = true
			continue
		}
		if inSection {
			if strings.HasPrefix(trimmed, "## ") {
				break
			}
			out = append(out, line)
		}
	}
	return out
}

// parseChecklistItem extracts the text and checked state from a markdown list
// item ("- [ ] text", "- [x] text", or "- text"). ok is false for non-items.
func parseChecklistItem(line string) (text string, done bool, ok bool) {
	trimmed := strings.TrimSpace(line)
	for _, bullet := range []string{"- ", "* "} {
		if !strings.HasPrefix(trimmed, bullet) {
			continue
		}
		rest := strings.TrimSpace(trimmed[len(bullet):])
		switch {
		case strings.HasPrefix(rest, "[ ] "):
			return strings.TrimSpace(rest[4:]), false, true
		case strings.HasPrefix(rest, "[x] "), strings.HasPrefix(rest, "[X] "):
			return strings.TrimSpace(rest[4:]), true, true
		default:
			if rest != "" {
				return rest, false, true
			}
		}
	}
	return "", false, false
}
