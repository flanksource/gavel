package main

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type fixtureSchemaProperty map[string]any
type fixtureJSONSchema map[string]any

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func writeFixturesSchema(w io.Writer) error {
	doc, err := fixturesSchemaDocument()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

func fixturesSchemaDocument() (map[string]any, error) {
	fixturesCommand, err := schemaCommand("fixtures")
	if err != nil {
		return nil, err
	}
	testCommand, err := schemaCommand("test")
	if err != nil {
		return nil, err
	}
	lintCommand, err := schemaCommand("lint")
	if err != nil {
		return nil, err
	}

	testSchema := commandSchema(testCommand, "paths", commandSchemaOptions{
		Exclude: map[string]bool{
			"addr":          true,
			"auto-stop":     true,
			"detach":        true,
			"diagnostics":   true,
			"idle-timeout":  true,
			"lint":          true,
			"lint-timeout":  true,
			"sync-todos":    true,
			"todo-template": true,
			"todos-dir":     true,
			"ui":            true,
		},
	})
	lintSchema := commandSchema(lintCommand, "files", commandSchemaOptions{
		Exclude: map[string]bool{
			"addr":                  true,
			"ai-fix":                true,
			"ai-fix-max-iterations": true,
			"allowed-tools":         true,
			"api-key":               true,
			"backend":               true,
			"bare":                  true,
			"budget":                true,
			"debug":                 true,
			"disallowed-tools":      true,
			"edit":                  true,
			"effort":                true,
			"fix":                   true,
			"group-by":              true,
			"hooks":                 true,
			"max-tokens":            true,
			"max-turns":             true,
			"mcp":                   true,
			"memory":                true,
			"model":                 true,
			"no-cache":              true,
			"no-hooks":              true,
			"no-mcp":                true,
			"no-memory":             true,
			"no-project":            true,
			"no-skills":             true,
			"no-user":               true,
			"permission-mode":       true,
			"profile":               true,
			"project":               true,
			"resume":                true,
			"skill-dir":             true,
			"skills":                true,
			"summary":               true,
			"summary-limit":         true,
			"sync-todos":            true,
			"temperature":           true,
			"triage":                true,
			"ui":                    true,
			"user":                  true,
			"yes":                   true,
		},
	})
	addStepDisplaySchema(testSchema)
	addStepDisplaySchema(lintSchema)

	return map[string]any{
		"schemaVersion": 1,
		"source":        "gavel fixtures --schema",
		"help":          fixturesSchemaHelpDocument(fixturesCommand),
		"frontmatter":   fixtureFrontmatterSchema(),
		"fences": map[string]any{
			"test": map[string]any{
				"schema":  testSchema,
				"aliases": []string{"yaml test"},
				"help":    fenceHelp("test"),
			},
			"lint": map[string]any{
				"schema":  lintSchema,
				"aliases": []string{"yaml lint"},
				"help":    fenceHelp("lint"),
			},
			"ai": map[string]any{
				"schema":  fixturePromptFenceSchema("Reviewer instructions passed to the AI verifier"),
				"aliases": []string{"prompt"},
				"help":    fenceHelp("ai"),
			},
			"exec": map[string]any{
				"schema":  fixtureExecFenceSchema(),
				"aliases": []string{"bash", "sh", "shell", "javascript", "js", "typescript", "ts", "python", "py", "go"},
				"help":    fenceHelp("exec"),
			},
		},
	}, nil
}

func schemaCommand(name string) (*cobra.Command, error) {
	cmd, _, err := rootCmd.Find([]string{name})
	if err != nil {
		return nil, fmt.Errorf("find gavel %s command: %w", name, err)
	}
	if cmd == nil || cmd.Name() != name {
		return nil, fmt.Errorf("find gavel %s command: got %v", name, cmd)
	}
	return cmd, nil
}

type commandSchemaOptions struct {
	Exclude map[string]bool
}

func commandSchema(cmd *cobra.Command, positionalName string, opts commandSchemaOptions) fixtureJSONSchema {
	properties := map[string]any{}
	order := []string{}

	if positionalName != "" {
		properties[positionalName] = fixtureSchemaProperty{
			"type":        "array",
			"title":       titleFromFlag(positionalName),
			"description": "Positional arguments for `gavel " + cmd.Name() + "`.",
			"items":       fixtureSchemaProperty{"type": "string"},
		}
		applyKnownFlagHints(positionalName, properties[positionalName].(fixtureSchemaProperty))
		order = append(order, positionalName)
	}

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden || opts.Exclude[flag.Name] {
			return
		}
		properties[flag.Name] = propertyFromFlag(flag)
		order = append(order, flag.Name)
	})

	return fixtureJSONSchema{
		"type":                 "object",
		"description":          commandSchemaDescription(cmd.Name()),
		"additionalProperties": false,
		"properties":           properties,
		"x-order":              order,
		"x-help":               commandSchemaHelp(cmd.Name()),
	}
}

