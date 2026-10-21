# Gavel Manual

`gavel` is a multi-purpose CLI for test orchestration, linting, AI-assisted review, pull-request operations, git history analysis, TODO automation, and a few long-running service flows.

This manual is organized around the actual CLI tree and answers two questions for each feature area:

- What problem is this part of the CLI meant to solve?
- Which command should you reach for first?

## How To Read This Manual

The CLI has two layers:

- Root-level feature commands such as `test`, `lint`, and `commit`.
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
|-- projects
|   |-- add
|   |-- delete
|   |-- get
|   |-- hide
|   |-- list
|   |-- unhide
|   `-- update
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
|   |-- history
|   |-- jest
|   |-- playwright
|   `-- vitest
|-- todos
|   |-- check
|   |-- get
|   |-- list
|   |-- run
|   `-- verify
|-- ui
|   `-- serve
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

- Live UI commands bind `0.0.0.0` by default, so the dashboard is reachable from another machine on the LAN out of the box. Pass `--addr localhost` to restrict a server to this machine.
- `gavel test --ui --detach` and `gavel ui serve` support `--auto-stop` and `--idle-timeout` so the server cleans itself up.
- `gavel ui serve` is the easiest way to inspect CI-produced JSON locally.
- The test and lint UIs share the same underlying server, so lint results can appear alongside test results when you run `gavel test --lint --ui`.
- `--dev` serves a UI from its Vite dev server instead of the bundle embedded in the binary, so source edits show up on a browser refresh. `gavel test --ui --dev`, `gavel fixtures --ui --dev` and `gavel ui serve --dev` start `pnpm dev` in `testrunner/ui` on port 5275, and `gavel pr list --ui --dev` / `gavel serve --dev` do the same in `pr/ui` on port 5273. A dev server already running on the port is reused. The Go server keeps every `/api/*` route and proxies everything else to Vite. It needs a source checkout of gavel, run from inside it, with the UI's `pnpm install` done.
- In dev mode, `@flanksource/clicky-ui` resolves from the sibling checkout's source (`../clicky-ui/packages/ui/src`) when that checkout exists, so clicky-ui edits show up on a refresh too. The aliases follow the sibling's `package.json` exports. Its stylesheets come from the sibling's `dist/`, so build it once (`task clicky-ui:build`). `GAVEL_UI_WATCH_CLICKY=0` keeps resolving the sibling's source but stops watching it.

### TODO Flows

Native PostgreSQL is the only runtime TODO store. Use `gavel todos` to create,
sync, inspect, run, and verify issues in the current registered workspace:

- `gavel todos create "Describe the work"`
- `gavel todos sync` to scan source `TODO`/`FIXME` comments into PostgreSQL
- `gavel todos list`, `get`, `run`, and `verify` for the native queue

Repository-local `.todos` Markdown is an explicit interchange format, not a
runtime provider. Use `gavel todos import [files...]` to load it into PostgreSQL
and `gavel todos export [refs...]` to write portable snapshots.

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
  model: claude-haiku-4-5
  groupModel: claude-sonnet-4-5
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
| `verify.model` | Default AI CLI or model name for generic AI-verification fixtures |
| `verify.prompt` | Optional custom verify prompt text |
| `verify.checks.disabled` | Disable specific verify checks by ID |
| `verify.checks.disabledCategories` | Disable whole verify categories |
| `lint.ignore` | Repo-wide or user-wide ignore rules matched by `source`, `rule`, and/or `file` |
| `lint.linters.<name>.enabled` | Force a linter on or off when Gavel would otherwise rely on detection/default behavior |
| `commit.model` | Model for commit-message and PR-content generation (fast/haiku-class); overridable with `--model` |
| `commit.groupModel` | Model for AI commit grouping (`gavel commit -A`, capable/sonnet-class); overridable with `--group-model`, falls back to `commit.model` |
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
| `commit.model`, `commit.groupModel` | Last non-empty value wins |
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

This is the biggest part of the CLI. The common pattern is: discover work automatically, run it with structured output, then optionally render it in a browser or export JSON/HTML artifacts.

### `gavel test`

Use `test` when you want Gavel to orchestrate repository test execution instead of calling framework CLIs yourself. It auto-detects supported frameworks, groups work by repo/module boundaries, and can combine tests, fixtures, benchmarks, UI output, caching, and change-aware selection.

Reach for it when you want one of these flows:

- Run everything Gavel can detect.
- Run only packages affected by local changes.
- Open a live dashboard while tests run.
- Re-run from a previous JSON baseline or only rerun failed targets.
- Re-run the failures a pull request hit in CI, without re-running the whole suite.

Common workflows:

```bash
gavel test
gavel test ./pkg/...
gavel test --lint --format "json=gavel-results.json,html=gavel-results.html"
gavel test --changed
gavel test --failed
gavel test --pr 42
gavel test --ui
gavel test --ui --detach --auto-stop=30m --idle-timeout=5m
gavel test --bench .
gavel test --fixtures
gavel test history
gavel test history ./pkg/foo
```

