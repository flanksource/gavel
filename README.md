<p align="left">
  <img src="assets/brand/gavel-logo.svg" alt="gavel" height="48"/>
</p>

# Gavel

A CLI toolkit for testing, linting, AI-powered code review, and CI automation.

## Install

```bash
# From source
go install github.com/flanksource/gavel/cmd/gavel@latest

# Or build locally
task build       # also: make build
task install     # installs to $GOPATH/bin
```

Pre-built binaries for Linux, macOS, and Windows are available on the [Releases](https://github.com/flanksource/gavel/releases) page.

## GitHub Action

Run `gavel test --lint` in CI, upload JSON + HTML artifacts, and post a sticky PR comment with the markdown summary.

### Minimal usage

```yaml
- uses: flanksource/gavel@main
  with:
    args: test --lint
```

### Full example

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write # required when comment=true so the action can post/update the sticky PR comment
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: flanksource/gavel@main
        with:
          args: test --lint
          version: latest
          json-file: gavel-results.json
          html-file: gavel-results.html
          summary-file: gavel-summary.md
          artifact-name: gavel-results
          comment: "true"
          comment-header: gavel
          fail-on-error: "true"
```

If you disable PR commenting (`comment: "false"`), `contents: read` is sufficient. Keep `pull-requests: write` when the action should post or update the sticky PR comment.

### Inputs

| Input | Default | Description |
|-------|---------|-------------|
| `args` | `test --lint` | Arguments passed to gavel |
| `working-directory` | `.` | Working directory for execution |
| `version` | `latest` | Release tag, `latest`, or `source` to use a pre-installed binary |
| `json-file` | `gavel-results.json` | Path for the JSON artifact |
| `html-file` | `gavel-results.html` | Path for the HTML artifact |
| `summary-file` | `gavel-summary.md` | Path for the compact markdown summary |
| `artifact-name` | `gavel-results` | Name of the uploaded artifact bundle |
| `comment` | `true` | Post/update a sticky PR comment with the summary |
| `comment-header` | `gavel` | Unique marker for the sticky comment (allows multiple comments per PR) |
| `github-token` | `${{ github.token }}` | Token used to post PR comments |
| `fail-on-error` | `true` | Fail the step when gavel exits non-zero (artifacts are still uploaded) |

### Outputs

| Output | Description |
|--------|-------------|
| `exit-code` | Exit code from gavel |
| `json-path` | Absolute path to the JSON artifact |
| `html-path` | Absolute path to the HTML artifact |
| `summary-path` | Absolute path to the markdown summary |
| `log-path` | Absolute path to the stderr log (`gavel.log`) — always present, even on crash |

The action is crash-resilient: if gavel exits before writing results, placeholder stubs with the exit code and log tail are written so the artifact upload and PR comment still succeed.

## Commands

### Testing & Linting

#### `gavel test`

Discover and run Go and Ginkgo tests with structured output, optional linting, benchmarks, fixture files, and a live browser UI.

```bash
gavel test ./pkg/...
gavel test --lint                                     # run linters in parallel with tests
gavel test --ui                                       # launch browser dashboard
gavel test --changed                                  # only packages affected by local changes
gavel test --cache                                    # skip packages unchanged since last pass
gavel test --bench .                                  # run benchmarks alongside tests
gavel test --fixtures                                 # discover and run *.fixture.md files
gavel test --format "json=out.json,html=report.html"  # write multiple output formats
gavel test --dry-run                                  # show what would run without executing
```

| Flag | Description |
|------|-------------|
| `[paths...]` | Package paths to test (e.g. `./pkg/...`). If empty, all packages are discovered. |
| `--lint` | Run linters in parallel with tests |
| `--ui` | Launch browser with real-time test progress dashboard |
| `--addr` | Interface to bind the UI server (default: `localhost`, use `0.0.0.0` for LAN) |
| `--cache` | Skip packages whose content fingerprint matches the last passing run |
| `--changed` | Only run packages affected by staged/unstaged/untracked changes vs `origin/main` |
| `--since` | Only run packages affected by the diff since `<ref>` (merge-base) |
| `--bench` | Run Go benchmarks matching this regex (`.` runs all) |
| `--fixtures` | Discover and run fixture files (also enabled via `.gavel.yaml`) |
| `--fixture-files` | Globs for fixture discovery (default: `**/*.fixture.md`) |
| `-p` / `--nodes` | Number of parallel Ginkgo nodes (0 = default, -1 = auto) |
| `-r` / `--recursive` | Recursive test discovery (default: true) |
| `--show-passed` | Show passed tests in output |
| `--show-stdout` | When to show stdout: `Never`, `OnFailure` (default), `Always` |
| `--show-stderr` | When to show stderr: `Never`, `OnFailure` (default), `Always` |
| `--skip-hooks` | Skip `.gavel.yaml` pre/post hooks (default: skip locally, run in CI) |
| `--sync-todos` | Sync test failures to TODO files |
| `--dry-run` | Print test commands without executing |
| `--auto-stop` | With `--ui`, fork a detached UI server that exits after this duration |
| `--idle-timeout` | With `--ui --auto-stop`, exit the detached UI after no HTTP requests |
| `--extra-args` | Additional arguments passed through to test runners |
| `--work-dir` | Working directory to run tests in |

#### `gavel lint`

Run linters on the project. Auto-detects which linters are installed and applicable.

Supported linters: golangci-lint, ruff, eslint, pyright, markdownlint, vale, jscpd, betterleaks.

```bash
gavel lint
gavel lint --fix                            # auto-fix violations
gavel lint --triage                         # interactive mode to select violations to ignore
gavel lint --changed                        # only new issues vs origin/main
gavel lint --ui                             # view violations in browser
gavel lint --dry-run                        # show linter commands without executing
gavel lint secrets                          # run betterleaks only (alias)
gavel lint jscpd eslint                     # run specific linters by name
gavel lint --sync-todos .todos              # sync violations to TODO files
```

| Flag | Description |
|------|-------------|
| `[linters/files...]` | Positional args: linter names run only those linters; file paths lint only those files |
| `--linters` | Comma-separated linter names or `*` for all (default: `*`) |
| `--fix` | Enable auto-fixing |
| `--triage` | Interactive mode to select violation types to ignore (saves to `.gavel.yaml`) |
| `--changed` | Only report new issues vs `origin/main` (or `$GAVEL_CHANGED_BASE`) |
| `--since` | Only report new issues since `<ref>` (merge-base with HEAD) |
| `--ignore` | Glob patterns to exclude from linting |
| `--ui` | Launch browser UI to view violations |
| `--addr` | Interface to bind the UI server (default: `localhost`) |
| `--sync-todos` | Sync violations to TODO files in directory |
| `--group-by` | Group synced TODOs by: `file`, `package`, `message` (default: `file`) |
| `--no-cache` | Disable caching/debounce |
| `--timeout` | Timeout per linter (default: `5m`) |
| `--dry-run` | Print linter commands without executing |

Config-file-gated linters (e.g. betterleaks) only run when their native config (`.betterleaks.toml` / `.gitleaks.toml`) is present, unless explicitly named. Disable betterleaks entirely via `secrets.disabled: true` in `.gavel.yaml`.

#### `gavel fixtures`

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

<details>
<summary>Front-matter reference</summary>

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

**Auto-injected variables** (available as `$VAR`, `{{.VAR}}`, and env vars):

| Variable | Description |
|----------|-------------|
| `GIT_ROOT_DIR` | Nearest parent directory containing `.git` |
| `GO_ROOT_DIR` | Nearest parent directory containing `go.mod` |
| `ROOT_DIR` | `GIT_ROOT_DIR` if available, else `GO_ROOT_DIR`, else working directory |
| `CWD` | Resolved working directory for the test |
| `GOOS` | Go runtime OS (e.g. `linux`, `darwin`) |
| `GOARCH` | Go runtime architecture (e.g. `amd64`, `arm64`) |
| `GOPATH` | Go workspace path |

**CWD resolution priority:** Test-level CWD → File-level CWD → SourceDir (directory of fixture file) → `--cwd` flag or current working directory.

See `gavel fixtures --help` for the full reference including CEL variables, validation shorthand, supported languages, and template syntax.

</details>

#### `gavel bench`

Run Go benchmarks and compare base vs head results for regression detection.

```bash
# Run benchmarks and write JSON
gavel bench run ./pkg/... --out base.json
gavel bench run ./pkg/... --out head.json --count 10

# Compare results
gavel bench compare --base base.json --head head.json
gavel bench compare --base base.json --head head.json --threshold 15 --ui
```

**`bench run` flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--count` | `6` | Number of times to run each benchmark |
| `--timeout` | `20m` | Test execution timeout |
| `--benchtime` | `1s` | Duration per benchmark |
| `--out` | stdout | Write JSON results to file |
| `--pattern` | `.` | Benchmark name regex |
| `--benchmem` | `true` | Include memory allocation stats |
| `--extra` | | Extra flags passed through to `go test` |

**`bench compare` flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--base` | | Path to base bench JSON (required) |
| `--head` | | Path to head bench JSON (required) |
| `--base-label` | `base` | Display label for the base run |
| `--head-label` | `head` | Display label for the head run |
| `--threshold` | `10` | Regression threshold in percent |
| `--ui` | `false` | Launch browser UI with the comparison |
| `--addr` | `localhost` | Interface to bind the UI server |

### Code Review & Commits

#### `gavel verify`

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
| `--model` | AI CLI: `claude`, `gemini`, `codex` (or a model name) |
| `--range` | Commit range to review |
| `--auto-fix` | Enable iterative AI fix loop |
| `--fix-model` | Separate model for fixes |
| `--max-turns` | Max verify-fix cycles (default: 3) |
| `--score-threshold` | Exit 0 if score >= this (default: 80) |
| `--disable-checks` | Check IDs to disable |
| `--sync-todos` | Create TODO files from findings |
| `--patch-only` | AI outputs patches instead of interactive tool-use |

#### `gavel commit`

Generate a conventional commit message via LLM and run pre-commit hooks from `.gavel.yaml`.

```bash
gavel commit                          # LLM-generated message, staged changes
gavel commit -A                       # split staged changes into multiple commits
gavel commit -m "chore: bump dep"     # explicit message, skip LLM
gavel commit --stage all --dry-run    # stage everything, print message
gavel commit --force                  # skip hooks
```

| Flag | Description |
|------|-------------|
| `--stage` | Which changes to commit: `staged` (default), `unstaged`, `all` |
| `-A` / `--commit-all` | Ask the LLM to group the selected change set into multiple commits; if nothing is staged, stage all first |
| `-m` / `--message` | Explicit commit message (skips LLM) |
| `--model` | Override LLM model from `.gavel.yaml` `commit.model` |
| `--dry-run` | Print the generated message without committing |
| `--force` | Skip pre-commit hooks |
| `--no-cache` | Bypass the LLM response cache |

Pre-commit hooks are configured in `.gavel.yaml` under `commit.hooks` — see [Configuration](#gavelyaml).

### Pull Requests

#### `gavel pr status`

Show GitHub Actions status for a pull request.

```bash
gavel pr status
gavel pr status 42
gavel pr status https://github.com/owner/repo/pull/123
gavel pr status --follow --interval 30s
gavel pr status --logs --tail-logs 50
gavel pr status --sync-todos
```

| Flag | Description |
|------|-------------|
| `-R` / `--repo` | GitHub repository (`owner/repo`) |
| `--follow` | Keep watching until all checks complete |
| `--interval` | Poll interval (default: `30s`) |
| `--logs` | Fetch and include failed job logs (uses extra API quota) |
| `--tail-logs` | Number of failed log lines to show per step (default: `100`) |
| `--sync-todos` | Sync TODO files for failed jobs to directory |

#### `gavel pr list`

List pull requests with filtering, a live browser dashboard, and macOS menu bar indicator.

```bash
gavel pr list
gavel pr list --ui                            # live dashboard in browser
gavel pr list --menu-bar                      # macOS menu bar status indicator
gavel pr list --state merged --since 30d
gavel pr list --any --status                  # all authors, with check status
gavel pr list owner/repo another/repo         # specific repos
gavel pr list --all --org myorg               # all repos in org
```

| Flag | Description |
|------|-------------|
| `--author` | GitHub username (default: `@me`) |
| `--since` | Show PRs updated since (e.g. `7d`, `now-30d`, `2024-01-01`) |
| `--state` | PR state: `open` (default), `closed`, `merged`, `all` |
| `--all` | List PRs across all repos in the org |
| `--any` | Show PRs from all authors |
| `--bots` | Include bot-authored PRs |
| `--org` | GitHub org for `--all` (auto-detected from git remote) |
| `--limit` | Maximum PRs to return (default: `50`) |
| `--status` | Show GitHub Actions check status counts |
| `--logs` | Fetch failed job log tails (requires `--status -v`) |
| `--url` | Show PR URL instead of number |
| `--ui` | Open PR dashboard in browser with live updates |
| `--menu-bar` | Show macOS menu bar status indicator |
| `--interval` | Poll interval for `--ui`/`--menu-bar` (default: `60s`) |

#### `gavel pr fix`

Sync TODOs from PR failures and interactively select which to fix with Claude Code.

```bash
gavel pr fix
gavel pr fix 42
gavel pr fix --group-by directory
gavel pr fix --dry-run
```

| Flag | Description |
|------|-------------|
| `-R` / `--repo` | GitHub repository (`owner/repo`) |
| `--dir` | TODOs directory (default: `.todos`) |
| `--group-by` | Group TODOs by: `file`, `directory`, `all`, `none` |
| `--max-retries` | Maximum retry attempts (default: `3`) |
| `--max-budget` | Maximum budget in USD |
| `--max-turns` | Maximum conversation turns |
| `--dirty` | Skip git stash/checkout |
| `--dry-run` | Print commands without executing |

### Git Analysis

#### `gavel git history`

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

#### `gavel git analyze`

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

#### `gavel git init-config`

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

#### `gavel git amend-commits`

Interactively improve commit messages using AI. Rewrites commits below a quality score threshold.

```bash
gavel git amend-commits
gavel git amend-commits --threshold 5.0 --base main
gavel git amend-commits --dry-run
```

### Infrastructure

#### `gavel ssh serve`

Start an SSH server that accepts `git push` and runs `gavel test --lint`. Developers add it as a git remote for local CI-like flows with real-time streaming results.

```bash
# Start the server
gavel ssh serve --port 2222

# From any project, add as a remote and push
git remote add gavel ssh://localhost:2222/myproject
git push gavel HEAD:main
```

Results stream back in real-time. Push is rejected on test failure. Repos are cached for fast incremental pushes.

| Flag | Description |
|------|-------------|
| `--port` | SSH server port (default: `2222`) |
| `--host` | Listen address (default: `0.0.0.0`) |
| `--host-key` | Path to SSH host key (default: `~/.gavel/ssh_host_key`) |
| `--repo-dir` | Directory for cached bare repos (default: `~/.gavel/repos`) |

The command run on push defaults to `gavel test --lint` but can be overridden via `ssh.cmd` in `.gavel.yaml`.

#### `gavel ssh install`

Install and enable a systemd unit for the SSH server (Linux only).

```bash
gavel ssh install                    # install and enable service
gavel ssh install --dry-run          # preview without making changes
gavel ssh install --port 3333 --user gavel
```

| Flag | Description |
|------|-------------|
| `--port` | SSH server port (default: `2222`) |
| `--host` | Listen address (default: `0.0.0.0`) |
| `--user` | System user to run the service as (default: `gavel`) |
| `--unit-path` | Path to write the systemd unit (default: `/etc/systemd/system/gavel-ssh.service`) |
| `--data-dir` | Directory for host key and cached repos (default: `/var/lib/gavel`) |
| `--binary` | Path to the gavel binary (default: current executable) |
| `--dry-run` | Print actions without writing |
| `--force` | Overwrite an existing unit file |

#### `gavel summary`

Build a compact markdown PR-comment summary from a gavel test/lint JSON result file.

```bash
gavel summary --input gavel-results.json
gavel summary --input results.json --output summary.md
```

| Flag | Description |
|------|-------------|
| `--input` | Path to gavel JSON result file |
| `--output` | Path to write compact markdown (default: stdout) |

#### `gavel ui serve`

Run a standalone UI server for replaying a previously captured test run.

```bash
gavel ui serve run.json
gavel ui serve run.json other-run.json --auto-stop=10m --idle-timeout=5m
```

| Argument / Flag | Description |
|------|-------------|
| `run.json [other-run.json ...]` | One or more JSON snapshots to load and merge in order |
| `--port` | Bind this port (0 = pick ephemeral) |
| `--addr` | Interface to bind (default: `localhost`) |
| `--auto-stop` | Hard wall-clock deadline from process start (default: `30m`) |
| `--idle-timeout` | Exit after this long with no HTTP requests (default: `5m`) |
| `--url-file` | Write the bound URL to this path for scripting |

#### `gavel repomap`

View the merged architecture configuration for a repository path.

```bash
gavel repomap view .
gavel repomap get src/main.go
```

### Task Management

#### `gavel todos`

Manage and execute TODO items with Claude Code integration.

```bash
gavel todos list
gavel todos list --status pending
gavel todos run .todos/fix-bug.md
gavel todos check .todos/fix-bug.md
```

## Output Formats

Many commands (`test`, `lint`, `bench compare`, `pr status`, `pr list`) support the `--format` flag for structured output:

```bash
# Single format
gavel test --format json

# Multiple formats written simultaneously
gavel test --format "json=results.json,html=report.html"

# Used by the GitHub Action internally
gavel test --lint --format "json=gavel-results.json,html=gavel-results.html"
```

Supported formats: `json`, `html`, `text` (default).

## Configuration

### `.gavel.yaml`

Project-level configuration placed at the repository root. Gavel also loads `~/.gavel.yaml` for user-level defaults and merges them (repo settings win).

```yaml
verify:
  model: claude                      # AI model for gavel verify
  checks:
    disabled: [SEC-1, PERF-2]        # disable specific check IDs
    disabledCategories: [performance] # disable entire categories

lint:
  ignore:
    - rule: SA1000                    # ignore a specific rule
      source: golangci-lint
    - source: eslint                  # ignore all eslint violations in vendor
      file: "vendor/**"
  linters:
    jscpd:
      enabled: true                   # opt in to jscpd duplicate detection

commit:
  model: claude                      # LLM model for gavel commit
  hooks:                             # pre-commit hooks
    - name: lint-staged
      run: "golangci-lint run --new-from-rev=HEAD~1"
      files: ["*.go"]

fixtures:
  enabled: true                      # auto-discover fixture files during gavel test
  files: ["tests/**/*.fixture.md"]   # override default glob (**/*.fixture.md)

ssh:
  cmd: "gavel test --lint --fixtures" # override the command run on git push

pre:                                 # hooks run before tests
  - name: generate
    run: "go generate ./..."

post:                                # hooks run after tests (non-blocking)
  - name: cleanup
    run: "rm -rf tmp/"

secrets:
  disabled: false                    # set true to disable betterleaks entirely
  configs: ["custom-leaks.toml"]     # additional betterleaks/gitleaks config files
```

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

<details>
<summary>CEL variables for ignore_commit_rules</summary>

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

</details>

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
task build       # build binary (also: make build)
task test        # run all tests
task test:unit   # run unit tests only
task lint        # run linters
task fmt         # format code
task ci          # fmt + lint + test + build
```

## License

See [LICENSE](LICENSE) for details.