func addStepDisplaySchema(schema fixtureJSONSchema) {
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return
	}
	properties["show-failed"] = fixtureSchemaProperty{
		"type":        "boolean",
		"title":       "Show failed",
		"description": "Show failing child nodes for this fixture runner step.",
		"default":     true,
		"x-help": fixtureHelpBlock(
			"Rendering",
			"`show-failed` adds failing tests or lint violations as child nodes. It defaults to true for fixture-editor steps.",
			"fixtures --help",
		),
	}
	if order, ok := schema["x-order"].([]string); ok {
		schema["x-order"] = append(order, "show-failed")
	}
}

func propertyFromFlag(flag *pflag.Flag) fixtureSchemaProperty {
	prop := fixtureSchemaProperty{
		"title":       titleFromFlag(flag.Name),
		"description": flag.Usage,
	}

	switch flag.Value.Type() {
	case "bool":
		prop["type"] = "boolean"
		if flag.DefValue != "" {
			prop["default"] = flag.DefValue == "true"
		}
	case "int", "int8", "int16", "int32", "int64":
		prop["type"] = "integer"
		if parsed, err := strconv.Atoi(flag.DefValue); err == nil {
			prop["default"] = parsed
		}
	case "float32", "float64":
		prop["type"] = "number"
		if parsed, err := strconv.ParseFloat(flag.DefValue, 64); err == nil {
			prop["default"] = parsed
		}
	case "stringSlice", "stringArray", "intSlice":
		prop["type"] = "array"
		itemType := "string"
		if flag.Value.Type() == "intSlice" {
			itemType = "integer"
		}
		prop["items"] = fixtureSchemaProperty{"type": itemType}
	default:
		prop["type"] = "string"
		if flag.DefValue != "" && flag.DefValue != "[]" {
			prop["default"] = flag.DefValue
		}
	}

	applyKnownFlagHints(flag.Name, prop)
	return prop
}

func applyKnownFlagHints(name string, prop fixtureSchemaProperty) {
	switch name {
	case "paths":
		prop["description"] = "Package paths to test. Empty means discover all packages, matching the `gavel test` CLI positional arguments."
		prop["examples"] = []any{[]string{"./testrunner/..."}, []string{"./cmd/gavel"}}
		prop["x-help"] = fixtureHelpBlock("Common test keys", "Package paths to test (empty = discover all).", "fixtures --help")
	case "framework":
		prop["x-array-display"] = "filter-pills"
		prop["items"] = fixtureSchemaProperty{
			"type": "string",
			"enum": frameworkSchemaValues(),
		}
		prop["description"] = "Restrict test discovery to selected frameworks. Leave empty to enable every detected framework."
		prop["x-help"] = fixtureHelpBlock("Common test keys", "Restrict to jest, vitest, playwright, go test, or ginkgo.", "fixtures --help")
	case "linters":
		prop["x-array-display"] = "filter-pills"
		prop["items"] = fixtureSchemaProperty{
			"type": "string",
			"enum": linterSchemaValues(),
		}
		prop["description"] = "Only run these linters. Leave empty to run every detected linter."
		prop["x-help"] = fixtureHelpBlock("Common lint keys", "Only run these linters (empty = every detected linter).", "fixtures --help")
	case "files":
		prop["description"] = "Target paths to lint, or a frontmatter glob that replicates fixture tests per matched file."
		prop["examples"] = []any{[]string{"cmd/gavel/fixtures.go"}, []string{"**/*.go"}}
		prop["x-help"] = fixtureHelpBlock("File expansion", "Set `files` in frontmatter to replicate each test per matching file; in lint steps it selects target paths.", "fixtures --help")
	case "changed", "since", "failed", "baseline":
		prop["x-help"] = fixtureHelpBlock("Common test/lint keys", "Narrow to changed files, prior failures, or new results compared with a baseline.", "fixtures --help")
	case "extra-args":
		prop["x-help"] = fixtureHelpBlock("Common test keys", "Arguments forwarded to the underlying test runner.", "fixtures --help")
	case "fix":
		prop["x-help"] = fixtureHelpBlock("Common lint keys", "Apply linter auto-fixes where the selected linter supports it.", "fixtures --help")
	case "group-by":
		prop["enum"] = []string{"file", "package", "message"}
	case "show-stdout", "show-stderr":
		prop["enum"] = []string{"Never", "OnFailure", "Always"}
		prop["x-help"] = fixtureHelpBlock("Output options", "Controls when command output is shown: Never, OnFailure, or Always.", "fixtures --help")
	case "show-passed":
		prop["x-help"] = fixtureHelpBlock("Rendering", "Add passing tests or clean linters as child nodes. The CLI also exposes this as `--show-passed`.", "fixtures --help")
	case "permission-mode":
		prop["enum"] = []string{"acceptEdits", "auto", "bypassPermissions", "default", "dontAsk", "plan"}
	case "timeout", "test-timeout", "lint-timeout", "auto-stop", "idle-timeout":
		prop["format"] = "duration"
		prop["examples"] = []any{"30s", "2m", "5m"}
		prop["x-help"] = fixtureHelpBlock("Execution", "Duration string for per-step or whole-run deadlines.", "fixtures --help")
	}
}

