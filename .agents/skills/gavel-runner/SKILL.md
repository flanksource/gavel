---
name: gavel-runner
description: Run gavel test and lint, focus on a subset, re-run only failures (defaults to .gavel/last.json), filter noise with baselines, and pull JSON/markdown results from finished or live runs
allowed-tools: [Bash, Read, Glob, Grep, WebFetch]
---

# Gavel Runner

Drive `gavel` for the everyday test + lint loop: run a subset, iterate on failures, suppress known noise with baselines, pull JSON/markdown out of a finished run, and attach to a live run via the UI server's HTTP/SSE API.

For *writing* fixture-style markdown tests, defer to the [`gavel-fixture-tester`](../gavel-fixture-tester/SKILL.md) skill instead — that skill is about authoring; this one is about running.

## When to use

- The user says "run gavel", "use gavel to lint", "rerun the failures", "what failed last time", "watch tests", "give me a JSON of results".
- The repo already has gavel installed and configured (a `.gavel.yaml`, a `.gavel/` directory, or a `make` target invokes `gavel`).

## Quick reference

| Goal | Command |
| --- | --- |
| All tests + lint | `gavel test --lint` |
| Only changed packages | `gavel test --changed` |
| Only previously-failed | `gavel test --failed`  *(uses `.gavel/last.json`)* |
| Suppress baseline noise | `gavel test --baseline baseline.json` |
| Save JSON + markdown | `gavel test --format "json=.tmp/g.json,markdown=.tmp/g.md"` |
| Browser UI + live SSE | `gavel test --ui` |
| PR-comment summary | `gavel summary --input results.json --output summary.md` |
| Lint only | `gavel lint` |

## Running a subset of tests

Three orthogonal mechanisms — pick the one that matches the user's intent:

**By path** — positional args. Recursive by default:
```bash
gavel test ./pkg/foo ./pkg/bar
```

**By change** — diff-based filtering:
```bash
gavel test --changed                    # vs origin/main
gavel test --since v1.2.0               # since merge-base with this ref
```

**By cache** — skip packages whose content fingerprint matches the last passing run:
```bash
gavel test --cache
```

**By framework** — restrict execution to one or more frameworks:
```bash
gavel test --framework go,ginkgo
```

**By name** — pass through to the underlying runner with `--`:
```bash
gavel test ./pkg/foo -- --focus "MyTest"             # Ginkgo focus
gavel test ./pkg/web -- --testNamePattern "MyTest"   # Vitest pattern
```

## Re-running only failures

`--failed` re-runs only the tests / linters that failed in a previous run. Pass it without a value to default to `.gavel/last.json`:

```bash
gavel test --failed                          # uses .gavel/last.json
gavel test --failed path/to/results.json     # explicit
gavel lint --failed                          # same on lint
```

If `.gavel/last.json` does not exist, gavel errors loudly — run `gavel test` once first. Don't fall back silently to a full run.

## Baselines (suppress known noise)

`--baseline <snapshot.json>` reports only what is *new* vs the baseline. Lint baseline keys are line-insensitive (`linter+file+rule`) so they survive line-shifting edits.

```bash
# Capture on main:
git checkout main && gavel test --lint --format "json=baseline.json"

# Run on the branch with main's noise suppressed:
git checkout my-branch && gavel test --lint --baseline baseline.json
```

## The `.gavel/` directory

Gavel persists every run here so the next invocation can compare or re-run.

| File | Shape | Purpose |
| --- | --- | --- |
| `.gavel/sha-<commitsha>[-<uncommitted-hash>].json` | Full snapshot | One per run; the actual data |
| `.gavel/last.json` | Pointer (`{path, sha, uncommitted}`) | Most recent run |
| `.gavel/<branch>.json` | Pointer | Most recent run per branch |
| `.gavel/<linter-binary>` | Binary | Auto-installed linters (e.g. `golangci-lint`) when config calls for them |
| `.gavel/ssh_host_key`, `.gavel/repos/` | — | Only present when `gavel ssh serve` has been used |

The pointer files are *pointers*, not snapshots — they reference a path inside `.gavel/`. To read the snapshot you have to follow the `path` field.

## Result schema (source of truth)

Field names below come straight from the Go structs. When in doubt, read the source — these are the canonical definitions:

- **`Test`** record — `testrunner/parsers/types.go` (`type Test struct`). Fields: `name`, `package`, `package_path`, `framework`, `failed`, `passed`, `skipped`, `duration`, `message`, `file`, `line`, `cached`, `timed_out`, …
- **`Snapshot`** envelope — `testrunner/ui/snapshot.go` (`type Snapshot struct`). Holds the flat `tests` array plus lint results and bench comparison.
- **Lint** `LinterResult` and `Violation` — `linters/` package. Keyed by `linter`, `file`, `rule`.
- **`Pointer`** — `snapshots/snapshots.go` (`type Pointer struct`): `{path, sha, uncommitted}`.

## `jq` recipes against `results.json`

```bash
# All failed tests, grouped by package:
jq -r '.tests[] | select(.failed) | "\(.package_path)\t\(.name)"' results.json | sort | uniq -c

# Counts by framework:
jq -r '.tests[] | select(.failed) | .framework' results.json | sort | uniq -c

# Slowest 10 tests:
jq '[.tests[] | {name, package_path, duration}] | sort_by(-.duration) | .[:10]' results.json

# Failure messages, package + first line of message:
jq -r '.tests[] | select(.failed) | "\(.package_path) :: \(.name)\n  \(.message | split("\n")[0])"' results.json

# Lint violation counts per linter:
jq -r '.lint_results[]? | "\(.linter)\t\(.violations | length)"' results.json

# CI gate: exit non-zero if anything failed.
test "$(jq '[.tests[] | select(.failed)] | length' results.json)" = 0
```

