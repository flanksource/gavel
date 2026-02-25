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
	prompt += buildTODOSection(todo, workDir, false)
	prompt += singleTODOInstructions
	return prompt
}

// BuildGroupPrompt constructs a combined prompt for multiple related TODOs.
func BuildGroupPrompt(todoList []*types.TODO, workDir string) string {
	prompt := "You are implementing multiple related fixes in a codebase.\nEach TODO below describes a separate task. Implement ALL of them.\n\n"
	for _, todo := range todoList {
		name := todo.Title
		if name == "" {
			name = filepath.Base(todo.FilePath)
		}
		prompt += buildTODOSection(todo, workDir, true)
	}
	prompt += "---\n"
	prompt += groupTODOInstructions
	return prompt
}

func buildTODOSection(todo *types.TODO, workDir string, grouped bool) string {
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
		section += fmt.Sprintf("## %s\n\n", heading)
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

1. Implement ALL TODOs listed above
3. Investigate the codebase to understand the root cause of each issue
4. Implement fixes that address the underlying issues
5. Do NOT run git add or git commit — gavel manages commits automatically

Your fixes should:
- Address root causes, not mask symptoms
- Follow existing code patterns and style
- Pass all verification tests (if specified)
- Be minimal and focused
`