func frameworkSchemaValues() []string {
	values := make([]string, 0, len(parsers.AllFrameworks))
	for _, framework := range parsers.AllFrameworks {
		values = append(values, string(framework))
	}
	return values
}

func linterSchemaValues() []string {
	return []string{
		"golangci-lint",
		"ruff",
		"eslint",
		"oxlint",
		"pyright",
		"tsc",
		"markdownlint",
		"vale",
		"jscpd",
		"betterleaks",
	}
}

func fixtureFrontmatterSchema() fixtureJSONSchema {
	return fixtureJSONSchema{
		"type":                 "object",
		"description":          "YAML frontmatter configures defaults for every fixture in the markdown document.",
		"additionalProperties": true,
		"x-help": fixtureHelpBlock(
			"File structure",
			"Each fixture file can start with YAML frontmatter containing global defaults such as build, daemon, exec, args, env, cwd, files, AI, and verify settings.",
			"fixtures --help",
		),
		"x-order": []string{
			"build", "daemon", "exec", "args", "env", "cwd", "terminal", "files", "codeBlocks",
			"timeout", "os", "arch", "skip", "ai", "verify",
		},
		"properties": map[string]any{
			"build":      withHelp(stringProp("Build", "Shell command run once before all tests."), "File structure", "Runs once before any tests. Template variables and shell-style `$VAR` expansion are supported.", "fixtures --help", "go build -o $workDir/myapp"),
			"daemon":     withHelp(stringProp("Daemon", "Background command run before tests and stopped afterwards."), "Execution", "Starts after build, waits for a free port to accept connections, and stops after all fixtures finish.", "fixtures --help", "go run ./server --port {{.port}}"),
			"exec":       withHelp(stringProp("Executable", "Default executable for command/table fixtures."), "Template variables", "The default executable supports shell-style `$VAR` and Go template syntax.", "fixtures --help", "./myapp", "$executablePath"),
			"args":       withHelp(stringArrayProp("Args", "Default arguments for exec."), "Template variables", "Default arguments support shell-style `$VAR` and Go template syntax.", "fixtures --help", []string{"--file", "$file"}),
			"env":        withHelp(stringMapProp("Environment", "Environment variables for all tests."), "File structure", "Environment variables are available to build, daemon, and test commands.", "fixtures --help"),
			"cwd":        withHelp(stringProp("Working directory", "Default working directory, resolved relative to the fixture file."), "CWD resolution", "Working directory priority: test-level cwd, file-level cwd, source directory, then `--cwd` or the current working directory.", "fixtures --help", "$GIT_ROOT_DIR/testdata", "./testdir"),
			"terminal":   withHelp(enumProp("Terminal", "Terminal mode. `pty` uses a pseudo-terminal and merges stdout/stderr.", []string{"pty"}), "File structure", "`pty` mode uses a pseudo-terminal, which is useful for terminal UI output and ANSI assertions.", "fixtures --help", "pty"),
			"files":      withHelp(stringProp("Files", "Glob pattern: replicate tests per matching file."), "File expansion", "Set `files` to replicate each test per matched file. File variables such as `file`, `filename`, `dir`, and `ext` become available.", "fixtures --help", "**/*.go"),
			"codeBlocks": withHelp(stringArrayProp("Code blocks", "Executable code fence languages."), "Supported languages", "Languages to execute from standalone code fences. Non-executable labels such as yaml/frontmatter/json are parsed as config.", "fixtures --help", []string{"bash", "python"}),
			"timeout":    withHelp(stringProp("Timeout", "Total timeout for fixture execution."), "Execution", "Total timeout for test execution. Individual command blocks can override this with YAML config or `timeout=N` fence attributes.", "fixtures --help", "30s"),
			"os":         withHelp(stringProp("OS", "OS constraint, e.g. linux or !darwin."), "File structure", "Skip fixtures on non-matching operating systems. Prefix with `!` to negate.", "fixtures --help", "linux", "!darwin"),
			"arch":       withHelp(stringProp("Arch", "Architecture constraint, e.g. amd64."), "File structure", "Skip fixtures on non-matching CPU architectures.", "fixtures --help", "amd64", "arm64"),
			"skip":       withHelp(stringProp("Skip", "Bash command; exit 0 skips the fixture."), "File structure", "The skip command runs in bash; exit code 0 means skip the fixture.", "fixtures --help", "! command -v docker"),
			"ai": fixtureJSONSchema{
				"type":                 "object",
				"description":          "AI verification fixture config.",
				"additionalProperties": true,
				"x-help":               fixtureHelpBlock("AI options", "Add `ai:` frontmatter to turn the markdown document into one AI verification step.", "fixtures --help"),
				"properties": map[string]any{
					"model":         withHelp(stringProp("Model", "Agent model name, falling back to the default model when unset."), "AI options", "Agent model name.", "fixtures --help", "claude-code-sonnet"),
					"temperature":   withHelp(numberProp("Temperature", "Sampling temperature.", 0, 2), "AI options", "Sampling temperature.", "fixtures --help", 0),
					"maxTokens":     withHelp(integerProp("Max tokens", "Maximum response tokens.", 0, 0), "AI options", "Maximum response tokens.", "fixtures --help", 10000),
					"maxConcurrent": withHelp(integerProp("Max concurrent", "Maximum concurrent AI checks.", 1, 0), "AI options", "Maximum concurrent AI checks.", "fixtures --help", 4),
					"cacheTTL":      withHelp(stringProp("Cache TTL", "Cache duration for AI calls."), "AI options", "Cache duration for AI calls.", "fixtures --help", "10m"),
					"noCache": fixtureSchemaProperty{
						"type":        "boolean",
						"title":       "No cache",
						"description": "Disable AI response cache.",
						"x-help":      fixtureHelpBlock("AI options", "Disable AI response cache.", "fixtures --help"),
					},
				},
			},
			"verify": verifySchema(),
		},
	}
}

