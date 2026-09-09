package types

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	captainapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/ghodss/yaml"
	"github.com/samber/lo"
)

func (todo TODO) AsYaml() (string, error) {
	// Serialize frontmatter
	frontmatterYAML, err := yaml.Marshal(&todo.TODOFrontmatter)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// Create file content
	var content strings.Builder
	content.WriteString("---\n")
	content.WriteString(string(frontmatterYAML))
	content.WriteString("---\n")

	// Add markdown content if available
	if todo.Implementation != "" {
		content.WriteString("\n")
		content.WriteString(todo.Implementation)
	}
	return content.String(), nil
}

func (t TODO) PrettyRow(opts interface{}) map[string]api.Text {
	title := t.Title
	if title == "" {
		title = t.Filename()
	}
	row := map[string]api.Text{
		"Title":    clicky.Text(title, "order-2"),
		"Status":   t.Status.Pretty().Styles("order-3"),
		"Priority": t.Priority.Pretty().Styles("order-4"),
	}
	if id := t.DisplayID(); id != "" {
		row["ID"] = clicky.Text(id, "order-0 text-muted")
	}
	if t.Workspace != "" {
		row["Workspace"] = clicky.Text(t.Workspace, "order-1 text-muted")
	}
	if t.LastRun != nil {
		row["Updated"] = clicky.Text(summaryAge(time.Since(*t.LastRun)), "order-5 text-muted")
	}
	// Always emitted, even when empty: clicky's NewTableFromRows takes its
	// headers from the first row alone, so a conditional column would appear or
	// vanish depending on whether the first todo happened to carry a label.
	// A predictable empty cell beats a column that comes and goes with sort order.
	row["Labels"] = t.labelChips().Text().Styles("order-9")
	// The phase columns are emitted unconditionally for the same reason, and in
	// pipeline order so a row reads left to right as the todo progressed.
	for index, phase := range Phases {
		row[phaseColumn(phase)] = t.phaseCell(phase).Styles(fmt.Sprintf("order-%d", 5+index))
	}
	return row
}

// phaseColumn is the header a phase renders under. Capitalised because clicky
// prints the map key verbatim.
func phaseColumn(phase Phase) string {
	return strings.ToUpper(string(phase)[:1]) + string(phase)[1:]
}

// labelChips returns the TODO's labels as resolved presentations. It prefers
// the definitions the provider resolved; a TODO from a source without a
// definition store still renders with the hashed palette colour rather than
// silently losing its labels.
func (t TODO) labelChips() labels.Definitions {
	if len(t.LabelDefinitions) > 0 {
		return t.LabelDefinitions
	}
	return labels.Derive(t.Labels)
}

