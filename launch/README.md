# Structured test and lint launch

Implementation status: implemented across Gavel, Captain, commons-db and Clicky UI in the local linked workspaces. This document records the public contract and runnable verification.

`POST /api/launch` starts test, lint, or both. `/api/launch/test` requires only `config.test`; `/api/launch/lint` requires only `config.lint`. `GET` on each route with a bare `?schema` returns the JSON Schema used to validate that route. Unknown fields and sibling sections are rejected. Native command schemas own test and linter options.

```json
{
  "workspace": { "cwd": "/enabled/workspace" },
  "config": {
    "test": { "paths": ["./..."] },
    "lint": {
      "files": ["src"],
      "linters": [
        { "id": "typescript", "linter": "oxlint", "options": {} },
        { "id": "design", "linter": "prompt", "options": {
          "prompt": "diagram-review", "promptVersion": "discovered-version",
          "variables": {}, "presets": ["lint"]
        } }
      ]
    }
  },
  "labels": { "page": "example", "reviewId": "example-design" }
}
```

Use an `Idempotency-Key` header for admission. Identical root and child requests reuse the execution ID; changed input with the same workspace and key returns 409. The service persists admission before returning 202 and dispatches independently of the HTTP connection. Lint runs before tests, and lint findings do not prevent tests from running. Stop cancels the exact execution and its native children. Recovery checks the recorded dispatcher before marking an abandoned run interrupted; it never replays it.

Workspace input is the shared `shell.Setup`. Local directory access is disabled unless enabled by server configuration and restricted to canonical allowed roots. Git launches create private worktrees pinned to a resolved commit. `workspace.patch` applies to a new worktree after a successful `git apply --check`; the source checkout remains unchanged. `POST /api/launch/prepare` returns a caller-scoped preparation handle, expiring after 15 idle minutes. Replace `workspace` with `preparationId` when launching after discovery. Active use leases the preparation until completion. Paths refer to the execution host.

`GET /api/launch/prompts?preparationId=...` discovers Captain prompts in the prepared workspace. Add `prompt=...` to retrieve a prompt, version, input schema and defaults. Prompt lint freezes the prompt and runtime before admission and executes through Captain's native prompt-run lifecycle. Input errors are 422 and version conflicts are 409. Lookup changes are excluded.

`GET /api/runs` lists executions with `limit`, `before` and `label.<key>` filters. `GET /api/runs/{executionId}` reads an exact snapshot; `/stream` follows revisioned SSE snapshots, including a full snapshot on reconnect; `POST /stop` cancels it. Terminal snapshots are immutable. `resultstore` owns common reader/writer contracts and typed test/lint views. JSON snapshots are the default adapter; SQLite uses commons-db recordstore ownership and schema. Backend switches require explicit import, with no dual write or fallback.

TODO orchestration remains native: the TODO fixture, `gavel todos check`, and `/api/todos/run` verification step own verdicts, target commit and lifecycle. Launch evidence must link to the exact attempt and iteration without replacing VerifyReport or nesting a parent scheduling lock.

Playground work persists run association and idempotency identity before submission, reconciles lost responses and restarts, and creates at most one result comment per run and review target. Clean and failed runs retain their associations. Public provenance excludes raw variables, rendered prompts, argv and setup secrets; only the patch digest is public.

Rejecting a design adds a `rejected` class in its source. Approval or revision removes it. Source and review-decision persistence use optimistic source checks and coordinated rollback. Rejected content, pins and controls are hidden by default, with a URL-backed reveal control. Existing classes and mapped review IDs are preserved.

Delivery checklist:

- [x] Shared workspace patch field and focused isolation/conflict tests.
- [x] Captain workspace-scoped discovery and frozen prompt preparation seam.
- [x] Common JSON and recordstore adapter contracts and round-trip checks.
- [x] Launch admission, recovery, cancellation, caller isolation and HTTP integration tests.
- [x] Dashboard and CLI wiring, configuration, OpenAPI and runnable examples.
- [x] Durable playground prompt association and structured lint result comments.
- [x] Source rejection mutation, reveal control and hidden-anchor handling.
- [x] Focused tests, builds and live desktop/compact visibility checks. Broad lint/type-check limitations are recorded below.

From a trusted local workspace, start the API with local access explicitly enabled:

```sh
gavel launch serve --root . --local-root . --addr 127.0.0.1:9330
curl --fail 'http://127.0.0.1:9330/api/launch?schema'
curl --fail 'http://127.0.0.1:9330/api/launch/lint?schema'
curl --fail -H 'Content-Type: application/json' -H 'Idempotency-Key: example-test-1' --data '{"workspace":{"cwd":"."},"config":{"test":{"paths":["./launch"],"framework":["ginkgo"]}}}' http://127.0.0.1:9330/api/launch/test
```

The last command assumes the Gavel repository is the enabled workspace. Follow the returned `snapshotUrl`, `streamUrl`, or `stopUrl` using that execution's exact ID. `GET /api/launch/openapi` returns the OpenAPI document. The PR dashboard exposes the same routes; configure it with `--launch-root`, `--launch-local-root` and `--launch-store`.

From the Clicky UI repository, use `gavel launch serve --root . --local-root .` for playground reviews. The playground connects to `http://localhost:9330` by default; `PLAYGROUND_GAVEL_URL` selects another server. Discovery returns the prompt version and a preparation handle used by the form submission. Reload discovery after a version conflict or expired preparation. Lookup extensions remain out of scope.

To switch adapters, stop the launch server, import its terminal executions, then restart with the selected adapter:

```sh
gavel launch import --root . --from json --to sqlite
gavel launch serve --root . --local-root . --store sqlite
```

Imports preserve IDs, revisions, native results, labels and dispatcher/admission identity. Repeating an import is idempotent; conflicting snapshots and active source runs fail. Keep the storage root and its private `.launch-key` together.

The executable definition of done is [verification.md](verification.md), also attached to TODO `31bf3fff-ecb6-418b-aecc-7a8f5cb9d60e`. It uses explicit local checkout paths to verify the linked dependency changes together. Run `gavel fixtures launch/verification.md` from Gavel, or `gavel todos check 31bf3fff` from Clicky UI. TODO verification continues to produce the native VerifyReport; these launch records do not replace its verdict or lifecycle.

Validation evidence includes JSON/SQLite contract and import tests, race-enabled launch lifecycle tests, native workspace patch and frozen prompt tests, 72 JavaScript helper/API/component checks, a native Ginkgo HTTP launch, root/child replay, 422/409 responses, terminal SSE, prompt form discovery, adapter import/reopening, and desktop/compact rejection visibility. Prompt preflight was verified without running a model-backed lint. The canonical TODO check passed all five definition-of-done checks and marked the item verified on 2026-10-27.

Broad gates have existing unrelated failures: Gavel's full linter reports the unused `testrunner/ui/traces.go` helper; Captain lint hits duplicate `main` functions in existing `hack/` files; Clicky UI lint reports unsafe optional chaining in `src/review/export.test.ts`. The full playground TypeScript check also fails on unrelated demo/sibling API drift. Scoped Go and TypeScript lint checks pass. Gavel, Captain and the shared Clicky UI build pass with the linked workspaces. Published dependency versions have not been released or verified in isolation.