func verifySchema() fixtureJSONSchema {
	return fixtureJSONSchema{
		"type":                 "object",
		"description":          "AI verification scoring options.",
		"additionalProperties": true,
		"x-help":               fixtureHelpBlock("Verify options", "Controls how AI verification scores and gates the fixture.", "fixtures --help"),
		"properties": map[string]any{
			"scope":     withHelp(enumProp("Scope", "Review scope, defaulting to the working-tree diff.", []string{"diff", "all", "changed"}), "Verify options", "Review scope, defaulting to the working-tree diff.", "fixtures --help", "diff"),
			"threshold": withHelp(integerProp("Threshold", "Minimum passing score, defaulting to 80.", 0, 100), "Verify options", "Minimum passing score, defaulting to 80.", "fixtures --help", 80),
			"disabled":  withHelp(stringArrayProp("Disabled checks", "Check IDs to disable for this fixture."), "Verify options", "Check IDs to disable for this fixture.", "fixtures --help", []string{"style"}),
		},
	}
}

func fixturePromptFenceSchema(description string) fixtureJSONSchema {
	return fixtureJSONSchema{
		"type":        "object",
		"description": description,
		"x-help":      fixtureHelpBlock("AI verification fixtures", "The first ```prompt or ```ai block becomes reviewer instructions for an AI verification fixture.", "fixtures --help"),
		"properties": map[string]any{
			"content": fixtureSchemaProperty{
				"type":        "string",
				"title":       "Content",
				"format":      "textarea",
				"description": description,
			},
		},
	}
}

