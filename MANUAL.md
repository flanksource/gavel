# Gavel Manual

`gavel` is a multi-purpose CLI for test orchestration, linting, AI-assisted review, pull-request operations, git history analysis, TODO automation, and a few long-running service flows.

This manual is organized around the actual CLI tree and answers two questions for each feature area:

- What problem is this part of the CLI meant to solve?
- Which command should you reach for first?

## How To Read This Manual

The CLI has two layers:

- Root-level feature commands such as `test`, `lint`, `verify`, and `commit`.
- Command groups such as `pr`, `git`, `todos`, `ssh`, `system`, and `repomap`.

Many commands also inherit shared output and execution behavior:

| Concept | What it means |
| --- | --- |
| `--cwd` | Resolve the command from another working directory without `cd`-ing first. |
| `--format` | Render structured output as `pretty`, `json`, `yaml`, `csv`, `html`, `markdown`, `pdf`, or `slack`, or write multiple sinks like `json=out.json,html=report.html`. |
| `--json`, `--yaml`, `--html`, `--markdown`, `--pdf`, `--csv` | Convenience output flags for common formats. |
| `--filter` | Apply a CEL filter to structured command output. |
| `-v`, `--log-level`, `--json-logs` | Raise verbosity or change log rendering. |
| `--no-progress`, `--no-color` | Make output quieter or more automation-friendly. |

Some commands are pure data commands, while others open browser UIs or run background processes. The manual calls that out where it matters.

## CLI Tree

```text
gavel
|-- bench
|   |-- compare
|   `-- run
|-- commit
|-- completion
|   |-- bash
|   |-- fish
|   |-- powershell
|   `-- zsh
|-- fixtures
|-- git
|   |-- amend-commits
|   |-- analyze
|   |-- history
|   `-- init-config
|-- help
|-- lint
|   |-- betterleaks
|   |-- eslint
|   |-- golangci
|   |-- golangci-lint
|   |-- jscpd
|   |-- markdownlint
|   |-- pyright
|   |-- ruff
|   |-- secrets
|   |-- tsc
|   |-- typescript
|   `-- vale
|-- pr
|   |-- fix
|   |-- list
|   `-- status
|-- repomap
|   |-- get
|   `-- view
|-- ssh
|   |-- install
|   `-- serve
|-- summary
|-- system
|   |-- install
|   |-- start
|   |-- status
|   |-- stop
|   `-- uninstall
|-- test
|   |-- ginkgo
|   |-- go
|   |-- jest
|   |-- playwright
|   `-- vitest
|-- todos
|   |-- check
|   |-- get
|   |-- list
|   `-- run
|-- ui
|   `-- serve
|-- verify
`-- version
```

## Shared Operating Model

### Config Files

Gavel uses a few repository-level files repeatedly:

| File | Used by |
| --- | --- |
| `.gavel.yaml` | `test`, `lint`, `commit`, `verify`, `ssh`, pre/post hooks, fixture discovery defaults |
| `arch.yaml` | `git analyze`, `repomap`, scope and technology detection |
| `.gitanalyze.yaml` | `git analyze` include/exclude filters |

### UI Modes

Gavel has three distinct UI patterns:

- Live run UIs: `gavel test --ui`, `gavel lint --ui`, `gavel bench compare --ui`, `gavel pr list --ui`
- Replay UIs: `gavel ui serve run.json`
- Managed background UI service: `gavel system start`, which runs the PR dashboard detached

### UI Guide

The UI is one of Gavel's strongest workflows. If you prefer drilling into failures interactively instead of reading long terminal output, start here.

| Command | UI type | Best for |
| --- | --- | --- |
| `gavel test --ui` | Live execution dashboard | Watching test progress, rerunning failed packages/tests, and keeping hooks/lint results in one view |
| `gavel test --ui --detach` | Live UI handed off to a child server | Starting a run from the CLI, then keeping the browser available after the parent process exits |
| `gavel lint --ui` | Live violations explorer | Browsing linter findings by file or rule, rerunning subsets, and writing ignore rules back to `.gavel.yaml` |
| `gavel bench compare --ui` | Comparison viewer | Inspecting benchmark regressions visually instead of reading a table dump |
| `gavel pr list --ui` | Polling PR dashboard | Watching many PRs, statuses, and refresh cycles in one place |
| `gavel ui serve run.json` | Snapshot replay UI | Re-opening saved JSON artifacts without rerunning work |
| `gavel system start` | Managed background PR UI | Keeping the PR dashboard alive across sessions |