The UI path is worth calling out separately:

- `gavel test --ui` starts a live browser dashboard before pre-hooks run, so hook status, test progress, reruns, and optional lint results all stay in one place.
- `gavel test --lint --ui` is the richest local feedback loop because the same UI can show both test and lint findings.
- `gavel test --ui --detach` turns the live run into a handoff flow: the CLI starts the run, then a child UI server stays up until `--auto-stop` or `--idle-timeout` expires.

Diagnosing hangs and timeouts:

- `--timeout` (default `20m`) bounds the whole run; `--test-timeout` bounds each test package and has no default unless `test.testTimeout` sets one. When a package passes its deadline, each Go test binary under it is sent `SIGQUIT` and given 3s to print its goroutines before the tree is killed; the dump lands on the timed-out attempt as `stack_trace` and `goroutine_count`.
- `--diagnostics` records the run's process tree, CPU/RSS/disk samples and stack dumps into `.gavel/traces/<run_id>` without a browser (`--ui` does the same). The watchdog checkpoints Gavel's goroutines there every 10 minutes (`periodic`), on `timeout`, on `package-timeout:<pkg>` and at the end (`final`). `--process-facets` adds each process's open files, sockets, ports and environment, and needs `--ui` or `--diagnostics`.
- `gavel ui serve <snapshot>` reopens the run's trace store read-only, so the Processes tab and stack dumps are browsable after the run.

Framework pinning is available in two forms:

- `gavel test --framework go --framework ginkgo`
- `gavel test go`, `gavel test ginkgo`, `gavel test jest`, `gavel test vitest`, `gavel test playwright`

Use the framework subcommands when you want the full `test` flag surface but do not want auto-detection to choose runners for you.

`gavel test history` reads completed run snapshots from `.gavel/run-*.json` and shows a package/file/suite outline of executable tests. Leaf rows include execution count, pass rate, min/avg/max duration, last passed, last failed, and the date the test first appeared in local history. Optional paths filter by package or file relative to `--cwd`.

### `gavel lint`

Use `lint` when you want a single entrypoint for repo linting across languages and tools. Gavel auto-detects installed linters and can narrow to changed files, show a browser UI, or auto-fix findings.

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
- `react-doctor`
- `pyright`
- `tsc`, `typescript`
- `markdownlint`
- `vale`
- `jscpd`
- `betterleaks`, `secrets`

Use those shortcuts when you want discoverability and stable command names in scripts.

#### Cap linters run outside gavel (`gavel lint install-shims`)

Editors, agents and scripts often call `golangci-lint`, `eslint` or `tsc` directly, running the full rule set with no bounds. `install-shims` writes PATH shims for them that run the real tool through gavel with cheap defaults and hard limits:

```bash
gavel lint install-shims --timeout 30s --cpu-limit 200 --mem-limit 2GB
export PATH="$HOME/.gavel/shims:$PATH"   # printed when the dir is not yet first on PATH
```

- `golangci-lint run` gets `--fast-only`, `--concurrency=<cores>` and `--allow-parallel-runners`, with `GOMAXPROCS` and `GOMEMLIMIT` set from the limits. golangci-lint otherwise runs only one lint per user at a time, so concurrent agents would queue and use up their timeout; each shimmed run is already capped.
- `eslint` gets `--cache`, with the cache kept in the user cache dir; `eslint` and `tsc` get `UV_THREADPOOL_SIZE` and `--max-old-space-size` set from the limits.
- Every shimmed linter is niced (`--nice`, default 10). It is killed after `--timeout` (exit 124), and killed when its process tree stays above `--cpu-limit` percent or exceeds `--mem-limit` (exit 137).
- Flags you pass explicitly (`--fast-only=false`, `--concurrency`, `--allow-serial-runners`, `--no-cache`, an existing `--max-old-space-size`) are left alone.
- `GAVEL_SHIMS=off golangci-lint run` runs the real tool unchanged. `gavel lint` itself always resolves the real binaries behind the shims.

Use `--tools` to choose which linters to shim, `--dir` to choose where the shims go, and `--dry-run` to preview. Re-run the command to change the limits.

### `gavel test outline`

`test outline` provides one read-only inventory across Go test, Ginkgo, Jest, Vitest, Playwright, and Markdown fixtures. It uses native framework collection where available, statically parses Jest declarations, and parses fixture Markdown without running setup or commands.

```bash
gavel test outline
gavel test outline --framework go,ginkgo ./pkg
gavel test outline --framework fixture --fixture-files 'examples/*.fixture.md'
```

