package verify

import "encoding/json"

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
	schema := object(
		"Root configuration for Gavel. Place .gavel.yaml (or .gavel.yml) in ~/, the git root, "+
			"or the target directory; layers merge in that order with later layers overriding earlier ones. "+
			"Run `gavel config [path]` to inspect the merged result.",
		map[string]any{
			"verify":   verifySchema(),
			"lint":     lintSchema(),
			"commit":   commitSchema(),
			"fixtures": fixturesSchema(),
			"ssh":      sshSchema(),
			"pre":      hookStepsSchema("Top-level hooks run before the main test/lint pipeline, in declaration order. Appended across layers."),
			"post":     hookStepsSchema("Top-level hooks run after the main pipeline as non-blocking cleanup/reporting. Appended across layers."),
			"secrets":  secretsSchema(),
			"procfile": procfileSchema(),
		},
	)
	schema["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	schema["$id"] = ConfigSchemaID
	schema["title"] = "Gavel configuration (.gavel.yaml)"
	return schema
}

func verifySchema() map[string]any {
	return object(
		"Settings for `gavel verify`, the AI code-review engine.",
		map[string]any{
			"model": stringWithDefault(
				"AI CLI / model used for review. Common values: claude, gemini, codex, or a fully "+
					"qualified model name. Last-write-wins across layers.", "claude"),
			"prompt": stringProp(
				"Optional repo-specific review policy appended to Gavel's built-in verify prompt. " +
					"Last-write-wins across layers."),
			"checks": object(
				"Selectively disable checks emitted by the verify engine.",
				map[string]any{
					"disabled": stringArray(
						"Individual check IDs to disable (e.g. SEC-1, PERF-2). Appended across layers."),
					"disabledCategories": stringArray(
						"Whole check categories to disable. Known categories: completeness, code-quality, " +
							"testing, consistency, security, performance. Appended across layers."),
				},
			),
		},
	)
}

func lintSchema() map[string]any {
	return object(
		"Settings for `gavel lint`.",
		map[string]any{
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
			"model": stringProp(
				"AI CLI / model used for commit-message generation and compatibility analysis. " +
					"Last-write-wins across layers."),
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
			"linkedDeps": checkModeObject(
				"Deprecated: superseded by commit.precommit. Retained for backward-compatible loading; "+
					"prefer commit.precommit.mode in new config.", "prompt"),
			"compatibility": checkModeObject(
				"Gate for the AI warning that surfaces removed functionality and backward-compatibility issues.",
				"skip"),
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
			"path": stringProp(
				"Override Procfile discovery. Relative paths resolve against the .gavel.yaml directory. " +
					"If omitted, the nearest Procfile up to the git root is used."),
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
