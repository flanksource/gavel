package verify

import (
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/prompts"
)

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

// commitTypeNames is the vocabulary offered for commit.types, read from the same
// source as the prompt's enum so the two cannot drift apart.
func commitTypeNames() []string {
	types := models.SelectableCommitTypes()
	out := make([]string, len(types))
	for i, t := range types {
		out[i] = string(t)
	}
	return out
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
			"types": enumStringArray(
				"Conventional-commit types AI message generation may choose from; they become the "+
					"generated message's type enum. Replaces the built-in list rather than adding to it. "+
					"Empty allows every type below.",
				commitTypeNames()),
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
			"fix": promptSpecSchema(prompts.PRFix,
				"AI spec for `gavel pr status --ai-fix`: the agent that repairs failing checks and "+
					"unresolved review comments. Its workflow.verify.commands are the loop's definition of "+
					"done (re-polling `gavel pr status`) and workflow.commits the per-turn commit policy."),
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
			"triage": promptSpecSchema(prompts.TodosTriage,
				"AI spec for the triage prompt: a read-only pass that compacts the TODO's description and "+
					"reviews its verification fixture, reporting the edits for gavel to apply."),
			"checkConcurrency": intProp(
				"How many definition-of-done checks run at once (`gavel todos check`, and the verification " +
					"phase after a bulk triage). Each check runs the TODO's fixture, so an unbounded fan-out " +
					"over a large selection thrashes the machine. Defaults to 4."),
			"verify": specNodeSchema(specSchema,
				"Spec a verification run executes as: `gavel todos check`, the dashboard's verify action, "+
					"and the acceptance-criteria grader inside a run's definition-of-done loop. It overrides "+
					"ai: and is overridden by the request. There is no prompt to override — the checklist is "+
					"generated from the todo's acceptance criteria.",
				"",
				"Catalog model slug the grader runs as. It must run on an agentic mechanism (mode agent, "+
					"cli or cmux — e.g. `agent:sonnet`): the grader is told to inspect the change with its "+
					"own tools, and an API model returns confident verdicts without reading the diff."),
			"steps": mapObject(
				"Spec layer for lifecycle steps that are not built in, keyed by step name: a `handoff` step "+
					"added under todos.lifecycle reads its project configuration from todos.steps.handoff, "+
					"exactly where todos.run sits for the built-in run step. The built-in steps (run, plan, "+
					"triage, verify) keep their own blocks and are rejected here.",
				specNodeSchema(specSchema,
					"Spec one custom lifecycle step runs as. It overrides the step's prompt frontmatter and "+
						"its lifecycle declaration, and is overridden by the todo's llm: and the request.",
					"",
					"Catalog model slug this step runs as (e.g. agent:sonnet). Overrides the base ai.model.")),
			"timeout": stringProp(
				"Wall-clock timeout for a run (e.g. 30m). Last-write-wins across layers."),
			"lifecycle": lifecycleSchema(),
			"baseUrl": stringProp(
				"Absolute origin this gavel dashboard is reachable at (e.g. https://gavel.example.com). " +
					"Todo bodies store attachments as server-relative links, so pushing a todo to an " +
					"external tracker rewrites them against this origin. A loopback origin only renders " +
					"for viewers on the same machine."),
		},
	)
}

func lifecycleSchema() map[string]any {
	return object(
		"Overrides the built-in todo lifecycle: the ordered steps (a captain prompt reference, a CEL "+
			"`when` over subject/runs/last, and ordered `outcomes` mapping the finished run onto a status). "+
			"Either `file` (a lifecycle YAML relative to the project) or an inline `name`/`subject`/`steps`. "+
			"A step named like a built-in step replaces it wholesale; a new name is appended. A `verify` "+
			"step must remain.",
		map[string]any{
			"file": stringProp(
				"Path to a lifecycle YAML document, relative to the project root. Mutually exclusive with the inline form."),
			"name": stringProp("Name of the lifecycle; defaults to the built-in `todos`."),
			"subject": map[string]any{
				"type": "object",
				"description": "Extra subject declarations a custom step's predicates may read, field → CEL type " +
					"(string, bool, int, double, dyn, list<string>, list<dyn>, map<string,dyn>).",
				"additionalProperties": map[string]any{"type": "string"},
			},
			"steps": map[string]any{
				"type": "array",
				"description": "Steps to replace or add: {name, prompt, envelope, when, inputs, spec, outcomes, auxiliary}. " +
					"Outcomes are `{status, when}` pairs evaluated in order; the first true wins and none true is an error.",
				"items": map[string]any{"type": "object"},
			},
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
			"timeout": stringWithDefault(
				"Wall-clock deadline for the whole test+lint run, as a Go duration. On timeout "+
					"diagnostics are captured and every subprocess is killed. Overridden by --timeout.", "10m"),
			"testTimeout": stringWithDefault(
				"Deadline for each test-package subprocess (one go test / ginkgo / vitest invocation), "+
					"as a Go duration. A suite that exceeds it is killed and reported as a timeout. "+
					"Overridden by --test-timeout.", "5m"),
			"lintTimeout": stringWithDefault(
				"Deadline for each linter subprocess when --lint is set, as a Go duration. "+
					"Overridden by --lint-timeout.", "5m"),
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
			"workflow": map[string]any{
				"$ref": "#/$defs/Workflow",
				"description": "Loop shape for operations that run one: verify commands (the definition of " +
					"done, whose exit code gates a re-run and whose output tail becomes the feedback) and " +
					"the commit policy. Overriding it replaces the operation's declared workflow field-wise.",
			},
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
		"Tests and lint appended to every todo's definition of done as `yaml test`/`yaml lint` fixture "+
			"steps: after an agent reports done, the run step's verify loop runs them and feeds failures "+
			"back to the agent until they pass or `todos.run.workflow.verify.maxIterations` is reached. "+
			"Opt-in — runs only when enabled here or by a TODO's frontmatter `checks`. Frontmatter "+
			"overrides these project defaults.",
		map[string]any{
			"enabled": boolProp(
				"Append the check steps. Omitted leaves them off unless a TODO's frontmatter enables them."),
			"retry": stringProp(
				"CEL definition-of-done predicate: while it is true the agent re-runs with the failing " +
					"nodes as feedback. It reads the verification report under the variable `verify` " +
					"(verify.state, verify.summary.{total,passed,failed,warned,skipped,timedout}, " +
					"verify.tests, verify.checklist[].{item,passed,message}) and can only add reasons to " +
					"re-run — a failing run is never talked into passing. Default: " + fixtures.DefaultRetryExpr),
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
