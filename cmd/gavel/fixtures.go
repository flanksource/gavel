package main

import (
	"fmt"
	"os"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/kinds/all"
	"github.com/spf13/cobra"
)

var (
	fixturesUpdateGolden    bool
	fixturesShowPassed      bool
	fixturesShowStdout      string
	fixturesShowStderr      string
	fixturesSchema          bool
	fixturesTraces          string
	fixturesBenchmark       bool
	fixturesBaseline        string
	fixturesProfile         bool
	fixturesMaxDeviationPct float64
	fixturesMaxTime         time.Duration
	fixturesMaxMemory       string
	fixturesMaxIORead       string
	fixturesMaxIOWrite      string
	fixturesMaxSQLTime      time.Duration
	fixturesMaxSQLQuery     time.Duration
	fixturesMaxSQLQueries   int
	fixturesMaxSlowSQL      int
)

var fixturesCmd = &cobra.Command{
	Use:          "fixtures [fixture-files...]",
	Short:        "Run fixture-based tests from markdown tables and command blocks",
	Args:         fixturesArgs,
	RunE:         runFixtures,
	SilenceUsage: true,
}

func fixturesArgs(cmd *cobra.Command, args []string) error {
	if fixturesSchema {
		return nil
	}
	return cobra.MinimumNArgs(1)(cmd, args)
}

