package claude

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

//go:embed todos-run.prompt
var todosRunPrompt string

// GroupPromptOptions configures BuildGroupPrompt.
type GroupPromptOptions struct {
	WorkDir string
	// Template is the resolved .gavel.yaml todos.runPrompt override (dotprompt
	// source); empty uses the embedded default. Resolve it with ResolveRunTemplate.
	Template string
}

// ResolveRunTemplate reads the .gavel.yaml todos.runPrompt override for dir,
// returning the inline/file template source or "" when unset (the embedded
// default is then used). A configured-but-missing file is a hard error.
func ResolveRunTemplate(dir string) (string, error) {
	cfg, err := verify.LoadGavelConfig(dir)
	if err != nil {
		return "", fmt.Errorf("load .gavel.yaml for todos.runPrompt: %w", err)
	}
	tmpl, err := cfg.Todos.RunPrompt.Resolve(dir, "")
	if err != nil {
		return "", fmt.Errorf("resolve todos.runPrompt override: %w", err)
	}
	return tmpl, nil
}

// BuildPrompt constructs a structured prompt from a TODO for Claude Code execution.
func BuildPrompt(todo *types.TODO, workDir string) string {
	prompt := "You are fixing a failing test in a Go codebase.\n\n"
	prompt += buildTODOSection(todo, workDir, false, 0)
	prompt += singleTODOInstructions
	return prompt
}

// BuildGroupPrompt renders the run prompt for a group of TODOs from the dotprompt
// template (opts.Template or the embedded default). The per-TODO sections are
// assembled in Go and injected as {{{body}}}; the template owns the framing and
// instructions so .gavel.yaml todos.runPrompt can override them. A single item
// keeps the plain framing (no numbering); several items are numbered so the agent
// can address each in turn.
func BuildGroupPrompt(todoList []*types.TODO, opts GroupPromptOptions) (string, error) {
	multiple := len(todoList) > 1
	var body strings.Builder
	for i, todo := range todoList {
		number := 0
		if multiple {
			number = i + 1
		}
		body.WriteString(buildTODOSection(todo, opts.WorkDir, true, number))
	}

	template := opts.Template
	if strings.TrimSpace(template) == "" {
		template = todosRunPrompt
	}
	req, _, err := prompt.Load(template).Render(map[string]any{
		"multiple": multiple,
		"count":    len(todoList),
		"body":     body.String(),
	}, nil)
	if err != nil {
		return "", fmt.Errorf("render todo run prompt: %w", err)
	}
	return req.Prompt.User, nil
}

// Prompts returns the overridable prompt templates owned by the claude todo
// runner: the run prompt assembled for a coding-agent session. The override is
// the typed todos.runPrompt field, resolved against Default by ResolveRunTemplate.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{
		ID:          prompts.TodosRun,
		Title:       "Todo run prompt",
		Description: "The agent prompt for `gavel todos run`: framing, the TODO items, and instructions.",
		ConfigPath:  "todos.runPrompt",
		Default:     todosRunPrompt,
	}}
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
