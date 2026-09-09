package main

import "github.com/flanksource/captain/pkg/api"

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
			"build", "daemon", "exec", "args", "env", "cwd", "terminal", "setup", "record", "files", "codeBlocks",
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
			"record":     fixtureRecordSchema(),
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

// fixtureRecordSchema describes the `record:` block: the diagnostic evidence a
// fixture captures beyond stdout, stderr and an exit code. Only the full mapping
// form is described here; the shorthands (`record: http`, `record: [ansi, http]`)
// are documented on the block itself, because a oneOf of a string, an array and
// an object would cost editor completion on the fields that matter.
func fixtureRecordSchema() fixtureJSONSchema {
	return fixtureJSONSchema{
		// Three surfaces, one schema: `record: http`, `record: [ansi, http]` and
		// the full mapping all parse. A type union rather than a oneOf so the
		// `properties` below keep driving completion for the mapping form, which
		// is the only surface with anything to complete.
		"type":                 []string{"string", "array", "object"},
		"description":          "Recorders capturing ANSI, HTTP and SQL traffic as artifacts under .gavel/recordings. Also accepts the shorthands `record: http` and `record: [ansi, http]`, or `record: none` to opt a fixture out of a file-level or --record default.",
		"examples":             []any{"http", []string{"ansi", "http"}, "none"},
		"items":                fixtureSchemaProperty{"type": "string", "enum": []string{"ansi", "http", "sql", "clients", "all", "none"}},
		"additionalProperties": true,
		"x-help": fixtureHelpBlock(
			"Recording",
			"`record:` captures evidence a failing fixture would otherwise not have: how the terminal rendered, what HTTP calls the child made, what SQL it issued. Artifacts land in `.gavel/recordings/` and their contents become CEL roots (`cast`, `http`, `sql`) the fixture can assert on. Nothing starts unless a fixture asks for it.",
			"fixtures --help",
		),
		"x-order": []string{"ansi", "http", "sql", "clients"},
		"properties": map[string]any{
			"ansi": fixtureJSONSchema{
				"type":                 "object",
				"description":          "Asciinema cast recorded from the fixture's PTY. Implies terminal: pty.",
				"additionalProperties": true,
				"x-help":               fixtureHelpBlock("Recording", "Records the terminal as an asciinema v2 cast, replayable with `asciinema play`. There is no ANSI to record from a pipe, so this implies `terminal: pty`.", "fixtures --help"),
				"x-order":              []string{"width", "height", "interval", "maxBytes"},
				"properties": map[string]any{
					"width":    withHelp(integerProp("Width", "PTY width in columns.", 1, 0), "Recording", "PTY width, which decides where output wraps.", "fixtures --help", 120),
					"height":   withHelp(integerProp("Height", "PTY height in rows.", 1, 0), "Recording", "PTY height.", "fixtures --help", 40),
					"interval": withHelp(stringProp("Snapshot interval", "How often the screen is sampled."), "Recording", "How often the rendered screen is sampled for the duplicate-line analysis.", "fixtures --help", "100ms"),
					"maxBytes": withHelp(stringProp("Max bytes", "Cap on the cast size, e.g. 4MiB."), "Recording", "Cap on the recorded cast. Accepts humanized sizes.", "fixtures --help", "4MiB"),
				},
			},
			"http": fixtureRecordHTTPSchema(),
			"sql": fixtureJSONSchema{
				"type":                 "object",
				"description":          "SQL captured as JSONL.",
				"additionalProperties": true,
				"x-help":               fixtureHelpBlock("Recording", "`proxy` sits between the child process and postgres and requires `sslmode=disable`. `inprocess` hooks gavel's own gorm logger and therefore sees nothing a child process does.", "fixtures --help"),
				"x-order":              []string{"mode", "dsn", "params", "scope"},
				"properties": map[string]any{
					"mode": withHelpDefault(enumProp("Mode", "Where SQL is observed.", []string{"off", "proxy", "inprocess"}), "proxy", "Recording", "`proxy` observes a child process; `inprocess` observes gavel's own queries.", "fixtures --help", "proxy"),
					"dsn":  withHelp(stringProp("DSN", "Upstream postgres the proxy forwards to. Read from the fixture environment when unset."), "Recording", "Upstream database. Left unset the recorder reads it from the fixture's own environment.", "fixtures --help", "$GAVEL_DB_DSN"),
					"params": fixtureSchemaProperty{
						"type":        "boolean",
						"title":       "Bind parameters",
						"description": "Keep bind parameter values in the artifact.",
						"x-help":      fixtureHelpBlock("Recording", "Off by default: bind parameters carry the row data, which is the most likely place for a secret.", "fixtures --help"),
					},
					"scope": fixtureRecordScopeProp(),
				},
			},
			"clients": fixtureJSONSchema{
				"type":                 "object",
				"description":          "HAR of gavel's own http.Client calls, as opposed to a child process's.",
				"additionalProperties": true,
				"x-help":               fixtureHelpBlock("Recording", "Records the HTTP calls gavel itself makes while running the fixture. Separate from `http:` because those clients live all over the gavel process rather than inside a fixture's call stack.", "fixtures --help"),
				"x-order":              []string{"bodies", "redact"},
				"properties": map[string]any{
					"bodies": withHelp(stringProp("Bodies", "Cap on captured body size, e.g. 64KiB."), "Recording", "How much of each body is written to the HAR.", "fixtures --help", "64KiB"),
					"redact": withHelp(stringArrayProp("Redact", "Extra header names to blank out."), "Recording", "Blanked in addition to the built-in denylist, which is always applied and cannot be disabled.", "fixtures --help", []string{"x-my-token"}),
				},
			},
		},
	}
}

