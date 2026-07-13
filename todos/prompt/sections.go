package prompt

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// buildTODOSection renders one TODO. grouped omits the per-todo PR context (the
// group framing carries it instead); number, when > 0, prefixes the heading with
// its position in the list so multi-todo runs read as a numbered checklist.
func buildTODOSection(todo *types.TODO, workDir string, grouped bool, number int) string {
	var section string

	if todo.Prompt != "" {
		section += fmt.Sprintf("## Prompt\n\n%s\n\n", todo.Prompt)
	}

	if !grouped && todo.PR != nil {
		section += "## PR Context\n\n"
		if todo.PR.URL != "" {
			section += fmt.Sprintf("- **PR:** [#%d](%s)\n", todo.PR.Number, todo.PR.URL)
		} else if todo.PR.Number > 0 {
			section += fmt.Sprintf("- **PR:** #%d\n", todo.PR.Number)
		}
		if todo.PR.Head != "" {
			section += fmt.Sprintf("- **Branch:** `%s` → `%s`\n", todo.PR.Head, todo.PR.Base)
		}
		if todo.PR.CommentAuthor != "" {
			section += fmt.Sprintf("- **Reviewer:** %s\n", todo.PR.CommentAuthor)
		}
		if todo.PR.CommentURL != "" {
			section += fmt.Sprintf("- **Comment:** %s\n", todo.PR.CommentURL)
		}
		section += "\n"
	}

	heading := todo.Title
	if heading == "" && len(todo.Path) > 0 {
		heading = todo.Path[0]
	}
	if heading != "" {
		if number > 0 {
			section += fmt.Sprintf("## %d. %s\n\n", number, heading)
		} else {
			section += fmt.Sprintf("## %s\n\n", heading)
		}
	}

	if refs := todo.PathRefs(); len(refs) > 0 && workDir != "" {
		for _, ref := range refs {
			src, err := ReadSourceLines(workDir, ref)
			if err != nil || src == "" {
				continue
			}
			lang := langFromExt(filepath.Ext(ref.File))
			section += fmt.Sprintf("```%s file=%s\n%s```\n\n", lang, ref.String(), src)
		}
	}

	if body := readTODOMarkdownBody(todo); body != "" {
		section += stripFileRefLine(body) + "\n\n"
	}

	section += buildCommentsSection(todo.ProviderEvents)

	if len(todo.StepsToReproduce) > 0 {
		section += "## Steps to Reproduce\n\nRun the following to reproduce the failure:\n\n"
		for _, node := range todo.StepsToReproduce {
			if node.Test != nil {
				section += fmt.Sprintf("```bash\n%s\n```\n\n", node.Test.String())
			}
		}
	}

	if todo.Implementation != "" {
		section += fmt.Sprintf("## Implementation\n\n%s\n\n", todo.Implementation)
	}

	if len(todo.Verification) > 0 {
		section += "## Verification\n\nAfter implementing your fix, verify it works by running:\n\n"
		for _, node := range todo.Verification {
			if node.Test != nil {
				section += fmt.Sprintf("```bash\n%s\n```\n\n", node.Test.String())
			}
		}
	}

	return section
}

// buildCommentsSection renders issue comments so the agent sees the discussion
// (clarifications, decisions, extra context) that accompanies the issue body.
// Only CommentAdded events with a non-empty body are included; other event kinds
// (label changes, status updates) are timeline noise for an implementation prompt.
func buildCommentsSection(events []types.ProviderEvent) string {
	var section string
	for _, event := range events {
		if event.Kind != "CommentAdded" {
			continue
		}
		body := strings.TrimSpace(event.Body)
		if body == "" {
			continue
		}
		if section == "" {
			section = "## Comments\n\n"
		}
		author := event.Actor
		if author == "" {
			author = "unknown"
		}
		section += fmt.Sprintf("**%s:**\n\n%s\n\n", author, body)
	}
	return section
}

var fileRefLineRegex = regexp.MustCompile(`(?m)^File: ` + "`[^`]+`" + `\s*\n?`)

func stripFileRefLine(body string) string {
	return strings.TrimSpace(fileRefLineRegex.ReplaceAllString(body, ""))
}

func langFromExt(ext string) string {
	switch ext {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".py":
		return "python"
	case ".sql":
		return "sql"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".sh", ".bash":
		return "bash"
	case ".rs":
		return "rust"
	case ".rb":
		return "ruby"
	default:
		return ""
	}
}

func readTODOMarkdownBody(todo *types.TODO) string {
	return strings.TrimSpace(todo.MarkdownBody)
}
