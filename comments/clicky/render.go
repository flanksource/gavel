package clicky

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/comments"
)

// listItem is one full-width cell: clicky's entity list renders rows only
// through TableProvider, so the whole thread is a single column.
var _ api.TableProvider = listItem{}

func (listItem) Columns() []api.ColumnDef {
	return []api.ColumnDef{api.Column("comment").Label("Comment").Build()}
}
func (c listItem) Row() map[string]any { return map[string]any{"comment": c.Pretty()} }

// listSelector is the union of the selectors hosts store on a target: code
// ranges (path + lines + quote), editor patches, and page element anchors.
type listSelector struct {
	StartLine int                   `json:"startLine"`
	EndLine   int                   `json:"endLine"`
	Quote     string                `json:"quote"`
	Patch     *comments.PatchAnchor `json:"patch"`
	Anchor    string                `json:"anchor"`
	Locators  []struct {
		Kind string `json:"kind"`
		Text string `json:"text"`
	} `json:"locators"`
}

type listContent struct {
	Page struct {
		URL string `json:"url"`
	} `json:"page"`
	AgentRun *struct {
		State  string `json:"state"`
		Error  string `json:"error"`
		Target struct {
			Label string `json:"label"`
		} `json:"target"`
	} `json:"agentRun"`
}

// Pretty renders a thread as its header, the lines it is anchored to, the
// full body, and only the metadata that is set.
func (c listItem) Pretty() api.Text {
	var selector listSelector
	_ = json.Unmarshal(c.Target.Selector, &selector) // hosts own the selector shape; unknown shapes render without a snippet
	var content listContent
	_ = json.Unmarshal(c.Content, &content) // content is free-form host JSON; only the known keys are shown

	out := api.Text{}.Append(c.Status, statusStyle(c.Comment)).Space().Append(c.location(selector, content), "font-bold")
	if c.Unread {
		out = out.Space().Append("●", "text-blue-500")
	}
	out = out.Space().Append(shortID(c.ID), "text-muted")
	if snippet := snippetOf(selector); !snippet.IsEmpty() {
		out = out.NewLine().Add(snippet)
	}
	out = out.NewLine().Append(strings.TrimSpace(c.Body)).NewLine().Add(c.details(content))
	for _, m := range c.Messages {
		out = out.NewLine().Append("↳ ", "text-muted").Append(actorName(m.Author), "font-bold").Space().Add(ago(m.CreatedAt).Styles("text-muted")).
			NewLine().Append("  " + strings.ReplaceAll(strings.TrimSpace(m.Body), "\n", "\n  "))
		if p := proposalOf(m); p != nil {
			out = out.NewLine().Append("  proposal: ", "text-muted").Append(string(p.State), "font-bold").Append(fmt.Sprintf(" · %d files", len(p.Files)), "text-muted")
		}
	}
	return out
}

// proposalOf is the patch a reply proposes, nil for a plain reply.
func proposalOf(m comments.Message) *comments.Proposal {
	var content map[string]json.RawMessage
	_ = json.Unmarshal(m.Content, &content) // content is free-form host JSON; only the proposal key is shown
	raw, ok := content[comments.ProposalContentKey]
	if !ok {
		return nil
	}
	var p comments.Proposal
	if err := json.Unmarshal(raw, &p); err != nil {
		return &comments.Proposal{State: comments.ProposalState("malformed: " + err.Error())}
	}
	return &p
}

// location is where the thread points: a project file with its line range, a
// page URL, or the raw target URI.
func (c listItem) location(s listSelector, content listContent) string {
	if rel, ok := strings.CutPrefix(c.Target.URI, c.Target.ScopeURI+"/files/"); ok {
		start, end := s.StartLine, s.EndLine
		if s.Patch != nil {
			start, end = s.Patch.StartLine, s.Patch.EndLine
		}
		switch {
		case start > 0 && end > start:
			return fmt.Sprintf("%s:%d-%d", rel, start, end)
		case start > 0:
			return fmt.Sprintf("%s:%d", rel, start)
		}
		return rel
	}
	if content.Page.URL != "" {
		return content.Page.URL
	}
	if key, ok := strings.CutPrefix(c.Target.URI, c.Target.ScopeURI+"/pages/"); ok {
		if page, err := url.PathUnescape(key); err == nil {
			return page
		}
	}
	return c.Target.URI
}

