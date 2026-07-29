package verify

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/todos/types"
)

// ConfigSchemaID is the canonical URL editors fetch to validate .gavel.yaml.
// Reference it from a file with a leading comment:
//
//	# yaml-language-server: $schema=https://raw.githubusercontent.com/flanksource/gavel/main/gavel.schema.json
const ConfigSchemaID = "https://raw.githubusercontent.com/flanksource/gavel/main/gavel.schema.json"

// ConfigJSONSchema renders the documented JSON Schema for the .gavel.yaml
// (a.k.a. .gavel.yml) configuration file. It is the single source of truth for
// the committed gavel.schema.json artifact and SCHEMA.md; regenerate the JSON
// with `go generate .` after changing GavelConfig.
func ConfigJSONSchema() (string, error) {
	b, err := json.MarshalIndent(gavelConfigSchema(), "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

func gavelConfigSchema() map[string]any {
	specSchema := captainSpecSchema()
	schema := object(
		"Root configuration for Gavel. Place .gavel.yaml (or .gavel.yml) in ~/, the git root, "+
			"or the target directory; layers merge in that order with later layers overriding earlier ones. "+
			"Run `gavel config [path]` to inspect the merged result.",
		map[string]any{
			"ai":       aiSchema(specSchema),
			"lint":     lintSchema(),
			"commit":   commitSchema(),
			"fixtures": fixturesSchema(),
			"ssh":      sshSchema(),
			"pre":      hookStepsSchema("Top-level hooks run before the main test/lint pipeline, in declaration order. Appended across layers."),
			"post":     hookStepsSchema("Top-level hooks run after the main pipeline as non-blocking cleanup/reporting. Appended across layers."),
			"secrets":  secretsSchema(),
			"procfile": procfileSchema(),
			"checks":   checksSchema(),
			"todos":    todosSchema(specSchema),
			"status":   statusSchema(),
			"test":     testSchema(),
			"pr":       prSchema(),
		},
	)
	schema["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	schema["$id"] = ConfigSchemaID
	schema["title"] = "Gavel configuration (.gavel.yaml)"
	schema["$defs"] = specSchema["$defs"]
	return schema
}

// aiSchema exposes the complete Captain spec because every field is a valid
// inherited default, including prompt.system, workspace, permissions, and env.
// x-prompt-picker tells the settings UI to replace the generic object form with
// the shared rich PromptPicker editor.
func aiSchema(schema map[string]any) map[string]any {
	node := specNodeSchema(schema,
		"Base AI spec inherited by every AI operation. Configure model, prompt, workspace, permissions, environment, and runtime defaults here.",
		DefaultAIModel,
		"Default catalog model slug for all AI operations (e.g. claude-sonnet-4-5).")
	node["x-prompt-picker"] = true
	return node
}

// specNodeSchema clones the captain Spec definition into a config node carrying
// its own description and model default. A config field typed api.Spec accepts
// every spec field, so the shape is the spec's — only the documentation and the
// default differ per operation.
func specNodeSchema(schema map[string]any, description, modelDefault, modelDescription string) map[string]any {
	defs, ok := schema["$defs"].(map[string]any)
	if !ok {
		panic("captain spec schema has no $defs")
	}
	spec, ok := defs["Spec"].(map[string]any)
	if !ok {
		panic("captain spec schema has no Spec definition")
	}
	node := cloneSchemaMap(spec)
	node["description"] = description

	props, ok := spec["properties"].(map[string]any)
	if !ok {
		panic("captain Spec definition has no properties")
	}
	props = cloneSchemaMap(props)
	model, ok := props["model"].(map[string]any)
	if !ok {
		panic("captain Spec definition has no model property")
	}
	model = cloneSchemaMap(model)
	model["default"] = modelDefault
	model["description"] = modelDescription
	props["model"] = model
	node["properties"] = props
	return node
}

func cloneSchemaMap(source map[string]any) map[string]any {
	clone := make(map[string]any, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func captainSpecSchema() map[string]any {
	raw, err := api.SchemaJSON(&api.Spec{})
	if err != nil {
		panic("generate Captain spec schema: " + err.Error())
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		panic("decode Captain spec schema: " + err.Error())
	}
	return schema
}

func lintSchema() map[string]any {
	return object(
		"Settings for `gavel lint`.",
		map[string]any{
			"fix": promptSpecSchema(prompts.LintFix,
				"AI spec for repairing lint violations from `gavel lint --ai-fix` and the `gavel commit` lint gate. "+
					"This is independent of commit.message because fixing source requires an editing-capable agent."),
			"ignore": arrayOf(
				"Rules that suppress matching lint violations. Each rule matches when every populated "+
					"field matches; an empty rule never matches. Appended across layers.",
				object(
					"A single lint-ignore rule. At least one of rule/source/file must be set.",
					map[string]any{
						"rule": stringProp(
							"Match the violation's rule ID. Accepts literals, `*` globs (\"acme-*\"), and " +
								"`!`-prefixed negations."),
						"source": stringProp(
							"Match the emitting linter (e.g. golangci-lint, eslint, betterleaks). Accepts " +
								"literals, `*` globs, and `!`-prefixed negations."),
						"file": stringProp(
							"Match the violation's file path using doublestar globs (e.g. \"pkg/**/*.go\")."),
					},
				),
			),
			"linters": mapObject(
				"Per-linter overrides keyed by linter name. Useful for opt-in linters like jscpd or "+
					"disabling a single tool.",
				object(
					"Override for one linter.",
					map[string]any{
						"enabled": boolProp(
							"Force the linter on (true) or off (false). Omit to keep the linter's built-in " +
								"default."),
					},
				),
			),
		},
	)
}

func commitSchema() map[string]any {
	return object(
		"Settings for `gavel commit`.",
		map[string]any{
			"message": promptSpecSchema(prompts.CommitMessage,
				"AI spec for commit-message generation. Model/prompt/budget/effort override the base "+
					"ai: spec field-wise."),
			"grouping": promptSpecSchema(prompts.CommitGrouping,
				"AI spec for AI commit grouping (`gavel commit -G`) that sorts uncommitted changes into "+
					"logical commits. The output schema (groups[]+ignore[]) is fixed; the prompt changes "+
					"only the instructions."),
			"summary": promptSpecSchema(prompts.CommitSummary,
				"AI spec for naming and summarising a group of commits (`gavel git analyze --summary`)."),
			"maxCommits": intProp(
				"Maximum number of commits AI grouping may produce. 0 uses the built-in default."),
			"hooks": arrayOf(
				"Hooks run during `gavel commit` before the final commit is written. Appended across "+
					"layers in declaration order.",
				object(
					"A commit hook.",
					map[string]any{
						"name": stringProp("Display name for the hook."),
						"run":  stringProp("Shell command to execute."),
						"files": stringArray(
							"Glob patterns; the hook runs only when staged files match one of them. Runs " +
								"unconditionally when omitted."),
					},
				),
			),
			"gitignore": stringArray(
				"Extra ignore globs applied when selecting files to commit. Appended and deduped across layers."),
			"allow": stringArray(
				"Paths allowed through even when a broader commit.gitignore glob matches. Useful for " +
					"generated artifacts you intentionally commit. Appended and deduped across layers."),
			"precommit": checkModeObject(
				"Gate for commit.gitignore prompts and linked-dependency checks (package.json file:/link: "+
					"refs and go.mod replace directives pointing outside the repo).", "prompt"),
			"lint": object(
				"Gates that run linters over the staged file set before the commit is created. CLI flags "+
					"--lint and --lint-secrets override these per invocation.",
				map[string]any{
					"enabled": boolWithDefault(
						"Toggle every non-secrets linter. Omit to keep off.", false),
					"secrets": boolWithDefault(
						"Toggle the betterleaks/secrets linter. Omit to keep on (the highest-value "+
							"pre-commit check).", true),
				},
			),
			"tidy": object(
				"Controls whether `gavel commit` runs `go mod tidy` in every Go module and stages the "+
					"resulting go.mod/go.sum changes. CLI flag --tidy overrides per invocation.",
				map[string]any{
					"enabled": boolWithDefault(
						"Toggle the tidy step. Omit to keep on.", true),
				},
			),
		},
	)
}

func prSchema() map[string]any {
	return object(
		"Settings for `gavel pr`.",
		map[string]any{
			"content": promptSpecSchema(prompts.PRContent,
				"AI spec for generating the PR title, body, and branch name."),
			"base": stringProp(
				"Base branch for the pull request (e.g. origin/main). Last-write-wins across layers."),
			"draft": boolProp("Open the pull request as a draft."),
		},
	)
}

func todosSchema(specSchema map[string]any) map[string]any {
	return object(
		"Settings for `gavel todos run`.",
		map[string]any{
			"run": promptSpecSchema(prompts.TodosRun,
				"AI spec for the todo run prompt: the framing, the TODO items injected as {{{body}}}, "+
					"and the instructions."),
			"plan": promptSpecSchema(prompts.TodosPlan,
				"AI spec for the plan-mode prompt: the read-only investigation framing that produces a "+
					"reviewable implementation plan."),
			"verify": specNodeSchema(specSchema,
				"Spec a verification run executes as: `gavel todos check`, the dashboard's verify action, "+
					"and the acceptance-criteria grader inside a run's definition-of-done loop. It overrides "+
					"ai: and is overridden by the request. There is no prompt to override — the checklist is "+
					"generated from the todo's acceptance criteria.",
				DefaultVerifyModel,
				"Catalog model slug the grader runs as. It must be agentic (e.g. claude-code-sonnet): the "+
					"grader is told to inspect the change with its own tools, and an API model returns "+
					"confident verdicts without reading the diff."),
			"driver": stringProp(
				"Execution mechanism for a run: cmux, cli, sdk, or api. The coding agent is derived from " +
					"the model. Last-write-wins across layers."),
			"timeout": stringProp(
				"Wall-clock timeout for a run (e.g. 30m). Last-write-wins across layers."),
			"groupBy": stringProp(
				"How todos are grouped into runs. Last-write-wins across layers."),
			"approvals": boolProp(
				"Gate Bash behind a human approval prompt. Unset means the entrypoint decides: the " +
					"dashboard can answer approvals, `gavel todos run` cannot. Enabling it where nothing " +
					"can answer is an error rather than a run that blocks forever."),
		},
	)
}

func statusSchema() map[string]any {
	return object(
		"Settings for `gavel status`.",
		map[string]any{
			"summary": promptSpecSchema(prompts.StatusSummary,
				"AI spec for the per-file summary prompt (`gavel status --ai`). Variable: {{details}} "+
					"(staged/unstaged diff or file contents). The output schema ({summary}) is fixed."),
		},
	)
}

func testSchema() map[string]any {
	return object(
		"Settings for `gavel test`.",
		map[string]any{
			"outlineSummary": promptSpecSchema(prompts.TestOutlineSummary,
				"AI spec for the per-test summary prompt (`gavel test outline --ai-summary`). Variables: "+
					"{{ids}} (the test ids), {{file}}, {{source}}. The output schema (tests[]) is fixed."),
		},
	)
}

// promptSpecSchema models a PromptSpec: one AI operation's captain api.Spec
// (model + fallbacks, budget, effort, prompt) plus an optional .prompt file. A
// bare string is shorthand for prompt.user, so the union `type` keeps both
// hand-written forms valid. desc documents the operation; promptID is stamped as
// x-prompt-id so the settings UI links this field to its prompts.Prompt
// descriptor (default + metadata).
func promptSpecSchema(promptID, desc string) map[string]any {
	return map[string]any{
		"description":          desc,
		"type":                 []any{"string", "object"},
		"additionalProperties": false,
		"x-prompt-id":          promptID,
		"properties": map[string]any{
			"model": stringProp(
				"Catalog model slug for this operation (e.g. claude-haiku-4-5). Overrides the base " +
					"ai.model."),
			"fallbacks": modelFallbacksSchema(),
			"effort":    effortSchema(),
			"budget":    budgetSchema(),
			"prompt":    promptBodySchema(),
			"file": stringProp(
				"Path to a .prompt file whose frontmatter/body supply this operation's spec. Relative " +
					"paths resolve against the .gavel.yaml directory."),
		},
	}
}

// modelFallbacksSchema documents Model.Fallbacks: alternative model slugs tried
// in order when the primary is unavailable.
func modelFallbacksSchema() map[string]any {
	return stringArray(
		"Alternative model slugs tried in order when the primary is unavailable or its provider " +
			"cannot be constructed.")
}

// effortSchema documents Model.Effort (reasoning effort for thinking-capable models).
func effortSchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": "Reasoning effort for thinking-capable models.",
		"enum":        []any{"", "low", "medium", "high", "xhigh", "max", "ultra"},
	}
}

// budgetSchema documents api.Budget: the per-operation resource ceilings.
func budgetSchema() map[string]any {
	return object(
		"Resource ceilings for the run. Zero/unset means unbounded.",
		map[string]any{
			"cost":      numberProp("Maximum spend in USD; 0 = no ceiling."),
			"maxTokens": intProp("Cap on output tokens per model call; 0 = backend default."),
			"maxTurns":  intProp("Cap on agent turns; 0 = backend default."),
		},
	)
}

// promptBodySchema documents api.Prompt's user/system fields. A bare string on
// the parent PromptSpec is shorthand for prompt.user.
func promptBodySchema() map[string]any {
	return object(
		"The prompt body. A bare string in place of the whole spec is shorthand for prompt.user.",
		map[string]any{
			"user":   stringProp("The user prompt / template body (Handlebars)."),
			"system": stringProp("Optional system-prompt framing."),
		},
	)
}

func checksSchema() map[string]any {
	return object(
		"Post-completion check loop for `gavel todos run --check`: after an agent reports done, run "+
			"these tests/lint and feed failures back to the agent until they pass. Opt-in — runs only when "+
			"enabled here, by a TODO's frontmatter `checks`, or by the --check flag. Frontmatter overrides "+
			"these project defaults.",
		map[string]any{
			"enabled": boolProp(
				"Turn the loop on. Omitted leaves it off unless --check or a TODO's frontmatter enables it."),
			"maxIterations": intWithDefault(
				"Maximum agent re-runs before giving up.", types.DefaultMaxCheckIterations),
			"retry": stringProp(
				"CEL definition-of-done predicate: while it is true the agent re-runs with the failing " +
					"nodes as feedback; when false the run is verified. Reads {results, test_results, " +
					"changed_files, iteration} where results carries checklist " +
					"([]{item, passed, message}). Default: " + types.DefaultRetryExpr),
			"test": object(
				"gavel test options for the check run. Omit to skip tests.",
				map[string]any{
					"paths":   stringArray("Package paths to test. Empty discovers all."),
					"changed": boolProp("Only test packages affected by the agent's changes."),
					"timeout": stringProp("Global wall-clock deadline (e.g. 5m)."),
				},
			),
			"lint": object(
				"gavel lint options for the check run. Omit to skip linting.",
				map[string]any{
					"linters": stringArray("Linters to run. Empty runs every detected linter."),
					"changed": boolProp("Only report new violations versus the base ref."),
					"timeout": stringProp("Per-linter deadline (e.g. 5m)."),
				},
			),
		},
	)
}

func fixturesSchema() map[string]any {
	return object(
		"Fixture-test discovery for `gavel test`.",
		map[string]any{
			"enabled": boolWithDefault(
				"Auto-discover fixture files when running `gavel test`. Sticky: once true in any layer it "+
					"stays true.", false),
			"files": stringArrayWithDefault(
				"Globs used to discover fixtures. Replaces (does not append to) the parent layer.",
				[]string{DefaultFixturesGlob}),
		},
	)
}

func sshSchema() map[string]any {
	return object(
		"SSH post-receive hook / push backend.",
		map[string]any{
			"cmd": stringWithDefault(
				"Command executed by the SSH post-receive hook. Last-write-wins; an empty override "+
					"inherits the parent value.", "gavel test --lint"),
		},
	)
}

func secretsSchema() map[string]any {
	return object(
		"betterleaks / gitleaks secret-scanning orchestration. Rule authoring lives in the TOML files "+
			"themselves; Gavel only discovers and merges them.",
		map[string]any{
			"disabled": boolWithDefault(
				"Disable the betterleaks linter even when the binary is on PATH. Sticky: once true in any "+
					"layer it stays true.", false),
			"configs": stringArray(
				"Additional .betterleaks.toml / .gitleaks.toml paths to merge in, beyond those discovered " +
					"in the home dir, git root, and cwd. Relative paths resolve against the .gavel.yaml " +
					"directory. Appended and deduped across layers."),
		},
	)
}

func procfileSchema() map[string]any {
	autoRestart := map[string]any{
		"description": "Default restart policy for every process. Accepts a bool " +
			"(true=on-failure, false=no) or an enum: no (never restart), on-failure " +
			"(restart only on a non-zero exit), or always (restart on any exit).",
		"default": "no",
		"oneOf": []any{
			map[string]any{"type": "boolean"},
			map[string]any{"type": "string", "enum": []any{"no", "on-failure", "always"}},
		},
	}
	return object(
		"Global defaults for `gavel proc`. Per-process settings live in the Procfile, whose "+
			"entries are either `name: command` or `name:` with command/default/autoRestart/cpu/mem/"+
			"profiles/env/maxRestarts. This section holds only defaults + the active profile.",
		map[string]any{
			"profile": stringProp(
				"Default active profile. A Procfile entry with `profiles` auto-starts only when one of " +
					"them is the active profile; `gavel proc --profile <name>` overrides this."),
			"autoRestart": autoRestart,
			"maxRestarts": intWithDefault(
				"Cap on automatic restarts per process. 0 means unlimited.", 0),
			"env": mapObject(
				"Environment injected into every process, on top of the parent environment and any sibling "+
					".env file. Merged key-by-key across layers.",
				map[string]any{"type": "string"}),
			"mem": stringProp(
				"Default resident-memory cap per process (e.g. \"512Mi\", \"2g\"). Empty disables it. " +
					"A process whose group exceeds it is killed."),
			"cpu": numberProp(
				"Default sustained CPU cap per process, as a percentage (100 = one full core). " +
					"0 disables it. A process that stays above it is killed."),
		},
	)
}

func hookStepsSchema(desc string) map[string]any {
	return arrayOf(desc, object(
		"A hook step.",
		map[string]any{
			"name": stringProp("Optional display name for the step."),
			"run":  stringProp("Shell command to execute."),
		},
	))
}

// --- leaf builders -------------------------------------------------------

func object(desc string, props map[string]any) map[string]any {
	return map[string]any{
		"type":                 "object",
		"description":          desc,
		"additionalProperties": false,
		"properties":           props,
	}
}

func mapObject(desc string, value map[string]any) map[string]any {
	return map[string]any{
		"type":                 "object",
		"description":          desc,
		"additionalProperties": value,
	}
}

func arrayOf(desc string, item map[string]any) map[string]any {
	return map[string]any{
		"type":        "array",
		"description": desc,
		"items":       item,
	}
}

func stringProp(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}

func stringWithDefault(desc, def string) map[string]any {
	m := stringProp(desc)
	m["default"] = def
	return m
}

func boolProp(desc string) map[string]any {
	return map[string]any{"type": "boolean", "description": desc}
}

func intProp(desc string) map[string]any {
	return map[string]any{"type": "integer", "description": desc}
}

func numberProp(desc string) map[string]any {
	return map[string]any{"type": "number", "description": desc}
}

func intWithDefault(desc string, def int) map[string]any {
	m := intProp(desc)
	m["default"] = def
	return m
}

func boolWithDefault(desc string, def bool) map[string]any {
	m := boolProp(desc)
	m["default"] = def
	return m
}

func stringArray(desc string) map[string]any {
	return map[string]any{
		"type":        "array",
		"description": desc,
		"items":       map[string]any{"type": "string"},
	}
}

func stringArrayWithDefault(desc string, def []string) map[string]any {
	m := stringArray(desc)
	defAny := make([]any, len(def))
	for i, s := range def {
		defAny[i] = s
	}
	m["default"] = defAny
	return m
}

// checkModeObject models a CheckMode field: the string values prompt/fail/skip,
// or the boolean false (an alias for skip).
func checkModeObject(desc, def string) map[string]any {
	return object(desc, map[string]any{
		"mode": map[string]any{
			"description": "Gate behavior. Use false as an alias for skip.",
			"default":     def,
			"oneOf": []any{
				map[string]any{"type": "string", "enum": []any{"prompt", "fail", "skip"}},
				map[string]any{"type": "boolean", "enum": []any{false}},
			},
		},
	})
}
