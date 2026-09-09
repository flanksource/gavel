package prompt

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

// buildTODOSection renders one TODO. grouped omits the per-todo PR context (the
// group framing carries it instead); number, when > 0, prefixes the heading with
// its position in the list so multi-todo runs read as a numbered checklist.
//
// rawFixture renders the verification section as its literal fixture source
// rather than as the bash commands it contains. An implementation run only needs
// to know what to execute, but a run that must REVIEW the fixture cannot judge
// what it cannot see: the frontmatter, the CEL predicate, and the yaml step
// structure are exactly what is being assessed, and the command projection
// discards all three.
func buildTODOSection(todo *types.TODO, workDir string, grouped bool, number int, rawFixture bool) string {
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

	section += buildVerificationSection(todo, rawFixture)

	return section
}

// buildVerificationSection renders the TODO's definition of done: its literal
// fixture source when the caller must review it, otherwise the commands to run.
func buildVerificationSection(todo *types.TODO, rawFixture bool) string {
	if rawFixture {
		fixture := strings.TrimSpace(todo.VerificationMarkdown)
		if fixture == "" {
			return "## Verification\n\nThis TODO has NO verification fixture — nothing can currently prove it done.\n\n"
		}
		return fmt.Sprintf("## Verification\n\nThe current fixture, verbatim:\n\n````markdown\n%s\n````\n\n", fixture)
	}
	var commands []string
	for _, node := range todo.Verification {
		if node.Test != nil {
			commands = append(commands, fixtureCommandProjection(*node.Test)...)
		}
	}
	if len(commands) == 0 {
		return ""
	}
	return "## Verification\n\nAfter implementing your fix, verify it works by running:\n\n" + strings.Join(commands, "")
}

// fixtureCommandProjection is the executable face of one fixture node — what
// the agent can run itself before the verify loop does. It is never the node's
// NAME: a bare ```exec fence is named `exec-block` by the parser, and that is
// not a command. An exec node projects to its command, a runner step to the
// fence gavel executes, and an AI checklist step — graded, not run — to nothing.
func fixtureCommandProjection(test fixtures.FixtureTest) []string {
	if test.IsAIStep() {
		return nil
	}
	if test.IsRunnerStep() {
		return []string{fmt.Sprintf("```yaml %s\n%s\n```\n\n", test.RunnerStep.Kind, strings.TrimSpace(test.RunnerStep.Config))}
	}
	if exec := strings.TrimSpace(test.ExecBase().Exec); exec != "" {
		return []string{fmt.Sprintf("```bash\n%s\n```\n\n", exec)}
	}
	return nil
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