// summaryAge renders only the largest useful unit so list timestamps stay
// scannable instead of showing exact multi-unit durations.
func summaryAge(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// GetID and GetName satisfy clicky's EntityItem, which is what lets TODOs be
// declared as an entity and get their CLI, REST routes and action catalog
// generated from one declaration. GetID returns the canonical id rather than
// DisplayID: it is the value a caller passes back to address this TODO.
func (t TODO) GetID() string { return t.ID }

func (t TODO) GetName() string { return t.Title }

func (t TODO) DisplayID() string {
	if t.ShortID != "" {
		return t.ShortID
	}
	if t.ID == "" {
		return ""
	}
	if len(t.ID) > 8 {
		return t.ID[:8]
	}
	return t.ID
}

func (t TODO) Filename() string {
	if t.Provider != "" && t.ID != "" {
		if t.Title != "" {
			return t.Title
		}
		if len(t.ID) > 8 {
			return t.ID[:8]
		}
		return t.ID
	}
	if t.FilePath == "" {
		if t.Title != "" {
			return t.Title
		}
		if t.ID != "" {
			return t.ID
		}
		return ""
	}
	file := filepath.Base(t.FilePath)
	return lo.PascalCase(file)
}

// Pretty returns a formatted text representation of the TODO item
func (t TODO) Pretty() api.Text {
	result := api.Text{}

	// Add title/file path
	if t.FilePath != "" {
		result = result.Add(icons.File).Append(" ", "").Append(t.Filename(), "text-blue-600 font-medium")
	}

	// Add priority if set
	if t.Priority != "" {
		result = result.Add(t.Priority.Pretty())
	}

	// Add status if set
	if t.Status != "" {
		result = result.Append(" ", "").Add(t.Status.Pretty())
	}

	return result
}

// PrettyDetailed returns a comprehensive formatted representation of the TODO with all sections and metadata.
func (t TODO) PrettyDetailed() api.Text {
	result := api.Text{}

	// Header line with filename, priority, status
	header := t.Pretty()
	result = result.Add(header).NewLine().NewLine()

	// File path
	result = result.Append("File: ", "text-gray-500").Append(t.FilePath, "").NewLine()

	if t.ID != "" {
		result = result.Append("ID: ", "text-gray-500").Append(t.ID, "").NewLine()
	}

	if t.Provider != "" {
		result = result.Append("Provider: ", "text-gray-500").Append(t.Provider, "").NewLine()
	}

	if chips := t.labelChips(); len(chips) > 0 {
		result = result.Append("Labels: ", "text-gray-500").Add(chips.Text()).NewLine()
	}

	// Only phases that have actually run are listed: a detail view has room to
	// say what happened, and silence is the honest rendering of a phase that
	// never started.
	for _, run := range t.PhaseRuns.Ordered() {
		result = result.Append(phaseColumn(run.Phase)+": ", "text-gray-500").Add(run.Pretty()).NewLine()
	}

	// Language
	if t.Language != "" {
		result = result.Append("Language: ", "text-gray-500").Add(t.Language.Pretty()).NewLine()
	}

	// Attempts
	if t.Attempts > 0 {
		result = result.Append("Attempts: ", "text-gray-500").Append(fmt.Sprintf("%d", t.Attempts), "").NewLine()
	}

	// Last run
	if t.LastRun != nil {
		result = result.Append("Last Run: ", "text-gray-500").Append(t.LastRun.Format(time.RFC3339), "").NewLine()
	}

	result = result.NewLine()

	if strings.TrimSpace(t.MarkdownBody) != "" {
		result = result.Append("Issue Body", "text-blue-600 font-bold").NewLine()
		result = result.Append(strings.TrimSpace(t.MarkdownBody), "").NewLine().NewLine()
	}

	if len(t.ProviderEvents) > 0 {
		result = result.Add(formatProviderEvents(t.ProviderEvents)).NewLine()
	}

	// Show fixture tests from the FileNode tree
	if t.FileNode != nil {
		result = result.Add(formatFixtureTree(t.FileNode, 0))
	}

	// LLM Configuration section
	if t.LLM != nil {
		result = result.Add(icons.Lambda).Append(" LLM Configuration", "text-blue-600 font-bold").NewLine()
		if t.LLM.Model != "" {
			result = result.Append("  Model: ", "text-gray-500").Append(t.LLM.Model, "").NewLine()
		}
		if t.LLM.MaxTokens > 0 {
			result = result.Append("  Max Tokens: ", "text-gray-500").Append(fmt.Sprintf("%d", t.LLM.MaxTokens), "").NewLine()
		}
		if t.LLM.SessionId != "" {
			result = result.Append("  Session ID: ", "text-gray-500").Append(t.LLM.SessionId, "").NewLine()
		}
	}

	return result
}

func formatProviderEvents(events []ProviderEvent) api.Text {
	result := api.Text{}
	result = result.Append("Event History", "text-blue-600 font-bold").NewLine()
	for _, event := range events {
		id := event.ShortID
		if id == "" && event.ID != "" {
			if len(event.ID) > 8 {
				id = event.ID[:8]
			} else {
				id = event.ID
			}
		}
		line := "  - "
		if !event.Timestamp.IsZero() {
			line += event.Timestamp.Format(time.RFC3339) + " "
		}
		if event.Kind != "" {
			line += event.Kind
		}
		if id != "" {
			line += " [" + id + "]"
		}
		if event.Actor != "" {
			line += " by " + event.Actor
		}
		result = result.Append(line, "").NewLine()
		switch {
		case event.OldLabel != "" || event.NewLabel != "":
			result = result.Append("    Label: ", "text-gray-500").Append(event.OldLabel+" → "+event.NewLabel, "").NewLine()
		case event.Label != "":
			result = result.Append("    Label: ", "text-gray-500").Append(event.Label, "").NewLine()
		}
		if event.Title != "" {
			result = result.Append("    Title: ", "text-gray-500").Append(event.Title, "").NewLine()
		}
		if strings.TrimSpace(event.Body) != "" {
			result = result.Append("    Body:", "text-gray-500").NewLine()
			for _, line := range strings.Split(strings.TrimSpace(event.Body), "\n") {
				result = result.Append("      "+line, "").NewLine()
			}
		}
	}
	return result
}

// formatFixtureTree recursively formats a fixture node tree
func formatFixtureTree(node *fixtures.FixtureNode, depth int) api.Text {
	result := clicky.Text("", "")

	// Skip the root file node itself
	if depth > 0 {
		indent := strings.Repeat("  ", depth-1)

		// Show section header
		if node.Name != "" && node.Type == fixtures.SectionNode {
			result = result.Append(indent, "").Add(icons.Folder).Append(" ", "").
				Append(node.Name, "text-blue-600 font-bold").NewLine()
		}

		// Show test
		if node.Test != nil {
			result = result.Append(indent, "").Append("  \u2514\u2500 ", "text-gray-500").
				Append(node.Test.Name, "").NewLine()
		}
	}

	// Recursively format children
	for _, child := range node.Children {
		result = result.Add(formatFixtureTree(child, depth+1))
	}

	return result
}

// Pretty returns a formatted text representation of the TODOFrontmatter
func (f TODOFrontmatter) Pretty() api.Text {
	result := api.Text{}

	// Start with a simple metadata display
	result = result.Append("📋 TODO Metadata", "text-blue-600 font-bold")

	// Add metadata line
	var metadata []api.Text
	if f.Priority != "" {
		metadata = append(metadata, clicky.Text("Priority: ", "text-gray-500").Add(f.Priority.Pretty()))
	}
	if f.Status != "" {
		metadata = append(metadata, clicky.Text("Status: ", "text-gray-500").Add(f.Status.Pretty()))
	}
	if f.Language != "" {
		metadata = append(metadata, clicky.Text("Language: ", "text-gray-500").Add(f.Language.Pretty()))
	}
	if f.Attempts > 0 {
		metadata = append(metadata, clicky.Text("Attempts: ", "text-gray-500").Append(fmt.Sprintf("%d", f.Attempts), "text-orange-600"))
	}
	if f.LastRun != nil {
		metadata = append(metadata, clicky.Text("Last Run: ", "text-gray-500").Append(f.LastRun.Format("2006-01-02 15:04"), "text-purple-600"))
	}

	if len(metadata) > 0 {
		result = result.Append("\n", "")
		for i, meta := range metadata {
			if i > 0 {
				result = result.Append(" | ", "text-gray-400")
			}
			result = result.Add(meta)
		}
	}

	return result
}

// Pretty returns a formatted text representation of the Priority with appropriate styling
func (p Priority) Pretty() api.Text {
	switch p {
	case PriorityHigh:
		return clicky.Text("").Add(icons.Error).Append(" HIGH", "text-red-600 font-bold")
	case PriorityMedium:
		return clicky.Text("").Add(icons.Warning).Append(" MEDIUM", "text-yellow-600 font-medium")
	case PriorityLow:
		return clicky.Text("").Add(icons.Pass).Append(" LOW", "text-green-600")
	default:
		return clicky.Text(string(p), "text-gray-500")
	}
}

// Pretty returns a formatted text representation of the Status with color coding
func (s Status) Pretty() api.Text {
	switch s {
	case StatusDraft:
		return clicky.Text("").Add(icons.Info).Append(" DRAFT", "text-gray-500")
	case StatusPending:
		return clicky.Text("").Add(icons.Info).Append(" PENDING", "text-gray-600")
	case StatusInProgress:
		return clicky.Text("").Add(icons.ArrowRight).Append(" IN PROGRESS", "text-blue-600 font-medium")
	case StatusReview:
		return clicky.Text("").Add(icons.Search).Append(" REVIEW", "text-amber-600 font-medium")
	case StatusAsk:
		return clicky.Text("").Add(icons.QuestionRed).Append(" ASK", "text-purple-600 font-medium")
	case StatusVerified:
		return clicky.Text("").Add(icons.Pass).Append(" VERIFIED", "text-emerald-600 font-medium")
	case StatusUnverified:
		return clicky.Text("").Add(icons.Warning).Append(" UNVERIFIED", "text-orange-600 font-medium")
	case StatusCompleted:
		return clicky.Text("").Add(icons.Pass).Append(" COMPLETED", "text-green-600 font-bold")
	case StatusFailed:
		return clicky.Text("").Add(icons.Fail).Append(" FAILED", "text-red-600 font-bold")
	case StatusSkipped:
		return clicky.Text("").Add(icons.Skip).Append(" SKIPPED", "text-yellow-600")
	default:
		return clicky.Text(string(s), "text-gray-500")
	}
}

// Pretty returns a formatted text representation of the Language with styling
func (l Language) Pretty() api.Text {
	switch l {
	case LanguageGo:
		return clicky.Text("").Add(icons.Package).Append(" Go", "text-blue-600 font-medium")
	case LanguageTypeScript:
		return clicky.Text("").Add(icons.File).Append(" TypeScript", "text-blue-500 font-medium")
	case LanguagePython:
		return clicky.Text("").Add(icons.Lambda).Append(" Python", "text-yellow-600 font-medium")
	default:
		return clicky.Text(string(l), "text-gray-500")
	}
}

// Pretty returns a formatted text representation of the TestResultInfo for markdown output.
func (r TestResultInfo) Pretty() api.Text {
	t := api.Text{Content: "## Latest Failure"}.NewLine().NewLine()

	t = t.Add(api.KeyValuePair{Key: "Run", Value: r.Timestamp.Format(time.RFC3339)}).NewLine()

	if r.Command != "" {
		t = t.Add(api.KeyValuePair{Key: "Command", Value: "`" + r.Command + "`"}).NewLine()
	}

	t = t.Add(api.KeyValuePair{Key: "CWD", Value: "`" + r.CWD + "`"}).NewLine()

	if r.GitBranch != "" {
		t = t.Add(api.KeyValuePair{Key: "Branch", Value: "`" + r.GitBranch + "`"}).NewLine()
	}

	if r.GitCommit != "" {
		commitVal := "`" + r.GitCommit + "`"
		if r.GitDirty {
			commitVal += " (dirty)"
		}
		t = t.Add(api.KeyValuePair{Key: "Commit", Value: commitVal}).NewLine()
	}

	resultStr := "PASSED"
	if !r.Passed {
		resultStr = "FAILED"
	}
	t = t.Add(api.KeyValuePair{
		Key:   "Result",
		Value: fmt.Sprintf("%s (%s)", resultStr, r.Duration.Round(time.Millisecond)),
	}).NewLine()

	if r.Output != "" {
		t = t.NewLine().Add(api.Code{Content: r.Output})
	}

	return t
}

// Pretty returns a formatted text representation of the CheckResult
func (c CheckResult) Pretty() api.Text {
	result := c.TODO.Pretty()

	if c.Error != nil {
		return result.Append(" ", "").Add(icons.Fail).Append(fmt.Sprintf(" Error: %v", c.Error), "text-red-600")
	}

	summary := captainapi.VerifySummary{}
	if c.Report != nil {
		summary = c.Report.Summary
	}
	counts := fmt.Sprintf(" %d/%d checks passed", summary.Passed, summary.Total)
	if c.AllPassed {
		result = result.Append(" ", "").Add(icons.Pass).Append(counts, "text-green-600")
	} else {
		result = result.Append(" ", "").Add(icons.Fail).Append(counts, "text-red-600")
	}

	if c.Duration > 0 {
		result = result.Append(fmt.Sprintf(" (%s)", c.Duration.Round(time.Millisecond)), "text-gray-500")
	}
	if c.Report == nil {
		return result
	}
	return result.NewLine().Add(VerifyReportText(*c.Report, "  "))
}