UI-related tips:

- Most live UI commands accept `--addr`; use `0.0.0.0` when you want the browser to connect from another machine on the LAN.
- `gavel test --ui --detach` and `gavel ui serve` support `--auto-stop` and `--idle-timeout` so the server cleans itself up.
- `gavel ui serve` is the easiest way to inspect CI-produced JSON locally.
- The test and lint UIs share the same underlying server, so lint results can appear alongside test results when you run `gavel test --lint --ui`.

### TODO Flows

Several features can turn failures into `.todos` entries:

- `gavel test --sync-todos`
- `gavel lint --sync-todos`
- `gavel verify --sync-todos`
- `gavel pr status --sync-todos`
- `gavel pr fix`

If you use TODOs heavily, `gavel todos` becomes the hub for listing, checking, and executing those files.

### `.gavel.yaml`

`.gavel.yaml` is Gavel's main project and user configuration file.

Load order:

- `~/.gavel.yaml`
- `<git-root>/.gavel.yaml`
- `<current-working-directory>/.gavel.yaml` when the working directory differs from the git root

That means you can keep personal defaults in your home directory, then refine or override them per repository.

Supported structure:

```yaml
verify:
  model: claude
  prompt: ""                       # optional custom prompt prefix/override
  checks:
    disabled: []                   # specific check IDs
    disabledCategories: []         # categories such as performance

lint:
  ignore:
    - source: golangci-lint        # optional
      rule: errcheck               # optional
      file: "pkg/**/*.go"          # optional doublestar glob
  linters:
    jscpd:
      enabled: true
    eslint:
      enabled: false

commit:
  model: claude
  hooks:
    - name: lint-staged
      run: golangci-lint run ./...
      files:
        - "**/*.go"
  precommit:
    mode: prompt
  compatibility:
    mode: prompt

fixtures:
  enabled: true
  files:
    - "tests/**/*.fixture.md"

ssh:
  cmd: "gavel test --lint"

pre:
  - name: deps
    run: make deps

post:
  - name: notify
    run: echo "done"

secrets:
  disabled: false
  configs:
    - ".betterleaks.toml"
    - "security/custom-gitleaks.toml"
```

Field reference:

| Key | Purpose |
| --- | --- |
| `verify.model` | Default AI CLI or model name for `gavel verify` |
| `verify.prompt` | Optional custom verify prompt text |
| `verify.checks.disabled` | Disable specific verify checks by ID |
| `verify.checks.disabledCategories` | Disable whole verify categories |
| `lint.ignore` | Repo-wide or user-wide ignore rules matched by `source`, `rule`, and/or `file` |
| `lint.linters.<name>.enabled` | Force a linter on or off when Gavel would otherwise rely on detection/default behavior |
| `commit.model` | Default model for `gavel commit` |
| `commit.hooks` | Pre-commit shell hooks run by `gavel commit` |
| `commit.hooks[].files` | Optional staged-file glob filter; hook runs only if any staged file matches |
| `commit.precommit.mode` | How `gavel commit` handles gitignore + linked-dependency precommit checks |
| `commit.compatibility.mode` | How `gavel commit` handles AI-detected removed functionality / compatibility issues |
| `fixtures.enabled` | Turn fixture discovery on for `gavel test` |
| `fixtures.files` | Replace the default fixture discovery glob list |
| `ssh.cmd` | Override the command run by `gavel ssh serve` after push |
| `pre` | Shell steps run before `gavel test` when hooks are enabled |
| `post` | Shell steps run after `gavel test`; failures are logged but do not replace the main test result |
| `secrets.disabled` | Disable the `betterleaks` / `secrets` linter entirely |
| `secrets.configs` | Extra betterleaks/gitleaks TOML files to merge into secrets scanning; relative paths resolve from the declaring `.gavel.yaml` |

Merge rules matter because `.gavel.yaml` can come from multiple layers:

| Section | Merge behavior |
| --- | --- |
| `verify.model`, `verify.prompt` | Last non-empty value wins |
| `verify.checks.disabled`, `verify.checks.disabledCategories` | Appended across layers |
| `lint.ignore` | Appended across layers |
| `lint.linters.<name>.enabled` | Later layer wins for that linter |
| `commit.model` | Last non-empty value wins |
| `commit.precommit.mode` | Last non-empty value wins |
| `commit.compatibility.mode` | Last non-empty value wins |
| `commit.hooks` | Appended across layers |
| `fixtures.enabled` | Any layer can enable it |
| `fixtures.files` | Later non-empty list replaces earlier list |
| `ssh.cmd` | Last non-empty value wins |
| `pre`, `post` | Appended in load order: home, then repo, then cwd |
| `secrets.disabled` | Sticky OR: once disabled by any layer, it stays disabled |
| `secrets.configs` | Appended and deduplicated |

Notes:

- `gavel test` skips `pre` and `post` hooks by default in local runs, but enables them automatically in CI and SSH-push contexts unless you pass `--skip-hooks`.
- `commit.hooks` run with `sh -c` from the repo root and stream output to stderr.
- `pre` and `post` steps also run with `sh -c`, in order, stopping on the first failure within that phase.
- `fixtures.files` defaults to `**/*.fixture.md` when unset.

## Testing And Quality

This is the biggest part of the CLI. The common pattern is: discover work automatically, run it with structured output, then optionally render it in a browser, export JSON/HTML artifacts, or sync failures into TODOs.

### `gavel test`

Use `test` when you want Gavel to orchestrate repository test execution instead of calling framework CLIs yourself. It auto-detects supported frameworks, groups work by repo/module boundaries, and can combine tests, fixtures, benchmarks, UI output, caching, and change-aware selection.

Reach for it when you want one of these flows:

- Run everything Gavel can detect.
- Run only packages affected by local changes.
- Open a live dashboard while tests run.
- Re-run from a previous JSON baseline or only rerun failed targets.
- Sync broken tests into `.todos`.

Common workflows:

```bash
gavel test
gavel test ./pkg/...
gavel test --lint --format "json=gavel-results.json,html=gavel-results.html"
gavel test --changed
gavel test --ui
gavel test --ui --detach --auto-stop=30m --idle-timeout=5m
gavel test --bench .
gavel test --fixtures
gavel test --sync-todos
```

The UI path is worth calling out separately:

- `gavel test --ui` starts a live browser dashboard before pre-hooks run, so hook status, test progress, reruns, and optional lint results all stay in one place.
- `gavel test --lint --ui` is the richest local feedback loop because the same UI can show both test and lint findings.
- `gavel test --ui --detach` turns the live run into a handoff flow: the CLI starts the run, then a child UI server stays up until `--auto-stop` or `--idle-timeout` expires.

Framework pinning is available in two forms:

- `gavel test --framework go --framework ginkgo`
- `gavel test go`, `gavel test ginkgo`, `gavel test jest`, `gavel test vitest`, `gavel test playwright`

Use the framework subcommands when you want the full `test` flag surface but do not want auto-detection to choose runners for you.

### `gavel lint`

Use `lint` when you want a single entrypoint for repo linting across languages and tools. Gavel auto-detects installed linters and can narrow to changed files, show a browser UI, auto-fix, or sync findings into TODOs.

It is the right command for:

- Local lint passes before commit or review.
- Changed-only or baseline-against-main linting.
- Interactive triage of noisy rules.
- Running one linter through a consistent interface.

Common workflows:

```bash
gavel lint
gavel lint --fix
gavel lint --triage
gavel lint --changed
gavel lint --ui
gavel lint --sync-todos
gavel lint eslint
gavel lint secrets
```

The UI flow is especially useful here:

- `gavel lint --ui` opens the same browser shell used by test runs, but focused on violations.
- The lint UI can rerun narrowed linter/file subsets.
- Ignore actions in the UI write back into `.gavel.yaml`, which makes the UI a practical triage tool instead of a read-only report.

Global lint ignore list examples for `.gavel.yaml`:

```yaml
lint:
  ignore:
    - source: golangci-lint
      rule: errcheck

    - source: eslint
      rule: no-console
      file: "web/src/legacy/**/*.ts"

    - source: markdownlint
      file: "docs/generated/**"

    - file: "vendor/**"
```

How matching works:

- `source` matches the linter name such as `golangci-lint`, `eslint`, `ruff`, or `markdownlint`
- `rule` matches the violation rule/method when the linter reports one
- `file` uses doublestar-style globs
- If you combine fields, all specified fields must match
- A file-only rule is a blunt instrument; prefer `source + rule + file` when you want a narrow suppression

The linter subcommands are shortcuts for pinning `--linters`:

- `golangci-lint`, `golangci`
- `ruff`
- `eslint`
- `pyright`
- `tsc`, `typescript`
- `markdownlint`
- `vale`
- `jscpd`
- `betterleaks`, `secrets`

Use those shortcuts when you want discoverability and stable command names in scripts.

### `gavel fixtures`

`fixtures` is Gavel's declarative testing mode. It runs tests described in Markdown using tables, command blocks, and CEL assertions. This is useful when the thing you want to test is easier to express as examples and expected output than as Go test code.

Typical use cases:

- API smoke tests
- CLI contract tests
- Reproducers for edge cases
- Mixed-language command validation in docs-like files

Common workflows:

```bash
gavel fixtures tests/api.fixture.md
gavel fixtures fixtures/**/*.md
gavel fixtures -v tests.md
gavel fixtures --no-progress tests.md
```

If your repo enables fixture discovery in `.gavel.yaml`, `gavel test --fixtures` can run the same files as part of the broader test pass.

### `gavel bench`

`bench` is for controlled Go benchmark runs and base-vs-head comparisons. Use it when you want a machine-readable benchmark artifact, not just terminal output from `go test -bench`.

The flow is deliberately split:

- `gavel bench run` creates structured JSON for one benchmark run.
- `gavel bench compare` compares two JSON files and fails on regression thresholds.

Common workflows:

```bash
gavel bench run ./pkg/... --out base.json
gavel bench run ./pkg/... --out head.json --count 10
gavel bench compare --base base.json --head head.json
gavel bench compare --base base.json --head head.json --threshold 15 --ui
```

Use `--ui` when a table is not enough and you want to inspect the comparison in the browser alongside the rest of Gavel's reporting patterns.

### `gavel summary`

`summary` is the artifact-to-comment bridge. It turns a Gavel JSON result file into a compact Markdown summary suitable for PR comments, CI summaries, or issue updates.

Use it after `test`, `lint`, or CI runs that already wrote JSON output:

```bash
gavel summary --input gavel-results.json
gavel summary --input results.json --output summary.md
```

### `gavel ui serve`

`ui serve` is for replay, not execution. It loads one or more previously captured JSON snapshots and serves the browser UI without rerunning tests or linters.

Use it when you want:

- A shareable local replay of CI artifacts
- A detached inspection server for one or more saved runs
- A stable URL written to disk for scripts

Common workflows:

```bash
gavel ui serve run.json
gavel ui serve run.json other-run.json --auto-stop=10m --idle-timeout=5m
gavel ui serve run.json --url-file /tmp/gavel-ui-url
```

This is the command to reach for when someone hands you `gavel-results.json` and you want the full UI instead of reading the raw artifact.

## Review, Authoring, And TODO Automation

This part of the CLI is about turning diffs and failure reports into actionable next steps: review findings, generated commit messages, and TODO execution loops.

### `gavel verify`

Use `verify` for AI-assisted review of local changes, commit ranges, branches, PRs, files, or directories. It applies a prescribed review structure across completeness, code quality, testing, consistency, security, and performance.

It fits best when you want:

- A review pass before opening or merging a PR
- A review of a specific commit range or PR
- An automated fix loop driven by the review findings
- Findings materialized as TODO files

Common workflows:

```bash
gavel verify
gavel verify main..HEAD
gavel verify #123
gavel verify path/to/file.go
gavel verify --model gemini
gavel verify --auto-fix --max-turns 5
gavel verify --sync-todos
```

Use `--patch-only` when the AI side should return patches instead of relying on interactive tool use.

### `gavel commit`

`commit` is for authoring commit messages and running pre-commit hooks described in `.gavel.yaml`. It can generate one conventional commit message, or plan multiple commits from a larger staged change set.

This is the right command when you want:

- An LLM-generated conventional commit message
- AI warnings for removed functionality or compatibility issues before the commit is written
- Hook execution before finalizing the commit
- AI-assisted splitting of a large change into multiple commits
- Optional follow-up push behavior

