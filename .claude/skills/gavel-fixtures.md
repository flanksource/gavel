---
name: gavel-fixtures
description: Create and run gavel fixture-based tests using markdown files with command blocks, tables, and CEL assertions
allowed-tools: [Read, Write, Edit, Bash, Glob, Grep, AskUserQuestion]
codeBlocks: [bash]
---

# Gavel Fixture Testing

Create fixture-based tests for CLI commands using gavel's markdown fixture format. Fixtures define test cases as markdown files with YAML front-matter, command blocks, tables, and CEL validation expressions.

This file is itself a valid fixture. Run it to verify all examples parse and execute:

    gavel fixtures .claude/skills/gavel-fixtures.md

## When to Use

Use gavel fixtures when:
- Testing CLI commands, shell scripts, or executables
- Tests are input/output pairs: run a command, check stdout/stderr/exitCode
- You want non-developers to read and add test cases
- Tests benefit from CEL expressions for flexible assertions

Use regular Go tests when:
- Testing Go functions directly (not CLI output)
- Tests need complex setup, mocking, or concurrency
- Assertions are simpler than what CEL provides

## Fixture File Format

### Front-Matter (YAML)

Every fixture file can start with `---` delimited YAML front-matter:

```yaml
---
build: go build -o myapp           # Run once before all tests
exec: ./myapp                      # Default executable (default: bash)
args: [--verbose]                  # Default arguments
env:                               # Environment variables
  LOG_LEVEL: debug
cwd: ./testdir                     # Working directory (relative to fixture file)
codeBlocks: [bash]                 # Languages to execute (default: [bash])
files: "**/*.go"                   # Glob: replicate tests per matching file
timeout: 30s                       # Total timeout
os: linux                          # OS constraint ("!darwin" = skip on macOS)
arch: amd64                        # Architecture constraint
skip: "test -z $CI"               # Bash command; exit 0 = skip fixture
---
```

### Format 1: Markdown Tables (Preferred)

Each row is a test. Column headers map to fixture fields. This is the preferred format — use it whenever tests share the same executable and differ only in arguments or expected output.

**Supported column headers** (case-insensitive):

Input: `name`, `cli`/`command`/`exec`, `args`, `cwd`, `query`

Expectations: `exit code`, `expected output`/`output`, `expected error`/`error`, `format`, `count`, `cel`/`validation`/`expr`

Unrecognized columns become `Properties`, accessible via `expectations.Properties["col"]`.

Use sections (`## Section Name`) to group related tables within a file.

#### Live examples — basic table tests

| Name | Command | Exit Code | CEL Validation |
|------|---------|-----------|----------------|
| Echo stdout | echo hello world | 0 | stdout.contains("hello") |
| Exit code zero | echo ok | 0 | exitCode == 0 |
| Non-zero exit | exit 1 | 1 | exitCode == 1 |
| Stderr output | echo "err msg" >&2 | 0 | stderr.contains("err msg") |
| Multiword contains | echo "the quick brown fox" | 0 | stdout.contains("quick") && stdout.contains("fox") |
| Negation check | echo "success" | 0 | !stdout.contains("error") |
| Regex match | echo "version 1.2.3" | 0 | stdout.matches("version [0-9]+\\.[0-9]+\\.[0-9]+") |

#### Live examples — JSON output in tables

| Name | Command | Exit Code | CEL Validation |
|------|---------|-----------|----------------|
| JSON object | echo '{"name":"test","count":3}' | 0 | json.name == "test" && json.count == 3.0 |
| JSON array | echo '[{"id":1},{"id":2}]' | 0 | size(json) == 2 && json[0].id == 1.0 |
| JSON nested | echo '{"data":{"items":["a","b"]}}' | 0 | size(json.data.items) == 2 |

### Format 2: Command Blocks

Use `### command: <test name>` headings for tests that need multi-line scripts, setup/teardown, or per-test YAML config.

YAML config block fields: `cwd`, `exitCode`, `env`, `timeout`

Validation bullet prefixes:
- `cel: <expr>` — Raw CEL expression
- `contains: <text>` — Converts to `stdout.contains("<text>")`
- `regex: <pattern>` — Converts to `stdout.matches("<pattern>")`
- `not: contains: <text>` — Converts to `!stdout.contains("<text>")`
- `not: <expr>` — Converts to `!(<expr>)`

Multiple validations are joined with `&&`.

#### Live examples — command blocks

### command: Multi-line script with setup and teardown

```yaml
exitCode: 0
```

```bash
TMPFILE=$(mktemp)
echo "hello from tempfile" > "$TMPFILE"
cat "$TMPFILE"
rm -f "$TMPFILE"
```

- contains: hello from tempfile

### command: Environment variables in script

```bash
export MY_VAR="fixture_value"
echo "var is $MY_VAR"
```

- cel: stdout.contains("fixture_value")
- not: contains: error

### command: YAML config with exitCode

```yaml
exitCode: 1
```

```bash
echo "expected failure" >&2
exit 1
```

- cel: stderr.contains("expected failure")

### command: Multiple validations joined

```bash
echo "line1: alpha"
echo "line2: beta"
echo "line3: gamma"
```

