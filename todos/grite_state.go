package todos

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// Grite has no frontmatter, so envelope-driven todo state maps onto its native
// primitives: short enums (plan status, run mode) become namespaced labels like
// the existing status:/priority:/session: ones, and long values (plan path,
// summary, questions) travel in a single marked comment whose trailing JSON
// blob round-trips exactly.

const (
	planLabelNamespace = "plan"
	modeLabelNamespace = "mode"
	agentStateMarker   = "<!-- gavel:state "
	agentStateSuffix   = " -->"
)

// griteAgentState is the marked-comment payload for values labels cannot carry.
type griteAgentState struct {
	PlanPath  string                `json:"planPath,omitempty"`
	Summary   string                `json:"summary,omitempty"`
	Questions []types.AgentQuestion `json:"questions,omitempty"`
}

// applyAgentStateUpdates handles the envelope-driven StateUpdate fields for a
// grite issue: label swaps for the enums, one marked comment for the rest.
func (p *GriteProvider) applyAgentStateUpdates(ctx context.Context, id string, todo *types.TODO, updates StateUpdate) error {
	if updates.PlanStatus != nil {
		if err := p.applyNamespacedLabel(ctx, id, todo, planLabelNamespace, string(*updates.PlanStatus)); err != nil {
			return err
		}
		todo.PlanStatus = *updates.PlanStatus
	}
	if updates.RunMode != nil {
		if err := p.applyNamespacedLabel(ctx, id, todo, modeLabelNamespace, string(*updates.RunMode)); err != nil {
			return err
		}
		todo.RunMode = *updates.RunMode
	}
	if updates.PlanPath == nil && updates.LastRunSummary == nil && updates.Questions == nil {
		return nil
	}
	if updates.PlanPath != nil {
		todo.PlanPath = *updates.PlanPath
	}
	if updates.LastRunSummary != nil {
		todo.LastRunSummary = *updates.LastRunSummary
	}
	if updates.Questions != nil {
		todo.Questions = *updates.Questions
	}
	comment, err := renderAgentStateComment(griteAgentState{
		PlanPath:  todo.PlanPath,
		Summary:   todo.LastRunSummary,
		Questions: todo.Questions,
	})
	if err != nil {
		return err
	}
	return p.comment(ctx, todo, comment)
}

// applyNamespacedLabel swaps the issue's <ns>:* label to value; an empty value
// just removes the namespace's labels.
func (p *GriteProvider) applyNamespacedLabel(ctx context.Context, id string, todo *types.TODO, ns, value string) error {
	want := ""
	if value != "" {
		want = ns + ":" + value
	}
	for _, label := range existingNamespacedLabels(todo.Labels, ns) {
		if label == want {
			continue
		}
		if _, err := p.run(ctx, "issue", "label", "remove", id, "--label", label, "--json"); err != nil {
			return err
		}
		todo.Labels = removeLabel(todo.Labels, label)
	}
	if want != "" && !hasLabel(todo.Labels, want) {
		if _, err := p.run(ctx, "issue", "label", "add", id, "--label", want, "--json"); err != nil {
			return err
		}
		todo.Labels = append(todo.Labels, want)
		sort.Strings(todo.Labels)
	}
	return nil
}

func existingNamespacedLabels(labels []string, ns string) []string {
	var out []string
	for _, label := range labels {
		if strings.HasPrefix(label, ns+":") {
			out = append(out, label)
		}
	}
	return out
}

func planStatusFromLabels(labels []string) types.PlanStatus {
	for _, label := range existingNamespacedLabels(labels, planLabelNamespace) {
		return types.PlanStatus(strings.TrimPrefix(label, planLabelNamespace+":"))
	}
	return ""
}

func runModeFromLabels(labels []string) types.RunMode {
	for _, label := range existingNamespacedLabels(labels, modeLabelNamespace) {
		return types.RunMode(strings.TrimPrefix(label, modeLabelNamespace+":"))
	}
	return ""
}

// renderAgentStateComment renders the human-readable agent state followed by
// the exact JSON blob the parse-back reads.
func renderAgentStateComment(s griteAgentState) (string, error) {
	var b strings.Builder
	b.WriteString("**Agent state**\n\n")
	if s.Summary != "" {
		b.WriteString("**Summary:** " + s.Summary + "\n\n")
	}
	if s.PlanPath != "" {
		b.WriteString("**Plan:** `" + s.PlanPath + "`\n\n")
	}
	if len(s.Questions) > 0 {
		b.WriteString("**Questions:**\n")
		for _, q := range s.Questions {
			b.WriteString("- " + q.Text + "\n")
		}
		b.WriteString("\n")
	}
	blob, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("marshal agent state: %w", err)
	}
	b.WriteString(agentStateMarker + string(blob) + agentStateSuffix)
	return b.String(), nil
}

// applyAgentStateFromEvents restores plan path/summary/questions from the LAST
// marked agent-state comment in the issue's event log.
func applyAgentStateFromEvents(todo *types.TODO, events []types.ProviderEvent) {
	var latest *griteAgentState
	for _, ev := range events {
		if ev.Kind != "CommentAdded" {
			continue
		}
		state, ok := parseAgentStateComment(ev.Body)
		if ok {
			latest = &state
		}
	}
	if latest == nil {
		return
	}
	todo.PlanPath = latest.PlanPath
	todo.LastRunSummary = latest.Summary
	todo.Questions = latest.Questions
}

func parseAgentStateComment(body string) (griteAgentState, bool) {
	idx := strings.LastIndex(body, agentStateMarker)
	if idx < 0 {
		return griteAgentState{}, false
	}
	rest := body[idx+len(agentStateMarker):]
	end := strings.Index(rest, agentStateSuffix)
	if end < 0 {
		return griteAgentState{}, false
	}
	var state griteAgentState
	if err := json.Unmarshal([]byte(rest[:end]), &state); err != nil {
		return griteAgentState{}, false
	}
	return state, true
}