An empty `--framework` includes every supported source. Positional paths filter all collectors; `--fixture-files` overrides the configured fixture patterns. Collection failures remain visible as error rows so one broken Node package does not hide the rest of the repository.

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
gavel fixtures --profile tests.md
gavel fixtures --no-progress tests.md
gavel fixtures outline fixtures/**/*.md
```

Use `gavel fixtures outline` to inspect fixture structure, source locations, and
fixture kind counts without running build commands, daemons, fixture commands,
test/lint runner steps, skip commands, or AI checks.

If your repo enables fixture discovery in `.gavel.yaml`, `gavel test --fixtures` can run the same files as part of the broader test pass.

`--profile` reads CPU, memory, and disk I/O for the fixture process tree from the run's `process_tree` and `process_samples` traces: each step's peaks are its command's process tree over the step's time, and its disk I/O is the step's delta of the cumulative counters. Save this example as `profile.fixture.md` in the Gavel repository, then run `gavel fixtures --profile --ui profile.fixture.md`:

````markdown
---
exec: sh
goProfiles:
  cpu:
    file: cpu.pprof
    env: CPU_PROFILE
---

# Go CPU profile

```sh
go test ./internal/safefs -run TestSafeRemove -cpuprofile="$CPU_PROFILE"
```
````

Gavel supplies profile destinations only with `--profile`, gives each fixture an isolated output directory, and ingests every file the command emits into the run's `goprof` trace, one window per file, listed with the fixture's traces. A missing file or empty directory is logged and skipped. An emitted file must parse as a Go pprof profile, and must stay inside the fixture's output directory. The entry's name (`cpu`, `heap`, `allocs`, `mutex`, `block`, `goroutine`) says which profile it is; another name is read from the profile's sample types. Commands that accept a filename argument can use `args: ["-cpuprofile={{.path}}"]` instead of `env:`; commands that emit several profiles can use `directory:` instead of `file:`. Each window downloads in the `pprof` export format (`.pb.gz`), which `go tool pprof` opens.

`--benchmark` records each fixture's and each prerequisite's (setup, build, daemon) duration, with its profile and SQL totals when collected, and writes the report to `.gavel/benchmarks/fixtures/run-<timestamp>.json`; copy one somewhere stable to use it as a `--baseline`. Use `--benchmark` with limits to fail individual fixture tests when they exceed a budget. `--max-time`, `--max-sql-time`, and `--max-sql-query` accept Go durations; `--max-memory`, `--max-io-read`, and `--max-io-write` accept byte sizes. `--max-sql-queries` and `--max-slow-sql` cap statement counts, including zero. `--max-deviation-pct` requires `--baseline` and checks positive regressions against each matching fixture's measured time, peak RSS, disk I/O, and SQL timings and counts. Metrics absent from the baseline are skipped; a missing current measurement fails when an absolute limit or a previously measured baseline value requires it.

```bash
gavel fixtures --benchmark --max-time 2s --max-memory 512MiB tests/performance.fixture.md
gavel fixtures --benchmark --max-sql-time 500ms --max-sql-query 100ms --max-slow-sql 0 tests/database.fixture.md
gavel fixtures --benchmark --baseline .gavel/benchmarks/fixtures/base.json --max-deviation-pct 15 tests/performance.fixture.md
```

When profiling or SQL limits are active, Gavel passes `SQL_PROFILE_FILE` to each command. Applications using commons-db's GORM `NewSqlLogger` export one JSON line per statement there, including raw SQL text, string-rendered bound parameters, duration in nanoseconds, row count, and slow/error flags. Gavel saves those statements with per-fixture totals in its benchmark report and shows them in the fixture benchmark view. Profile files and benchmark reports can contain query data. A fixture with a SQL limit fails if its command never creates the SQL profile file.

#### YAML and JSON fixtures

A fixture can also be a YAML or JSON document (`*.fixture.yaml`, `*.fixture.yml`, `*.fixture.json`, discovered alongside `*.fixture.md`). Instead of Markdown tables and code blocks, it lists typed **steps**. Each step has an `id` and a `kind`, which selects what the step does and which fields it takes. Steps run in order; once one fails, the later steps are reported as skipped. `gavel fixtures --schema` prints the JSON schema of every kind.

```yaml
kind: Fixture
name: exec steps share exported env
defaults:
  wait: { interval: 50ms, backoff: { factor: 2, max: 500ms }, jitter: 20%, timeout: 10s }
  timeout: 1m
env:
  GREETING: hello
steps:
  - id: mint
    kind: exec
    script: 'printf "{\"token\":\"t-%s\"}" 42'
    cel: json.token == 't-42'
    export:
      TOKEN: "{{.outputs.json.token}}"

  - id: use
    kind: exec
    script: 'echo "token=$TOKEN"'
    cel: stdout.contains('token=t-42') && steps.mint.json.token == 't-42'