func fixtureRecordHTTPSchema() fixtureJSONSchema {
	return fixtureJSONSchema{
		"type":                 "object",
		"description":          "HTTP proxy the child process is pointed at, written as a HAR 1.2 document.",
		"additionalProperties": true,
		"x-help": fixtureHelpBlock(
			"Recording",
			"Points the child at a recording proxy through HTTP_PROXY/HTTPS_PROXY and writes what it sees as a HAR, openable in Chrome DevTools. In the default `connect` mode TLS payloads stay encrypted and each tunnel is recorded as one entry with its host, duration and byte counts; plain HTTP is always recorded in full.",
			"fixtures --help",
		),
		"x-order": []string{"mode", "hosts", "bodies", "redact", "requireEntries", "scope"},
		"properties": map[string]any{
			"mode":   withHelpDefault(enumProp("Mode", "How much of the traffic is visible.", []string{"off", "connect", "mitm"}), "connect", "Recording", "`connect` records TLS tunnels without decrypting them. `mitm` terminates TLS with a generated, ephemeral CA — opt-in, because whether a given runtime trusts that CA is best-effort. The CA is never installed into a system trust store.", "fixtures --help", "connect"),
			"hosts":  withHelp(stringArrayProp("Hosts", "Glob patterns limiting what is recorded. Empty records everything."), "Recording", "Only matching hosts are recorded, and under `mitm` only matching hosts are decrypted.", "fixtures --help", []string{"*.github.com"}),
			"bodies": withHelp(stringProp("Bodies", "Cap on captured body size, e.g. 64KiB. Zero in connect mode — a tunnel has no readable body."), "Recording", "How much of each request and response body is written to the HAR. Bodies over the cap are truncated and marked.", "fixtures --help", "64KiB"),
			"redact": withHelp(stringArrayProp("Redact", "Extra header names to blank out."), "Recording", "Blanked in addition to the built-in denylist (authorization, proxy-authorization, cookie, set-cookie, x-api-key, x-auth-token), which is always applied on disk and cannot be disabled.", "fixtures --help", []string{"x-my-token"}),
			"requireEntries": withHelp(integerProp("Require entries", "Fail the fixture when fewer than N entries were recorded.", 0, 0), "Recording",
				"Guards against the silent failure: if the child cannot verify the generated CA its TLS handshake fails, nothing is recorded, and the fixture passes with an empty HAR. Set this and that becomes a red test.", "fixtures --help", 1),
			"scope": fixtureRecordScopeProp(),
		},
	}
}

func fixtureRecordScopeProp() fixtureSchemaProperty {
	return withHelpDefault(enumProp("Scope", "Whether the recorder is shared by the whole file or started per test.", []string{"file", "test"}), "file", "Recording",
		"Under `file` one recorder serves every test in the document, so attributing an entry to a single test is a time-slice heuristic — the file's tests run concurrently and overlap. `test` gives each test its own recorder at the cost of one listener per test.", "fixtures --help", "file")
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
		"x-order":              []string{"mode", "base", "uncommitted", "ignored", "prefix", "path", "keep"},
		"properties": map[string]any{
			"mode": withHelp(enumProp("Mode", "Worktree lifecycle.", []string{"none", "new", "existing"}), "Setup", "`new` creates a disposable worktree and removes it afterwards; `existing` reuses the one at `path`; `none` runs in the checkout itself.", "fixtures --help", "new"),
			"base": withHelpDefault(stringProp("Base", "Commit-ish the new worktree branches from."), "HEAD", "Setup", "Defaults to HEAD so the start commit is the tree you are looking at, independent of `checkout.ref`.", "fixtures --help", "HEAD"),
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