## Reading `.gavel/last.json` (two-step)

```bash
# Resolve the pointer, then jq the snapshot it points at:
SNAP="$(jq -r '.path' .gavel/last.json)"
jq '.tests[] | select(.failed) | .name' "$SNAP"
```

## Live runs via the UI server (HTTP + SSE)

There is no `gavel watch`. The live workflow is `gavel test --ui`, which starts an HTTP server with a JSON snapshot endpoint and an SSE stream. The port is printed on startup.

Endpoints (from `testrunner/ui/handler.go`):

| Method | Path | Use |
| --- | --- | --- |
| GET | `/api/tests` | Current snapshot as JSON (same shape as `--format json`) |
| GET | `/api/tests/stream` | SSE stream of live test updates |
| GET | `/api/diagnostics` | Timeout artifacts + resource stats |
| GET | `/api/diagnostics/collect` | Trigger an on-demand diagnostics capture |
| GET | `/api/process/metrics` | Process CPU/memory metrics |
| POST | `/api/rerun` | Trigger a rerun with filters (test names / frameworks) |
| GET | `/api/rerun/stream` | SSE stream of rerun updates |
| POST | `/api/stop` | Stop the in-flight run |
| GET | `/api/benchmarks` | Bench comparison JSON |
| GET | `/?format=json\|html\|markdown` | Export the current report |

Recipes:
```bash
# Failed tests in a live run:
curl -s http://localhost:$PORT/api/tests | jq '.tests[] | select(.failed) | {pkg: .package_path, name}'

# Stream live updates:
curl -Ns http://localhost:$PORT/api/tests/stream | grep --line-buffered '^data:'

# Trigger a rerun of two specific tests:
curl -X POST http://localhost:$PORT/api/rerun \
     -H 'Content-Type: application/json' \
     -d '{"tests":["TestFoo","TestBar"]}'
```

Replay a finished run (no live execution): `gavel ui serve run.json` exposes the same endpoints against a saved snapshot.

## Markdown / HTML / JSON output

Multiple formats in one invocation, one path per format:

```bash
gavel test --lint --format "json=.tmp/g.json,html=.tmp/g.html,markdown=.tmp/g.md"
```

For a compact PR-comment-shaped markdown (counts + first few failures only):
```bash
gavel summary --input .tmp/g.json --output .tmp/summary.md
```

Use the JSON for full failure detail; the markdown is sized for PR comments.

## Timeouts (call out — silent kills are the #1 confusion)

Gavel layers four wall-clock deadlines on `gavel test`. Each kills its subprocess when it fires (no graceful drain).

| Flag | Default | Scope |
| --- | --- | --- |
| `--timeout` | `10m` | Global, for the entire `gavel test` run (test + lint). On hit, diagnostics are captured and **every** subprocess is killed. |
| `--test-timeout` | `5m` | Per-test-package subprocess (one go-test/ginkgo/vitest invocation). |
| `--lint-timeout` | `5m` | Per-linter subprocess when `--lint` is set. |
| `gavel lint --timeout` | `5m` | Per-linter when running `gavel lint` standalone. Duration string (`30s`, `2m`, …). |

Two more deadlines apply only to the detached UI (`gavel test --ui --detach`):

- `--auto-stop` (default `30m`) — hard wall-clock for the detached UI child.
- `--idle-timeout` (default `5m`) — exit the detached UI after this long with no HTTP requests.

Tuning recipes:
```bash
# Long integration suite — bump global and per-test caps:
gavel test --timeout 30m --test-timeout 15m ./pkg/integration

# Stricter per-linter cap:
gavel lint --timeout 2m
```

When a global timeout fires, gavel dumps goroutines and a resource snapshot into `.gavel/`. Read those before bumping the timeout — the test that ran out of time is usually the bug, not the deadline.

## Common recipes

```bash
# Run everything once, save artifacts:
gavel test --lint --format "json=.tmp/gavel.json,markdown=.tmp/gavel.md"

# Iterate on failures:
gavel test --failed

# Only NEW lint violations on this branch:
gavel lint --baseline .gavel/<main-pointer>.json

# Trigger a rerun in a live UI session:
curl -X POST http://localhost:$PORT/api/rerun -d '{"tests":["TestFoo"]}'

# CI gate on failed-test count:
jq '[.tests[] | select(.failed)] | length' results.json
```

## Gotchas

- Use `.tmp/` (not `/tmp`) for scratch artifacts in this repo.
- The markdown summary is sized for PR comments; use the JSON for full failure detail.
- There is no `gavel watch`. Reach for `--ui` (live SSE) or `--changed` (rerun-on-edit via your editor) instead.
- `--failed` errors loudly when `.gavel/last.json` is missing — don't fall back silently to a full run; either pass an explicit baseline or run `gavel test` once first.
- `.gavel/last.json` is a *pointer*, not the snapshot. Follow `.path` before parsing test data.
- Lint baseline keys are line-insensitive — `linter+file+rule` only — so they survive line-shifting edits but won't distinguish two violations of the same rule in the same file.
