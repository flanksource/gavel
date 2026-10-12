# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

gavel is **Flanksource's test runner** (module `github.com/flanksource/gavel`): it discovers, runs,
filters, and renders test suites (Go, vitest, Playwright, lint adapters, …) for the terminal, a web
dashboard, and a GitHub Action. For end-user usage — commands, output formats, configuration, filter
sets, the action — read **[README.md](README.md)** and **[MANUAL.md](MANUAL.md)**. This file captures
the architecture and the integration facts a consumer or contributor needs.

## Architecture map

- `testrunner/` — the engine. `runner.go`/`streamer.go` drive runs; `parsers/` defines the result
  model (`Test`, `TestSummary`) and rendering; `runners/` are per-tool adapters; `registry.go`,
  `selector.go`, `outline/`, `history/` handle discovery/replay; `tree_builder.go` rolls child status
  up to parents.
- `testrunner/ui/` — the React/Vite test-runner dashboard, published as `@flanksource/gavel` (subpath
  export `./testrunner`). It produces **two** build outputs: `dist/testui.js` (the `//go:embed`-ed
  single bundle) and `dist/lib/*` (the consumable library: `index.js`, `testrunner.js`, `.d.ts`).
- `models/`, `linters/`, `service/`, `serve/`, `status/`, `cmd/`, `ai/`, `site/` — data model, lint
  adapters, server, status reporting, CLI, AI helpers, and the landing site.

## The result model (read before touching test state)

`parsers.Test` carries independent state flags, **not** a single enum — get the precedence right:

- **Pending** = queued / not yet started → static hollow `codicon:circle-large-outline` (gray).
- **Running** = started, unfinished → blue `svg-spinners:ring-resize` spinner. Distinct from Pending
  (before the flag existed the two states read inverted). `Test.Sum()` and the TS
  `sum()`/`sumNonTaskTests()` count Running separately, and `TestSummary` carries it.
- **Warned** (amber) = a soft failure that **never** flips a parent to Failed (a failed child still
  does). In `Test.Sum()` the order is Failed → Warned → Skipped; `tree_builder.go`
  `propagateFailureStatusRecursive` raises `Warned` only under `!Failed`, so amber never masks red. TS
  mirrors this: `statusIcon`/`statusColor`/`testStatus` put the `warned` branch *after* `failed`.
- Prefer the `StatusCounts` helpers (`emptyCounts`/`addCounts`/`countsFromSummary`/`countsFromLeaf`)
  in `testrunner/ui/src/utils.ts` over re-inlining count literals.

## Embedded UI & consuming gavel (from project memory)

- **Both embedded bundles are committed despite being gitignored.** `testrunner/ui/dist/testui.js`
  and `pr/ui/dist/prui.js` are `//go:embed`-ed; since c7f713a1 (first tag v0.0.38) they are tracked so
  a module fetched by tag/SHA satisfies the embed, and `release.yml` rebuilds and `git add -f`s them
  before tagging. `.gitignore` still lists them but has no effect on tracked files, so every local
  `make build` (Taskfile `pr-ui`/`test-ui` regenerate them) shows them modified — that diff is
  build noise: leave it out of feature commits and let release CI refresh it. Refs older than v0.0.38
  lack the bundles and fail with `pattern dist/testui.js: no matching files found`.
- **The frontend has two builds — `build` ≠ `build:lib`.** `pnpm run build` emits only the embed
  bundle (`dist/testui.js`); the consumable library (`dist/lib/testrunner.{js,cjs,d.ts}`, used via the
  `./testrunner` subpath) comes from `pnpm run build:lib` (`vite -c vite.lib.config.ts`, ESM +
  `--mode cjs`). A consumer that links `testrunner/ui` via a `file:` dep and runs its own
  `pnpm install` re-copies the package from *current source state* — if `dist/lib/` is missing the
  store copy ends up dist-less and the consumer's `tsc -b` fails with `Cannot find module
  '@flanksource/gavel/testrunner'`. Always `build:lib` before reinstalling downstream.
- **Multi-run servers compose; `Server` stays single-run.** Concurrent runs are served by
  `testrunner/ui/multiserver.go` `MultiServer`, which holds one ordinary `*Server` per runID:
  `BeginRun(runID, kind) *Server` returns it and `Handler()` strips the leading path segment as the
  runID before delegating to `srv.Handler()`. `Server.Done()` lets a MultiServer evict only finished
  runs (cap + idle-TTL, done-runs only, never a live run). The React `useTestRun` hook takes an
  optional `runId` that scopes every request URL to `${base}/${runId}/api/...`. Keep `Server`
  single-run — don't make it run-keyed.