```

Save this as `env.fixture.yaml` and run `gavel fixtures env.fixture.yaml`. The full version is `fixtures/testdata/kinds/exec-env.fixture.yaml`.

**Environment and exports.** A fixture has one environment, made of the process env plus `env:`. Every step's subprocesses receive it. Inside a step, `{{.env.NAME}}` and `$NAME` read from it. A step's `export:` map renders templates over that step's outputs (`{{.outputs.…}}`) and adds the results to the environment of every later step. Some kinds also export defaults, for example `TODO_ID` and `GAVEL_URL`, so URLs, ids and session names are mapped once rather than repeated. Exporting a name that an earlier step already set is an error unless the step sets `override: true`. Referencing an unset `$NAME` fails the step. A later step's CEL and templates can also read an earlier step's outputs through `steps.<id>.*`.

**Templates.** Step fields are gomplate templates with access to `.env`, `.steps`, `.tmp` (a per-fixture scratch directory) and `.fixture.dir`. Shell scripts keep `$NAME` and `$?` as plain shell, so only `{{…}}` is expanded in them.

**Wait, retry and timeout.** These work the same way for every kind:

- `wait` acts the step out once, then re-observes its state on a schedule until `until` is true. The schedule starts at `interval`, grows by `backoff.factor` up to `backoff.max`, and varies each delay by `jitter`, given either as a percentage or a duration. The step fails early when `failWhen` is true, and fails with the last observed outputs once `timeout` passes. Every observation is recorded in the step result, so a timeout shows how the state changed over time.
- `retry` re-runs a failed step up to `attempts` times, but only while the CEL `when` holds. Its `delay`, `backoff` and `jitter` work the same way as for `wait`.
- `timeout` caps the whole step.

A fixture's `defaults:` block sets the `wait`, `retry` and `timeout` that every step inherits.

**Background steps.** A step with `background: true` starts and keeps running while later steps execute. Its verdict is collected at teardown.

| Kind | Does | Default exports |
|------|------|-----------------|
| `exec` | Runs `script` in a shell. CEL sees `stdout`, `stderr`, `exitCode` and `json` (stdout parsed as JSON). | – |
| `ai` | Runs a captain AI check from `prompt`/`criteria`/`checklist` with an optional `model`. | – |
| `services` | Starts named services in order and tears them down in reverse. Each entry is one of `deps: postgres` (a deps catalog service), `exec:` (a supervised command, with `{{.port}}` allocated and readiness checked by `ready: {http\|tcp\|exec}`), or `git:` (a seeded temporary repository). Later entries can read earlier ones through `{{.services.<name>.url}}`. | `<NAME>_URL`, `<NAME>_PORT` |
| `todo` | Creates a gavel TODO (`create: {title, body, triage}`), or runs `approve`, `reject` or `get` on one, over `$GAVEL_URL`. CEL sees `todo`, `triage`, … | `TODO_ID`, `TODO_URL` |
| `todo_prompt` | Launches a lifecycle `step` (`plan`, `run` or `verify`) and observes it. CEL sees `launch`, `todo`, `plan`, `verify`, `session.errors`. | – |
| `browser` | Starts Chrome with a profile (`profile: {fresh: true}`) for later steps to attach to over CDP. | `BROWSER_CDP_URL`, `BROWSER_WS_URL`, `BROWSER_SESSION`, `BROWSER_PROFILE_DIR` |
| `playwright` | Runs a Playwright spec, either from `file:` or inline `script:`, against the shared Chrome (`import { test } from '../fixtures/cdp'`). It reports each test as a child result. With `background: true`, `file: ../../e2e/watch-session.ts` watches a page and fails at teardown on any session error alert, stream error banner, console error, page error, or `/api/` response with status 400 or higher. | – |
| `agent-browser` | Runs `commands:` with `agent-browser --engine chrome` attached to the shared Chrome. Setting `observe: snapshot` re-takes the snapshot on each `wait` observation. | – |

`pr/ui/testdata/todo-lifecycle/claude.fixture.yaml` drives a TODO end to end: create, triage, plan, approve, run and verify. It runs the real `claude` CLI against captain's scripted AI mock and watches the Session tab for errors throughout. Run it with `gavel fixtures pr/ui/testdata/todo-lifecycle/claude.fixture.yaml`. It needs `captain`, `claude` and Chrome on `PATH`, and the sibling `clicky-ui` checkout for `serve --dev`. `codex.fixture.yaml` runs the same flow with the `codex` CLI.

Ready-made workflow templates live in `examples/` (they run against the bundled `examples/sample-app`): `precommit.fixture.md` (lint + test the files you changed), `pre-release.fixture.md` (build, then the full suite and every linter), `smoke-test.fixture.md` (fast smoke tests + a `/health` probe), and `ai-review.fixture.md` (an AI reviewer scores the change against an acceptance-criteria checklist — needs `captain configure`). Run one with `gavel fixtures examples/precommit.fixture.md`, or copy it into your project.

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

### `gavel todos check`

Use `todos check` to run a TODO's complete fixture-backed definition of done
without starting an implementation agent. It executes configured test/lint
checks, the persisted `## Verification` fixture, and any acceptance-criteria
checklist step through the same CEL verdict used inside `todos run`.

```bash
gavel todos check 3f2a1b                 # check one TODO
gavel todos check 3f2a1b --timeout 10m   # bound this verification run
```

### `gavel commit`

`commit` is for authoring commit messages and running pre-commit hooks described in `.gavel.yaml`. It can generate one conventional commit message, or plan multiple commits from a larger staged change set.

