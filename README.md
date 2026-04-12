<p align="left">
  <img src="assets/brand/gavel-logo.svg" alt="gavel" height="48"/>
</p>

# Gavel

A CLI toolkit for git analysis, AI-powered code review, and markdown-based test fixtures.

## Install

```bash
# From source
go install github.com/flanksource/gavel/cmd/gavel@latest

# Or build locally
task build
task install  # installs to $GOPATH/bin
```

Pre-built binaries for Linux, macOS, and Windows are available on the [Releases](https://github.com/flanksource/gavel/releases) page.

## Commands

### `gavel git history`

Retrieve and filter commit history from a git repository.

```bash
gavel git history --path .
gavel git history --since 2024-01-01 --until 2024-06-01
gavel git history --author "alice" --message "feat*"
gavel git history abc1234                    # specific commit
gavel git history main..feature-branch       # commit range
```

| Flag | Description |
|------|-------------|
| `--path` | Path to git repository (default: `.`) |
| `--since` | Start date for filtering |
| `--until` | End date for filtering |
| `--author` | Filter by author name/email (repeatable) |
| `--message` | Filter by commit message |
| `--show-patch` | Include diffs |

### `gavel git analyze`

Analyze commits with scope/technology detection, Kubernetes resource change tracking, severity scoring, and optional AI-powered analysis.

```bash
gavel git analyze
gavel git analyze --ai --model claude-sonnet-4-20250514
gavel git analyze --scope backend --tech kubernetes
gavel git analyze --summary --summary-window week
gavel git analyze --include bots --exclude merges --verbose
gavel git analyze --input previous-run.json   # re-analyze from JSON
```

| Flag | Description |
|------|-------------|
| `--ai` | Enable AI-powered analysis |
| `--model` | AI model to use |
| `--scope` | Filter by scope types |
| `--commit-types` | Filter by conventional commit types |
| `--tech` | Filter by technologies |
| `--summary` | Generate a tree-based summary |
| `--summary-window` | Grouping window: `day`, `week`, `month`, `year` |
| `--include` | Include named filter sets from `.gitanalyze.yaml` |
| `--exclude` | Exclude named filter sets from `.gitanalyze.yaml` |
| `--verbose` | Show what was skipped and why |
| `--input` | Load from previous JSON output (repeatable) |
| `--short` | Show condensed file-change summary |

### `gavel git init-config`

Create a `.gitanalyze.yaml` with sensible defaults, then optionally spawn an AI CLI to analyze the repo and recommend additional rules.

```bash
gavel git init-config                          # create defaults + AI recommendations via claude
gavel git init-config --model gemini           # use gemini instead
gavel git init-config --model codex            # use codex instead
gavel git init-config --model none             # just create the defaults file, no AI
```

| Flag | Description |
|------|-------------|
| `--path` | Path to git repository (default: `.`) |
| `--model` | AI CLI to use: `claude`, `gemini`, `codex`, or `none` (default: `claude`) |
| `--debug` | Enable debug logging |

### `gavel git amend-commits`

Interactively improve commit messages using AI. Rewrites commits below a quality score threshold.

```bash
gavel git amend-commits
gavel git amend-commits --threshold 5.0 --base main
gavel git amend-commits --dry-run
```

### `gavel verify`

AI-powered code review with structured checks across completeness, code quality, testing, consistency, security, and performance.

```bash
gavel verify
gavel verify --range main..HEAD
gavel verify --model gemini --disable-checks SEC-1,PERF-2
gavel verify --auto-fix --max-turns 5
gavel verify --sync-todos
```

| Flag | Description |
|------|-------------|
| `--model` | AI CLI: `claude`, `gemini`, `codex`, or a model name |
| `--range` | Commit range to review |
| `--auto-fix` | Enable iterative AI fix loop |
| `--fix-model` | Separate model for fixes |
| `--max-turns` | Max verify-fix cycles (default: 3) |
| `--score-threshold` | Exit 0 if score >= this (default: 80) |
| `--disable-checks` | Check IDs to disable |
| `--sync-todos` | Create TODO files from findings |

### `gavel fixtures`

Run declarative tests defined in markdown files using tables, command blocks, or standalone code blocks.

```bash
gavel fixtures tests.md
gavel fixtures fixtures/**/*.md
gavel fixtures -v tests.md                  # verbose (stderr on pass, stdout+stderr on fail)
gavel fixtures -vv tests.md                 # more verbose
gavel fixtures --no-progress tests.md       # disable progress display
```

Fixtures support three formats in a single markdown file:

**Table format** (preferred) — each row is a test. Custom columns become template variables in `exec`/`args`:

```markdown
---
exec: bash
args: ["-c", "curl {{.flags}} {{.baseUrl}}{{.path}}"]
baseUrl: https://api.example.com
flags: "-s"
---

| Name       | path    | CEL                      |
|------------|---------|--------------------------|
| get users  | /users  | json.size() > 0          |
| get health | /health | json.status == "ok"      |
```

**Command blocks** — use when commands are multi-line or need per-test setup:

````markdown
### command: my test
```yaml
exitCode: 0
env:
  KEY: value
```

```bash
echo "hello"
```

* contains: hello
````

**Standalone code blocks** — auto-detected from headings:

````markdown
## Smoke Tests

```bash
echo "auto-detected"
```

* cel: stdout.contains("auto-detected")
````

#### Front-matter reference

File-level front-matter applies to all tests. Fields marked with **†** can also be set per-test via table columns or command block YAML.

```yaml
---
build: go build -o myapp           # run once before all tests
exec: ./myapp                      # † default executable (default: bash)
args: [--verbose]                  # † default arguments
env:                               # † environment variables
  LOG_LEVEL: debug
cwd: ./testdir                     # † working directory (relative to fixture file)
terminal: pty                      # † pseudo-terminal mode (merges stdout/stderr)
files: "**/*.go"                   # glob: replicate tests per matching file
codeBlocks: [bash, python]         # languages to execute (default: [bash])
timeout: 30s                       # † total timeout
os: linux                          # † skip on other OSes (prefix ! to negate: !darwin)
arch: amd64                        # † skip on other architectures
skip: "! command -v docker"        # † skip if command exits 0
---
```

#### Auto-injected variables

Available using `$VAR` syntax, Go template syntax (`{{.VAR}}`), and as environment variables:

| Variable | Description |
|----------|-------------|
| `GIT_ROOT_DIR` | Nearest parent directory containing `.git` |
| `GO_ROOT_DIR` | Nearest parent directory containing `go.mod` |
| `ROOT_DIR` | `GIT_ROOT_DIR` if available, else `GO_ROOT_DIR`, else working directory |
| `CWD` | Resolved working directory for the test |
| `GOOS` | Go runtime OS (e.g. `linux`, `darwin`) |
| `GOARCH` | Go runtime architecture (e.g. `amd64`, `arm64`) |
| `GOPATH` | Go workspace path |

#### CWD resolution priority

1. Test-level CWD (per-test frontmatter or table column)
2. File-level CWD (YAML front-matter)
3. SourceDir (directory containing the fixture file)
4. `--cwd` flag or current working directory

See `gavel fixtures --help` for the full reference including CEL variables, validation shorthand, supported languages, and template syntax.

### `gavel test`

Discover and run Go and Ginkgo tests with structured output.

```bash
gavel test --path ./...
gavel test --path ./pkg/... --name "TestFoo" -r
gavel test --format json
```

| Flag | Description |
|------|-------------|
| `--path` | Test path pattern |
| `--name` | Filter by test name |
| `-r` / `--recursive` | Recursive test discovery |
| `--format` | Output format (`json`, `text`) |

### `gavel todos`

Manage and execute TODO items with Claude Code integration.

```bash
gavel todos list
gavel todos list --status pending
gavel todos run .todos/fix-bug.md
gavel todos check .todos/fix-bug.md
```

### `gavel pr watch`

Monitor GitHub Actions status for pull requests.

```bash
gavel pr watch
gavel pr watch --follow --interval 30s
gavel pr watch --sync-todos
```

### `gavel repomap`

View the merged architecture configuration for a repository path.

```bash
gavel repomap view .
gavel repomap get src/main.go
```

## Configuration

### `arch.yaml`

Placed at the repository root (or any directory — gavel walks up to find it). Defines scope classification, technology detection, severity rules, and build/git settings.

```yaml
scopes:
  rules:
    backend:
      - path: "cmd/**"
      - path: "pkg/**"
    frontend:
      - path: "web/**"

tech:
  rules:
    kubernetes:
      - path: "deploy/*.yaml"

severity:
  default: medium
  rules:
    'kubernetes.kind == "Secret"': critical
    'change.type == "deleted"': critical
    'commit.line_changes > 500': critical

git:
  version_field_patterns:
    - "**.image"
    - "**.tag"
    - "**.version"
```

Embedded defaults detect common scope types (ci, dependency, docs, test) and technologies (Go, Node.js, Python, Kubernetes, Terraform, Docker, etc.).

### `.gitanalyze.yaml`

Controls which commits, files, authors, and resources are excluded from `gavel git analyze`. Placed at the repository root.

```yaml
# Named filter sets toggled with --include / --exclude
filter_sets:
  bots:
    ignore_authors:
      - "dependabot*"
      - "renovate*"
      - "github-actions*"

  noise:
    ignore_files:
      - "*.lock"
      - "go.sum"
      - "package-lock.json"

  generated:
    ignore_resources:
      - kind: ConfigMap
        name: "*-generated"
      - kind: Secret

# Which filter sets are active by default
includes:
  - bots
  - noise

# Top-level filters (always applied)
ignore_commits:
  - "fixup!*"
  - "squash!*"

ignore_files:
  - ".idea/*"
  - "*.svg"

ignore_commit_types:
  - "chore"
  - "ci"

# CEL expressions — skip when true
ignore_commit_rules:
  - cel: "commit.is_merge"
  - cel: "commit.line_changes > 10000"

# Kubernetes resource filters
ignore_resources:
  - kind: Secret
  - kind: ConfigMap
    name: "*-generated"
```

**CEL variables** available in `ignore_commit_rules`:

| Variable | Type | Description |
|----------|------|-------------|
| `commit.author` | string | Author name |
| `commit.author_email` | string | Author email |
| `commit.subject` | string | Commit subject line |
| `commit.body` | string | Commit body |
| `commit.type` | string | Conventional commit type |
| `commit.scope` | string | Conventional commit scope |
| `commit.is_merge` | bool | True if subject starts with "Merge " |
| `commit.files_changed` | int | Number of files changed |
| `commit.line_changes` | int | Total lines added + deleted |
| `commit.additions` | int | Lines added |
| `commit.deletions` | int | Lines deleted |
| `commit.files` | list | File paths changed |
| `commit.tags` | list | Commit tags |
| `commit.is_tagged` | bool | True if commit has tags |

**Embedded defaults** skip bot authors (`dependabot*`, `renovate*`, `github-actions*`), lock files (`*.lock`, `go.sum`, etc.), and merge commits.

Override at the CLI:

```bash
gavel git analyze --exclude bots      # include bot commits
gavel git analyze --include generated  # activate the "generated" filter set
gavel git analyze --verbose            # show skip reasons
```

## Agent Skills

Gavel ships [Agent Skills](https://agentskills.io/) that give AI coding agents the ability to create and run fixture-based tests. Skills are auto-discovered from `.agents/skills/` by any compatible agent (Claude Code, VS Code Copilot, Cursor, Gemini CLI, and [others](https://agentskills.io/)).

| Skill | Description |
|-------|-------------|
| `gavel-fixture-tester` | Create and run fixture-based tests using markdown files with command blocks, tables, and CEL assertions |

### Install

```bash
npx skills add flanksource/gavel
```

Use `-g` to install globally (all projects) or omit for project-only. Preview available skills first with `npx skills add flanksource/gavel -l`.

See [.agents/skills/README.md](.agents/skills/README.md) for alternative installation methods.

## Development

```bash
task build       # build binary
task test        # run all tests
task test:unit   # run unit tests only
task lint        # run linters
task fmt         # format code
task ci          # fmt + lint + test + build
```

## License

See [LICENSE](LICENSE) for details.
