package main

import (
	"fmt"
	"os"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var (
	fixturesUpdateGolden bool
	fixturesShowPassed   bool
	fixturesShowStdout   string
	fixturesShowStderr   string
)

var fixturesCmd = &cobra.Command{
	Use:          "fixtures [fixture-files...]",
	Short:        "Run fixture-based tests from markdown tables and command blocks",
	Args:         cobra.MinimumNArgs(1),
	RunE:         runFixtures,
	SilenceUsage: true,
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
		Append("  YAML fields: ", "text-muted").Add(code("cwd, exitCode, env, timeout, terminal, os, arch, skip")).NewLine().NewLine().
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
		Add(kv("timeout", "Per-linter timeout (e.g. 5m)")).
		Add(kv("group-by", "Group synced TODOs by file, package or message")).NewLine().
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

	// CWD resolution
	t = t.Add(h("CWD RESOLUTION")).
		Append("  Working directory is resolved with the following priority:").NewLine().NewLine().
		Append("    1. ", "text-yellow-400").Append("Test-level CWD").Append(" (per-test frontmatter or table column)", "text-muted").NewLine().
		Append("    2. ", "text-yellow-400").Append("File-level CWD").Append(" (YAML front-matter at top of file)", "text-muted").NewLine().
		Append("    3. ", "text-yellow-400").Append("SourceDir").Append(" (directory containing the fixture file)", "text-muted").NewLine().
		Append("    4. ", "text-yellow-400").Add(code("--cwd")).Append(" flag or current working directory", "text-muted").NewLine().NewLine().
		Append("  Relative CWD paths are resolved from SourceDir.", "text-muted").NewLine()

	// Execution
	t = t.Add(h("EXECUTION")).
		Append("  Tests run in parallel with a 2-minute default timeout per test").NewLine().
		Append("  and 5-minute timeout for the build step. The build command runs").NewLine().
		Append("  once before any tests. A daemon command, when configured, starts").NewLine().
		Append("  after build, waits for its free port to accept connections, and is").NewLine().
		Append("  stopped after all fixtures finish.").NewLine()

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
		Add(code("  gavel fixtures --no-progress tests.md")).Add(dim("    # Disable progress display")).NewLine().
		Add(renderHelpFlags("FLAGS", cmd.NonInheritedFlags())).
		Add(renderHelpFlags("GLOBAL FLAGS", cmd.InheritedFlags()))

	return t
}

func runFixtures(cmd *cobra.Command, args []string) error {
	wd, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	runner, err := fixtures.NewRunner(fixtures.RunnerOptions{
		Paths:          args,
		Format:         clicky.Flags.ResolveFormat(),
		NoColor:        clicky.Flags.NoColor,
		WorkDir:        wd,
		MaxWorkers:     clicky.Flags.MaxConcurrent,
		Logger:         logger.StandardLogger(),
		ExecutablePath: executablePath,
		UpdateGolden:   fixturesUpdateGolden,
		Display: lo.ToPtr(fixtures.DisplayOptionsForVerbosity(clicky.Flags.LevelCount, fixtures.DisplayOptions{
			ShowPassed: fixturesShowPassed,
			ShowStdout: fixtures.ParseOutputMode(fixturesShowStdout),
			ShowStderr: fixtures.ParseOutputMode(fixturesShowStderr),
		}, cmd.Flags().Changed("show-stdout"), cmd.Flags().Changed("show-stderr"))),
	})
	if err != nil {
		return fmt.Errorf("failed to create fixture runner: %w", err)
	}

	tree, runErr := runner.Run()
	if tree != nil {
		if len(tree.Children) == 1 {
			fmt.Println(clicky.MustFormat(*tree.Children[0]))
		} else {
			fmt.Println(clicky.MustFormat(*tree))
		}
	}
	return runErr
}

func init() {
	fixturesCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr, fixturesHelp(cmd).ANSI())
	})
	fixturesCmd.Flags().BoolVar(&fixturesUpdateGolden, "update-golden", false,
		"Rewrite @file stdout/stderr expectations with actual output instead of failing on mismatch")
	fixturesCmd.Flags().BoolVar(&fixturesShowPassed, "show-passed", false, "Show passed fixture results in output")
	fixturesCmd.Flags().StringVar(&fixturesShowStdout, "show-stdout", string(fixtures.OutputOnFailure),
		"When to show stdout: Never, OnFailure, Always")
	fixturesCmd.Flags().StringVar(&fixturesShowStderr, "show-stderr", string(fixtures.OutputOnFailure),
		"When to show stderr: Never, OnFailure, Always")
	rootCmd.AddCommand(fixturesCmd)
}