This is the right command when you want:

- An LLM-generated conventional commit message
- AI warnings for removed functionality or compatibility issues before the commit is written
- Hook execution before finalizing the commit
- Interactive file selection with an in-picker `/` filter for path, status, language, or scope
- AI-assisted splitting of a large change into multiple logical commits (`-A`), with a separate chore commit for lock/generated files. `-A` feeds the LLM a `gavel status` table (scope, file, status, line counts), groups by logical change (scope is a hint, not a hard boundary), and caps the result at `--max-commits` (default 7, excluding the chore commit) — rendered into the grouping prompt's output schema as `maxItems` and enforced by captain's `schemaStrictness=retry` policy. Each group is committed as soon as its message is ready
- Optional follow-up push behavior

By default (`--stage session`) commit scopes itself to the running agent's edits: it resolves a session id from `GAVEL_SESSION_ID`, `CLAUDE_SESSION_ID`, or `CODEX_SESSION_ID` and stages only the files that Claude or Codex session touched. With no session id in the environment it falls back to the previous default of committing the already-staged set. Pass `--stage staged|unstaged|all` or an explicit session id to override.

Common workflows:

```bash
gavel commit
gavel commit -t
gavel commit -A
gavel commit -A --max-commits=3
gavel commit --max-commits=3
gavel commit -m "chore: bump dep"
gavel commit --stage all --dry-run
gavel commit --force
```

### `gavel todos`

`todos` is the task-execution side of Gavel's failure-to-fix loop. Its runtime
queue is stored in native PostgreSQL and shared by the CLI and UI.

Use it when you want to:

- Create or sync native issues
- List issues by status or grouping
- Inspect one issue in detail
- Re-run verification checks
- Drive an AI coding agent through the project's todo lifecycle
- Combine several overlapping issues into one
- Explicitly import or export portable `.todos` Markdown

Common workflows:

```bash
gavel todos list
gavel todos list --status pending
gavel todos get 3f2a1b
gavel todos create "Fix the parser race"
gavel todos sync ./pkg/parser
gavel todos steps 3f2a1b                   # which lifecycle steps apply to this todo now
gavel todos check 3f2a1b
gavel todos run 3f2a1b
gavel todos run 3f2a1b --step plan         # propose a plan; the TODO parks in `review`
gavel todos plan approve 3f2a1b --run      # accept the plan and implement it
gavel todos plan revise 3f2a1b --feedback "split the migration in two"
gavel todos plan reject 3f2a1b             # discard the plan; the TODO returns to pending
gavel todos merge 3f2a1b 9c4d2e --dry-run  # preview folding two TODOs into one
gavel todos merge 3f2a1b 9c4d2e            # write the merge
gavel todos import --dir ./archive/todos   # explicit Markdown → PostgreSQL
gavel todos export 3f2a1b --dir ./backup   # explicit PostgreSQL → Markdown
```

#### `gavel todos merge <id> <id>...`