func fixturesHelp(cmd *cobra.Command) api.Text {
	h := func(title string) api.Text { return clicky.Text("\n"+title, "font-bold text-cyan-400").NewLine() }
	sh := func(title string) api.Text { return clicky.Text("  "+title, "font-bold text-blue-400").NewLine() }
	code := func(s string) api.Text { return clicky.Text(s, "text-green-400") }
	dim := func(s string) api.Text { return clicky.Text(s, "text-muted") }
	kv := func(k, v string) api.Text {
		return clicky.Text("    ").Append(k, "text-yellow-400").Append("  " + v).NewLine()
	}

	t := clicky.Text("Run fixture-based tests defined in markdown files.", "font-bold").NewLine().
		NewLine().
		Append("Fixtures are markdown files that define test cases using tables or command blocks.").NewLine().
		Append("Each file can have optional YAML front-matter for global configuration.").NewLine()

	t = t.Add(h("USAGE")).
		Add(code("  gavel fixtures [flags] <fixture-file-or-glob> [fixture-file-or-glob...]")).NewLine().
		Add(code("  gavel fixtures --ui checks.fixture.md")).Add(dim("  # stream the typed execution tree in the test UI")).NewLine().
		Add(h("ARGUMENTS")).
		Add(kv("fixture-files", "One or more markdown fixture files or doublestar glob patterns. At least one is required.")).
		Append("    ").Add(dim("Quote globs when you want Gavel to expand them instead of your shell.")).NewLine()

	// File structure
	t = t.Add(h("FILE STRUCTURE")).
		Add(code(`  ---
  build: go build -o myapp`)).Add(dim("           # Shell command run once before all tests")).NewLine().
		Add(code("  daemon: go run ./server --port {{.port}}")).Add(dim("  # Background command run before tests; stopped after")).NewLine().
		Add(code("  exec: ./myapp")).Add(dim("                      # Default executable (default: bash)")).NewLine().
		Add(code("  args: [--verbose]")).Add(dim("                  # Default arguments for exec")).NewLine().
		Add(code("  env:")).Add(dim("                               # Environment variables for all tests")).NewLine().
		Add(code("    LOG_LEVEL: debug")).NewLine().
		Add(code("  cwd: ./testdir")).Add(dim("                     # Default working directory")).NewLine().
		Add(code("  terminal: pty")).Add(dim("                      # Use pseudo-terminal (merges stdout/stderr)")).NewLine().
		Add(code("  setup:")).Add(dim("                             # Environment prepared once per file (see SETUP)")).NewLine().
		Add(code("    dotenv: [.env.test]")).NewLine().
		Add(code("    checkout: {mode: local, worktree: {mode: new}}")).NewLine().
		Add(code("  traces: [http_proxy]")).Add(dim("               # Arm trace kinds around each fixture (see TRACES)")).NewLine().
		Add(code("  files: \"**/*.go\"")).Add(dim("                   # Glob pattern: replicate tests per matching file")).NewLine().
		Add(code("  codeBlocks: [bash, python]")).Add(dim("         # Languages to execute (default: [bash])")).NewLine().
		Add(code("  timeout: 30s")).Add(dim("                       # Total timeout for test execution")).NewLine().
		Add(code("  os: linux")).Add(dim("                          # Skip on other OSes (prefix ! to negate: !darwin)")).NewLine().
		Add(code("  arch: amd64")).Add(dim("                        # Skip on other architectures")).NewLine().
		Add(code("  skip: \"! command -v docker\"")).Add(dim("        # Skip if command exits 0")).NewLine().
		Add(code("  ai:")).Add(dim("                                # AI verification fixture config")).NewLine().
		Add(code("    model: \"provider/model\"")).NewLine().
		Add(code("    temperature: 0")).NewLine().
		Add(code("    maxTokens: 10000")).NewLine().
		Add(code("    maxConcurrent: 4")).NewLine().
		Add(code("    cacheTTL: 10m")).NewLine().
		Add(code("    noCache: false")).NewLine().
		Add(code("  verify:")).Add(dim("                            # AI verification scoring options")).NewLine().
		Add(code("    scope: diff")).NewLine().
		Add(code("    threshold: 80")).NewLine().
		Add(code("    disabled: [style]")).NewLine().
		Add(code("  ---")).NewLine()

	// Format 1: Markdown tables
	t = t.Add(h("FORMAT 1: MARKDOWN TABLES")).
		Append("  Each row defines a test. Column headers map to fixture fields:").NewLine().NewLine().
		Add(code("  | Name | CLI        | Args     | Exit Code | CEL               |")).NewLine().
		Add(code("  |------|------------|----------|-----------|-------------------|")).NewLine().
		Add(code("  | test | ./myapp    | --help   | 0         | stdout.contains() |")).NewLine().NewLine().
		Add(sh("Input columns")).
		Add(kv("name, test name", "Test name (required)")).
		Add(kv("cli, command, exec", "Executable to run")).
		Add(kv("cli args, args, arguments", "Arguments (space-separated)")).
		Add(kv("cwd, dir, working directory", "Working directory")).
		Add(kv("terminal, term", "Terminal mode (\"pty\" for pseudo-terminal)")).
		Add(kv("os", "OS constraint (e.g. \"linux\", \"!darwin\")")).
		Add(kv("arch", "Architecture constraint (e.g. \"amd64\")")).
		Add(kv("skip", "Bash command; exit 0 = skip test")).
		Add(kv("traces, trace", "Trace kinds for this row (comma-separated), or none")).
		Add(kv("query", "Query string")).NewLine().
		Add(sh("Expectation columns")).
		Add(kv("exit code, exitcode", "Expected exit code (default: 0, \"-\" to skip)")).
		Add(kv("expected output, output, stdout", "Expected stdout (literal or @file reference)")).
		Add(kv("stderr", "Expected stderr (literal or @file reference)")).
		Add(kv("expected error, error", "Expected stderr substring (implies non-zero exit)")).
		Add(kv("expected format, format", "Output format validation (json, yaml)")).
		Add(kv("expected count, count", "Expected result count for fixture types that expose counts")).
		Add(kv("expected matches, matches", "Expected output/result matcher value")).
		Add(kv("expected results, results", "Expected structured result value")).
		Add(kv("expected files, files", "Expected file result value")).
		Add(kv("template output", "Expected templated output value")).
		Add(kv("cel, validation, expr", "CEL validation expression")).NewLine().
		Append("  Unrecognized columns become Properties available in CEL.", "text-muted").NewLine()

	// Format 2: Command blocks
	t = t.Add(h("FORMAT 2: COMMAND BLOCKS")).
		Append("  Use heading ").Add(code("### command: <test name>")).Append(" followed by code blocks:").NewLine().NewLine().
		Add(code("  ### command: my test\n  ```yaml\n  cwd: ./testdir\n  exitCode: 0\n  terminal: pty\n  os: linux\n  env:\n    KEY: value\n  ```\n  ```bash\n  echo \"hello world\"\n  ```")).NewLine().NewLine().
		Append("  YAML fields: ", "text-muted").Add(code("cwd, exitCode, env, timeout, terminal, os, arch, skip, traces")).NewLine().NewLine().
		Add(sh("Validations")).
		Append("    ").Add(code("* cel: stdout.contains(\"hello\")")).NewLine().
		Append("    ").Add(code("* contains: hello")).NewLine().
		Append("    ").Add(code("* regex: .*world.*")).NewLine().
		Append("    ").Add(code("* not: contains: error")).NewLine()

	t = t.Add(h("FORMAT 3: AI VERIFICATION FIXTURES")).
		Append("  Add ").Add(code("ai:")).Append(" front-matter to turn the markdown document into one AI verification step.").NewLine().
		Append("  The first ").Add(code("```prompt")).Append(" or ").Add(code("```ai")).Append(" block becomes reviewer instructions.").NewLine().
		Append("  GitHub task-list items become scored acceptance criteria.").NewLine().NewLine().
		Add(sh("Example")).
		Add(code("  ---\n  ai:\n    model: \"claude-code-sonnet\"\n  verify:\n    threshold: 80\n  ---\n\n  # Verify feature X\n\n  Some prose describing the expected change.\n\n  ```prompt\n  Focus on the new parser path and acceptance criteria.\n  ```\n\n  ## Acceptance Criteria\n  - [ ] Parser errors are actionable\n  - [ ] Existing fixture formats still pass\n\n  - cel: json.score >= 80")).NewLine().NewLine().
		Add(sh("AI options")).
		Add(kv("model", "Agent model name (falls back to the default model)")).
		Add(kv("temperature", "Sampling temperature")).
		Add(kv("maxTokens", "Maximum response tokens")).
		Add(kv("maxConcurrent", "Maximum concurrent AI checks")).
		Add(kv("cacheTTL", "Cache duration for AI calls")).
		Add(kv("noCache", "Disable AI response cache")).NewLine().
		Add(sh("Verify options")).
		Add(kv("scope", "Review scope, defaulting to the working-tree diff")).
		Add(kv("threshold", "Minimum passing score, defaulting to 80")).
		Add(kv("disabled", "Check IDs to disable for this fixture")).NewLine()

	t = t.Add(h("FORMAT 4: TEST / LINT STEPS")).
		Append("  A ").Add(code("```yaml test")).Append(" or ").Add(code("```yaml lint")).Append(" fence (bare ").Add(code("```test")).Append(" / ").Add(code("```lint")).Append(" also work; the").NewLine().
		Append("  leading ").Add(code("yaml")).Append(" is optional and only there for editor highlighting) runs the test / lint").NewLine().
		Append("  engine. The YAML body is unmarshalled directly onto the ").Add(code("gavel test")).Append(" / ").Add(code("gavel lint")).Append(" options, so").NewLine().
		Append("  every CLI flag is available as a key (kebab-case). Each test / violation becomes a child node.").NewLine().NewLine().
		Add(sh("Example")).
		Add(code("  ## Run the engine tests\n  ```yaml test\n  paths: [./testrunner/...]\n  framework: [go]\n  test-timeout: 2m\n  show-passed: true\n  ```\n\n  ## Lint the changed files\n  ```lint\n  linters: [golangci-lint, ruff]\n  changed: true\n  fix: false\n  ```")).NewLine().NewLine().
		Add(sh("Common test keys")).Append("    ").Add(dim("(full list: ")).Add(code("gavel test --help")).Add(dim(")")).NewLine().
		Add(kv("paths", "Package paths to test (empty = discover all)")).
		Add(kv("framework", "Restrict to jest, vitest, playwright, go, ginkgo")).
		Add(kv("changed, since, failed, baseline", "Narrow to changed / prior-failed / new-vs-baseline packages")).
		Add(kv("extra-args", "Arguments forwarded to the underlying runner")).
		Add(kv("test-timeout, timeout", "Per-package / whole-run deadline (e.g. 2m)")).
		Add(kv("lint", "Also run linters in parallel with the tests")).NewLine().
		Add(sh("Common lint keys")).Append("    ").Add(dim("(full list: ")).Add(code("gavel lint --help")).Add(dim(")")).NewLine().
		Add(kv("linters", "Only run these linters (empty = every detected linter)")).
		Add(kv("files", "Target paths to lint")).
		Add(kv("ignore", "Glob patterns to exclude")).
		Add(kv("fix", "Apply auto-fixes")).
		Add(kv("changed, since, baseline, failed", "Only report new / prior violations")).
		Add(kv("timeout", "Per-linter timeout (e.g. 5m)")).NewLine().
		Add(sh("Rendering")).
		Add(kv("show-passed", "Add passing tests / clean linters as child nodes (default: only failures)")).
		Add(kv("show-failed", "Add failing tests / violations as child nodes (default: true)")).NewLine().
		Append("  UI / detach options (", "text-muted").Add(code("ui")).Append(", ", "text-muted").Add(code("addr")).Append(", ", "text-muted").Add(code("detach")).Append(") are ignored, and a test step never", "text-muted").NewLine().
		Append("  recurses into fixture discovery.", "text-muted").NewLine()

	// Supported languages
	t = t.Add(h("SUPPORTED LANGUAGES")).
		Add(kv("bash, sh, shell", "bash -c <content>")).
		Add(kv("python, py, python3", "python -c <content>")).
		Add(kv("typescript, ts", "ts-node -e <content>")).
		Add(kv("javascript, js", "node -e <content>")).
		Add(kv("pwsh, powershell", "pwsh -Command <content>")).
		Add(kv("go", "go <content>")).
		Add(kv("test, lint (or yaml test/yaml lint)", "Run the test / lint engine (see FORMAT 4)")).NewLine().
		Append("  Non-executable labels (parsed as config): ", "text-muted").Add(code("yaml, frontmatter, json")).NewLine()

	// Inline code fence attributes
	t = t.Add(h("INLINE CODE FENCE ATTRIBUTES")).
		Append("  Attributes on the opening fence override YAML block values:").NewLine().NewLine().
		Append("    ").Add(code("```bash exitCode=1 timeout=30")).NewLine().
		Append("    ").Add(code("exit 1")).NewLine().
		Append("    ").Add(code("```")).NewLine().NewLine().
		Append("  Supported: ", "text-muted").Add(code("exitCode=N")).Append(" (integer), ", "text-muted").Add(code("timeout=N")).Append(" (seconds).", "text-muted").NewLine()

	// Validation shorthand
	t = t.Add(h("VALIDATION SHORTHAND")).
		Append("  Bullet lists after a code block define validations (joined with &&):").NewLine().NewLine().
		Add(kv("cel: <expr>", "Raw CEL expression")).
		Add(kv("contains: <text>", "stdout.contains(\"<text>\")")).
		Add(kv("regex: <pattern>", "stdout.matches(\"<pattern>\")")).
		Add(kv("not: contains: <text>", "!stdout.contains(\"<text>\")")).
		Add(kv("not: <expr>", "!(<expr>)")).NewLine()

	// CEL Validation
	t = t.Add(h("CEL VALIDATION")).
		Append("  Expressions must evaluate to ").Add(code("true")).Append(".").NewLine().NewLine().
		Add(sh("Variables")).
		Add(kv("stdout", "string    Process stdout")).
		Add(kv("output", "string    Alias for process stdout")).
		Add(kv("stderr", "string    Process stderr")).
		Add(kv("exitCode", "int       Process exit code")).
		Add(kv("json", "any       Auto-parsed JSON (when stdout starts with { or [)")).
		Add(kv("name", "string    Test name")).
		Add(kv("sourceDir", "string    Directory containing the fixture file")).
		Add(kv("query", "string    Query string")).
		Add(kv("expectations", "object    Expected values and custom table properties")).
		Add(kv("workDir", "string    Working directory")).
		Add(kv("executablePath", "string    Path to the gavel binary")).NewLine().
		Add(sh("Auto-injected variables (use as $VAR or {{.VAR}})")).
		Add(kv("GIT_ROOT_DIR", "string    Nearest parent with .git")).
		Add(kv("GO_ROOT_DIR", "string    Nearest parent with go.mod")).
		Add(kv("ROOT_DIR", "string    GIT_ROOT_DIR > GO_ROOT_DIR > workDir")).
		Add(kv("CWD", "string    Resolved working directory")).
		Add(kv("GOOS", "string    Go runtime OS (linux, darwin, ...)")).
		Add(kv("GOARCH", "string    Go runtime arch (amd64, arm64, ...)")).
		Add(kv("GOPATH", "string    Go workspace path")).NewLine().
		Add(sh("ANSI detection")).
		Add(kv("ansi.has_color", "bool   Output contains ANSI color codes")).
		Add(kv("ansi.has_any", "bool   Output contains any ANSI escape sequences")).
		Add(kv("ansi.has_updates", "bool   Output contains cursor movement codes")).NewLine().
		Add(kv("ansi.has_cursor_hide", "bool   Output hides the cursor")).
		Add(kv("ansi.has_cursor_show", "bool   Output shows the cursor")).
		Add(kv("ansi.has_reset", "bool   Output contains an SGR reset")).
		Add(kv("ansi.stray_controls", "bool   Output contains unexpected control bytes")).
		Add(kv("ansi.final_text", "string Final settled terminal text")).
		Add(kv("ansi.duplicate_lines", "list   Duplicate settled terminal lines")).
		Add(kv("ansi.has_duplicates", "bool   Duplicate settled terminal lines were found")).NewLine().
		Add(sh("File expansion variables")).
		Add(kv("file", "string    Relative path to matched file")).
		Add(kv("filename", "string    Filename without extension")).
		Add(kv("dir", "string    Directory containing the file")).
		Add(kv("absfile", "string    Absolute file path")).
		Add(kv("absdir", "string    Absolute directory")).
		Add(kv("basename", "string    Full filename with extension")).
		Add(kv("ext", "string    File extension")).NewLine().
		Add(sh("Temp file variables")).
		Append("  ", "text-muted").Append("(when ", "text-muted").Add(code("temp_files")).Append(" configured)", "text-muted").NewLine().
		Add(kv("<name>.path", "string    Path to temp file")).
		Add(kv("<name>.content", "string    File content")).
		Add(kv("<name>.ext", "string    File extension")).
		Add(kv("<name>.detected", "string    Detected type (text, json, xml, yaml)")).
		Add(kv("<name>.json", "any       Parsed JSON (if content is JSON)")).NewLine().
		Add(sh("Built-in CEL functions")).
		Append("    ").Add(code("string.contains(s)  startsWith(s)  endsWith(s)  matches(regex)")).NewLine().
		Append("    ").Add(code("size(list)  list.all(x, pred)  list.exists(x, pred)  list.filter(x, pred)")).NewLine().
		Append("    ").Add(code("has_color(s)  has_ansi(s)  has_cursor_updates(s)")).NewLine().
		Add(sh("Extended gomplate functions")).
		Append("    ").Add(code("strings.*  math.*  regexp.*  conv.*  coll.*  data.*  file.*  time.*")).NewLine()

	// Template variables
	t = t.Add(h("TEMPLATE VARIABLES")).
		Append("  The ").Add(code("exec")).Append(", ").Add(code("daemon")).Append(", ").Add(code("build")).Append(", ").Add(code("args")).Append(", and ").Add(code("cwd")).Append(" fields support shell-style $VAR and Go template syntax:").NewLine().NewLine().
		Add(code("    exec: $executablePath")).NewLine().
		Add(code("    daemon: go run ./server --port {{.port}}")).NewLine().
		Add(code("    args: [--file, \"$file\"]")).NewLine().
		Add(code("    build: go build -o $workDir/myapp")).NewLine().
		Add(code("    cwd: $GIT_ROOT_DIR/testdata")).NewLine().
		Add(kv("port", "Free TCP port available to daemon commands and fixture templates")).NewLine()

	// File expansion
	t = t.Add(h("FILE EXPANSION")).
		Append("  Set ").Add(code("files")).Append(" in front-matter to replicate each test per matching file:").NewLine().NewLine().
		Add(code("  ---\n  files: \"**/*.go\"\n  exec: golint\n  args: [\"{{.file}}\"]\n  ---")).NewLine()

	// Setup
	t = t.Add(h("SETUP")).
		Append("  ").Add(code("setup:")).Append(" declares the environment a file's tests need — dotenv files, environment").NewLine().
		Append("  variables, cloud/Kubernetes connections, and a git checkout or worktree to run in.").NewLine().NewLine().
		Append("  It is ", "text-muted").Append("file-level front-matter only", "font-bold").Append(": prepared once, shared by every test in the file,", "text-muted").NewLine().
		Append("  and torn down after the last one finishes. A per-test ", "text-muted").Add(code("setup:")).Append(" is an error. Because the", "text-muted").NewLine().
		Append("  file's tests still run concurrently in one group, a worktree isolates the file from the", "text-muted").NewLine().
		Append("  rest of the repository — it does not isolate the file's tests from each other.", "text-muted").NewLine().NewLine().
		Add(sh("Example")).
		Add(code("  ---\n  setup:\n    dotenv: [.env.test]\n    envVars:\n      - name: LOG_LEVEL\n        value: debug\n    checkout:\n      mode: local          # none | local | remote\n      worktree:\n        mode: new          # none | new | existing\n        base: HEAD         # commit-ish to branch from\n        uncommitted: clone # clone | skip — staged + unstaged + untracked\n        ignored: clone     # clone | skip — gitignored content\n        keep: false\n    connections:\n      aws: {connection: \"connection://aws/sandbox\", region: us-east-1}\n  build: go build -o ./bin/app\n  exec: ./bin/app\n  ---")).NewLine().NewLine().
		Add(sh("Options")).
		Add(kv("cwd", "Where the file's commands run (default: the markdown file's directory)")).
		Add(kv("baseDir", "Where clones and worktrees land (default: a per-file dir under the user cache)")).
		Add(kv("dotenv", "Dotenv files to load, resolved relative to the markdown file")).
		Add(kv("envVars", "Explicit name/value (or valueFrom) environment entries")).
		Add(kv("checkout", "Git checkout: mode none|local|remote, url, ref, path, connection, since")).
		Add(kv("checkout.worktree", "Disposable worktree: mode none|new|existing, base, uncommitted, ignored, prefix, keep")).
		Add(kv("connections", "Named cloud/Kubernetes connections; connection:// refs need a database")).NewLine().
		Add(sh("Guarantees")).
		Append("    * ", "text-muted").Append("The source repository is never mutated — nothing is stashed, moved, or restored.").NewLine().
		Append("    * ", "text-muted").Append("Ordering is setup → build → daemon → tests → daemon stop → setup cleanup, so").NewLine().
		Append("      ", "text-muted").Add(code("build:")).Append(" compiles the same tree the tests exercise.").NewLine().
		Append("    * ", "text-muted").Add(code("worktree.uncommitted")).Append(" defaults to ").Add(code("clone")).Append(" only when ").Add(code("base")).Append(" is HEAD; branching").NewLine().
		Append("      ", "text-muted").Append("elsewhere degrades it to ").Add(code("skip")).Append(" with a warning, because uncommitted work is a diff").NewLine().
		Append("      ", "text-muted").Append("against your HEAD. ").Add(code("worktree.ignored")).Append(" defaults to ").Add(code("clone")).Append(" so node_modules, .env and").NewLine().
		Append("      ", "text-muted").Append("build caches come along; set ").Add(code("ignored: skip")).Append(" when that tree is large.").NewLine().
		Append("    * ", "text-muted").Add(code("$SETUP_DIR")).Append(" (template var and child env var) is the prepared directory, and").NewLine().
		Append("      ", "text-muted").Add(code("GIT_ROOT_DIR")).Append("/").Add(code("ROOT_DIR")).Append(" are re-rooted onto it. ").Add(code("{{.setup}}")).Append(" carries commit,").NewLine().
		Append("      ", "text-muted").Append("worktree, path and dirtyFiles.").NewLine().
		Append("    * ", "text-muted").Append("Golden ").Add(code("@file")).Append(" expectations still resolve next to the markdown, never inside the").NewLine().
		Append("      ", "text-muted").Append("worktree — the commands move, their expectations do not.").NewLine().
		Append("    * ", "text-muted").Append("Environment precedence, highest first: fixture ").Add(code("env:")).Append(" > setup env > injected roots >").NewLine().
		Append("      ", "text-muted").Append("the inherited process environment. Setup env is additive, not hermetic.").NewLine()

	// Traces
	t = t.Add(h("TRACES")).
		Append("  ").Add(code("traces:")).Append(" arms commons-db trace kinds around a fixture and keeps their records in the").NewLine().
		Append("  run's store, ").Add(code(".gavel/traces/<run>")).Append(" (the newest 20 runs are kept). Each kind's summary of the").NewLine().
		Append("  fixture's window is the CEL root named by the kind. Nothing starts unless a fixture asks.").NewLine().NewLine().
		Add(sh("Surfaces")).
		Add(code("  traces: http_proxy")).Add(dim("                  # one kind, default params, test scope")).NewLine().
		Add(code("  traces: [http_proxy, terminal]")).Add(dim("      # several")).NewLine().
		Add(code("  traces: []")).Add(dim("                          # opt out of a file-level or --traces default")).NewLine().
		Add(code("  traces:")).Add(dim("                             # full form")).NewLine().
		Add(code("    - kind: http_proxy")).NewLine().
		Add(code("      params: {mode: mitm, hosts: [\"*.github.com\"]}")).NewLine().
		Add(code("      scope: test")).Add(dim("                     # test (default) | file | run")).NewLine().
		Add(code("      require: count > 0")).Add(dim("              # CEL over the summary; false fails a passing fixture")).NewLine().NewLine().
		Append("  Also settable per test (command-block YAML, a ", "text-muted").Add(code("Traces")).Append(" table column with the compact form or", "text-muted").NewLine().
		Append("  ", "text-muted").Add(code("none")).Append(", a fixture document or one of its steps) and run-wide with ", "text-muted").Add(code("--traces")).Append(". A per-test value", "text-muted").NewLine().
		Append("  replaces the file's outright. An unknown kind or params the kind refuses fail the parse.", "text-muted").NewLine().NewLine().
		Add(sh("Kinds")).
		Add(kv("http_proxy", "Recording proxy every child is pointed at (HTTP(S)_PROXY, CA trust vars under mitm)")).
		Add(kv("pg_proxy", "Postgres wire proxy in front of dsn (DATABASE_URL, PG* rewritten for children)")).
		Add(kv("terminal", "Runs the command under a PTY and records it; exports as an asciinema cast")).
		Add(kv("http", "gavel's own outbound HTTP (feature gavel)")).
		Add(kv("sql_statements", "gavel's own SQL (connections: [self])")).
		Add(kv("process_tree, process_samples", "Processes under a root and their CPU/memory")).NewLine().
		Add(sh("Scopes")).
		Append("    * ", "text-muted").Add(code("test")).Append(" starts a capture around the fixture; its window is the whole stream.").NewLine().
		Append("    * ", "text-muted").Add(code("file")).Append(" and ").Add(code("run")).Append(" share one capture across the file or run; each fixture's window is").NewLine().
		Append("      ", "text-muted").Append("marked before and after it, so overlapping rows of one file see each other's records.").NewLine().
		Append("    * ", "text-muted").Append("Every process a fixture starts — exec, services, browser — inherits the exports; a").NewLine().
		Append("      ", "text-muted").Append("variable the setup, front matter or a step sets itself wins.").NewLine().NewLine().
		Add(sh("Asserting")).
		Add(code("  cel: http_proxy.entries == 2 && http_proxy.errors == 0")).NewLine().
		Add(code("  cel: http_proxy.requests.filter(r, r.host == \"api.github.com\").size() == 2")).NewLine().
		Add(code("  cel: terminal.exit_code == 0 && terminal.duplicates.size() == 0")).NewLine().
		Add(code("  cel: pg_proxy.by_op[\"INSERT\"] == 1 && pg_proxy.errors == 0")).NewLine().NewLine().
		Append("  A trace's root exists only when it was armed; its result (kind, stream, window, count,", "text-muted").NewLine().
		Append("  summary) rides on the fixture result under ", "text-muted").Add(code("traces")).Append(".", "text-muted").NewLine().NewLine().
		Append("  The store also holds kinds gavel fills itself, which a fixture does not arm: ", "text-muted").Add(code("stacks")).NewLine().
		Append("  (goroutine/native stack dumps) and ", "text-muted").Add(code("goprof")).Append(" (the ", "text-muted").Add(code("goProfiles:")).Append(" a command wrote under --profile,", "text-muted").NewLine().
		Append("  listed with the fixture's traces but not a CEL root).", "text-muted").NewLine()

	t = t.Add(h("PROFILING AND BENCHMARKS")).
		Add(sh("--profile")).
		Append("  Reads the fixture process tree's CPU, memory and disk I/O from the run's ").Add(code("process_tree")).Append(" and").NewLine().
		Append("  ").Add(code("process_samples")).Append(": each step's peaks over its own time, its disk I/O as the delta of the counters.").NewLine().
		Append("  It also hands each command the destinations its ").Add(code("goProfiles:")).Append(" declare and ingests what it wrote:").NewLine().NewLine().
		Add(code("  goProfiles:")).NewLine().
		Add(code("    cpu:")).NewLine().
		Add(code("      file: cpu.pprof")).Add(dim("                # or directory: — every file written under it")).NewLine().
		Add(code("      env: CPU_PROFILE")).Add(dim("               # and/or args: [\"-cpuprofile={{.path}}\"]")).NewLine().NewLine().
		Append("    * ", "text-muted").Append("Paths are relative to a per-fixture dir under ").Add(code(".gavel/profiles/fixtures/")).Append(".").NewLine().
		Append("    * ", "text-muted").Append("The name (cpu, heap, allocs, mutex, block, goroutine) says which profile it is; any").NewLine().
		Append("      ", "text-muted").Append("other name is read from the profile's sample types. A file not written is logged and skipped;").NewLine().
		Append("      ", "text-muted").Append("one that is not a Go pprof profile fails the fixture.").NewLine().
		Append("    * ", "text-muted").Append("Each file is one ").Add(code("goprof")).Append(" window; download it as ").Add(code("pprof")).Append(" (.pb.gz) for ").Add(code("go tool pprof")).Append(".").NewLine().
		Append("    * ", "text-muted").Add(code("SQL_PROFILE_FILE")).Append(" (also set by any --max-sql-* limit) is passed to every command for").NewLine().
		Append("      ", "text-muted").Append("commons-db's GORM SQL logger to write one JSON line per statement.").NewLine().NewLine().
		Add(sh("--benchmark")).
		Append("  Records each fixture's and prerequisite's (setup, build, daemon) duration, with its profile and").NewLine().
		Append("  SQL totals, into ").Add(code(".gavel/benchmarks/fixtures/run-<timestamp>.json")).Append(". Pass a saved report as").NewLine().
		Append("  ").Add(code("--baseline")).Append(" to compare against it. Limits fail individual fixtures:").NewLine().NewLine().
		Add(kv("--max-time", "Duration")).
		Add(kv("--max-memory", "Peak RSS byte size (512MiB)")).
		Add(kv("--max-io-read, --max-io-write", "Observed disk byte sizes")).
		Add(kv("--max-sql-time, --max-sql-query", "Total SQL time, slowest statement")).
		Add(kv("--max-sql-queries, --max-slow-sql", "Statement counts (zero allowed)")).
		Add(kv("--max-deviation-pct", "Regression vs --baseline: time, peak RSS, disk I/O, SQL")).NewLine()

	// CWD resolution
	t = t.Add(h("CWD RESOLUTION")).
		Append("  Working directory is resolved with the following priority:").NewLine().NewLine().
		Append("    1. ", "text-yellow-400").Append("Test-level CWD").Append(" (per-test frontmatter or table column)", "text-muted").NewLine().
		Append("    2. ", "text-yellow-400").Append("File-level CWD").Append(" (YAML front-matter at top of file)", "text-muted").NewLine().
		Append("    3. ", "text-yellow-400").Append("Prepared setup").Append(" (the checkout or worktree from ", "text-muted").Add(code("setup:")).Append(")", "text-muted").NewLine().
		Append("    4. ", "text-yellow-400").Append("SourceDir").Append(" (directory containing the fixture file)", "text-muted").NewLine().
		Append("    5. ", "text-yellow-400").Add(code("--cwd")).Append(" flag or current working directory", "text-muted").NewLine().NewLine().
		Append("  Relative CWD paths are resolved from the prepared setup when the file declared one,", "text-muted").NewLine().
		Append("  otherwise from SourceDir.", "text-muted").NewLine()

	// Execution
	t = t.Add(h("EXECUTION")).
		Append("  Tests run in parallel with a 2-minute default timeout per test").NewLine().
		Append("  and 5-minute timeout for the build step. Each file's ").Add(code("setup:")).Append(" is prepared").NewLine().
		Append("  first, so the build command — which runs once before any tests —").NewLine().
		Append("  compiles inside the prepared tree. A daemon command, when configured,").NewLine().
		Append("  starts after build, waits for its free port to accept connections, and").NewLine().
		Append("  is stopped after all fixtures finish; setups are torn down last.").NewLine()

	t = t.Add(h("OUTPUT OPTIONS")).
		Add(kv("-v", "Show passed fixture results")).
		Add(kv("-vv", "Also show executed commands")).
		Add(kv("-vvv", "Also show CEL variables and stdout/stderr by default")).
		Add(kv("--show-passed", "Show passed fixture results without increasing log verbosity")).
		Add(kv("--show-stdout", "When to show stdout: Never, OnFailure, Always")).
		Add(kv("--show-stderr", "When to show stderr: Never, OnFailure, Always")).
		Add(kv("--update-golden", "Rewrite mismatched @file stdout/stderr expectations")).NewLine()

	// Examples
	t = t.Add(h("EXAMPLES")).
		Add(code("  gavel fixtures tests.md")).Add(dim("                  # Run a single fixture file")).NewLine().
		Add(code("  gavel fixtures fixtures/**/*.md")).Add(dim("          # Run with glob")).NewLine().
		Add(code("  gavel fixtures -v tests.md")).Add(dim("               # Show passed fixtures")).NewLine().
		Add(code("  gavel fixtures -vv tests.md")).Add(dim("              # Also show commands")).NewLine().
		Add(code("  gavel fixtures -vvv tests.md")).Add(dim("             # Also show stdout/stderr")).NewLine().
		Add(code("  gavel fixtures --benchmark tests.md")).Add(dim("  # Save and show stage durations")).NewLine().
		Add(code("  gavel fixtures --benchmark --baseline .gavel/benchmarks/fixtures/run-...json tests.md")).NewLine().
		Add(code("  gavel fixtures --benchmark --max-time 2s --max-memory 512MiB tests.md")).NewLine().
		Add(code("  gavel fixtures --benchmark --baseline base.json --max-deviation-pct 15 tests.md")).NewLine().
		Add(code("  gavel fixtures --profile --ui tests.md")).Add(dim("  # Process metrics and declared Go pprof files")).NewLine().
		Add(code("  gavel fixtures --traces http_proxy tests.md")).Add(dim("  # Record every HTTP exchange; assert http_proxy.*")).NewLine().
		Add(code("  gavel fixtures --no-progress tests.md")).Add(dim("    # Disable progress display")).NewLine().
		Add(code("  gavel fixtures outline tests.md")).Add(dim("          # Parse and outline without running fixtures")).NewLine().
		Add(code("  gavel fixtures --schema")).Add(dim("                  # Print the fixture/test JSON schemas and exit")).NewLine().
		Add(renderHelpFlags("FLAGS", cmd.NonInheritedFlags())).
		Add(renderHelpFlags("GLOBAL FLAGS", cmd.InheritedFlags()))

	return t
}

