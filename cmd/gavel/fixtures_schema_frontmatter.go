package main

import (
	"fmt"

	"github.com/flanksource/captain/pkg/api"

	"github.com/flanksource/gavel/fixtures"
)

func fixtureFrontmatterSchema() fixtureJSONSchema {
	captain := api.Schema(&api.Spec{})
	spec, ok := captain.Definitions["Spec"]
	if !ok {
		panic("Captain spec schema has no Spec definition")
	}
	spec.Description = "Complete grader runtime snapshot. Replaces runner defaults; flat ai options override fields they set. Conversation, workflow, prompt body and result schema are supplied by the grader."
	spec.Required = nil
	return fixtureJSONSchema{
		"$defs":                captain.Definitions,
		"type":                 "object",
		"description":          "YAML frontmatter configures defaults for every fixture in the markdown document.",
		"additionalProperties": true,
		"x-help": fixtureHelpBlock(
			"File structure",
			"Each fixture file can start with YAML frontmatter containing global defaults such as build, daemon, exec, args, env, cwd, files, AI, and verify settings.",
			"fixtures --help",
		),
		"x-order": []string{
			"build", "daemon", "exec", "args", "env", "cwd", "terminal", "setup", "traces", "goProfiles", "files", "codeBlocks",
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
			"setup":      fixtureSetupSchema(),
			"traces":     fixtureTracesSchema(),
			"goProfiles": fixtureGoProfilesSchema(),
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
					"spec":            spec,
					"criteriaSection": stringProp("Criteria section", "Heading whose task list the grader evaluates; empty reads the whole document."),
					"model":           withHelp(stringProp("Model", "Agent model name, falling back to the default model when unset."), "AI options", "Agent model name.", "fixtures --help", "claude-code-sonnet"),
					"temperature":     withHelp(numberProp("Temperature", "Sampling temperature.", 0, 2), "AI options", "Sampling temperature.", "fixtures --help", 0),
					"maxTokens":       withHelp(integerProp("Max tokens", "Maximum response tokens.", 0, 0), "AI options", "Maximum response tokens.", "fixtures --help", 10000),
					"maxConcurrent":   withHelp(integerProp("Max concurrent", "Maximum concurrent AI checks.", 1, 0), "AI options", "Maximum concurrent AI checks.", "fixtures --help", 4),
					"cacheTTL":        withHelp(stringProp("Cache TTL", "Cache duration for AI calls."), "AI options", "Cache duration for AI calls.", "fixtures --help", "10m"),
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

// fixtureSetupSchema describes the file-level `setup:` block: the environment
// every fixture in the document runs in. It is prepared once per markdown file,
// before `build:`, and torn down after the file's last test.
func fixtureSetupSchema() fixtureJSONSchema {
	return fixtureJSONSchema{
		"type":                 "object",
		"description":          "Environment prepared once per markdown file, before build, and torn down after its last test.",
		"additionalProperties": true,
		"x-help": fixtureHelpBlock(
			"Setup",
			"`setup:` prepares dotenv files, environment variables, cloud/Kubernetes connections and a git checkout before any fixture in the document runs. It is file-level only — every test in the file shares one prepared tree.",
			"fixtures --help",
		),
		"x-order": []string{"cwd", "baseDir", "dotenv", "envVars", "checkout", "connections"},
		"properties": map[string]any{
			"cwd":     withHelp(stringProp("Working directory", "Directory the file's fixtures run in, defaulting to the markdown file's own directory."), "Setup", "Where the file's fixtures run. Relative paths resolve against the markdown file. A `checkout:` with a worktree overrides this with the prepared tree.", "fixtures --help", "."),
			"baseDir": withHelp(stringProp("Base directory", "Where clones and worktrees land, defaulting to a per-file directory under the user cache."), "Setup", "Clones and worktrees are written here. Defaults to a hash of the markdown path under the user cache directory so runs never write into the repository under test.", "fixtures --help", ".gavel/fixtures"),
			"dotenv":  withHelp(stringArrayProp("Dotenv files", "`.env` files loaded into the fixture environment, resolved relative to the markdown file."), "Setup", "Dotenv files loaded into the environment of every fixture in the document.", "fixtures --help", []string{".env.test"}),
			"envVars": fixtureSchemaProperty{
				"type":        "array",
				"title":       "Environment variables",
				"description": "Environment variables, either literal or sourced from a secret via valueFrom.",
				"x-help":      fixtureHelpBlock("Setup", "Environment variables for every fixture in the document. `valueFrom` resolves against a configured secret store.", "fixtures --help"),
				"items": fixtureJSONSchema{
					"type":                 "object",
					"additionalProperties": true,
					"properties": map[string]any{
						"name":  stringProp("Name", "Variable name."),
						"value": stringProp("Value", "Literal value."),
					},
				},
			},
			"checkout": fixtureSetupCheckoutSchema(),
			"connections": fixtureJSONSchema{
				"type":                 "object",
				"description":          "Cloud and Kubernetes connections whose credentials are injected into the fixture environment.",
				"additionalProperties": true,
				"x-help":               fixtureHelpBlock("Setup", "Named connections (aws, azure, gcp, kubernetes, …). Each accepts inline credentials or a `connection://namespace/name` reference, which requires a configured database.", "fixtures --help"),
			},
		},
	}
}

// fixtureTracesSchema describes `traces:`: fixtures.TracesSchema, whose kinds
// and params come from the trace host's catalog, with its help block.
func fixtureTracesSchema() fixtureJSONSchema {
	schema, err := fixtures.TracesSchema()
	if err != nil {
		panic(fmt.Sprintf("fixture traces schema: %v", err))
	}
	schema["x-help"] = fixtureHelpBlock(
		"Traces",
		"`traces:` arms commons-db trace kinds around a fixture: a recording proxy its children are pointed at (`http_proxy`, `pg_proxy`), its terminal (`terminal`), gavel's own HTTP and SQL (`http`, `sql_statements`) and its processes (`process_tree`, `process_samples`). Each kind's summary of the fixture's window is the CEL root named by the kind, and the records stay in the run's store under `.gavel/traces/<run>`.",
		"fixtures --help",
	)
	return schema
}

func fixtureSetupCheckoutSchema() fixtureJSONSchema {
	return fixtureJSONSchema{
		"type":                 "object",
		"description":          "Git checkout prepared before the file's fixtures run.",
		"additionalProperties": true,
		"x-help":               fixtureHelpBlock("Setup", "Checks out a local or remote repository before the file's fixtures run. With `worktree:` the fixtures run in a disposable tree; the source repository is never mutated.", "fixtures --help"),
		"x-order":              []string{"mode", "url", "path", "connection", "ref", "depth", "since", "worktree"},
		"properties": map[string]any{
			"mode":       withHelp(enumProp("Mode", "Checkout source. Inferred from url/path when unset.", []string{"none", "local", "remote"}), "Setup", "`local` checks out the repository at `path`; `remote` clones `url`; `none` disables the checkout.", "fixtures --help", "local"),
			"url":        withHelp(stringProp("URL", "Repository URL, with mode: remote."), "Setup", "Repository to clone in remote mode.", "fixtures --help", "https://github.com/flanksource/gavel"),
			"path":       withHelp(stringProp("Path", "Local repository path, with mode: local, resolved relative to the markdown file."), "Setup", "Local repository to check out, resolved relative to the markdown file.", "fixtures --help", "."),
			"connection": withHelp(stringProp("Connection", "Stored git credentials, as a `connection://namespace/name` reference."), "Setup", "Credentials for a private repository. Resolving a `connection://` reference requires a configured database.", "fixtures --help", "connection://git/github"),
			"ref":        withHelp(stringProp("Ref", "Branch, tag or commit to check out."), "Setup", "Branch, tag or commit to check out.", "fixtures --help", "v1.0.42"),
			"depth":      withHelp(integerProp("Depth", "Shallow-clone depth.", 0, 0), "Setup", "Shallow-clone depth for remote checkouts.", "fixtures --help", 1),
			"since":      withHelp(stringProp("Since", "Commit-ish whose merge-base diff against HEAD is folded into the reported changed files."), "Setup", "Informational only: widens the reported `dirtyFiles` to include the merge-base diff against this commit. It does not change what the worktree contains.", "fixtures --help", "main"),
			"worktree":   fixtureSetupWorktreeSchema(),
		},
	}
}

func fixtureSetupWorktreeSchema() fixtureJSONSchema {
	return fixtureJSONSchema{
		"type":                 "object",
		"description":          "Disposable worktree the file's fixtures run in.",
		"additionalProperties": true,
		"x-help":               fixtureHelpBlock("Setup", "A worktree isolates the file from the rest of the repository — not its tests from each other, since every test in the file shares the same tree. Nothing is ever stashed: the source repository is never mutated.", "fixtures --help"),
		"x-order":              []string{"mode", "base", "uncommitted", "ignored", "prefix", "path", "branch", "keep"},
		"properties": map[string]any{
			"mode":   withHelp(enumProp("Mode", "Worktree lifecycle.", []string{"none", "new", "existing", "branch"}), "Setup", "`new` creates a disposable worktree and removes it afterwards; `existing` reuses the one at `path`; `branch` checks out the existing `branch` as it is; `none` runs in the checkout itself.", "fixtures --help", "new"),
			"branch": withHelp(stringProp("Branch", "Existing branch to check out, with mode: branch."), "Setup", "The branch must exist and not be checked out in another worktree. It is checked out as it is and never deleted at teardown.", "fixtures --help"),
			"base":   withHelpDefault(stringProp("Base", "Commit-ish the new worktree branches from."), "HEAD", "Setup", "Defaults to HEAD so the start commit is the tree you are looking at, independent of `checkout.ref`.", "fixtures --help", "HEAD"),
			// No `default:` for uncommitted — its default is conditional on
			// base, and a static `default: clone` in the schema would tell
			// editor completion something that is only sometimes true.
			"uncommitted": withHelp(enumProp("Uncommitted changes", "Whether staged, unstaged and untracked changes are carried into the new worktree. Defaults to `clone` when `base` is HEAD, otherwise `skip` — uncommitted work is a diff against your HEAD, so replaying it onto a worktree branched elsewhere applies to the wrong context.", []string{"clone", "skip"}), "Setup", "`clone` carries staged, unstaged and untracked changes across. `skip` gives a pristine tree at `base`. The source repository is never mutated either way.", "fixtures --help", "clone"),
			"ignored":     withHelpDefault(enumProp("Ignored files", "Whether gitignored content is copied into the new worktree.", []string{"clone", "skip"}), "clone", "Setup", "`git worktree add` never brings gitignored content, so without `clone` the worktree is missing node_modules/, .env and build caches. Set `skip` in repositories with a very large ignored tree.", "fixtures --help", "clone"),
			"prefix":      withHelp(stringProp("Branch prefix", "Prefix for the generated branch name. Leave unset unless you keep worktrees."), "Setup", "The branch is disposable and its name is generated. A prefix only helps you recognise leftovers from `keep: true`.", "fixtures --help"),
			"path":        withHelp(stringProp("Path", "Existing worktree path, with mode: existing."), "Setup", "Worktree to reuse in existing mode.", "fixtures --help"),
			"keep": fixtureSchemaProperty{
				"type":        "boolean",
				"title":       "Keep",
				"description": "Keep the worktree after the run for post-mortem inspection.",
				"x-help":      fixtureHelpBlock("Setup", "Worktrees are removed on cleanup unless kept.", "fixtures --help"),
			},
		},
	}
}

// withHelpDefault is withHelp for a property whose default the runtime actually
// applies, so editor completion shows what will happen rather than a blank.
func withHelpDefault(prop fixtureSchemaProperty, def any, section, body, source string, examples ...any) fixtureSchemaProperty {
	prop["default"] = def
	return withHelp(prop, section, body, source, examples...)
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

func fixtureGoProfilesSchema() fixtureJSONSchema {
	return fixtureJSONSchema{
		"type":        "object",
		"description": "Named Go pprof outputs checked after each exec when --profile is enabled. Missing outputs are valid.",
		"additionalProperties": fixtureJSONSchema{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"file":      stringProp("File", "Relative pprof filename under this fixture's isolated artifact directory."),
				"directory": stringProp("Directory", "Relative directory whose emitted files are Go pprof profiles."),
				"args":      stringArrayProp("Arguments", "Arguments appended to the exec only with --profile; use {{.path}} for the absolute destination."),
				"env":       stringProp("Environment variable", "Variable set to the absolute destination only with --profile."),
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
			"content", "exitCode", "cel", "stdout", "stderr", "output", "error", "format", "count", "timeout", "goProfiles", "traces", "properties",
		},
		"properties": map[string]any{
			"goProfiles": fixtureGoProfilesSchema(),
			"traces":     fixtureTracesSchema(),
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
