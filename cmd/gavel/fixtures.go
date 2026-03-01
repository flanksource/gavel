package main

import (
	"fmt"
	"os"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/spf13/cobra"
)

var fixturesCmd = &cobra.Command{
	Use:   "fixtures [fixture-files...]",
	Short: "Run fixture-based tests from markdown tables and command blocks",
	Long:  fixturesHelp,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runFixtures,
	Example: `  # Run a single fixture file
  gavel fixtures tests.md

  # Run multiple fixtures with glob
  gavel fixtures fixtures/**/*.md

  # Run with verbose output
  gavel fixtures -v tests.md`,
	SilenceUsage: true,
}

var fixturesHelp = `Run fixture-based tests defined in markdown files.

Fixtures are markdown files that define test cases using tables or command blocks.
Each file can have optional YAML front-matter for global configuration.

FILE STRUCTURE
  ---
  build: go build -o myapp           # Shell command run once before all tests
  exec: ./myapp                      # Default executable (default: bash)
  args: [--verbose]                  # Default arguments for exec
  env:                               # Environment variables for all tests
    LOG_LEVEL: debug
  cwd: ./testdir                     # Default working directory
  files: "**/*.go"                   # Glob pattern: replicate tests per matching file
  codeBlocks: [bash, python]         # Languages to execute (default: [bash])
  timeout: 30s                       # Total timeout for test execution
  ---

FORMAT 1: MARKDOWN TABLES

  Each row defines a test. Column headers map to fixture fields:

  | Name | CLI        | Args     | Exit Code | CEL               |
  |------|------------|----------|-----------|-------------------|
  | test | ./myapp    | --help   | 0         | stdout.contains() |

  Supported column headers (case-insensitive):

  Input columns:
    name, test name          Test name (required, rows without names are skipped)
    cli, command, exec       Executable to run
    cli args, args           Arguments (space-separated)
    cwd, working directory   Working directory
    query                    Query string

  Expectation columns:
    exit code, exitcode      Expected exit code (default: 0, use "-" to skip)
    expected output, output  Expected stdout (exact match)
    expected error, error    Expected stderr substring (implies non-zero exit)
    expected format, format  Output format validation (e.g., json, yaml)
    expected count, count    Expected count (use "-" to skip)
    expected matches         Expected output substring
    expected results         Expected output substring
    expected files           Expected output substring
    template output          Expected output substring
    cel, validation, expr    CEL validation expression

  Any unrecognized column header is stored in Properties and available
  in CEL expressions via expectations.Properties["column_name"].

FORMAT 2: COMMAND BLOCKS

  Use heading "### command: <test name>" followed by code blocks:

  ### command: my test
  ` + "```" + `yaml
  cwd: ./testdir
  exitCode: 0
  env:
    KEY: value
  timeout: 60
  ` + "```" + `

  ` + "```" + `bash
  echo "hello world"
  ` + "```" + `

  Validations:
  * cel: stdout.contains("hello")
  * contains: hello
  * regex: .*world.*
  * not: contains: error

  Command block YAML fields: cwd, exitCode, env (map), timeout.

FORMAT 3: STANDALONE CODE BLOCKS

  Executable code blocks outside a "command:" heading are auto-detected
  as tests. The test name comes from the nearest preceding heading.

  ## My Tests

  ` + "```" + `bash
  echo "auto-detected test"
  ` + "```" + `

  * contains: auto-detected

SUPPORTED LANGUAGES

  Language         Executor    Args format
  bash, sh, shell  bash        -c <content>
  python, py       python      -c <content>
  typescript, ts   ts-node     -e <content>
  javascript, js   node        -e <content>
  pwsh, powershell pwsh        -Command <content>
  go               go          <content>

  Non-executable labels (parsed as config): yaml, frontmatter

INLINE CODE FENCE ATTRIBUTES

  Attributes on the code fence override YAML block values:

  ` + "```" + `bash exitCode=1 timeout=30
  exit 1
  ` + "```" + `

  Supported: exitCode=N (integer), timeout=N (seconds).

VALIDATION SHORTHAND

  In bullet lists following a code block, these prefixes are supported:

  * cel: <expr>               Raw CEL expression
  * contains: <text>          Converts to stdout.contains("<text>")
  * regex: <pattern>          Converts to stdout.matches("<pattern>")
  * not: contains: <text>     Converts to !stdout.contains("<text>")
  * not: <expr>               Converts to !(<expr>)

  Multiple validations are joined with &&.

CEL VALIDATION

  CEL expressions must evaluate to true (boolean or string "true").

  Available variables:
    stdout         string    Process stdout
    output         string    Alias for stdout
    stderr         string    Process stderr
    exitCode       int       Process exit code
    json           any       Auto-parsed JSON (when stdout starts with { or [)
    name           string    Test name
    sourceDir      string    Directory containing the fixture file
    query          string    Query string (if set)
    expectations   object    Expected values from the table/config
    executablePath string    Path to the gavel binary
    workDir        string    Working directory

  File expansion variables (when "files" front-matter is set):
    file           string    Relative path to matched file
    filename       string    Filename without extension
    dir            string    Directory containing the file
    absfile        string    Absolute file path
    absdir         string    Absolute directory
    basename       string    Full filename with extension
    ext            string    File extension

  Temp file variables (when temp_files configured):
    <name>.path    string    Path to temp file
    <name>.content string    File content
    <name>.ext     string    File extension
    <name>.detected string   Detected type (text, json, xml, yaml)
    <name>.json    any       Parsed JSON (if content is JSON)

  Built-in CEL functions:
    string.contains(s)           Check substring
    string.startsWith(s)         Check prefix
    string.endsWith(s)           Check suffix
    string.matches(regex)        Regex match
    size(list)                   List/string length
    list.all(x, predicate)       All elements match
    list.exists(x, predicate)    Any element matches
    list.filter(x, predicate)    Filter elements

  Extended functions (via gomplate):
    strings.Contains, strings.TrimSpace, strings.Split, ...
    math.Abs, math.Max, math.Min, ...
    regexp.Match, regexp.FindAll, regexp.Replace, ...
    conv.ToInt, conv.ToString, conv.ToBool, ...
    coll.Has, coll.Keys, coll.Values, ...
    data.JSON, data.YAML, data.CSV, ...
    file.Exists, file.Read, file.IsDir, ...
    time.Now, time.Parse, ...

TEMPLATE VARIABLES

  The exec, build, and args fields support Go template syntax:

    exec: {{.executablePath}}
    args: [--file, "{{.file}}"]
    build: go build -o {{.workDir}}/myapp

  Available: .executablePath, .workDir, .name, .query, .file,
  .filename, .dir, .absfile, .absdir, .basename, .ext

FILE EXPANSION

  Set "files" in front-matter to replicate each test per matching file:

  ---
  files: "**/*.go"
  exec: golint
  args: ["{{.file}}"]
  ---

  Each test copy gets file template variables and is named
  "<original name> [<relative path>]".

EXECUTION

  Tests run in parallel with a 2-minute default timeout per test
  and 5-minute timeout for the build step. The build command runs
  once before any tests. Working directory resolves relative to the
  fixture file directory (sourceDir), falling back to --workdir.`

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
	})
	if err != nil {
		return fmt.Errorf("failed to create fixture runner: %w", err)
	}

	return runner.Run()
}

func init() {
	rootCmd.AddCommand(fixturesCmd)
}