func init() {
	fixturesCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// Cobra inherits a help func down the tree, so without this guard
		// `gavel fixtures outline --help` would print the parent's page and the
		// subcommand's own description would never be reachable. Delegating to
		// the parent chain resolves to cobra's default renderer.
		if cmd != fixturesCmd {
			fixturesCmd.Parent().HelpFunc()(cmd, args)
			return
		}
		fmt.Fprintln(os.Stderr, fixturesHelp(cmd).ANSI())
	})
	fixturesCmd.Flags().BoolVar(&fixturesUpdateGolden, "update-golden", false,
		"Rewrite @file stdout/stderr expectations with actual output instead of failing on mismatch")
	fixturesCmd.Flags().BoolVar(&fixturesShowPassed, "show-passed", false, "Show passed fixture results in output")
	fixturesCmd.Flags().StringVar(&fixturesShowStdout, "show-stdout", string(fixtures.OutputOnFailure),
		"When to show stdout: Never, OnFailure, Always")
	fixturesCmd.Flags().StringVar(&fixturesShowStderr, "show-stderr", string(fixtures.OutputOnFailure),
		"When to show stderr: Never, OnFailure, Always")
	fixturesCmd.Flags().BoolVar(&fixturesSchema, "schema", false, "Print fixture editor JSON schemas and exit")
	fixturesCmd.Flags().StringVar(&fixturesTraces, "traces", "",
		"Trace kinds armed, test-scoped with default params, for fixtures that declare no `traces:` of their own: http_proxy, pg_proxy, terminal, http, sql_statements, process_tree, process_samples (comma-separated)")
	fixturesCmd.Flags().BoolVar(&fixturesBenchmark, "benchmark", false, "Record and show fixture and prerequisite durations in .gavel/benchmarks/fixtures/run-<timestamp>.json")
	fixturesCmd.Flags().StringVar(&fixturesBaseline, "baseline", "", "Compare benchmark metrics with a saved fixture benchmark JSON report (see --max-deviation-pct)")
	fixturesCmd.Flags().BoolVar(&fixturesProfile, "profile", false, "Record the fixture process tree's CPU, memory and disk I/O, and ingest declared goProfiles pprof files into the run's goprof trace")
	fixturesCmd.Flags().Float64Var(&fixturesMaxDeviationPct, "max-deviation-pct", 0, "Fail fixtures whose measured metrics regress beyond this percentage from --baseline")
	fixturesCmd.Flags().DurationVar(&fixturesMaxTime, "max-time", 0, "Fail fixtures taking longer than this duration")
	fixturesCmd.Flags().StringVar(&fixturesMaxMemory, "max-memory", "", "Fail fixtures exceeding this peak RSS byte size (for example 512MiB)")
	fixturesCmd.Flags().StringVar(&fixturesMaxIORead, "max-io-read", "", "Fail fixtures exceeding this observed disk read byte size")
	fixturesCmd.Flags().StringVar(&fixturesMaxIOWrite, "max-io-write", "", "Fail fixtures exceeding this observed disk write byte size")
	fixturesCmd.Flags().DurationVar(&fixturesMaxSQLTime, "max-sql-time", 0, "Fail fixtures exceeding this total SQL duration")
	fixturesCmd.Flags().DurationVar(&fixturesMaxSQLQuery, "max-sql-query", 0, "Fail fixtures exceeding this longest SQL query duration")
	fixturesCmd.Flags().IntVar(&fixturesMaxSQLQueries, "max-sql-queries", 0, "Fail fixtures exceeding this SQL statement count")
	fixturesCmd.Flags().IntVar(&fixturesMaxSlowSQL, "max-slow-sql", 0, "Fail fixtures exceeding this slow SQL statement count")
	rootCmd.AddCommand(fixturesCmd)
}
