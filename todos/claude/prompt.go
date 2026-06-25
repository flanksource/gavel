package claude

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// BuildPrompt constructs a structured prompt from a TODO for Claude Code execution.
func BuildPrompt(todo *types.TODO, workDir string) string {
	prompt := "You are fixing a failing test in a Go codebase.\n\n"
	prompt += buildTODOSection(todo, workDir, false, 0)
	prompt += singleTODOInstructions
	return prompt
}

// BuildGroupPrompt constructs a combined prompt for multiple related TODOs,
// structured as intro → numbered list of items → outro instructions. A single
// item keeps the plain framing (no numbering); several items are numbered so the
// agent can address each in turn.
func BuildGroupPrompt(todoList []*types.TODO, workDir string) string {
	var prompt string
	if len(todoList) > 1 {
		prompt = fmt.Sprintf("You are implementing the %d todo items listed below. Each is a separate task — implement ALL of them.\n\n", len(todoList))
	} else {
		prompt = "You are implementing the todo item below in a codebase.\n\n"
	}
	for i, todo := range todoList {
		number := 0
		if len(todoList) > 1 {
			number = i + 1
		}
		prompt += buildTODOSection(todo, workDir, true, number)
	}
	prompt += "---\n"
	prompt += groupTODOInstructions
	return prompt
}

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
	if todo.MarkdownBody != "" {
		return strings.TrimSpace(todo.MarkdownBody)
	}
	if todo.FilePath == "" {
		return ""
	}
	parsed, err := todos.ParseFrontmatterFromFile(todo.FilePath)
	if err != nil {
		return ""
	}
	return parsed.MarkdownContent
}

const singleTODOInstructions = `## Instructions

1. Analyze the test failure and reproduction steps
2. Investigate the codebase to understand the root cause
3. Implement a fix that addresses the underlying issue
4. Run verification tests (if any) to confirm the fix works
5. Do NOT run git add or git commit — gavel manages commits automatically

Your fix should:
- Address the root cause, not mask symptoms
- Follow existing code patterns and style
- Pass all verification tests
- Be minimal and focused
`

const groupTODOInstructions = `## Instructions

1. Implement ALL todo items listed above
2. Investigate the codebase to understand the root cause of each issue
3. Implement fixes that address the underlying issues
4. Do NOT run git add or git commit — gavel manages commits automatically

Your fixes should:
- Address root causes, not mask symptoms
- Follow existing code patterns and style
- Pass all verification tests (if specified)
- Be minimal and focused
`