func fixtureExecFenceSchema() fixtureJSONSchema {
	return fixtureJSONSchema{
		"type":                 "object",
		"description":          "Shell command or script fixture block with optional execution expectations.",
		"additionalProperties": true,
		"x-help": fixtureHelpBlock(
			"Command blocks",
			"Executable code blocks can be raw scripts, or YAML objects with `content` plus expectations such as exitCode, stdout/stderr, CEL, timeout, and properties.",
			"fixtures --help",
		),
		"x-order": []string{
			"content", "exitCode", "cel", "stdout", "stderr", "output", "error", "format", "count", "timeout", "properties",
		},
		"properties": map[string]any{
			"content": fixtureSchemaProperty{
				"type":        "string",
				"title":       "Content",
				"format":      "textarea",
				"description": "Shell command or script body.",
				"x-help":      fixtureHelpBlock("Command blocks", "The command body runs using the fence language, such as bash, sh, python, javascript, typescript, or go.", "fixtures --help"),
			},
			"exitCode": withHelp(integerProp("Exit code", "Expected process exit code. Use `-` in table fixtures to skip exit-code checking.", 0, 0), "Expectation columns", "Expected exit code (default: 0, `-` to skip).", "fixtures --help", 0, 1),
			"stdout":   withHelp(stringProp("Stdout", "Expected stdout content or @golden file reference."), "Expectation columns", "Expected stdout, literal content, or an @file reference.", "fixtures --help", "@testdata/output.txt"),
			"stderr":   withHelp(stringProp("Stderr", "Expected stderr content or @golden file reference."), "Expectation columns", "Expected stderr, literal content, or an @file reference.", "fixtures --help"),
			"output":   withHelp(stringProp("Output", "Expected output substring in stdout or stderr."), "Expectation columns", "Expected output substring.", "fixtures --help", "success"),
			"error":    withHelp(stringProp("Error", "Expected non-zero failure text in stderr."), "Expectation columns", "Expected stderr substring; implies non-zero failure.", "fixtures --help", "permission denied"),
			"format":   withHelp(stringProp("Format", "Optional output format hint, commonly json or yaml."), "Expectation columns", "Output format validation.", "fixtures --help", "json", "yaml"),
			"count":    withHelp(integerProp("Count", "Expected result count.", 0, 0), "Expectation columns", "Expected result count for fixture types that expose counts.", "fixtures --help", 1),
			"timeout": fixtureSchemaProperty{
				"type":        "string",
				"title":       "Timeout",
				"format":      "duration",
				"description": "Maximum execution time for this command.",
				"examples":    []any{"30s", "2m"},
				"x-help":      fixtureHelpBlock("Inline code fence attributes", "Can be set in YAML config or as a fence attribute: ```bash timeout=30.", "fixtures --help"),
			},
			"cel": fixtureSchemaProperty{
				"type":        "string",
				"title":       "CEL",
				"format":      "textarea",
				"description": "CEL assertion over stdout, stderr, exitCode, ansi, and temp files.",
				"examples":    []any{`stdout.contains("hello")`, `json.score >= 80`, `exitCode == 0 && ansi.has_color`},
				"x-help":      fixtureHelpBlock("CEL validation", "Expressions must evaluate to true. Variables include stdout, stderr, exitCode, json, name, sourceDir, query, expectations, workDir, executablePath, ansi, and temp files.", "fixtures --help"),
			},
			"properties": fixtureSchemaProperty{
				"type":                 "object",
				"title":                "Properties",
				"description":          "Additional named values available to fixture templating and CEL.",
				"additionalProperties": true,
				"x-help":               fixtureHelpBlock("Markdown tables", "Unrecognized table columns become custom template variables and are exposed in CEL through expectations.Properties.", "fixtures --help"),
			},
		},
	}
}

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

func stringProp(title, description string) fixtureSchemaProperty {
	return fixtureSchemaProperty{"type": "string", "title": title, "description": description}
}

func stringArrayProp(title, description string) fixtureSchemaProperty {
	return fixtureSchemaProperty{
		"type":        "array",
		"title":       title,
		"description": description,
		"items":       fixtureSchemaProperty{"type": "string"},
	}
}

func stringMapProp(title, description string) fixtureSchemaProperty {
	return fixtureSchemaProperty{
		"type":                 "object",
		"title":                title,
		"description":          description,
		"additionalProperties": fixtureSchemaProperty{"type": "string"},
	}
}

func enumProp(title, description string, values []string) fixtureSchemaProperty {
	return fixtureSchemaProperty{
		"type":        "string",
		"title":       title,
		"description": description,
		"enum":        values,
	}
}

func numberProp(title, description string, min, max float64) fixtureSchemaProperty {
	prop := fixtureSchemaProperty{"type": "number", "title": title, "description": description}
	prop["minimum"] = min
	if max > min {
		prop["maximum"] = max
	}
	return prop
}

func integerProp(title, description string, min, max int) fixtureSchemaProperty {
	prop := fixtureSchemaProperty{
		"type":        "integer",
		"title":       title,
		"description": description,
		"minimum":     min,
		"multipleOf":  1,
	}
	if max > min {
		prop["maximum"] = max
	}
	return prop
}

func titleFromFlag(name string) string {
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func sortedSchemaPropertyNames(schema fixtureJSONSchema) []string {
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return nil
	}
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