// details are the key: value pairs worth reading; empty values are omitted.
func (c listItem) details(content listContent) api.Text {
	type pair struct {
		key   string
		value api.Text
	}
	pairs := []pair{{"author", api.Text{}.Append(actorName(c.Author))}}
	if c.Assignee != nil {
		pairs = append(pairs, pair{"assignee", api.Text{}.Append(actorName(*c.Assignee))})
	}
	if run := content.AgentRun; run != nil {
		agent := api.Text{}.Append(strings.TrimSpace(run.State + " " + run.Target.Label))
		if run.Error != "" {
			agent = agent.Append(": "+run.Error, "text-red-500")
		}
		pairs = append(pairs, pair{"agent", agent})
	}
	if c.Resolution != nil && c.Resolution.Summary != "" {
		pairs = append(pairs, pair{"resolution", api.Text{}.Append(c.Resolution.Summary)})
	}
	if c.Outcome != nil && c.Outcome.Summary != "" {
		pairs = append(pairs, pair{"outcome", api.Text{}.Append(c.Outcome.Summary)})
	}
	if len(c.Attachments) > 0 {
		pairs = append(pairs, pair{"attachments", api.Text{}.Append(len(c.Attachments))})
	}
	pairs = append(pairs, pair{"updated", ago(c.UpdatedAt)})
	out := api.Text{}
	for i, kv := range pairs {
		if i > 0 {
			out = out.Append(" · ", "text-muted")
		}
		out = out.Append(kv.key+": ", "text-muted").Add(kv.value)
	}
	return out
}

// snippetOf numbers the anchored lines, with the patch's surrounding context muted.
func snippetOf(s listSelector) api.Text {
	switch {
	case s.Patch != nil && len(s.Patch.Lines) > 0:
		start := s.Patch.StartLine - len(s.Patch.Before)
		lines := append(append(append([]string{}, s.Patch.Before...), s.Patch.Lines...), s.Patch.After...)
		return numbered(lines, start, s.Patch.StartLine, s.Patch.EndLine)
	case s.Quote != "" && s.StartLine > 0:
		return numbered(strings.Split(strings.TrimSuffix(s.Quote, "\n"), "\n"), s.StartLine, s.StartLine, s.EndLine)
	}
	for _, l := range s.Locators {
		if l.Kind == "text" && strings.TrimSpace(l.Text) != "" {
			return api.Text{}.Append("> ", "text-muted").Append(strings.TrimSpace(l.Text), "italic")
		}
	}
	if s.Anchor != "" {
		return api.Text{}.Append("> ", "text-muted").Append(s.Anchor, "text-muted")
	}
	return api.Text{}
}

func numbered(lines []string, first, start, end int) api.Text {
	width := len(fmt.Sprint(first + len(lines) - 1))
	out := api.Text{}
	for i, line := range lines {
		if i > 0 {
			out = out.NewLine()
		}
		n := first + i
		if n >= start && n <= end {
			out = out.Append(fmt.Sprintf("> %*d | ", width, n), "text-yellow-600").Append(line)
		} else {
			out = out.Append(fmt.Sprintf("  %*d | ", width, n), "text-muted").Append(line, "text-muted")
		}
	}
	return out
}

func statusStyle(c comments.Comment) string {
	switch c.Phase {
	case comments.Active:
		return "text-yellow-600 font-bold"
	case comments.AwaitingOutcome:
		return "text-blue-500 font-bold"
	default:
		return "text-green-600 font-bold"
	}
}

func actorName(a comments.Actor) string {
	if a.Label != "" {
		return a.Label
	}
	return a.ID
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func ago(t time.Time) api.Text {
	return api.Human(time.Since(t).Truncate(time.Minute)).Append(" ago")
}
