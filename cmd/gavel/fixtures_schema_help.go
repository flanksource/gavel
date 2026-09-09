package main

import (
	"regexp"

	"github.com/spf13/cobra"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func fixturesSchemaHelpDocument(cmd *cobra.Command) map[string]any {
	return map[string]any{
		"command": "gavel fixtures --help",
		"text":    stripANSI(fixturesHelp(cmd).ANSI()),
		"sections": []map[string]any{
			helpSection("Usage", "Run fixture-based tests from markdown tables and command blocks.", []string{
				"gavel fixtures [flags] <fixture-file-or-glob> [fixture-file-or-glob...]",
			}),
			helpSection("Arguments", "One or more markdown fixture files or doublestar glob patterns. Quote globs when Gavel should expand them.", []string{
				"gavel fixtures tests.md",
				"gavel fixtures 'fixtures/**/*.md'",
			}),
			helpSection("File structure", "Optional YAML frontmatter configures file-wide defaults.", []string{
				"build: go build -o myapp",
				"daemon: go run ./server --port {{.port}}",
				"exec: ./myapp",
				"codeBlocks: [bash, python]",
			}),
			helpSection("Markdown tables", "Each row defines a test. Unrecognized columns become template variables and CEL properties.", []string{
				"| Name | CLI | Args | Exit Code | CEL |",
			}),
			helpSection("Command blocks", "Use `### command: <test name>` followed by YAML config, executable code, and validation bullets.", []string{
				"```yaml\nexitCode: 0\n```",
				"```bash\necho hello\n```",
				"- cel: stdout.contains(\"hello\")",
			}),
			helpSection("AI verification fixtures", "AI frontmatter plus a prompt/ai fence turns the markdown document into one AI verification step.", []string{
				"```prompt\nFocus on the new parser path and acceptance criteria.\n```",
				"- cel: json.score >= 80",
			}),
			helpSection("Test / lint steps", "A `yaml test` or `yaml lint` fence runs the Gavel test/lint engine; bare `test` and `lint` fences also work.", []string{
				"```yaml test\npaths: [./testrunner/...]\nframework: [go test]\n```",
				"```lint\nlinters: [golangci-lint, ruff]\nchanged: true\n```",
			}),
			helpSection("Supported languages", "Executable code fences include bash, sh, shell, python, javascript, typescript, powershell, and go.", []string{
				"```bash exitCode=1 timeout=30\nexit 1\n```",
			}),
			helpSection("CEL validation", "CEL expressions must evaluate to true and can inspect stdout, stderr, exitCode, json, ansi, expectations, and fixture paths.", []string{
				`stdout.contains("hello") && exitCode == 0`,
				`json.score >= 80`,
			}),
			helpSection("Template variables", "`exec`, `daemon`, `build`, `args`, and `cwd` support shell-style variables and Go template syntax.", []string{
				"daemon: go run ./server --port {{.port}}",
				"cwd: $GIT_ROOT_DIR/testdata",
			}),
			helpSection("Output options", "Verbosity and output flags control when passed fixtures, commands, stdout, stderr, and CEL vars are shown.", []string{
				"gavel fixtures -vvv tests.md",
				"gavel fixtures --show-stdout Always tests.md",
			}),
		},
	}
}

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

func helpSection(title, body string, examples []string) map[string]any {
	return map[string]any{
		"title":    title,
		"body":     body,
		"examples": examples,
	}
}

func commandSchemaDescription(name string) string {
	switch name {
	case "commit":
		return "Commit selected project files with schema-driven `gavel commit` options."
	case "test":
		return "Fixture test step. YAML body maps to `gavel test` options; each test becomes a child node."
	case "lint":
		return "Fixture lint step. YAML body maps to `gavel lint` options; each violation becomes a child node."
	default:
		return "Fixture runner step options."
	}
}

func commandSchemaHelp(name string) map[string]any {
	switch name {
	case "commit":
		return fixtureHelpBlock(
			"Commit options",
			"Runs `gavel commit` over the selected project files. Interactive, push, fixup, and history-rewrite options are excluded from the dashboard action.",
			"commit --help",
		)
	case "test":
		return fixtureHelpBlock(
			"Test / lint steps",
			"A `yaml test` fence runs the Gavel test engine. The YAML body is unmarshalled onto `gavel test` options, so CLI flags are available as kebab-case keys.",
			"fixtures --help",
		)
	case "lint":
		return fixtureHelpBlock(
			"Test / lint steps",
			"A `yaml lint` or `lint` fence runs the Gavel lint engine. The YAML body is unmarshalled onto `gavel lint` options, so CLI flags are available as kebab-case keys.",
			"fixtures --help",
		)
	default:
		return fixtureHelpBlock("Fixture step", "Fixture runner step options.", "fixtures --help")
	}
}

func fenceHelp(kind string) map[string]any {
	switch kind {
	case "test":
		return fixtureHelpBlock("Test / lint steps", "Runs `gavel test` from a markdown fence; each test becomes a child node.", "fixtures --help")
	case "lint":
		return fixtureHelpBlock("Test / lint steps", "Runs `gavel lint` from a markdown fence; each violation becomes a child node.", "fixtures --help")
	case "ai":
		return fixtureHelpBlock("AI verification fixtures", "Reviewer instructions for AI verification. GitHub task-list items become scored acceptance criteria.", "fixtures --help")
	case "exec":
		return fixtureHelpBlock("Command blocks", "Executable code fence with optional expectations and CEL validation.", "fixtures --help")
	default:
		return fixtureHelpBlock("Fixture fence", "Fixture markdown code fence.", "fixtures --help")
	}
}

func fixtureHelpBlock(section, body, source string) map[string]any {
	return map[string]any{
		"source":  source,
		"section": section,
		"body":    body,
	}
}

func withHelp(prop fixtureSchemaProperty, section, body, source string, examples ...any) fixtureSchemaProperty {
	prop["x-help"] = fixtureHelpBlock(section, body, source)
	if len(examples) > 0 {
		prop["examples"] = examples
	}
	return prop
}
