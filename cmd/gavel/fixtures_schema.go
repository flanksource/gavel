package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type fixtureSchemaProperty map[string]any
type fixtureJSONSchema map[string]any

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
			"addr":         true,
			"auto-stop":    true,
			"detach":       true,
			"diagnostics":  true,
			"idle-timeout": true,
			"lint":         true,
			"lint-timeout": true,
			"ui":           true,
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
		"react-doctor",
		"pyright",
		"tsc",
		"markdownlint",
		"vale",
		"jscpd",
		"betterleaks",
	}
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