- contains: alpha
- contains: beta
- contains: gamma
- not: contains: delta

### command: ExitCode via YAML config

```yaml
exitCode: 1
```

```bash
exit 1
```

### command: Timeout via YAML config

```yaml
exitCode: 0
timeout: 10
```

```bash
echo "completed within timeout"
```

- contains: completed

### Format 3: Standalone Code Blocks

Executable code blocks outside `command:` headings are auto-detected. Test name comes from the nearest preceding heading.

#### Standalone echo test

```bash
echo "standalone block detected"
```

- contains: standalone block detected

### Inline Code Fence Attributes

Override YAML block values on the code fence info string: `exitCode=N`, `timeout=N` (seconds).

## CEL Variables

| Variable | Type | Description |
|----------|------|-------------|
| `stdout` / `output` | string | Process stdout |
| `stderr` | string | Process stderr |
| `exitCode` | int | Process exit code |
| `json` | any | Auto-parsed JSON (when stdout starts with `{` or `[`) |
| `name` | string | Test name |
| `sourceDir` | string | Directory containing fixture file |
| `query` | string | Query string |
| `expectations` | object | Expected values |
| `executablePath` | string | Path to gavel binary |
| `workDir` | string | Working directory |

File expansion variables (when `files:` is set): `file`, `filename`, `dir`, `absfile`, `absdir`, `basename`, `ext`

## CEL Functions

Built-in: `contains()`, `startsWith()`, `endsWith()`, `matches(regex)`, `size()`, `all()`, `exists()`, `filter()`

Extended (gomplate): `strings.*`, `math.*`, `regexp.*`, `conv.*`, `coll.*`, `data.JSON`, `data.YAML`, `file.*`, `time.*`

## Template Variables

`exec`, `build`, and `args` support Go template syntax:

```yaml
exec: "{{.executablePath}}"
args: [--file, "{{.file}}"]
```

Available: `.executablePath`, `.workDir`, `.name`, `.query`, `.file`, `.filename`, `.dir`, etc.

## Supported Languages

| Language | Executor |
|----------|----------|
| bash, sh, shell | bash -c |
| python, py | python -c |
| typescript, ts | ts-node -e |
| javascript, js | node -e |
| pwsh, powershell | pwsh -Command |
| go | go (direct) |

Non-executable (config): `yaml`, `frontmatter`

## Running Fixtures

    gavel fixtures <fixture-files...>
    gavel fixtures tests.md
    gavel fixtures fixtures/**/*.md
    gavel fixtures -v tests.md                  # Verbose
    gavel fixtures --json tests.md              # JSON output
    gavel fixtures --json tests.md 2>/dev/null  # JSON only, no logs

Tests run in parallel (2-minute timeout per test, 5-minute for build). Working directory resolves relative to the fixture file.

## Process

### Step 1: Understand what to test

Read the command or executable being tested. Identify:
- The CLI command and its flags/arguments
- Expected outputs (stdout, stderr, exit codes)
- Edge cases and error conditions

### Step 2: Choose fixture format

- **Tables** (preferred) for most tests — especially when tests share a command pattern and differ in arguments/expectations
- **Command blocks** when a test needs multi-line scripts, setup/teardown, or per-test YAML config (env, cwd overrides)
- **Standalone code blocks** for simple one-liners under section headings
- You can mix tables and command blocks in the same file — use tables as the default and command blocks only when needed

### Step 3: Write the fixture file

Place fixture files alongside the code they test, typically in `fixtures/testdata/` or a `testdata/` directory near the relevant package.

File naming: `<feature-name>.md` using kebab-case.

### Step 4: Write assertions

Prefer CEL expressions for flexible validation:
- Use `stdout.contains("text")` for substring checks
- Use `json.field == value` when testing JSON output
- Use `exitCode == N` for exit code checks
- Use `!stdout.contains("text")` for negative assertions
- Combine with `&&` for multiple conditions

### Step 5: Run and verify

    gavel fixtures <your-fixture-file>.md

Fix failures and iterate until all tests pass.

## Rules

- PREFER markdown tables over command blocks — use tables as the default format for all tests that can be expressed as single-line commands with exit code and CEL validation
- Only use command blocks (`### command: <name>`) when tests need multi-line scripts, setup/teardown, or per-test YAML config (env, cwd overrides)
- Use sections (`## Section Name`) to group related tables within a file
- Place YAML config blocks BEFORE the executable code block within a command section
- Use `codeBlocks: [bash]` in front-matter when mixing executable and non-executable code blocks
- Default exit code expectation is 0 — only specify `exitCode` when expecting non-zero
- CEL expressions referencing JSON output require stdout to start with `{` or `[` for auto-parsing
- When testing gavel itself, use `2>/dev/null` to suppress log output from stdout: `gavel cmd --json 2>/dev/null`
- Prefer `contains:` shorthand over `cel: stdout.contains(...)` for simple substring checks (command blocks only — tables use `CEL Validation` column)
- Keep fixture files focused: one feature or command per file
- When a command block test needs setup/teardown (temp files, config files), include cleanup in the same bash block
- For tests that should only run on specific platforms, use `os:` and `arch:` front-matter constraints