Common workflows:

```bash
gavel commit
gavel commit -A
gavel commit -A --max=5
gavel commit -m "chore: bump dep"
gavel commit --stage all --dry-run
gavel commit --force
```

### `gavel todos`

`todos` is the task-execution side of Gavel's failure-to-fix loop. Other commands create `.todos`; this command helps you inspect, verify, group, and execute them.

Use it when you already have TODO files and want to:

- List them by status or grouping
- Inspect one TODO in detail
- Re-run verification checks
- Execute them through the Claude Code integration

Common workflows:

```bash
gavel todos list
gavel todos list --status pending
gavel todos get .todos/fix-bug.md
gavel todos check
gavel todos run
gavel todos run --interactive
gavel todos run --group-by directory
```

`gavel pr fix` builds on the same execution engine, but starts from a pull request instead of an existing `.todos` directory.

## Pull Requests, Git History, And Repository Mapping

These commands help when the unit of work is no longer "run tests" but "understand what changed", "watch a PR", or "classify a repository".

### `gavel pr status`

Use `pr status` to inspect or follow GitHub Actions state for a pull request. It can resolve the current branch's PR automatically, fall back to the most recent PR, and optionally fetch failed job logs.

This is the fastest command for:

- Checking whether a PR is green
- Following checks until completion
- Pulling failed log tails into the CLI
- Syncing failed jobs or PR comments into TODO files

Common workflows:

```bash
gavel pr status
gavel pr status 42
gavel pr status https://github.com/owner/repo/pull/123
gavel pr status --follow --interval 30s
gavel pr status --logs --tail-logs 50
gavel pr status --sync-todos
```

### `gavel pr list`

`pr list` is for portfolio-level PR visibility. It can show your own PRs, all PRs in an org, a live dashboard in the browser, or a macOS menu bar indicator.

Use it when you want:

- A multi-repo personal PR queue
- A lightweight operations view for open PRs
- A background dashboard with periodic refresh

Common workflows:

```bash
gavel pr list
gavel pr list --state merged --since 30d
gavel pr list --any --status
gavel pr list owner/repo another/repo
gavel pr list --all --org myorg
gavel pr list --ui
gavel pr list --menu-bar
```

If you want the UI-first version of PR operations, this is it. `--ui` is for an explicit browser dashboard; `--menu-bar` is for ambient status on macOS.

### `gavel pr fix`

`pr fix` is the bridge from PR breakage to interactive remediation. It fetches PR status, syncs TODOs from failures and comments, lets you choose what to work on, and then executes the selected TODOs.

Use it when your workflow is "show me what's broken on this PR and let me start fixing it now."

Common workflows:

```bash
gavel pr fix
gavel pr fix 42
gavel pr fix --group-by directory
gavel pr fix --dry-run
```

### `gavel git history`

Use `git history` when you want filtered commit data rather than a terminal `git log` stream. It is the raw history retrieval command that powers the more semantic `git analyze` workflows.

Common workflows:

```bash
gavel git history --path .
gavel git history --since 2024-01-01 --until 2024-06-01
gavel git history --author alice --message "feat*"
gavel git history abc1234
gavel git history main..feature-branch
```

### `gavel git analyze`

`git analyze` is the semantic layer on top of commit history. It classifies changes by scope and technology, applies severity rules, understands repository architecture, and can optionally use AI to enrich analysis or produce summaries.

This is the right command when you want:

- A higher-level explanation of what a series of commits changed
- Scope or technology filtering over history
- Summary views over a day, week, month, or year
- Re-analysis from previously saved JSON

Common workflows:

```bash
gavel git analyze
gavel git analyze --ai --model claude-sonnet-4-20250514
gavel git analyze --scope backend --tech kubernetes
gavel git analyze --summary --summary-window week
gavel git analyze --include bots --exclude merges --verbose
gavel git analyze --input previous-run.json
```

The quality of `git analyze` depends heavily on `arch.yaml` and `.gitanalyze.yaml`, so this command becomes much more useful after the repo has been classified well.

### `gavel git init-config`

Use `git init-config` to bootstrap `.gitanalyze.yaml`. It writes sensible defaults first, then can optionally ask an AI CLI to recommend repo-specific additions.

Common workflows:

```bash
gavel git init-config
gavel git init-config --model gemini
gavel git init-config --model codex
gavel git init-config --model none
```

### `gavel git amend-commits`

`git amend-commits` is a history-rewriting tool for improving low-quality commit messages. It reviews commits below a score threshold, suggests replacements, and then interactively rewords them through rebase.

Use it carefully on branches you control.

Common workflows:

```bash
gavel git amend-commits
gavel git amend-commits --threshold 5.0 --ref main
gavel git amend-commits --dry-run
```

### `gavel repomap`

`repomap` exposes the merged architecture and file-mapping rules that Gavel uses elsewhere. If `git analyze` or other classification-aware features look surprising, this is the inspection command to run next.

It has two complementary modes:

- `gavel repomap view` shows the merged effective config for a path.
- `gavel repomap get` shows the resolved file map for a file or directory.

Common workflows:

```bash
gavel repomap view .
gavel repomap view ./cmd
gavel repomap get ./main.go
gavel repomap get src/main.go
```

## Infrastructure And Long-Running Services

These commands are for environments where Gavel is not just a one-shot CLI, but part of a longer-lived developer or team workflow.

### `gavel ssh`

The `ssh` group turns Gavel into a git-push backend. Developers push to a Gavel SSH remote; the server accepts the push, runs `gavel test --lint` by default, and streams results back like a lightweight local CI service.

Use `ssh serve` when you want to host that flow in the foreground:

```bash
gavel ssh serve --port 2222
git remote add gavel ssh://localhost:2222/myproject
git push gavel HEAD:main
```

Use `ssh install` on Linux when you want the same backend managed by systemd:

```bash
gavel ssh install
gavel ssh install --dry-run
gavel ssh install --port 3333 --user gavel
```

### `gavel system`

The `system` group manages a detached PR dashboard service. Concretely, it is about running `gavel pr list --all --ui --menu-bar` in the background, keeping its service definition installed, and checking whether the daemon is healthy.

Use it when you want an always-on PR dashboard instead of launching `pr list --ui` manually each time.

Common workflows:

```bash
gavel system install
gavel system start
gavel system status
gavel system stop
gavel system uninstall
```

Command intent:

- `install`: create the user-level service and persist daemon config such as database mode and GitHub token
- `start`: launch the detached daemon immediately
- `status`: inspect service state, DB mode, live daemon health, and recent logs
- `stop`: terminate the daemon process
- `uninstall`: remove the installed service definition

## Utility Commands

### `gavel completion`

Use `completion` to generate shell completion scripts for Bash, Fish, PowerShell, or Zsh.

Typical workflows:

```bash
gavel completion zsh > "${fpath[1]}/_gavel"
gavel completion bash > /etc/bash_completion.d/gavel
```

### `gavel version`

Use `version` when you need the exact build identifier, commit, build date, and Go toolchain version:

```bash
gavel version
```

## Common End-To-End Workflows

### Local quality pass before opening a PR

```bash
gavel test --lint --changed
gavel verify main..HEAD
gavel commit
```

### CI-style artifacts with local replay

```bash
gavel test --lint --format "json=gavel-results.json,html=gavel-results.html"
gavel summary --input gavel-results.json --output gavel-summary.md
gavel ui serve gavel-results.json
```

### Triage and fix a broken PR

```bash
gavel pr status --logs --sync-todos
gavel todos list
gavel pr fix
```

### Stand up a persistent PR dashboard

```bash
gavel system install
gavel system start
gavel system status
```

## Choosing The Right Command

If you only remember a few entrypoints, use this map:

| Goal | Start here |
| --- | --- |
| Run tests across the repo | `gavel test` |
| Run linters across the repo | `gavel lint` |
| Run declarative Markdown tests | `gavel fixtures` |
| Review a diff with AI | `gavel verify` |
| Generate a commit message | `gavel commit` |
| Check PR status | `gavel pr status` |
| Browse many PRs | `gavel pr list` |
| Analyze commit history semantically | `gavel git analyze` |
| Understand file/scope classification | `gavel repomap` |
| Work through synced TODOs | `gavel todos` |
| Replay a saved UI snapshot | `gavel ui serve` |
| Run a background PR dashboard | `gavel system` |
| Turn pushes into local CI | `gavel ssh` |