Folds several TODOs into one. A single AI pass (the `todos.merge` prompt, see
[SCHEMA.md](SCHEMA.md#prompt-overrides)) reads every selected TODO — body,
acceptance criteria, verification fixture verbatim, and existing plan — and
returns one of each. Gavel performs the writes:

- the **survivor** (the first id given, or `--into <id>`) has its title, body,
  verification fixture, labels and severity replaced by the merged versions, and
  the merged plan saved as a new *unapproved* revision — merging changed the
  scope the previous approval was given for, so the plan goes back through
  `gavel todos plan approve|revise`;
- every other TODO gets a comment saying why it is part of the merged work, a
  `related_to` link to the survivor, and is then **soft-deleted** (cancelled):
  its history, comments and runs stay intact;
- a TODO the model judges is *not* the same work is reported as `excluded` and
  left completely untouched.

It is not a lifecycle step and records no run. Every selected TODO must belong
to the same workspace, and the merge is refused before any write when the model
drops a fixture or a plan a source TODO had, or returns a fixture that does not
parse. `--dry-run` prints the proposed title, body, fixture and plan and writes
nothing. The same action is on the dashboard's selection toolbar (Merge), where
the survivor is the first row selected.

#### The todo lifecycle

Every todo moves through a **lifecycle**: an ordered list of steps declared in
`todos/lifecycle/todos.yaml` (overridable from `.gavel.yaml` `todos.lifecycle`,
see [SCHEMA.md](SCHEMA.md#todoslifecycle)). Each step is:

- a captain **prompt** reference (`todos.run`, `todos.plan`, `todos.triage`,
  `todos.verify`, or `file:<path>` for a project-owned template),
- a CEL **`when`** predicate over `subject` (the todo), `runs` (its run
  history), and `last` (the latest run per step) that decides whether the step
  currently applies,
- the **`events`** its result may report, and the **`tools`** sets it runs with,
- an ordered list of **`outcomes`** — `{status, when}` pairs evaluated in
  order, first true wins — that decide which status the finished run lands the
  todo in. No outcome matching is an error: a run the definition cannot
  classify must not silently keep the status it started under.

Every step's agent ends with the same result envelope, `{success, event, comment}`: `success` says whether the step did its job, `event` is one of the step's declared `events` (the schema is built per step, so any other value is refused), and `comment` is the markdown that becomes the run's history entry. Outcomes read them as `envelope.success` and `envelope.event`, next to `run` (the run's own fate) and `verify` (the definition-of-done report). The `verify` step generates nothing; its envelope is synthesised from the report with the events `passed` and `failed`.

A step changes TODOs only through the tools it declares, and each tool writes when it is called — gavel applies nothing from the result afterwards:

| Set | Tools | Policy |
| --- | --- | --- |
| `todos` | `todo_search`, `todo_get`, `todo_update`, `todo_labels`, `todo_link`, `todo_link_pr`, `todo_set_parent`, `todo_comment`, `todo_plan` (plan steps only) | allow |
| `todos` | `todo_merge`, `todo_close` | ask — a person approves each call |
| `ask` | `todo_ask` | allow; it raises its own question through the run's approval broker |
| `pr` | `pr_thread_addressed` (only on a run that pushes to the todo's PR) | allow |

`gavel todos run --preview` runs the agent, but `todo_merge` and `todo_close` report what they would do and write nothing. Each finished run records one `lifecycle_outcome` event: its body is the comment plus a verification line, and its payload carries `envelope: {success, event}` and `tools: [{name, ref}]`, every write the run's tools made. Todo prompts deny the native `AskUserQuestion`; a question goes through `todo_ask`, which blocks inside the run until the answer arrives. `gavel todos run` from a terminal answers approvals and `todo_ask` questions at that terminal; with stdin redirected, or from the dashboard, they go to Captain's durable broker and are answered on the dashboard. Caller tools are exposed on the `agent`, `api`, `cli` and `cmux` runtimes (claude and codex); preflight refuses a step that declares tools on a runtime that cannot expose them.

The built-in lifecycle's steps, in the order `gavel todos run` considers them,
are `triage` (auxiliary — never picked automatically, only run by name),
`plan`, `verify`, and `run`. `review` and `ask` are **human-facing statuses**,
not steps: a todo in `review` is waiting for `gavel todos plan
approve|reject|revise`, and one in `ask` has a run waiting on an approval
or a `todo_ask` question — nothing in the lifecycle is "next" for either
until a person acts.

`gavel todos run [--step <name>]` runs one lifecycle step for each selected
todo. With no `--step`, the host picks the first non-auxiliary step whose
`when` predicate holds (`Next` in `todos/lifecycle/engine.go`) and prints that
step's name and its reason next to it before dispatching. Passing `--step
<name>` runs that named step regardless of what the lifecycle would have
picked — including an auxiliary step like `triage`, which is never chosen
automatically. Pass `--dry-run` to print the rendered prompt, the resolved spec
layer stack (lowest precedence first), and the resolved Captain spec, without
dispatching anything.

`gavel todos steps [todo]` lists the lifecycle itself when given no argument
(every step, its prompt reference, whether it is auxiliary, and its `when`
predicate), or — given a todo — reports where that todo stands: whether each
step applies to it now and why, which one `gavel todos run` would pick next,
and how that step's last run ended.

`gavel todos check [ids...] [--concurrency N]` **is** the lifecycle's `verify`
step, run standalone: the same fixture/CEL definition of done a `run` step's
in-loop check performs, dispatched by name instead of by the lifecycle picking
it. `--concurrency` bounds how many todos are checked at once (0 uses
`.gavel.yaml` `todos.checkConcurrency`, else 4 — each check runs that todo's
fixture, so unbounded fan-out thrashes the machine).

`--model` selects the model (and, via its compact `mode:model:effort` form,
the execution mechanism) together: `cli:opus:high`, `api:sonnet`. Empty uses
the step's `.prompt` frontmatter default.

The `run` step creates a new worktree by default
(`setup.checkout.worktree.mode: new` in `todos/lifecycle/todos.yaml`) and
commits its work with `commits[].stage: worktree` — the run commits from
inside that worktree rather than the caller's working tree. `--dirty` is a
request-layer shorthand for a worktree that carries the working tree's
uncommitted and gitignored content across, equivalent to declaring:

```yaml
setup:
  checkout:
    worktree:
      mode: new
      uncommitted: clone
      ignored: clone
```

It applies even when the project configures no checkout block of its own, so
the flag can never be a silent no-op.

#### Retired flags and config keys

The old ad-hoc run-mode/driver model is gone. These flags and `.gavel.yaml`
keys no longer exist:

| Retired | Replacement |
| --- | --- |
| `todos run --driver` | the compact model form, e.g. `--model cli:opus:high` |
| `todos run --mode` | `--step` (a run is one lifecycle step for one todo) |
| `todos run --prompt` | `--step` |
| `todos run --group-by` | `--step` (grouping is gone; runs dispatch per todo) |
| `todos run --check` | `.gavel.yaml checks.enabled` (the checks are part of the definition of done, rendered from configuration; the loop budget is `todos.run.workflow.verify.maxIterations`) |
| `todos.driver` | `ai.model` with the compact `mode:model:effort` form, e.g. `ai.model: "cli:opus:high"` |
| `todos.prompts` | a lifecycle step under `todos.<step>` |
| `todos.groupBy` | nothing — grouping was removed; runs dispatch per todo |
| `todos.approvals` | `permissions.mode: default` on the step (the dashboard brokers each tool call) |
| `checks.maxIterations` | `todos.run.workflow.verify.maxIterations` |

A `.gavel.yaml` that still declares one of the removed keys fails to load with
the exact message above (`<path>: <key> is no longer supported; use
<replacement>`), rather than silently falling back to built-in defaults.

#### Run the tests and linters when the agent is done (`checks:`)

After an agent reports a TODO done, the configured `checks:` suite runs your real
test and lint suite and, if anything fails, feeds a compact failure summary back
into the *same* agent session so it can fix the issues — re-running the suite
until it passes or a `todos.run.workflow.verify.maxIterations` cap is hit. This
closes the loop: the agent doesn't just claim it's finished, it has to leave the
suite green.

```bash
gavel todos run 3f2a1b
gavel todos run 3f2a1b --model cmux:opus   # primary path: resumes the live agent REPL
```

The loop is **opt-in**: it is part of the todo's definition of done only when
one of these enables it:

- a project default in `.gavel.yaml` under `checks:` (`enabled: true`),
- the TODO's own `checks:` front matter.

Configure what runs in `.gavel.yaml`:

```yaml
checks:
  enabled: true
  test:                     # omit to skip tests
    changed: true           # only packages affected by the agent's changes
    timeout: 5m
  lint:                     # omit to skip linting
    changed: true           # only new violations vs the base ref
```

When enabled with neither `test` nor `lint` set, both run against changed files. The
re-run budget is `todos.run.workflow.verify.maxIterations` (default 3).
Feedback works best on the cmux runtime (`--model cmux:opus`, which resumes the live agent session); the
inline agent resumes via its session id, and agents that can't resume fall back to
reporting the failures without iterating. See the [`checks`](SCHEMA.md#checks)
schema for every field.

## Pull Requests, Git History, And Repository Mapping

These commands help when the unit of work is no longer "run tests" but "understand what changed", "watch a PR", or "classify a repository".

### `gavel pr status`

Use `pr status` to inspect or follow GitHub Actions state for a pull request. It can resolve the current branch's PR automatically, fall back to the most recent PR, and optionally fetch failed job logs.

This is the fastest command for:

- Checking whether a PR is green
- Following checks until completion
- Pulling failed log tails into the CLI
- Getting the exact command that re-runs the PR's failures locally
- Seeing exactly which files a conflicting PR conflicts on

A PR that GitHub reports as `CONFLICTING` renders a **Merge conflicts** section naming every conflicting path, how it conflicts, and the commands that clear it:

```
✖ Merge conflicts main ← feat/widget (2 files)
├── go.mod (content)
├── docs/api.md (modify/delete)
╰── Resolve locally
    ├── $ git fetch origin main
    ├── $ git switch feat/widget
    ╰── $ git merge origin/main
```

No GitHub API returns the conflicting paths, so gavel replays the merge locally with `git merge-tree` against the same two commits GitHub merges — the PR head and the base branch's *current* tip. It needs a checkout whose remote points at the PR's repository; without one the section says so rather than showing an empty list. Nothing is written to your working tree, index, or refs.

The exit code is `1` whenever the PR cannot merge — a failing check, a failing job, a failing gavel shard, or a merge conflict — and `0` only when it is clear on all four. A merge conflict is the one signal `--actions` and `--comments` never scope away: it blocks the merge whichever checks you asked about. `--fail-fast` is unaffected; it still stops only on a definitive check or job failure, because a conflict is cleared by pushing a merge rather than by waiting.

When the PR published gavel results, each failing shard renders its failing tests and lint violations plus a **Reproduce locally** block:

```
└─ ✖ gavel-test  passed: 41 failed: 2  total: 43
   ├─ Test failures (2)
   ├─ Lint summary: 3 of 47 violations
   ├─ Reproduce locally
   │  $ gavel test --pr 42
   │  $ gavel lint --pr 42
   └─ View full results: https://github.com/owner/repo/actions/runs/…
```

`--pr` downloads every gavel artifact on the PR, merges the shards into `.gavel/pr-<n>.json`, and narrows the run to those failures — the same narrowing `--failed` applies to a local run. It accepts `42`, `#42`, `owner/repo#42`, or a PR URL, and cannot be combined with `--failed` or `--baseline`.

Common workflows:

```bash
gavel pr status
gavel pr status 42
gavel pr status https://github.com/owner/repo/pull/123
gavel pr status --follow --interval 30s
gavel pr status --follow --actions 'CI / Test' --fail-fast   # stop the moment Test goes red
gavel pr status --follow --comments '@coderabbit'            # wait until those threads resolve
gavel pr status --logs --tail-logs 50
gavel test --pr 42        # re-run just the tests that failed on PR #42
gavel lint --pr 42        # re-run just the linters that flagged PR #42
```

`--follow` gates on the dimensions you filtered on, ANDed together: `--actions` waits for
the selected checks to reach `COMPLETED`, `--comments` waits for the selected review threads
to be resolved, and with neither the gate is the whole status rollup. `--fail-fast` returns
at the first definitive failure (`FAILURE`, `TIMED_OUT`, `STARTUP_FAILURE`) rather than
holding the watch open for slower checks — useful when a later scan takes minutes longer
than the check you care about.

A selector that matches nothing is an error under `--follow` too, not a silent success: an
empty filtered set used to satisfy the gate, so a typo — or a poll issued before GitHub had
registered a freshly pushed commit's checks — returned exit 0 immediately.

When stderr is not a terminal, each poll prints a one-line heartbeat rather than repainting
the whole status table; the full report is still printed once, at the end.

#### Fixing a PR with `--ai-fix`

`--ai-fix` hands the status snapshot to an agent through the `pr.fix` prompt. The agent edits files, gavel commits and pushes each turn to the PR's head branch, and the prompt's `workflow.verify.commands` re-poll `gavel pr status <url>` until the PR is green or `--ai-fix-max-iterations` runs out.

A PR that conflicts with its base is merged first. Before the first turn, gavel starts `git merge --no-commit` of the base commit GitHub merges against and leaves the conflicted paths in the tree. The prompt opens with a **Merge conflicts** section listing those paths. The agent resolves the markers and `git add`s each file, and gavel concludes the merge commit and pushes it. Until nothing is left conflicted, two cheap git checks run ahead of the CI poll, so an unresolved file is fed back to the agent by name.

Without `--worktree` the fix runs in your checkout, which must be clean and contain the PR head. `--worktree` moves the run into a fresh git worktree branched from the PR head under the user cache directory (`…/gavel/worktrees/<repo>/pr-<n>-<ts>`), so your branch and uncommitted work are never touched. The worktree is removed when the run ends. It is ordinary captain `setup:` isolation, so the same block can be declared under `pr.fix` in `.gavel.yaml` instead.

```bash
gavel pr status 42 --ai-fix              # fix in the current checkout
gavel pr status 42 --ai-fix --worktree   # fix in an isolated worktree at the PR head
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

### `gavel projects`

`projects` manages the workspace projects the dashboard shows: named local directories, each backing zero or more GitHub repos, stored in `~/.config/gavel/projects.json`. The dashboard's `/api/projects` serves the same store.

```bash
gavel projects list
gavel projects add acme ~/src/acme --repo acme/acme
gavel projects update acme --repo acme/acme --repo acme/acme-ui
gavel projects hide acme
gavel projects unhide acme
gavel projects delete acme
```

#### Hiding a project

A hidden project stays registered but drops out of the dashboard: the project picker, the Projects sidebar, processes, sessions, TODO lists and the git summary all leave it out, and the git state tracker stops scanning its repository. Its stored git rows are kept but no longer read.

- In the dashboard, hover a project in the project picker and click the hide (closed eye) button. Hidden projects are listed under **Manage hidden** in the same picker, where the eye button shows them again.
- From the CLI, `gavel projects hide <name>` and `gavel projects unhide <name>` set the same flag. `gavel projects list` lists every project, hidden ones included, with a `hidden` column.

Hiding from the dashboard stops git tracking immediately. Hiding from the CLI only takes effect for a running dashboard's tracker when it restarts, because the tracker lives in the dashboard process. Editing a project in the dashboard's form keeps its hidden flag.

One exception is intentional: a view that reads stored git data for a specific repository, such as the diff stats of a TODO run whose worktree is in a hidden project's repository, registers that repository with the tracker again so the data it shows is current. Hiding the project again, or restarting the dashboard, stops it.

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
gavel git analyze --ai --ai-model api:haiku
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

- `install`: create the user-level service, launch it through the user's login and interactive shell so `.zshrc` PATH changes apply, and persist daemon config such as database mode and GitHub token
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
gavel pr status --logs
gavel todos create "Fix the failing PR check" --body "Paste the relevant failure details"
gavel todos list
gavel todos run 3f2a1b
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
| Re-run a TODO's definition of done | `gavel todos check` |
| Generate a commit message | `gavel commit` |
| Check PR status | `gavel pr status` |
| Browse many PRs | `gavel pr list` |
| Analyze commit history semantically | `gavel git analyze` |
| Understand file/scope classification | `gavel repomap` |
| Work through native TODOs | `gavel todos` |
| Replay a saved UI snapshot | `gavel ui serve` |
| Run a background PR dashboard | `gavel system` |
| Turn pushes into local CI | `gavel ssh` |
