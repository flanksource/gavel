# Gavel multi-tenancy: implementation spec

> **First action after approval:** copy this file verbatim to `docs/plans/multi-tenancy.md` in the
> gavel repo. That copy is the hand-off document for the implementing session.

---

## 0. How to use this document (read first, implementing session)

This spec turns gavel into a hosted, multi-tenant service. It is written so that a session that was
not part of the design discussion can implement it milestone by milestone. Every design decision
below **has already been made with Batman**. Do not re-open them. If something in the code
contradicts this spec, **stop and ask Batman**; don't improvise.

**Working rules.** These come from `~/.claude/CLAUDE.md` and apply throughout:

- Address the user as **Batman**.
- **TDD for every change.**
  1. Write the failing test.
  2. Run it and see it fail.
  3. Write the minimal code.
  4. Run it and see it pass.
- **Ask Batman before running any build or test command**, every time.
- **Never commit, push, amend or rebase** unless Batman explicitly asks for that specific operation.
- Every new code file starts with a **2-line comment** saying what the file does.
- Names and comments describe **what the code does**. No "new", "legacy", "improved", "wrapper",
  "unified", and no implementation details in names.
- Make the smallest reasonable change. Match the surrounding style. No backward-compatibility shims
  without approval.
- Never delete a failing test. Raise it with Batman instead.
- Tests use real logic, never mocked behaviour. E2E tests use real services: embedded Postgres, a
  real `git`, real docker, a kind cluster.

**Milestones.** Each milestone is independently reviewable and leaves `--auth none` (today's
behaviour) working. Do them **in order**: M0 → M1 → … → M10. Each one lists:

- **Goal**
- **Files**
- **Exact changes**
- **Tests (write first)**
- **Acceptance**
- **Stop-and-ask points**

**Spikes.** Some milestones start with a spike: a short investigation whose result you must
**report to Batman before continuing**. Don't skip them. Write the result into
§13 "Spike log" of `docs/plans/multi-tenancy.md`.

**Progress tracking.** Keep the §14 checklist at the bottom of `docs/plans/multi-tenancy.md` up to
date as you finish items.

---

## 1. Goals and decisions (final)

| # | Decision |
|---|---|
| D1 | **Fully hosted multi-tenant mode** behind `gavel serve --auth clerk`. `--auth none` is the default and is exactly today's behaviour. |
| D2 | **One Postgres for all orgs.** |
| D3 | **Routes are org-neutral.** No org in the URL. The active org comes from the authenticated session. |
| D4 | **Browser login via Clerk.** The UI shows the logged-in user and has an **org switcher** (Clerk `OrganizationSwitcher`). |
| D5 | **`org_id uuid NOT NULL` on every tenant-owned table**, with a built-in **local org** (fixed UUID) that `--auth none` uses. `org_id` is part of every PK or unique key that previously identified a row by repo, path or URL. `user_id` stays nullable (attribution only). |
| D6 | **`git push gavel`** is authenticated with **registered SSH keys** (key → user). Repos are addressed as `<org-slug>/<repo>`. Only members of that org may push. |
| D7 | **GitHub access in hosted mode is a GitHub App.** Users log in via the GitHub App (Clerk's GitHub connection configured with the App's OAuth credentials), install the App on their GitHub org/account and add repos to it. Serve calls GitHub with **installation tokens**. **GitHub Actions → gavel** uploads (test results etc.) authenticate with a **GitHub Actions OIDC token**. |
| D8 | **Blob store** (local FS, S3 or GCS via commons-db) for: todo attachments, test/lint run history and snapshots, fixture recordings, attempt transcripts, and job logs (moved out of Postgres `bytea`). |
| D9 | **Retention** is property-controlled and defaults to **60 days**, with a periodic **cleanup job**. |
| D10 | **commons-db gaps are fixed upstream** in `flanksource/commons-db`, and gavel bumps the dependency. No workarounds in gavel. |
| D11 | **In `--auth clerk` everything that executes code is sandboxed**: gavel test, lint and commit, push checks, and Captain agent runs. `--sandbox kubernetes` is the default and runs **Kubernetes Jobs**; `--sandbox docker` and `--sandbox podman` are options. |
| D12 | **Checkouts** for sandboxed runs come from **pushed repos** or **GitHub (App installation)**. Projects pointing at host directories are rejected in clerk mode. |
| D13 | **Sandbox network:** public egress is allowed; internal networks are blocked (private ranges, metadata IP, cluster CIDRs, serve host). The only exception is the serve run API. |
| D14 | **Agent runs execute entirely inside the pod.** The pod has **no DB and no GitHub credentials**. It talks back to serve through a **run-scoped token API**. Serve persists everything. |
| D15 | **agent-substrate is out of scope.** The `sandbox.Executor` interface is where it can plug in later. |

---

## 2. Glossary

- **Org.** A row in `orgs`. In clerk mode it maps 1:1 to a Clerk organization. The **local org** is
  `00000000-0000-0000-0000-000000000001` (`issuer='local'`, `clerk_org_id='local'`,
  `slug='local'`, `name='Local'`).
- **Scope.** `tenant.Scope{OrgID, UserID, Role, GitHubLogin}` for the current request or run.
- **Auth mode.** `none` or `clerk`, selected by `gavel serve --auth`.
- **Org server.** One `*ui.Server` per org in clerk mode, composed by `ui.OrgServers`. This mirrors
  the existing `testrunner/ui/multiserver.go` pattern.
- **Installation.** A GitHub App installation linked to exactly one gavel org.
- **Sandbox run.** A row in `sandbox_runs`. It executes in a K8s Job or a container and talks to
  serve with a **run token**.
- **Run token.** An HMAC-signed `{run_id, org_id, exp}`, valid only on `/api/runs/{id}/*`.

---

## 3. What already exists (uncommitted working tree on branch `api-auth`)

- **`api/`:**
  - Clerk device-flow client: `NewAuthClient`, `BeginLogin`, `CompleteLogin`, `WhoAmI`, and the
    unexported `verifyIdentity`, which fetches JWKS on every call with no cache.
  - Credentials file: `api/credentials.go`, `CredentialPath` →
    `os.UserConfigDir()/gavel/clerk-session.json` (mode 0600).
  - `IdentityStore.Sync(ctx, Identity) (Identity, Ownership, error)`, which upserts `users` and
    `orgs`.
  - `ClerkOrganizations{SecretKey}`, a backend org resolver.
- **`cmd/gavel/auth.go`:** `login`, `whoami` and `logout`, built as plain cobra commands and
  MCP-excluded. Flags `--issuer`, `--client-id`, `--no-browser`. Env `GAVEL_CLERK_ISSUER`,
  `GAVEL_CLERK_CLIENT_ID`.
- **Schema:**
  - `identity.hcl`: `users`, `orgs`, and the sidecars `session_owners`, `api_token_owners`,
    `runtime_profile_owners`, `runtime_preset_owners`.
  - `git_agent_ownership.hcl`.
  - The sidecars have no FKs onto Captain's tables, because gavel installs no DDL on them. Every
    migration deletes rows whose Captain resource is gone (`sweepOwnershipOrphans` in
    `database.go`).
  - **Nullable** `user_id`/`org_id` columns plus `<t>_owner_idx` on 14 root tables, and attribution
    columns on the todo tables. M1 changes these to NOT NULL.
- **Docs:** `api/README.md`, `api/OWNERSHIP.md`, and a README section.

Known problems you will fix along the way, each in the milestone named:

1. `printAuthIdentity` discards `Ownership` and runs migrations on every `whoami` (M0).
2. Cache upserts use `UpdateAll`, which would overwrite owner columns (M1).
3. Keys don't include org, so orgs collide (M1).
4. `serve/server.go` accepts every SSH key, and `cleanRepoPath` allows `..` traversal (M6). The
   e2e test writes the developer's real `~/.gavel/ssh_host_key` (M6).
5. The post-receive hook `eval`s `ssh.cmd` from the pushed `.gavel.yaml` on the host (M9).
6. `pr/ui` has no auth, binds `0.0.0.0`, and serves attachments with `Access-Control-Allow-Origin: *`
   and `Cache-Control: public` (M3).
7. `validateTodoRunSpec` doesn't strip client-sent `setup`, `sandbox` or `cliArgs` (M3).
8. `pr/ui/dist/prui.js`, `testrunner/ui/dist/testui.js`, both `pnpm-lock.yaml` files and the
   `pnpm-workspace.yaml` catalog additions (ajv, jsonc-parser) are **rebuild noise unrelated to
   auth** (M0).

---

## 4. Architecture in clerk mode (target)

```
                 Browser (Clerk JS, __session cookie)          gavel CLI (Clerk OAuth access token)
                              │                                          │
                              ▼                                          ▼
 ┌──────────────────────────── gavel serve --auth clerk --sandbox kubernetes ─────────────────────┐
 │ authMiddleware ─► tenant.Scope{org,user} ─► OrgServers[org] ─► *ui.Server (one per org)        │
 │   │                                            │  PR poller (GitHub App installation token)    │
 │   │ public: /, /_assets, /api/auth/config      │  todos (runtime.Provider ForOrg)              │
 │   │ OIDC:   /api/ci/*   (GitHub Actions)       │  project actions ─► sandbox.Executor          │
 │   │ run token: /api/runs/{id}/*  (pods)        │                                               │
 │   │ webhook-less GitHub App setup: /api/github/*                                               │
 │ SSH push server (in-process, --ssh-addr) ─► key→user, org membership ─► bare repo ─► Executor  │
 │ retention ticker ─► DB rows + blob store                                                       │
 └──────────────┬───────────────────────────────┬───────────────────────────────┬─────────────────┘
                │                               │                               │
            Postgres (all orgs)        Blob store (S3/GCS/local)      K8s Jobs / docker / podman
                                                                    (gavel runner exec, no DB/GH creds)
```

---

## M0. Housekeeping (small, do first)

**Goal.** Get a clean base for the auth work.

1. ~~Stop and ask Batman whether to drop or split the unrelated working-tree changes listed in §3
   item 8.~~ **Decided (2026-10-13):** keep them in the working tree. The `pnpm-workspace.yaml`
   catalog entries (`ajv`, `jsonc-parser`) are required: the linked sibling `clicky-ui` checkout
   declares them as `catalog:` specifiers, so the local install fails without them. The lockfiles
   and `dist/` bundles are regenerated artefacts of that install and build. Rebuild `dist/` from the
   final source before each PR rather than hand-picking these diffs.
2. Change `cmd/gavel/auth.go` `printAuthIdentity` to take a `migrate bool` parameter.
   - `login` passes `true`.
   - `whoami` passes `false`, and calls `database.Open(ctx)` without `WithMigrations()`.
   - If the DB isn't migrated (the `users` table is missing), `whoami` skips Sync and prints the
     identity only. Detect the missing table with `to_regclass('public.users') IS NULL`.
3. Write the test first: `cmd/gavel/auth_commands_test.go` asserts that `whoami` never requests
   migrations. Inject a package-level `var openAuthDatabase = database.Open` and assert on the
   options it receives.

**Acceptance.** The `whoami` path doesn't call `WithMigrations()`, and the existing auth tests pass.

---

## M1. Tenant core and schema

**Goal.**
- Every tenant-owned row has `org_id NOT NULL`, defaulting to the local org.
- Stores are org-scoped through a `ForOrg` handle.
- `--auth none` behaves exactly as today.

### M1.1 Package `tenant/` (new)

File `tenant/tenant.go`:

```go
// Package tenant carries the organization and user a request or run acts for.
// The local organization owns all data when gavel runs without authentication.
package tenant

var LocalOrgID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type Scope struct {
    OrgID       uuid.UUID
    UserID      *uuid.UUID // nil for local mode and for system jobs
    Role        string     // Clerk org role, e.g. "org:admin", "org:member"; "" in local mode
    GitHubLogin string     // viewer login for the @me PR filter; "" when unknown
}

func Local() Scope                                     // {OrgID: LocalOrgID}
func WithScope(ctx context.Context, s Scope) context.Context
func FromContext(ctx context.Context) Scope            // Local() when absent
func (s Scope) IsLocal() bool
```

Tests in `tenant/tenant_test.go`:
- `FromContext` with no scope returns `Local()`.
- A round trip through `WithScope`/`FromContext` returns the same scope.

### M1.2 Schema

**Rule for every tenant-owned table:**

```hcl
column "org_id" {
  null    = false
  type    = uuid
  default = sql("'00000000-0000-0000-0000-000000000001'::uuid")
}
```

- **Remove the `org_id` FK blocks from the HCL.** They move to SQL (see below). The reason:
  - Atlas adds the column and the FK in one `ALTER TABLE`.
  - The FK would fail on existing rows, because the local org row doesn't exist until the post
    phase.
- **Keep the `user_id` FKs in the HCL.** The column is nullable, so existing rows pass.
- Keep the `<t>_owner_idx` index.

Tables that get `org_id NOT NULL` (change the existing nullable column):

- `todo_workspaces`, `todo_labels`, `session_owners`, `api_token_owners`, `runtime_profile_owners`,
  `runtime_preset_owners`, `git_agent_task_owners`
- `http_cache_entries`, `workflow_run_caches`, `job_log_caches`, `workflow_def_caches`, `seen_prs`,
  `test_run_caches`, `test_run_cursors`
- `git_repos` (unique key becomes `git_repos_org_root_dir_key (org_id, root_dir)`; worktrees,
  branches and range stats inherit through `repo_id`)
- `file_scans`, `linter_executions`, `debounce_metadata`, `task_run_history`
- Add a **new** `org_id` to `violations` (same block). It is needed for the composite FK.

Unchanged:
- `favicon_caches` stays shared, with no org.
- `todo_issues`, `todo_issue_*`, `todo_workspace_paths` and `todo_issue_aliases` inherit their org
  through `workspace_id`. Don't add `org_id` to them.

**PK and unique-key changes** (edit the HCL):

| Table | New key |
|---|---|
| `http_cache_entries` | PK `(org_id, url, method)` |
| `workflow_run_caches` | PK `(org_id, run_id)` |
| `job_log_caches` | PK `(org_id, job_id)` |
| `workflow_def_caches` | PK `(org_id, repo, workflow_id, sha)` |
| `seen_prs` | PK `(org_id, repo, number)` |
| `test_run_caches` | PK `(org_id, workspace_dir, run_id)` |
| `test_run_cursors` | PK `(org_id, workspace_dir)` |
| `file_scans` | PK `(org_id, file_path)` |
| `violations` | FK `violations_file_scan_fkey (org_id, file_path)` → `file_scans(org_id, file_path)` `ON DELETE CASCADE` (replaces `violations_file_path_fkey`) |
| `debounce_metadata` | PK `(org_id, linter_name, work_dir)` |
| `linter_executions` | unique `linter_executions_unique` → `(org_id, linter_name, work_dir, executed_at)` |
| `todo_workspaces` | unique `todo_workspaces_repo_key_key` → `(org_id, repo_key)` |
| `todo_labels` | `todo_labels_global_name_key` → `(org_id, name) WHERE workspace_id IS NULL`. Keep `todo_labels_workspace_name_key (workspace_id, name)`. |

`task_run_history` keeps PK `id`. Run IDs are UUID-like and globally unique; filter reads by
`org_id`.

**New tables.** Add these in a new file `internal/database/schema/tenancy.hcl`. Use the same column
style as `identity.hcl`:

| Table | Columns |
|---|---|
| `org_memberships` | `user_id uuid NOT NULL` FK users ON DELETE CASCADE; `org_id uuid NOT NULL` FK orgs ON DELETE CASCADE; `role text NOT NULL`; `synced_at timestamptz NOT NULL DEFAULT now()`. PK `(user_id, org_id)`. Index `(org_id)`. |
| `user_ssh_keys` | `id uuid PK DEFAULT gen_random_uuid()`; `user_id uuid NOT NULL` FK users ON DELETE CASCADE; `name text NOT NULL DEFAULT ''`; `fingerprint text NOT NULL` (unique, SHA256 form as from `gossh.FingerprintSHA256`); `public_key text NOT NULL` (authorized_keys line); `created_at timestamptz NOT NULL DEFAULT now()`; `last_used_at timestamptz NULL` |
| `repos` | `id uuid PK DEFAULT gen_random_uuid()`; `org_id uuid NOT NULL` FK orgs RESTRICT; `name text NOT NULL` (`<repo>` for pushed, `<owner>/<repo>` for GitHub); `source text NOT NULL CHECK (source IN ('push','github'))`; `github_repo_id bigint NULL`; `installation_id bigint NULL` FK github_installations ON DELETE SET NULL; `default_branch text NOT NULL DEFAULT 'main'`; `created_at`; `updated_at`. Unique `(org_id, source, name)`. Unique `(github_repo_id) WHERE github_repo_id IS NOT NULL`. |
| `github_installations` | `id bigint PK` (GitHub installation id); `org_id uuid NOT NULL` FK orgs RESTRICT; `account_login text NOT NULL`; `account_type text NOT NULL`; `created_by_user_id uuid NULL` FK users SET NULL; `created_at`; `suspended_at timestamptz NULL` |
| `org_settings` | `org_id uuid PK` FK orgs CASCADE; `search_config jsonb NOT NULL DEFAULT '{}'` (the `ui.SearchConfig` JSON); `updated_at`. Replaces `~/.config/gavel/ui.settings.json` in clerk mode. |
| `sandbox_runs` | `id uuid PK`; `org_id uuid NOT NULL` FK orgs RESTRICT; `user_id uuid NULL` FK users SET NULL; `repo_id uuid NULL` FK repos SET NULL; `kind text NOT NULL CHECK (kind IN ('test','lint','push-check','agent'))`; `ref text NOT NULL DEFAULT ''`; `sha text NOT NULL DEFAULT ''`; `todo_id uuid NULL`; `status text NOT NULL CHECK (status IN ('queued','running','succeeded','failed','cancelled'))`; `exit_code int NULL`; `backend text NOT NULL`; `backend_ref text NOT NULL DEFAULT ''` (Job name or container id); `log_key text NOT NULL DEFAULT ''` (blob key); `created_at`; `started_at NULL`; `finished_at NULL`. Index `(org_id, created_at DESC)`. |

**Migration scripts** (in `internal/database/schema/`). Remember that commons-db re-runs a script
whenever its content changes, so every script must be idempotent.

- `095_tenancy_prepare.sql`, `-- phase: pre`:
  - drop `violations_file_path_fkey` if both the table and the constraint exist;
  - guard every statement with `to_regclass(...) IS NOT NULL` and `pg_constraint` existence checks.
  - This must run before Atlas changes the `file_scans` PK. Otherwise the PK drop fails because the
    old FK depends on it.
- `151_tenancy_local_org.sql`, `-- phase: post`:
  - `INSERT INTO orgs (id, issuer, clerk_org_id, slug, name) VALUES ('00000000-0000-0000-0000-000000000001','local','local','local','Local') ON CONFLICT (id) DO NOTHING;`
  - Then a `DO $$ ... $$` loop that adds
    `<table>_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.orgs (id) ON DELETE RESTRICT` for
    every table in the "org_id NOT NULL" list (and `repos`, `github_installations`, `sandbox_runs`
    if you also keep their org FKs out of the HCL), when the constraint is missing.
- `internal/database/database.go` `applyGavelSchema`: add every `<table>.<table>_org_id_fkey`
  installed by 151 to the `migrate.WithExclude(...)` list, **one entry per table** (don't rely on
  wildcards). Without that, Atlas sees them as unknown and drops them.

**Write the tests first.** Extend `internal/database/identity_schema_test.go`. Also extend
`identity_integration_test.go`, which uses embedded
Postgres:

1. Create the pre-tenancy schema: check out the HCL from `main` into a temp dir and apply it via
   `migrate.Apply`, or use `AutoMigrate` on the old structs as the existing test does. Insert one
   row into `seen_prs`, `file_scans` + `violations`, and `todo_workspaces`.
2. Run `database.Open(ctx, database.WithMigrations())`.
3. Assert:
   - those rows now have `org_id = LocalOrgID`;
   - the local org row exists;
   - the FKs exist;
   - inserting an `org_id` that doesn't exist fails;
   - two orgs can each have `seen_prs(repo='a/b', number=1)`.
4. Run the migrations a **second time** and assert nothing fails. This checks idempotency.

**Stop and ask** if Atlas plans any DROP of a column or table you didn't intend. Print the plan:
commons-db `migrate` has a debug/dry-run option; check `WithDryRun` or similar in
`$GOMODCACHE/github.com/flanksource/commons-db@v0.1.39/migrate`.

### M1.3 Org-scoped store handles

**Pattern.**
- Each store gets an `org uuid.UUID` field.
- Existing constructors and singletons default it to `tenant.LocalOrgID`, so every existing caller
  stays correct with **no signature change**.
- A `ForOrg(org uuid.UUID) *X` method returns a shallow copy with the org set.
- Every query adds `org_id = ?`. Every insert sets `org_id`.
- Conflict targets include `org_id`, and conflict updates **never** assign `org_id`, `user_id` or
  `created_*`.

**`github/cache`** (`Store`, db.go:34).
- Add `org uuid.UUID`. `Open()` and `Shared()` set `LocalOrgID`. Add `func (s *Store) ForOrg(org uuid.UUID) *Store`.
  - It shares `db` and `disabled`.
  - It **shares** `writeMu`: hold it by pointer. Change the field to `writeMu *sync.Mutex`,
    allocated in `Open`.
- Add `OrgID uuid.UUID \`gorm:"primaryKey;type:uuid"\`` and `UserID *uuid.UUID \`gorm:"type:uuid"\``
  to:
  - `HTTPCacheEntry`, `WorkflowRunCache`, `JobLogCache`, `WorkflowDefCache` (models.go)
  - `SeenPR` (seen.go)
  - `TestRunCache`, `TestRunCursor` (test_runs.go)
- Replace every `UpdateAll: true` with an explicit
  `Columns: [org_id + previous key columns], DoUpdates: clause.AssignmentColumns([...payload columns...])`.
  The payload column lists are below (the struct fields that aren't keys, owners or `created_*`):
  - `http_cache_entries`: `etag, last_modified, status_code, body, headers, fetched_at, expires_at`
  - `workflow_run_caches`: `repo, status, conclusion, payload, fetched_at`
  - `job_log_caches`: `repo, logs_gz, fetched_at`. In M7 this becomes `repo, log_key, fetched_at`.
  - `workflow_def_caches`: `path, yaml, fetched_at`
  - `test_run_caches`: `path, kind, repo, sha, started_ts, ended_ts, passed, failed, skipped, warned, total, lint_violations, lint_linters, frameworks, synced_at`
  - `test_run_cursors`: `last_run_ts, synced_at`
- Every `Where(...)`, `First`, `Find` and `Delete` adds `org_id = ?` with `s.org`. Every `Create`
  sets `OrgID: s.org`. This includes the `pruneOnOpen` delete.
  - `pruneOnOpen` is a global maintenance job. It deliberately prunes **all orgs**, so leave it
    unscoped and add a comment saying so.
- **Tests** (embedded PG via `commons-db/dbtest`) in `github/cache/tenancy_test.go`. For every
  public method:
  - org A writes;
  - org B reads nothing;
  - org B writes the same key, and both rows coexist;
  - org A's read is unchanged.

**`internal/cache`** (`ViolationCache`, `LinterStats`).
- Add `org uuid.UUID`, defaulting to local in `NewViolationCache` and `NewLinterStats`, plus a
  `ForOrg` method.
- Struct fields: `FileScan.OrgID` (primaryKey), `violationRecord.OrgID`, `LinterExecution.OrgID`,
  `DebounceMetadata.OrgID` (primaryKey).
- Update the two `OnConflict` clauses: `RecordExecution` → columns `org_id, linter_name, work_dir, executed_at`;
  `updateDebounceMetadata` → `org_id, linter_name, work_dir`.
- `GetStats` uses `pg_total_relation_size`, which is table-wide. Leave it and comment it as
  "storage size across all orgs".
- Tests: same cross-org pattern.

**`internal/taskhistory/store.go`.**
- Add `org` to `Store` with `ForOrg`.
- In `importRecord`, add `org_id` to the INSERT column list and values. Don't add it to the
  `DO UPDATE SET`.
- Add `AND org_id = ?` to `ListRuns`, `Snapshot` and the `ON CONFLICT ... WHERE`. Use
  `task_run_history.org_id = EXCLUDED.org_id` so another org can't overwrite a row.
- `Prune` stays global.
- Tests: same pattern.

**`todos/native` (`Repository`).**
- Add `org uuid.UUID`. `NewRepository` defaults it to local. Add `func (r *Repository) ForOrg(org uuid.UUID) *Repository`.
- `LaunchCoordinator` builds `NewRepository(captainTx.Gorm())` inside transactions (launch.go).
  Propagate the org: give `LaunchCoordinator` an `org` field set from the repository passed to
  `NewLaunchCoordinator`, and use `NewRepository(tx).ForOrg(c.org)`. Grep todos/native for every
  `NewRepository(` call and propagate the org the same way.
- `CreateWorkspace`: add `org_id` and `user_id` to the INSERT. `user_id` comes from
  `tenant.FromContext(ctx).UserID`. The unique-violation error message stays the same.
- Every query that reads `todo_workspaces` adds `AND workspace.org_id = ?`. The queries that reach
  issues **through** a workspace (`GetIssue`, `GetIssueByRef`, `GetIssueByGlobalRef`,
  `GetIssueBySessionRef`, `ListIssues`, global lookups) must join `todo_workspaces` and filter by
  org. Find them all with:
  ```
  grep -n "todo_issues\|todo_workspaces\|todo_labels\|todo_issue_" todos/native/*.go | grep -v _test
  ```
  Each read query must end up org-filtered. Writes that take an issue ID must first verify the
  issue's workspace org: one `SELECT 1 ... JOIN todo_workspaces ... WHERE org_id = ?`, returning
  `ErrNotFound` on mismatch.
- `SetLabelDefinition`:
  - With a workspace: set `org_id = (SELECT org_id FROM todo_workspaces WHERE id = ?)`. The
    conflict target `(workspace_id, name) WHERE workspace_id IS NOT NULL` is unchanged.
  - Without a workspace: set `org_id = r.org`, and the conflict target becomes
    `(org_id, name) WHERE workspace_id IS NULL`.
  - `ListGlobalLabelDefinitions` adds `AND org_id = ?`.
- Attribution:
  - `CreateIssue` sets `created_by_user_id`;
  - `insertEvent` sets `actor_user_id`;
  - `LinkPromptRun` sets `requested_by_user_id`;
  - `LinkPlan` sets `created_by_user_id`;
  - `SetLabelDefinition` sets `updated_by_user_id`.
  - All of these come from `tenant.FromContext(ctx).UserID`; `insertEvent` gets it through its
    caller's ctx. NULL in local mode.
- **Captain sidecar writes.** Wherever gavel calls `captainTx.CreateOrGetSession`, also run
  `INSERT INTO session_owners (session_id, org_id, user_id) VALUES (?,?,?) ON CONFLICT (session_id) DO NOTHING`
  in the same transaction. The call sites are:
  - `todos/native/launch.go:139,147,360,366`
  - `create_issue_session.go:49,83,91`
  - `create_issue_plan.go:47,55`
  - `todos/runtime/lifecycle_runs.go:248`
  Add one helper, `recordSessionOwner(tx *gorm.DB, sessionID uuid.UUID, scope tenant.Scope) error`,
  in `todos/native/ownership_sessions.go`.
- Tests (embedded PG) in `todos/native/tenancy_test.go`. Workspace `repo_key='github.com/a/b'`
  exists in org A and org B independently, and:
  - org B's `GetIssue(A's issue id)` → `ErrNotFound`;
  - org B's `ListIssues` → empty;
  - global labels with the same name coexist across orgs;
  - `session_owners` rows are written with the right org.

**`todos/runtime/provider.go`.**
- `Open`, `OpenGlobal` and `New` use the local org.
- Add `OpenForOrg(ctx, options, scope tenant.Scope) (*Provider, error)`, which builds the repository
  with `ForOrg(scope.OrgID)`.

**`api.IdentityStore`.**
- `Sync` keeps its signature, but `Ownership.OrgID` becomes `uuid.UUID`.
  - When the identity has no org, return `tenant.LocalOrgID` **only in local mode**. In clerk mode
    the middleware rejects requests without an active org first (M3).
- Add `func (s IdentityStore) SyncMemberships(ctx, userID uuid.UUID, memberships []Membership) error`.
  It replaces the user's rows in `org_memberships` and upserts the `orgs` rows.

**M1 acceptance:**
- All existing tests pass with `--auth none` semantics.
- The new cross-org tests pass.
- Running the migrations twice is a no-op.
- The `api/OWNERSHIP.md` table is updated to describe NOT NULL plus the local org, and keys that
  include org. This is the user-facing record of the model.

---

## M2. Token verification (`api/`)

**Goal.** Verify, on the server:
- (a) Clerk **session tokens** from the browser;
- (b) the CLI's Clerk **OAuth access tokens**;
- (c) **GitHub Actions OIDC** tokens.

All three share one JWKS cache.

### M2.0 Spike: CLI access token format (report to Batman)

Use a real Clerk dev instance. Run `gavel login`, then inspect `Session.Token.AccessToken`. Record
the answers:

1. Is it a JWT? Count the dots, then decode the header and payload. If it is, which `iss`, `aud`,
   `sub` and org claims does it carry?
2. If it's opaque, does `GET <issuer>/oauth/userinfo` with `Authorization: Bearer <token>` return
   `sub` and `email`?

**Decision rule.**
- If it is a JWT verifiable with the instance JWKS, verify it locally.
- Otherwise use `/oauth/userinfo`, caching the response for 60 seconds keyed by
  `sha256(token)` (never by the raw token).
- In both cases the **org for CLI requests comes from the `X-Gavel-Org` header** (an org slug or a
  Clerk org ID), checked against `org_memberships`.

### M2.1 Files

- `api/jwks.go`: `type JWKSCache struct{...}` with `func NewJWKSCache(url string, client *http.Client, ttl time.Duration) *JWKSCache`
  and `func (c *JWKSCache) Key(ctx, kid string) (*rsa.PublicKey, error)`.
  - Refetch on an unknown `kid`, at most once per 30 seconds.
  - The TTL is 10 minutes.
  - Move the RSA-key parsing out of `verifyIdentity` into this file, keeping the same validation:
    a modulus of at least 2048 bits, and the exponent checks.
- Refactor `AuthClient.verifyIdentity` to use `JWKSCache`. Behaviour must not change; the existing
  `api/auth_test.go` must still pass.
- `api/session_token.go`: `type SessionVerifier struct { Issuer string; AuthorizedParties []string; JWKS *JWKSCache }`,
  with `func (v SessionVerifier) Verify(ctx, raw string) (SessionClaims, error)`.
  - The JWKS URL is `<issuer>/.well-known/jwks.json`.
  - Requirements: RS256, `iss == Issuer`, `exp`/`nbf` checked with 5 seconds of leeway, `sub`
    present, and `azp` in `AuthorizedParties` when `azp` is present.
  - `SessionClaims{UserID string; OrgID, OrgSlug, OrgRole string}` supports both Clerk claim
    versions: v2 `o: {id, slg, rol}` and v1 `org_id`/`org_slug`/`org_role`.
  - `OrgRole` always has the `org:` prefix. v2's `rol` comes without it and gets it added, so roles
    compare equal to v1 claims and to Backend API memberships.
  - A session whose `sts` is `pending` is rejected; Clerk's SDKs treat it as signed out (§15).
- `api/access_token.go`: the CLI verifier, chosen by the spike. Signature:
  `func (v AccessTokenVerifier) Verify(ctx, raw string) (clerkUserID string, err error)`.
  - `func NewAccessTokenVerifier(issuer, clientID string, jwks *JWKSCache, clerk ClerkOrganizations, ttl time.Duration) *AccessTokenVerifier`.
    A constructor rather than a struct literal, because the verifier owns its result cache. The
    M2.0 spike showed Clerk OAuth access tokens are RS256 JWTs signed by the instance JWKS, so the
    first checks are local.
  - Requirements: header `typ == "at+jwt"`, RS256, `iss == Issuer`, `client_id == ClientID`,
    `exp`/`nbf` with 5 seconds of leeway, and `sub` present. There is no `aud` claim.
  - The `typ` and `client_id` checks matter: ID tokens and browser session tokens are signed by the
    same keys, and neither may be accepted as a CLI bearer token.
  - **Revocation.** After the local checks pass, call
    `POST https://api.clerk.com/v1/oauth_applications/access_tokens/verify` with the server's
    secret key (confirmed 2026-10-14, §13). Accept only a token object with `revoked` and
    `expired` false whose `subject` and `client_id` match the JWT. Reject `{"active": false}`, a
    404, and a revoked token. Any other response is an error.
  - **Cache.** Cache the result with `commons-db/cache.NewCache[T]`, keyed by `sha256(token)`
    (never the raw token). Each entry expires at the earlier of the token's `exp` and the check
    time plus `auth.token.cache.ttl` (`properties.Duration`, default 10m). Requests never extend
    an entry, so a revoked token stops working within one TTL. Only successful checks are cached.
- `SessionVerifier` rejects a token whose header `typ` is `at+jwt` or that carries `client_id`, so an
  OAuth access token never passes as a browser session.
- **Errors.** All three verifiers return an error wrapping `api.ErrInvalidToken` for a bad token
  (M3 answers 401; check it with `errors.Is`). Each says why, e.g. `invalid token: the Clerk
  session is pending: finish signing in, e.g. choose an organization`, or `... was issued to OAuth
  client "x", not "y"`. Messages never include the token or the secret key.
  A JWKS fetch failure or a Backend API failure comes back as a different error, so M3 can tell
  "invalid" from "couldn't check". Shared checks (RS256, kid lookup, 5s leeway, `exp` required)
  live in `parseSigned` in `api/jwks.go`.
- `api/github_oidc.go`: `type GitHubOIDCVerifier struct { Audience string; JWKS *JWKSCache }`.
  - Issuer `https://token.actions.githubusercontent.com`; JWKS
    `https://token.actions.githubusercontent.com/.well-known/jwks`.
  - `Verify` returns
    `GitHubOIDCClaims{Repository string; RepositoryID int64; RepositoryOwnerID int64; SHA, Ref, RunID, Workflow string}`.
  - `repository_id` arrives as a string claim; parse it.
- `api/clerk_organizations.go`: add `func (c ClerkOrganizations) UserMemberships(ctx, clerkUserID string) ([]Membership, error)`.
  - It calls `GET https://api.clerk.com/v1/users/{id}/organization_memberships?limit=100` and
    paginates with `offset` until it has `total_count` memberships.
  - `Membership{OrgID, OrgSlug, OrgName, Role string}`.
  - Check the response shape against Clerk's Backend API docs before coding; record any surprises
    in the spike log.
- `api/clerk_users.go`: `func (c ClerkOrganizations) GitHubToken(ctx, clerkUserID string) (string, error)`
  calls `GET /v1/users/{id}/oauth_access_tokens/oauth_github` and returns the token of the first
  entry. Used in M5.

**Tests** use `api/clerktest`, which signs with a generated RSA key and also serves the Backend
API endpoints above. Its `BackendClient()` routes `api.clerk.com` requests to the fake, so
`ClerkOrganizations` keeps the real URL. Cover:
- a valid token;
- an expired token;
- an ID token or session token presented as an access token, and an access token presented as a
  session token;
- an access token for another `client_id`;
- a revoked access token rejected on a cache miss; a cached token accepted without a Clerk call
  until its entry expires, then re-checked; an entry never outliving the token's `exp`;
- the wrong issuer;
- the wrong `azp`;
- an unknown `kid` that triggers a refetch;
- v1 and v2 org claims;
- an OIDC token with the wrong audience;
- membership pagination.

---

## M3. `gavel serve --auth clerk`

**Goal.** An authenticated, org-scoped dashboard. `--auth none` is unchanged.

### M3.1 Flags

`PRListOptions` is shared by `gavel pr list`, so **don't add the flags there**.
- Create `ServeOptions` in `cmd/gavel/serve_dashboard.go`, embedding `PRListOptions`.
  `clicky.AddNamedCommand` supports struct embedding; check how clicky handles embedded structs in
  `$GOMODCACHE/github.com/flanksource/clicky@v1.21.70`. **If it doesn't, stop and ask Batman.**
- Rename `runServeDashboard` to take `ServeOptions`.

Fields (`flag` tags as in the existing code):

| Flag | Default | Notes |
|---|---|---|
| `--auth` | `none` | `none` or `clerk` |
| `--public-url` | | Required in clerk mode. Example: `https://gavel.example.com`. Used for `azp`, the OIDC audience, the GitHub setup URL and the pod run API URL. |
| `--clerk-issuer` | env `GAVEL_CLERK_ISSUER` | |
| `--clerk-publishable-key` | env `GAVEL_CLERK_PUBLISHABLE_KEY` | |
| `--clerk-client-id` | env `GAVEL_CLERK_CLIENT_ID` | The public OAuth client, so `gavel login --server` can discover it. |
| `--trusted-proxy` | | CIDRs whose `X-Forwarded-*` headers are trusted. |
| `--sandbox` | `kubernetes` | `kubernetes`, `docker` or `podman`. Only valid with clerk. |
| `--sandbox-image` | | |
| `--sandbox-namespace` | `gavel-runs` | |
| `--kubeconfig` | | |
| `--sandbox-secret-env` | | Repeated. Env var names copied from serve into pods; see M10. |
| `--ssh-addr` | | Example: `:2222`. Starts the push server in-process, clerk mode only. |
| `--ssh-host-key` | | |
| `--data-dir` | `/var/lib/gavel` | Holds repos, mirrors, transcripts. |
| `--github-app-id` | | |
| `--github-app-slug` | | |
| `--github-app-private-key` | env `GAVEL_GITHUB_APP_PRIVATE_KEY_FILE` | |

Secrets come **from env only**, never flags:
- `CLERK_SECRET_KEY`
- `GAVEL_SECRET_KEY`: 32+ random bytes, base64. Signs run tokens and GitHub install state.

**Validation** (`func (o ServeOptions) Validate() error`, written test-first):
- clerk mode requires `public-url`, the issuer, the publishable key, `CLERK_SECRET_KEY` and
  `GAVEL_SECRET_KEY`;
- `--sandbox`, `--ssh-addr` and the GitHub App flags are rejected with `--auth none`;
- a configured DB is required in clerk mode (`database.Require`).

### M3.2 `OrgServers` (`pr/ui/orgservers.go`, new)

This mirrors `testrunner/ui/multiserver.go`. `Server` stays **single-org**; never make it org-keyed
internally.

```go
// OrgServers serves one dashboard Server per organization in authenticated mode.
// The request's tenant.Scope selects the Server; idle organizations are evicted.
type OrgServers struct {
    mu      sync.Mutex
    servers map[uuid.UUID]*orgEntry   // entry: srv *Server, poller *Poller, cancel, lastUsed
    factory func(ctx context.Context, scope tenant.Scope) (*Server, *Poller, error)
    idleTTL time.Duration             // 30m
}
func NewOrgServers(factory ...) *OrgServers
func (o *OrgServers) Handler() http.Handler   // scope := tenant.FromContext(r.Context()); srv := get-or-create; srv.Handler().ServeHTTP
func (o *OrgServers) Close()
```

- Cache each org's `srv.Handler()`; build it once per Server.
- Eviction runs on a 1-minute ticker. It cancels the poller context of servers idle longer than
  `idleTTL` **and** with no running project actions. Add `func (s *Server) Busy() bool`, which
  checks `projectActions` and `commitQueues`.
- **The factory**, in `cmd/gavel/serve_dashboard.go`, builds a Server for an org:
  - `ghOpts := github.Options{TokenSource: installationTokenSource(org)}` (M5). Before M5 lands,
    the factory returns a Server with the PR poller **disabled** (no GitHub), and the PR tab shows
    "Connect GitHub". Add `func (s *Server) DisablePRPolling()` if needed.
  - The config comes from `org_settings.search_config`, falling back to the default
    `SearchConfig{}`.
  - Wire the server's stores to the org: `srv.SetOrg(scope.OrgID)`. Add this setter; it stores the
    org, and every handler that opens a store uses `.ForOrg(s.org)`. Grep `pr/ui` for:
    - `cache.Shared()`
    - `database.Shared`
    - `todoruntime.Open`, `OpenGlobal` and `New`
    - `taskhistory.NewStore`
    - `NewViolationCache`
    and route each one through the org.
- In `handleConfig` POST, clerk mode persists to `org_settings` instead of `SaveSettings`. Branch on
  `s.org != tenant.LocalOrgID`; don't add a separate flag.

### M3.3 Auth middleware (`pr/ui/auth.go`, new)

```go
// Authenticator resolves the caller's identity and active organization for each request
// and rejects requests that have neither.
type Authenticator struct {
    Session   api.SessionVerifier
    Access    api.AccessTokenVerifier
    Identity  api.IdentityStore           // with OrganizationResolver = api.ClerkOrganizations{SecretKey}
    Clerk     api.ClerkOrganizations
    Public    func(path string) bool      // public paths
    Trusted   []netip.Prefix
}
func (a *Authenticator) Middleware(next http.Handler) http.Handler
```

**Flow per request:**

1. **Public paths** pass through with no scope: `/`, `/prs`, `/todos`, every SPA page route that
   `handleRoute` serves, `/_assets/`, `/favicon.svg`, `/brand/`, `/manifest.webmanifest`,
   `/api/auth/config`, `/api/github/setup` (it has its own state check), `/api/ci/` (OIDC, M5),
   and `/api/runs/` (run token, M10). The page shell is public; the SPA gates itself.
2. **Token source:**
   - `Authorization: Bearer <t>`. If it verifies as a session token, use the browser flow;
     otherwise try the access-token verifier (CLI).
   - Else the `__session` cookie.
   - Neither present: `401 {"error":"unauthenticated"}`.
3. **Browser.** Verify the session. Then:
   - `Identity.Sync`, which upserts the user and the active org.
   - Membership: an active org claim is authoritative for the session. Also upsert the
     `org_memberships` row with the claim's role.
   - No active org: `409 {"error":"no_active_org"}`.
4. **CLI.**
   - Resolve the user. If the user isn't in `users`, sync it with `Clerk.UserMemberships` plus the
     user fetch.
   - Read the org from `X-Gavel-Org`. If it's missing and the user has exactly one membership, use
     that one; otherwise `409 no_active_org`.
   - Membership comes from `org_memberships`. Resync from Clerk when `synced_at` is older than
     5 minutes.
   - Not a member: `403`.
5. Put `tenant.Scope{OrgID, UserID, Role, GitHubLogin}` on the context. `GitHubLogin` comes from
   the Clerk user's GitHub external account, fetched once at Sync and stored on `users`: add a
   nullable `github_login text` column to `users`.
6. **CSRF.** For unsafe methods (POST, PUT, PATCH, DELETE) with cookie auth, require
   `Origin == --public-url` (or `Sec-Fetch-Site: same-origin`). Return `403` otherwise.
7. **Forwarded headers.** Fix the request-origin logic at `pr/ui/handler.go:579-603`: only honour
   `X-Forwarded-Proto` and `X-Forwarded-Host` when `RemoteAddr` is in `Trusted`.

**Wiring.**
- `cmd/gavel/serve_dashboard.go`: in clerk mode, `http.Serve(listener, auth.Middleware(orgServers.Handler()))`.
- In none mode, keep the exact current path (`srv.Handler()`).
- `ui.Server.Handler()` itself is unchanged, so the existing handler tests that use `&Server{}` keep
  passing.

**Endpoints** (`pr/ui/me.go`, new):
- `GET /api/auth/config` (public): `{"mode":"none"|"clerk","publishableKey":"...","issuer":"...","clientId":"..."}`.
  In none mode it returns `{"mode":"none"}`.
- `GET /api/me`:
  `{"user":{"id","name","email","githubLogin"},"org":{"id","slug","name","role"},"orgs":[...memberships]}`.
  - In none mode it returns the CLI-local identity, if `api.LoadSession` succeeds and verifies
    offline via the stored ID token claims (no network call), with `org: {"id": local, "slug":"local","name":"Local"}`.
  - Otherwise it returns `{"user":null,"org":{local}}`.
- Register both on the Server mux (`Handler()`), so they work in both modes. `/api/auth/config`
  needs the auth mode, so add `func (s *Server) SetAuthConfig(cfg AuthConfigResponse)`.

**Features off in clerk mode: an allowlist (Batman, 2026-10-14).** `Server.SetHostedMode(true)`
serves only the routes in `hostedRoutes` (`pr/ui/hosted.go`); every other route returns `404`.
A route added later stays off in hosted mode until someone adds it, after checking it is
org-scoped. The gate sits inside the SSE hub (`hub.Guard(gate(mux))`), so a hub subscription
can't reach a route that is turned off. The list for M3 (`hostedRoutes` is the authoritative
copy):

- **Shell and assets:** `/` (the SPA routes `handleRoute` serves), `/_assets/`, `/favicon.svg`,
  `/brand/`, `/manifest.webmanifest`, and the SSE hub's endpoints.
- **Auth:** `/api/auth/config`, `/api/me`.
- **PR dashboard** (the org's GitHub token, M5): `/api/prs`, `/api/prs/stream`, `/refresh`,
  `/bots`, `/closed`, `/pause`, `/detail`, `/commits/diff`, `/files/diff`, `/merge`, `/approve`,
  `/update-branch`, `/job-logs`, `/seen`; `/api/config`, `/api/repos`, `/api/repos/favicon`,
  `/api/orgs`.
- **Not in M3: todos, projects and runtime presets.** Every todo path resolves its workspace
  from a host directory: `openTodoProvider(ctx, dir)` calls `ProjectForDir` with a
  client-supplied `dir`, the entity's `Workspaces`/`DefaultDir`/`ProjectDir` read the host's
  project registry, and its `DB` hook is the unscoped database. `GET /api/projects` and the
  runtime presets read the host registry too. They return in M6, once workspaces and projects
  come from the org's repositories (`repos`), and M6's tests include "org A's todo ID requested
  with org B's session returns 404".
- **Answered without running anything:** `/api/todos/run`, `/plan/approve`, `/plan/revise`,
  `/answer` and `/api/v1/todo/{id}/{run|plan|triage}` return
  `501 {"error":"agent runs are not available in hosted mode yet"}` (M10 lifts this).
- **Projects:** only `POST /api/projects/{name}/actions` (`test`/`lint` 501, `commit` 400,
  answered before the host's project registry is read), and `POST /api/projects`,
  `PUT`/`DELETE /api/projects/{name}` answer `400 projects are repos in hosted mode`. `POST .../commit-queue` and `.../branch/pr` (open
  a PR) answer `400 commits come from agent runs in hosted mode`.
- **Settings:** `/api/settings/schema`. `/api/config` POST saves to `org_settings` for any org
  but the local one (`SaveOrgSearchConfig`), never to the host's settings file.
- **Status:** `/api/status`, with the database section reduced to health (no DSN source, no table
  counts) when the cache store's org isn't the local org.

Off in hosted mode until a milestone makes them org-aware, among others: `/api/captain/sessions/`,
`/api/sessions`, `/api/todos/session/*` (every org's sessions); `/api/git/*`,
`/api/projects/git-summary`, `/api/projects/{name}/git|diff|branch/*`, `/api/tests*`,
`/results/`, `/api/todos/land`, `/api/todos/commits/*` (the host's files and checkouts);
`/api/activity*`, pprof, metrics, ingest stats, `/api/openapi.json`, `/api/proc/*`,
`/api/v1/tasks/*` (process-wide state); `/api/chat` (the assistant runs with the host's keys and
tools); `/api/settings/gavel*` and `/api/settings/prompts*` (host files); `/api/todos/github`,
`/api/todos/transfer`, `/api/todos/criteria`, `/api/todos/verification/*`,
`/api/todos/run/context|preview`, `/api/project-runs/`, `/react-grab*`.

Also:
- `startServeRuntime` doesn't start the Captain session monitor in clerk mode. Add a mode
  parameter.
- **Execution is off in hosted mode until it can be sandboxed** (D11). Merged states between M3 and
  M10 must never run pushed or user code on the serve host.
  - Agent, todo and fixture runs return `501 {"error":"agent runs are not available in hosted mode yet"}`
    from `/api/todos/run`, `/plan/approve`, `/plan/revise`, `/answer` and
    `/api/v1/todo/{id}/{run|plan|triage}`. M10 lifts this.
  - Project actions (`test`, `lint`) return `501 {"error":"project actions are not available in hosted mode yet"}`.
    M9.5 lifts this.
  - Project `commit` and `open-pr` return `400 {"error":"commits come from agent runs in hosted mode"}`
    permanently (see §15).
- Projects: in hosted mode, `handleProjects` POST/PUT reject any `dir` (`400 "projects are repos in hosted mode"`).
  The project list comes from the `repos` table (M5 and M6). Implement `projectsForOrg(ctx)`
  returning `[]Project{Name: repo.Name, Repos: [repo.Name]}` with `Dir` empty.
- `validateTodoRunSpec` (`pr/ui/todo_run.go:419`) clears `spec.Setup`, `spec.Sandbox` and
  `spec.CLIArgs` (use the exact field names on captain `api.Spec`) **in all modes**. The server
  decides those.
- Attachments, in all modes:
  - drop `Access-Control-Allow-Origin: *` and the OPTIONS handling in `handleTodoAttachmentUpload`;
  - serve with `Cache-Control: private, max-age=86400`.

**Tests** (Ginkgo, `pr/ui/auth_ginkgo_test.go`):
- the middleware with a fake signer (real RSA and JWKS through `httptest`);
- public path, missing token, bad token, no active org, CLI with an `X-Gavel-Org` it isn't a member
  of, CSRF;
- the scope reaches the handler;
- org A's todo ID requested with org B's session returns 404.
- hosted mode: todo/agent run endpoints and project `test`/`lint` return 501; project `commit`/`open-pr` return 400.

For `OrgServers`:
- two orgs get distinct Servers;
- eviction skips busy servers.

For `Validate`, a table-driven test.

---

## M4. UI (`pr/ui/src`)

**Goal.** Clerk sign-in, a user badge, and the org switcher in clerk mode. A local identity badge in
none mode.

1. Add the `@clerk/clerk-react` dependency to `pr/ui/package.json` (React 19 compatible; pin an
   exact version that's at least two weeks old). Run `pnpm install` in `pr/ui` (ask Batman first).
2. `src/authConfig.ts`: `export async function loadAuthConfig(): Promise<AuthConfig>`, which
   fetches `/api/auth/config`.
   - Don't rely on `window.__GAVEL__`. Dev mode proxies to Vite and has no `__GAVEL__`.
   - Type: `{ mode: 'none' | 'clerk'; publishableKey?: string; issuer?: string; clientId?: string }`.
3. `src/index.tsx`: wrap rendering in `loadAuthConfig().then(cfg => render(cfg))`.
   - In clerk mode: render `<ClerkProvider publishableKey={cfg.publishableKey} afterSignOutUrl="/">`
     around the existing tree, then `<AuthGate>` inside `QueryClientProvider`.
   - `AuthGate` (`src/components/AuthGate.tsx`):
     - `<SignedOut><SignIn routing="hash" /></SignedOut>`;
     - `<SignedIn>`: if `useOrganization().organization` is null, render a centered
       `<OrganizationList hidePersonal afterSelectOrganizationUrl="/" />`; otherwise render
       `children`.
   - The gate wraps `<App/>` **outside** `App`. `App` has early returns for `isProcessesPage` and
     `isTodoNewPage` (App.tsx:541,553), so a gate inside `AppShell` wouldn't cover them.
4. `src/components/UserMenu.tsx`:
   - Clerk mode: `<OrganizationSwitcher hidePersonal afterSelectOrganizationUrl="/" />` plus
     `<UserButton />`.
   - None mode: a small badge from `useQuery(meQuery())` showing `user?.name ?? 'local'`, with no
     switcher.
   - On org change, call `queryClient.clear()` and `window.location.assign('/')`. A full reload
     restarts the SSE `EventSource`s (six of them) with the new session cookie.
5. Insert `<UserMenu />` in `App.tsx` `AppShell actions`, immediately **before** `<ThemeToggle />`
   (App.tsx ~594). Also add it to `MenubarView` (App.tsx:824) and the `ProcessesPage` header
   (App.tsx:750).
6. Rename the visible label of the GitHub `OrgChooser` (`src/components/OrgChooser.tsx`) to
   "GitHub scope", so it isn't confused with the gavel org. Only the title and aria label change;
   keep its behaviour. Update `OrgChooser.test.tsx` accordingly.
7. `src/query.ts`: add `queryKeys.me = () => ['me'] as const`. Add `meQuery()` in
   `components/oneShotQueries.ts`.
8. A 401 from any `fetchJSON` in clerk mode should reload the page so Clerk refreshes the session.
   Add the handling in `query.ts` `request()`: `if (res.status === 401 && authMode === 'clerk') location.reload()`.
   Hold `authMode` in a module variable set by `index.tsx`.
9. **Vitest tests:**
   - `AuthGate.test.tsx`: mock Clerk components with `vi.mock('@clerk/clerk-react')`. This tests
     our gating logic, not Clerk's.
   - `UserMenu.test.tsx`: none mode renders the local badge.
10. After building, run `pnpm run build` in `pr/ui` (ask first). If `testrunner/ui` changed, also
    run its `build` **and** `build:lib`, per the CLAUDE.md note.

---

## M5. GitHub App and GitHub Actions OIDC

### M5.0 Spike (report to Batman)

1. In Clerk, configure the **GitHub social connection** with the **GitHub App's** OAuth client ID
   and secret ("Use custom credentials").
2. Sign in, then call `ClerkOrganizations.GitHubToken(user)` (M2).
3. With that token, does `GET https://api.github.com/user/installations` list the App's
   installations?
   - GitHub only allows this with a token from the App's own user authorization.
   - If it doesn't work, **stop**. The fallback is a separate GitHub App OAuth "user authorization"
     redirect handled by gavel. Batman decides.

### M5.1 GitHub client token source

- `github/client.go` `Options`: add `TokenSource func() (string, error)`.
- In `token()`, if `TokenSource != nil`, return its result and **never** fall back to the env
  variables or `auth.json`. A hosted org server must not use the host's token.
- Test: with `TokenSource` set and `GITHUB_TOKEN` in the env, the `TokenSource` value is used.

### M5.2 App installation tokens

File `github/app.go`:

```go
// App mints GitHub App JWTs and caches installation access tokens.
type App struct { ID int64; Slug string; Key *rsa.PrivateKey; HTTP *http.Client; mu sync.Mutex; tokens map[int64]cachedToken }
func LoadApp(id int64, slug, keyPath string) (*App, error)
func (a *App) InstallationToken(ctx context.Context, installationID int64) (string, error)
func (a *App) Installation(ctx context.Context, installationID int64) (Installation, error)
func (a *App) InstallationRepos(ctx context.Context, installationID int64) ([]InstallationRepo, error)
```

- `InstallationToken`:
  - the JWT is RS256 with `iat = now-60s`, `exp = now+9m`, `iss = app id`, signed with
    `golang-jwt` (already a dependency);
  - it calls `POST /app/installations/{id}/access_tokens`;
  - it caches until `expires_at - 5m`.
- `Installation` calls `GET /app/installations/{id}` with the App JWT.
- `InstallationRepos` calls `GET /installation/repositories?per_page=100`, paginated, with the
  installation token.
- Tests: an `httptest` GitHub. Check the JWT claims, the caching, and that pagination returns all
  repositories.

### M5.3 Linking an installation to an org

- `GET /api/github/install` (authenticated) responds `302` to
  `https://github.com/apps/<slug>/installations/new?state=<state>`.
  - `state = base64url(json{org, user, exp: now+15m}) + "." + base64url(HMAC-SHA256(GAVEL_SECRET_KEY, payload))`.
- `GET /api/github/setup?installation_id=&setup_action=&state=`:
  - This path is public in the middleware, but it still requires the session cookie. Re-run the
    verifier inline.
  - Verify the state HMAC and expiry, and that `state.user` matches the session user.
  - Verify the user can access the installation: the user's GitHub token from Clerk →
    `GET /user/installations` must contain the installation ID.
  - Then:
    - `INSERT INTO github_installations (id, org_id, account_login, account_type, created_by_user_id) ... ON CONFLICT (id) DO NOTHING`.
    - If the installation is **already linked to another org**, return `409`.
    - Sync the repos: `INSERT INTO repos (org_id, name, source, github_repo_id, installation_id, default_branch) ... ON CONFLICT (github_repo_id) WHERE github_repo_id IS NOT NULL DO UPDATE SET name, installation_id, default_branch, updated_at`.
  - Redirect to `/`.
- **Repo sync** also runs every 10 minutes per active org server, and on `POST /api/github/sync`.
  Repos that are gone from the installation are deleted when nothing references them; otherwise
  `installation_id` is set to NULL.

### M5.4 Org server PR polling

- In the `OrgServers` factory (M3.2), `ghOpts.TokenSource` mints a token for the org's
  **first** installation.
- `SearchConfig.Repos` is fixed to the org's GitHub repos (`repos.source = 'github'`). Set
  `All = false`. The GitHub scope chooser is hidden in clerk mode.
- When an org has several installations, PR search runs per installation and the results are
  merged. Implement this as one `searchFn` that loops over the installations; the poller's
  `searchFn` signature stays the same.
- `snapshot.viewer` (the `@me` filter) comes from `scope.GitHubLogin` of the requesting user, not
  from `s.auth.Login`. `s.auth` is the GitHub token probe, which doesn't apply to installations. In
  clerk mode, set `viewer` per request in `handleJSON`/`handleSSE` from the context.

### M5.5 GitHub Actions → gavel uploads (OIDC)

- **Server.** `POST /api/ci/results`, authenticated with `api.GitHubOIDCVerifier{Audience: --public-url}`.
  - The body is the snapshot JSON (`testui.Snapshot`), up to 50 MB.
  - Map `repository_id` → `repos.github_repo_id` → `org_id`. If no row matches, return `403`.
  - Store the snapshot through the same ingest function as sandbox results (M7/M9):
    `ingestRunSnapshot(ctx, org, repoID, sha, ref, snapshot)`. It writes the blob plus a
    `test_run_caches` row with `workspace_dir = "repo:" + repoID`.
- **CLI.** `gavel ci upload --server URL --file path.json`, a clicky command in
  `cmd/gavel/ci_upload.go`. It reads `ACTIONS_ID_TOKEN_REQUEST_URL` and
  `ACTIONS_ID_TOKEN_REQUEST_TOKEN`, requests a token with `&audience=<server>`, and POSTs.
- **Action.** `action.yml`:
  - new input `gavel-server` (optional);
  - when set, add a step after the results are written:
    `gavel ci upload --server ${{ inputs.gavel-server }} --file <json result path>`.
  - Document that the workflow needs `permissions: id-token: write`.
- Tests:
  - the verifier (M2);
  - the handler with a signed fake OIDC token: repo known → 201, unknown → 403;
  - CLI token-request URL composition.

---

## M6. SSH keys and authorized push

### M6.1 Key management

- **API**, scoped to the user (not the org):
  - `GET /api/ssh-keys` → `[{id,name,fingerprint,createdAt,lastUsedAt}]`
  - `POST /api/ssh-keys {name, publicKey}`:
    - parse with `gossh.ParseAuthorizedKey`;
    - reject anything that isn't `ssh-ed25519`, `ecdsa-sha2-*`, or `ssh-rsa` of at least 3072 bits;
    - store the fingerprint from `gossh.FingerprintSHA256`;
    - a duplicate fingerprint returns `409`.
  - `DELETE /api/ssh-keys/{id}` succeeds only for the owner.
- **UI.** A "SSH keys" section in the global settings dialog (where the settings cog opens), in
  clerk mode only: a list, an add form (textarea plus name), and delete.
- **CLI.** Only if the M2.0 spike gave a working CLI token. Add a cobra group `gavel ssh-key` with
  clicky leaves `add [path]` (default `~/.ssh/id_ed25519.pub`), `list` and `rm <id>`. Model it on
  `cmd/gavel/projects.go`.
  - These call the server with `Authorization: Bearer <access token>` and `X-Gavel-Org`.
  - Add `"^ssh-key"` to `gavelMCPExcludedCommands()`.
- **`gavel login --server URL`.**
  - Fetch `URL/api/auth/config` to discover the issuer and client ID.
  - Store `Server string` in `api.Session` (add the field) and `Org string`, the selected org slug.
  - Add `gavel org use <slug>` to set `Session.Org`, plus `gavel org list`.
  - All CLI API calls send `X-Gavel-Org: Session.Org`.

### M6.2 Push server changes (`serve/`)

- **Path traversal, both modes.**
  - `cleanRepoPath` returns `(string, error)`.
  - Reject empty segments, `.`, `..`, and anything outside `[A-Za-z0-9._-]` per segment.
  - At most 2 segments in clerk mode and 1 segment in none mode. Keep the existing default
    behaviour for a single segment.
  - `HandleGitReceive` re-verifies that `filepath.Rel(repoDir, bare)` doesn't start with `..`.
- **`Options` additions:**
  ```go
  Authorizer Authorizer // nil keeps accept-all for --auth none
  OnPushed   func(ctx context.Context, push Push)  // clerk mode: called after receive (always set in clerk mode); nil keeps hook behaviour
  ```
  ```go
  // Authorizer maps SSH keys to users and decides who may push to which repository.
  type Authorizer interface {
      UserForKey(ctx context.Context, key ssh.PublicKey) (uuid.UUID, bool)
      AuthorizePush(ctx context.Context, user uuid.UUID, repoPath string) (Target, error) // Target{OrgID, RepoID uuid.UUID; BareDir string}
  }
  ```
- **`PublicKeyHandler`.**
  - With `Authorizer == nil`, it returns `true`, as today.
  - Otherwise it calls `UserForKey`. If found, it does `ctx.SetValue(userKey, id)`, updates
    `last_used_at`, and returns `true`.
- **`handleSession`.**
  - With an Authorizer, call `AuthorizePush(user, path)`. Errors are written to stderr with exit 1:
    "repository not found or access denied" (deliberately ambiguous).
  - In `AuthorizePush`, in `serve/authorizer.go` with the DB-backed implementation:
    - `org = SELECT id FROM orgs WHERE slug = ? AND issuer = <clerk issuer>`;
    - check membership in `org_memberships`, resyncing from Clerk when stale (5 minutes);
    - any role may push;
    - upsert `repos(org_id, name, source='push')`;
    - `BareDir = <data-dir>/repos/<org_id>/<repo_id>.git`.
- **Clerk mode receive.**
  - Run `git -c core.hooksPath=/dev/null receive-pack <bare>`, so no hooks run.
  - Before and after, snapshot `git for-each-ref --format='%(refname) %(objectname)'` and diff the
    two to compute the updated refs.
  - Then call `OnPushed(ctx, Push{OrgID, RepoID, UserID, Refs: []RefUpdate{Name, Old, New}, Out: sess.Stderr()})`.
- **Clerk mode never runs hooks on the host**, even before M9. Until the dispatcher exists, the
  clerk-mode `OnPushed` writes `push checks are not available yet` to the session's stderr. M9.5
  replaces it.
- **None mode.** The hook path is unchanged.
- **Run the SSH server inside `gavel serve`** when `--auth clerk --ssh-addr` is set, sharing the DB,
  the Clerk client and the executor. `gavel ssh serve` stays as the none-mode server.

**Tests**, extending the Ginkgo `serve/e2e_test.go`:
- Always set `HostKeyPath` to a temp file (fixes §3 item 4).
- New `Describe("authorized push")` using a fake `Authorizer`. This tests our handler logic, not
  the authorizer:
  - unknown key → push fails with the message;
  - known key, but `AuthorizePush` denies → fails;
  - allowed → succeeds;
  - `OnPushed` receives the right refs.
- `cleanRepoPath` table test: `../x`, `a/../../b`, `a//b`, `'a/b'`, and `/a`.
- `serve/authorizer_test.go` (embedded PG): a member can push, a non-member can't, and an unknown
  slug is denied.

---

## M7. Blob store

### M7.0 Upstream commons-db (D10). Stop and ask Batman to approve the PR scope.

In `github.com/flanksource/commons-db`, as one PR:
1. Add `Delete(ctx, path string) error` to `fs.FilesystemRW` and implement it for local, S3, GCS,
   SFTP and SMB.
2. Make S3 honour `AWSConnection.Endpoint` (via `s3.Options.BaseEndpoint`), `S3Connection.UsePathStyle`
   and `S3Connection.ObjectPath` as a key prefix.
3. Fix the file-descriptor leak in `fs/local.go` `Write`: close the created file.
4. Make GCS with empty credentials use Application Default Credentials, not
   `option.WithoutAuthentication()`.
5. Tests for each, using MinIO through testcontainers or the commons-db test conventions.

After it's released, bump gavel's `go.mod` to the new commons-db version.

### M7.1 Package `blob/` (new)

```go
// Package blob stores run results, attachments, recordings and logs on local disk or object storage.
// Keys are slash-separated and never absolute.
type Store interface {
    Write(ctx context.Context, key string, r io.Reader) error
    Read(ctx context.Context, key string) (io.ReadCloser, error)  // ErrNotFound
    Exists(ctx context.Context, key string) (bool, error)
    Delete(ctx context.Context, key string) error                 // ErrNotFound is not an error
    List(ctx context.Context, prefix string) ([]Object, error)    // Object{Key string; Size int64; ModTime time.Time}
}
func Open(ctx context.Context, cfg Config) (Store, error)
func Local(root string) Store
func WithPrefix(s Store, prefix string) Store
```

- `Config` is the `storage:` section of `verify.GavelConfig`:
  ```yaml
  storage:
    type: local | s3 | gcs        # default local
    path: /var/lib/gavel/blobs    # local only; default: per-call-site default (see below)
    bucket: my-bucket
    prefix: gavel/
    s3: { region: us-east-1, endpoint: "", usePathStyle: false, accessKey: {valueFrom: {...}}, secretKey: {...} }
    gcs: { credentials: {valueFrom: {...}} }
  ```
- `Open` maps the config onto `connection.S3Connection` or `connection.GCSConnection`. It then calls
  `Populate`/`HydrateConnection` with `dbcontext.NewContext(ctx)` and
  `connection.GetFilesystemForConnection`.
- `List` uses `ReadDir(prefix + "**")`. Check the doublestar semantics in commons-db.
- Key validation rejects empty keys, a leading `/`, `..` segments, and backslashes.
- **Where the config comes from.** The serve host's **global** `~/.gavel.yaml` only, via the
  existing global-config loader in `verify/config_load.go` (find the function that loads only the
  home layer).
  - Never read it from a repo's `.gavel.yaml`. Pushed code must not redirect storage.
  - Add `Storage StorageConfig` to `GavelConfig` (`verify/config.go:296`).
  - Update `verify/config_schema.go` and `verify/config_schema_sections.go` and their tests.
- **Org prefix.** Callers in clerk mode use `blob.WithPrefix(store, "orgs/"+org.String()+"/")`.
- **Tests:**
  - a conformance suite run against `Local(t.TempDir())`: write, read, exists, list, delete, and
    the not-found cases;
  - the same suite against MinIO when `GAVEL_BLOB_S3_TEST_ENDPOINT` is set;
  - a key-validation table test.

### M7.2 Call-site migrations

Each item is its own TDD step. **When no `storage:` is configured, local mode keeps exactly today's
file layout.** The default store for each call site is `blob.Local(<today's dir>)`.

1. **Snapshots** (`snapshots/`).
   - Add `var OpenStore = func(workDir string) (blob.Store, error) { return blob.Local(filepath.Join(workDir, Dir)), nil }`.
   - When the global config has `storage`, it returns the configured store, prefixed with
     `workspaces/<sha256(abs workDir)[:16]>/`.
   - `Save`, `SavePerRun`, `uniqueRunPath` (use `Exists`), `writePointer`, `LoadPointer`,
     `LoadByPointer`, `ResolveLast`, `LastRun` and `ListRuns` go through the store. Keys are the
     file base names (`run-<ts>.json`, `sha-<sha>.json`, `last.json`, `<branch>.json`).
   - `Pointer.Path` stays `.gavel/<name>` for compatibility with existing pointer files. Strip the
     `.gavel/` prefix to get the key.
   - **`running.json` stays on local disk.** It is in-flight state, so leave `writeRunningJSON` as
     is.
   - Port `testrunner/history/history.go` `loadRunSnapshots` to `snapshots.ListRuns`, keeping its
     sort and filters.
   - Port `pr/ui/test_runs.go` `resolveRunPath`/`readRunSnapshot` to read through the store. The
     run-ID validation is unchanged.
   - Port `pr/ui/project_action_runs.go:28,37`.
   - `fixtures/types/runartifact.go:52` goes through `SavePerRun`, so it's covered.
2. **Todo attachments** (`pr/ui/todo_attachments.go`).
   - `Server` gets an `attachments blob.Store`, defaulting to `blob.Local(attachmentsDir)`.
   - `storeAttachment` writes the key `<uuid><ext>`. `handleTodoAttachment` reads it.
   - In clerk mode the store is org-prefixed, so another org's attachment ID is a 404.
3. **Fixture recordings** (`fixtures/record/store.go`).
   - `Create` and `CreateSidecar` keep returning `*os.File` (the callers stream into it), writing
     to a temp file as today.
   - Add `func (s *Store) Persist(ctx) error`. It uploads every file created during the run to the
     blob store under `recordings/<name>`, and is called from the recorder's end-of-run in
     `fixtures/recorders.go`.
   - When `storage` is unconfigured, `Persist` is a no-op, because the files already sit in
     `.gavel/recordings`.
4. **Attempt transcripts** (`todos/attempts.go`). `writeTranscript` has **no production caller**:
   `saveAttempt` is only called from `todos/attempts_test.go:133`. **Stop and ask Batman** whether
   to delete it as dead code or migrate it. Don't migrate dead code without an answer.
5. **Job logs.**
   - Schema: `job_log_caches` gets `log_key text NOT NULL DEFAULT ''`. Keep `logs_gz` and make it
     nullable during the transition.
   - `PutJobLogs` gzips and writes `joblogs/<job_id>.log.gz` to the org's store, then upserts a row
     with `log_key`.
   - `GetJobLogs` reads `log_key` when it's non-empty. Otherwise it reads the legacy `logs_gz`,
     then writes it to the store and nulls `logs_gz` (lazy migration).
   - `github/cache.Store` gets `blobs blob.Store`. `Shared()` uses
     `blob.Local(<UserCacheDir>/gavel/blobs)` when unconfigured.
   - **Stop and ask Batman** before adding a later migration that drops `logs_gz`.

---

## M8. Retention and cleanup

- **Properties.** Use `github.com/flanksource/commons/properties`:
  `properties.Duration(def, key)` exists (`commons@v1.59.1/properties/properties.go:235`).
  - Load an optional properties file in `gavel serve` with `properties.LoadFile` from
    `--properties` (default `/etc/gavel/gavel.properties` if it exists).
  - Call `properties.BindFlags(cmd.Flags())` if the existing CLI setup allows it; check how clicky
    builds the flag set first.

  | Key | Default | What it removes |
  |---|---|---|
  | `retention.job_logs` | 60d | `job_log_caches` rows whose `fetched_at` is older, plus their blobs |
  | `retention.test_runs` | 60d | `test_run_caches` rows with `kind='test'` older than the window (by `started_ts`), plus the blob at `path` |
  | `retention.lint_runs` | 60d | the same for `kind='lint'` |
  | `retention.recordings` | 60d | blob keys under `recordings/` older than the window (by `ModTime`) |
  | `retention.linter_executions` | 60d | `linter_executions` rows by `executed_at` |
  | `retention.task_history` | 30d | Replaces the constant `taskhistory.Retention` (history.go:16), keeping its current 30-day value. Use it in both `prune` and `Import` (`expires_at`). |
  | `retention.sandbox_runs` | 60d | `sandbox_runs` rows by `finished_at`, plus their `log_key` blob |
  | `retention.interval` | 1h | how often the cleanup job runs |

  The duration format is whatever `properties.Duration` parses. Check whether it accepts `60d`; if
  it doesn't, document `1440h`.
- **Not cleaned by age:**
  - todo attachments, which belong to their todo;
  - `file_scans` and `violations`, which are replaced per file;
  - `seen_prs` and commit stats, which are small;
  - `http_cache_entries`, which keep their existing `pruneOnOpen` 30-day env-controlled prune,
    unchanged.
- **Package `internal/retention/`** (new): `func Run(ctx context.Context, db *gorm.DB, blobs func(org uuid.UUID) blob.Store, now time.Time) (Report, error)`.
  - Deletions are batched at 500 rows. For each row, delete the blob first and then the row. A
    blob `ErrNotFound` is fine.
  - It works across all orgs; each blob delete uses that row's org store.
  - `Report` records counts per category and is logged at Info.
- **Wiring.** `gavel serve` (both modes) starts a ticker every `retention.interval` that runs it as
  a clicky task named "Retention cleanup", so it appears in the task manager.
- **Tests** (embedded PG plus local blob store): seed rows and blobs older and newer than the
  window across two orgs, run, and assert exactly the old ones are gone. Override the properties
  with `properties.Set`.

---

## M9. Sandboxed execution: executors, test/lint/commit, push checks

### M9.1 Package `sandbox/` (new)

```go
// Package sandbox runs gavel jobs in isolated Kubernetes Jobs or containers for hosted mode.
// Jobs never receive database or GitHub credentials; they report back through the run API.
type Job struct {
    RunID     uuid.UUID
    Scope     tenant.Scope
    Image     string
    Command   []string            // e.g. ["gavel","runner","exec","--run","<id>"]
    Env       map[string]string   // non-secret: GAVEL_SERVER, GAVEL_RUN_ID, GAVEL_DB=off, CI=1
    SecretEnv map[string]string   // GAVEL_RUN_TOKEN, model API keys
    Limits    Limits              // CPU "2", Memory "4Gi", Timeout 60m (defaults; per-kind overrides later)
}
type Handle interface {
    Logs(ctx context.Context) (io.ReadCloser, error) // follows until exit
    Wait(ctx context.Context) (Result, error)        // Result{ExitCode int}
    Cancel(ctx context.Context) error
    Ref() string                                     // Job name / container id
}
type Executor interface {
    Submit(ctx context.Context, job Job) (Handle, error)
    Name() string
}
```

**`sandbox/kubernetes.go`.**
- `NewKubernetes(ctx, KubernetesOptions{Kubeconfig, Namespace string}) (Executor, error)`. It gets
  the clientset from commons-db `connection.KubernetesConnection{KubeconfigConnection{Kubeconfig: ...}}.Populate(dbctx, false)`.
- `Submit` creates:
  1. A **Secret** `gavel-run-<runid>` holding `SecretEnv`, with an `ownerReference` to the Job,
     which is set after the Job is created. Create the Job first, then the Secret, then patch.
     Alternatively label it and delete it in `Wait`.
  2. A **batch/v1 Job** `gavel-run-<runid-first-8>`:
     - `backoffLimit: 0`, `ttlSecondsAfterFinished: 3600`, `activeDeadlineSeconds: Timeout`;
     - labels `app.kubernetes.io/name=gavel-runner`, `gavel.dev/org=<org>`, `gavel.dev/run=<run>`;
     - pod: `restartPolicy: Never`, `automountServiceAccountToken: false`, `enableServiceLinks: false`;
     - `securityContext`: `runAsNonRoot: true`, `runAsUser: 1000`, `fsGroup: 1000`,
       `seccompProfile: RuntimeDefault`;
     - container: `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem: true`,
       `capabilities.drop: [ALL]`;
     - volumes: `emptyDir` for `/workspace`, `/tmp` and `/home/gavel`, with `sizeLimit: 10Gi` on
       workspace;
     - resources: requests = limits = `Limits`;
     - `envFrom` the Secret, plus `env` from `Env`.
- `Logs` waits for the pod to be Running or finished, then uses commons-db
  `logs/k8s.Follow(ctx, client, Target{Pod, Container:"runner"}, Request{}, emit)`, adapted to an
  `io.Pipe`.
- `Wait` watches the Job until `Complete` or `Failed`, reads the container exit code from pod
  status, and deletes the Secret.
- `Cancel` deletes the Job with `propagationPolicy: Background`.
- **NetworkPolicy.** On startup, ensure `gavel-runner-egress` exists in the namespace. Apply it
  idempotently with server-side apply.
  - Pod selector: `app.kubernetes.io/name=gavel-runner`.
  - `policyTypes: [Ingress, Egress]`, with no ingress rules (deny all).
  - Egress:
    - `to: [{ipBlock: {cidr: 0.0.0.0/0, except: [10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 100.64.0.0/10, 169.254.0.0/16, 127.0.0.0/8]}}]`;
    - DNS to `kube-system` pods labelled `k8s-app=kube-dns` on ports 53/UDP and 53/TCP;
    - the gavel serve pods (selector from `--serve-pod-selector`, default
      `app.kubernetes.io/name=gavel`) on the serve port.
  - Also add `--sandbox-extra-deny` CIDRs for the cluster pod and service CIDRs when they aren't
    RFC1918.
  - NetworkPolicy only works with a CNI that enforces it (Calico, Cilium). Log a warning at
    startup reminding operators.
- **Tests** with `k8s.io/client-go/kubernetes/fake`: `Submit` creates a Job and Secret whose specs
  match golden expectations (security context, labels, limits, envFrom). `Cancel` deletes. This
  tests our spec generation, not Kubernetes.
- **E2E** (`GAVEL_SANDBOX_KIND_TEST=1`): against a kind cluster, run `sh -c 'echo hi; exit 3'` in
  `busybox` and assert the logs contain `hi` and the exit code is 3.

**`sandbox/container.go`.**
- `NewContainer(binary string /* "docker"|"podman" */, opts ContainerOptions{Network string /* "gavel-sandbox" */}) Executor`.
- `Submit` runs:
  ```
  <bin> run -d --name gavel-run-<id> --network gavel-sandbox --user 1000:1000 --read-only
    --tmpfs /tmp --tmpfs /workspace:exec,size=10g --tmpfs /home/gavel
    --cap-drop ALL --security-opt no-new-privileges --pids-limit 1024
    --cpus <cpu> --memory <mem> -e K=V... --env-file <0600 temp file with SecretEnv, removed after start>
    --label gavel.dev/run=<id> --label gavel.dev/org=<org> <image> <command...>
  ```
- `Logs` is `<bin> logs -f <name>`. `Wait` is `<bin> wait <name>` (which prints the exit code)
  followed by `<bin> rm <name>`. `Cancel` is `<bin> rm -f <name>`.
- A timeout is enforced by `Wait` with a context deadline, which calls `Cancel`.
- **Network check at startup.**
  - `docker network inspect gavel-sandbox` must exist; if it doesn't, fail with setup
    instructions.
  - Verify the host firewall blocks private ranges from that bridge. Look for the gavel rules with
    `iptables -C DOCKER-USER -i <bridge> -d 10.0.0.0/8 -j REJECT -m comment --comment gavel-sandbox`
    and the equivalents for 172.16/12, 192.168/16, 100.64/10, 169.254/16, plus an `ACCEPT` for the
    serve IP and port inserted **before** the rejects.
  - If the rules are missing, **refuse to start** and print the exact commands. They need root, and
    the binary doesn't run them itself.
  - Document them in `docs/sandbox.md`. Rootless podman isn't supported in this milestone; reject
    it with a clear error if detected (`podman info --format '{{.Host.Security.Rootless}}'`).
- **Tests:**
  - argument construction as a pure function, `containerRunArgs(job) []string`, with a table test;
  - an e2e test against real docker (`GAVEL_SANDBOX_DOCKER_TEST=1`) with busybox, checking logs and
    exit code.

### M9.2 Run records, run token and run API (`pr/ui/runs_api.go`, new; `sandbox/token.go`)

- **Run token.** `sandbox.SignRunToken(key []byte, runID, org uuid.UUID, exp time.Time) string` and
  `sandbox.VerifyRunToken(key, token) (runID, org uuid.UUID, err error)`. The format is
  `base64url(json) + "." + base64url(HMAC-SHA256)`. Compare with `hmac.Equal`.
- **Dispatcher** (`sandbox/dispatch.go`), `func (d *Dispatcher) Start(ctx, scope tenant.Scope, spec RunSpec) (uuid.UUID, error)`:
  1. Insert `sandbox_runs` with status `queued`.
  2. Build the `Job`:
     - `GAVEL_SERVER = --public-url`;
     - the run token with expiry = timeout + 15 minutes;
     - `Command = ["gavel","runner","exec","--run",id]`.
  3. `Executor.Submit`.
  4. Set status `running` and `backend_ref`.
  5. Stream the logs into a blob key `orgs/<org>/runs/<id>/log.txt` (tee'd to an optional
     `io.Writer`, which is how the SSH stderr gets them), and into a clicky task so the task
     manager shows them.
  6. `Wait`, then set status and `exit_code`.
- **Run API**, authenticated **only** by the run token in `Authorization: Bearer`. The token's run
  ID must equal the path ID. Mounted outside the Clerk middleware.
  - `GET /api/runs/{id}/spec` → `RunSpec{Kind, Args []string, RepoID, Ref, SHA, TodoID, Step}`, for
    the runner.
  - `GET /api/runs/{id}/source` → a `git bundle` stream of the ref (M9.3).
  - `POST /api/runs/{id}/results` → one snapshot JSON (`testui.Snapshot`), handled by
    `ingestRunSnapshot` (M5.5).
- **Tests:** token sign/verify, tampering, expiry, a token for another run (403), and every
  endpoint with a valid token against an embedded PG and a local blob store.

### M9.3 Source bundles

`sandbox/source.go`, `func (s *Sources) Bundle(ctx, repo Repo, ref string, w io.Writer) (sha string, err error)`:

- **source=push:** run `git -C <bare> bundle create - <ref>` from
  `<data-dir>/repos/<org>/<repo>.git`.
- **source=github:** maintain a mirror at `<data-dir>/mirrors/<org>/<github_repo_id>.git`:
  - `git clone --mirror` the first time;
  - `git fetch --prune origin` each time, authenticating with
    `-c http.extraHeader="Authorization: Basic base64(x-access-token:<installation token>)"`.
    The token is never written to git config.
  - Then `bundle create - <ref>`.
- Also resolve and return the SHA (`git rev-parse <ref>`).

### M9.4 `gavel runner exec` (hidden command, runs inside the pod)

File `cmd/gavel/runner.go`. Cobra command `runner exec --run <id>`, `Hidden: true`, MCP-excluded.
It reads `GAVEL_SERVER`, `GAVEL_RUN_TOKEN` and `GAVEL_RUN_ID`, and requires `GAVEL_DB=off`.

1. `GET spec`.
2. `GET source` into `/workspace/src.bundle`, then `git clone /workspace/src.bundle /workspace/src`
   and `git checkout <sha>`.
3. Depending on `spec.Kind`:
   - `test` / `lint`: exec `gavel test|lint --work-dir /workspace/src <spec.Args...>` (no `--ui`).
     Afterwards, `POST results` for every `.gavel/run-*.json` created.
   - `push-check`: if `.gavel.yaml` has `ssh.cmd`, run `sh -c "$SSH_CMD"` in `/workspace/src`;
     otherwise run `gavel test --lint`. `POST results` as for test.
   - `agent`: see M10.
4. Exit with the job's exit code.


### M9.5 Routing existing execution through the dispatcher (clerk mode only)

- **Project actions.** `pr/ui/project_lifecycle.go:126`'s `executeProjectAction` is a package
  `var`.
  - In hosted mode, `Server` uses an injected
    `actionRunner func(ctx, scope, project, args []string, output io.Writer, group *clickytask.Group, opts ...) clickytask.TypedTask[cexec.ExecResult]`
    that calls `Dispatcher.Start` with `Kind` from `args[0]` (`test` or `lint`) and waits. This
    lifts the M3.3 `501` for project actions. `commit` and `open-pr` keep returning `400` (§15).
  - Don't mutate the package var per org. Add a Server field `actionExecutor projectActionExecutor`,
    defaulting to `executeProjectAction`, and use it at the call sites (L242,
    `project_commit_queue.go:140`).
  - Hosted projects have no `Dir`: pass the repo ID through `project.Name` → `repos` lookup.
- **Push checks.** `OnPushed` (M6.2) replaces its placeholder and calls `Dispatcher.Start(Kind: push-check)` for each updated
  branch ref (skip deletions). It streams the logs to the SSH session's stderr until the run
  finishes (`tee`), then prints PASSED or FAILED. The push itself has already been accepted
  (post-receive semantics).
- **Todo, agent and fixture runs.** See M10. They keep the M3.3 `501` until M10 lands.
- **Runner image** (`sandbox/runner.Dockerfile`, new):
  - multi-stage: build gavel, then `FROM` a toolchain image;
  - Go 1.26, Node 22 + pnpm, Python 3.12, git, yq, and the agent CLIs (`npm i -g @anthropic-ai/claude-code @openai/codex`; pin exact versions);
  - user `gavel` uid 1000, `HOME=/home/gavel`, workdir `/workspace`;
  - `ENTRYPOINT ["gavel"]`.
  - **Stop and ask Batman** whether to base this on captain's `captain container` generator
    (`captain/pkg/container`, `Generate`/`Build`) instead.
  - Add a `make runner-image` target if a Makefile exists; check the repo's build tooling first.

**Tests:**
- the dispatcher with an in-memory `Executor` that runs a Go func: status transitions, log blob
  written, task created;
- project action in hosted mode → dispatcher invoked with the right kind and args;
- push-check `OnPushed` → one run per updated branch;
- E2E with docker (`GAVEL_SANDBOX_DOCKER_TEST=1` plus an image built locally): push a small Go repo
  with a failing test, see FAILED and a stored snapshot. Then confirm from inside the container that
  `curl http://169.254.169.254` and `nc -z <postgres-ip> 5432` both fail, and that
  `curl https://proxy.golang.org` succeeds.

---

## M10. Agent runs inside the pod

### M10.0 Spikes (report each to Batman before coding)

1. **Interfaces the run uses.** List every interface the lifecycle type-asserts on the provider:
   ```
   grep -rn "\.(todos\.\|\.(interface" todos/lifecycle todos/run todos/*.go
   ```
   The full set of `todos.Provider` plus capability interfaces is what the pod-side remote
   provider must implement. Known so far: `Provider`, `RunLifecycleProvider`,
   `RunProgressProvider`, `RunNoticeProvider`, `RunIterationProvider`, `PlanContentProvider`
   (`todos/provider.go:51-200`), plus whatever plan persistence `OnOutcome` uses
   (`PersistAndSelectPlan`). Produce the exact list.
2. **Transcript ingest.** Captain only ingests transcript files located under
   `~/.claude/projects` and `~/.codex/sessions` (`pkg/monitor/ingest.go` `ingestFile` and
   `hookTranscriptRoot`). Its builder functions `claudeIngestInput` and `codexIngestInput` are
   unexported.
   - Propose an **upstream captain change**: export
     `monitor.BuildIngestInput(source, path string) (database.IngestTranscriptInput, error)`,
     which takes any path.
   - Batman must approve the captain PR. Gavel then bumps captain.
3. **Approval broker.** Interactive tool approvals (`/api/todos/session/approve`, `ApprovalBroker`)
   cross the pod boundary. Confirm hosted agent runs use `permissions.mode` without interactive
   approval for v1, or design a long-poll endpoint. **Batman decides.**

### M10.1 Remote provider (`todos/remote/`, new; pod side)

```go
// Package remote implements the TODO provider over the gavel run API so agent runs
// inside a sandbox persist state without database credentials.
type Provider struct { Server string; RunID uuid.UUID; Token string; HTTP *http.Client }
```

- It implements exactly the interfaces from spike 1.
- Each method is a `POST /api/runs/{id}/provider/<method>` with a JSON body mirroring the Go
  arguments, returning a JSON result.
  - `*types.TODO` round-trips as JSON. Check that `types.TODO` marshals losslessly, and add a
    round-trip test.
- **Serve side** (`pr/ui/runs_provider_api.go`): each endpoint decodes the body and calls the same
  method on `todoruntime.OpenForOrg(ctx, opts, scope{org from token})`. The token's run is bound to
  one todo (`sandbox_runs.todo_id`), and every call must reference that todo's ID; otherwise `403`.
- **Transcripts:**
  - In the pod, after the agent CLI exits and every 10 seconds during the run, the runner scans
    `~/.claude/projects/**` and `~/.codex/sessions/**` for files modified since the last scan.
  - It POSTs them to `POST /api/runs/{id}/transcripts?source=claude|codex&name=<relative path>`
    with the raw file bytes (full file each time; ingest is idempotent by source path and offset).
  - Serve writes them to `<data-dir>/transcripts/<org>/<run>/<source>/<name>`, then calls the
    upstream-exported `monitor.BuildIngestInput(source, path)` followed by
    `captaindb.IngestTranscript`, and records `session_owners` for the resulting session.
- **Run flow in the pod** (`runner exec`, kind `agent`):
  1. Fetch the spec and source as in M9.4.
  2. Build `run.Request{Provider: remote.Provider{...}, Todo: <GET via provider>, Dir: /workspace/src, Options: spec.Options}`.
  3. Call `run.Start(req)` and await `Done`.
  4. Commit hooks (`commit.AgentHooks`) run inside the pod, so they're sandboxed. Force
     `Push=false` in hosted mode. The runner then runs
     `git -C /workspace/src bundle create /workspace/out.bundle <base-sha>..HEAD` and posts it to
     `POST /api/runs/{id}/bundle`; serve pushes (below).
  5. Upload the transcripts a final time, then exit.
- **Serve side, starting runs.** In clerk mode, `/api/todos/run` and its siblings stop calling
  `run.Start` in-process. They call `Dispatcher.Start(Kind: agent, TodoID, Step, Options)` and
  return the same `StartResult`-shaped JSON. `SessionID` comes from `PrepareRun`, which the pod
  calls through the provider API. For the HTTP response, return the run ID and let the UI pick up
  the session through existing polling. **Check what the UI expects from `/api/todos/run`'s
  response (`pr/ui/src`) before changing it, and stop and ask if it needs the session ID
  synchronously.**
- **Serve side of `POST /api/runs/{id}/bundle`** (run-token authenticated, like the rest of the
  run API):
  - `git -C <bare or mirror> fetch <bundle> HEAD:refs/gavel/runs/<id>`.
  - Then push to the target:
    - pushed repo: update the target branch in the bare repo with `git update-ref`;
    - GitHub: `git push` with the installation-token header to `spec.Branch`. When the run asks
      for a PR, also create it via the API with the installation token.
  - Check: reuse `commit/push.go` and `pr_create.go` logic where it is git-only. Read them first.
    **If the PR-creation logic can't run without a working tree, stop and ask Batman.**
- **Model credentials.** `--sandbox-secret-env ANTHROPIC_API_KEY --sandbox-secret-env OPENAI_API_KEY`
  copies those env values from serve into the run's Secret. Per-org keys are out of scope.

**Tests:**
- remote provider ↔ serve endpoints round trip (an `httptest` server backed by an embedded-PG
  runtime provider): `PrepareRun`, `RecordRunStart`, `UpdateState`, `SaveAttempt`, and wrong-todo
  `403`;
- transcript upload → ingest (after the captain bump), against a fixture claude JSONL;
- E2E (kind or docker, with a stub agent CLI on `PATH` in the image that echoes a scripted
  transcript): a todo run reaches a completed state and its transcript is visible in the dashboard.

---

## 11. Out of scope (don't build)

- agent-substrate integration.
- Per-org secrets (model keys and similar).
- Postgres row-level security.
- Rootless podman.
- GitHub webhooks: repo sync is polling.
- API tokens for non-GitHub CI.
- A pre-receive hook that rejects failed pushes. Pushes are accepted and checks report
  asynchronously.

## 12. Verification matrix (final, after M10)

Ask Batman before running anything.

| Area | How |
|---|---|
| Unit | `go test ./tenant/... ./api/... ./blob/... ./sandbox/... ./serve/... ./pr/ui/... ./todos/remote/... ./internal/retention/...` |
| DB / tenancy | `go test ./internal/database/... ./github/cache/... ./internal/cache/... ./internal/taskhistory/... ./todos/native/... ./todos/runtime/... ./git/gitstate/...` (embedded PG via `commons-db/dbtest`, always on) |
| Frontend | `pnpm --dir pr/ui test` and `pnpm --dir pr/ui run check` |
| Docker sandbox | `GAVEL_SANDBOX_DOCKER_TEST=1 go test ./sandbox/... ./serve/...` |
| K8s sandbox | `GAVEL_SANDBOX_KIND_TEST=1 go test ./sandbox/...` (kind with Calico or Cilium for NetworkPolicy) |
| Manual, none mode | `gavel serve` behaves exactly as before: PRs, todos, project actions, proc, attachments. |
| Manual, clerk mode | `gavel serve --auth clerk --sandbox docker --ssh-addr :2222 ...` against a Clerk dev instance, then run the checks in the table below. |

Clerk-mode manual checks:

| # | Step | Expected |
|---|---|---|
| 1 | Sign in. | The badge shows your name. |
| 2 | Switch org. | The PR, todo and run lists change. |
| 3 | Connect GitHub (App install). | Repos appear. |
| 4 | Add an SSH key in settings, then `git push gavel` to `ssh://host:2222/<org-slug>/demo`. | The push check runs in a container; logs stream. |
| 5 | Push from an unregistered key. | Rejected. |
| 6 | Request org A's todo with org B's session. | 404. |
| 7 | Run a GitHub Action with `gavel-server` set. | Results show up. |
| 8 | Run the retention job with `retention.test_runs=1m`. | Old runs are gone. |

## 13. Spike log

*(The implementing session fills this in. One entry per spike: date, finding, Batman's decision.)*

- **2026-10-13: planning decisions (Batman).**
  - `retention.task_history` defaults to **30d**, the existing `taskhistory.Retention` value, so
    none mode is unchanged. Every other retention key keeps 60d.
  - M3.3's all-mode changes are accepted on purpose: stripping `setup`/`sandbox`/`cliArgs` in
    `validateTodoRunSpec`, and dropping the attachment CORS header.
  - Milestones may be reordered when that keeps each merged state complete and safe.
  - Delivery is phased across two PRs (three until the 2026-10-14 entry below). Each merged state
    must leave none mode working and clerk mode safe:
    1. **PR 1: M0–M6, on branch `api-auth`.** Tenancy in the schema and stores, Clerk login, org
       switcher, GitHub App, CI uploads, SSH keys, authorized push. Execution in clerk mode
       returns 501. The migration is one-way (PK changes break an older binary's upserts);
       Batman accepted this: nobody goes back to an older binary after running this one.
    2. **PR 2: M7–M10, on a separate branch.** Blob store, retention, sandboxes, agents.
  - Reorders so PR 1 is safe on its own: the execution 501s move from M9.5 into M3.3, and
    clerk-mode push never runs host hooks (M6.2 placeholder `OnPushed`).
  - Hosted `commit`/`open-pr` project actions return 400; commits come only from agent runs.
    Bundle upload and serve-side push move from M9 into M10, so model keys
    (`--sandbox-secret-env`) are only needed in M10. Recorded in §15 to revisit.
  - M0: the dist/lockfile/catalog changes stay (see M0 item 1). The existing `api-auth` working
    tree is **not** committed first; Batman considers it incomplete, and M1 builds on it in place.
- **2026-10-13: test infrastructure (Batman).** Auth tests use a fake Clerk issuer,
  `api/clerktest`. It serves OAuth discovery, the device and token endpoints, and JWKS over TLS,
  signing with a generated key. Extend it, not ad-hoc `httptest` handlers, for M2–M6 tests: session
  tokens, Backend API memberships, and GitHub OAuth tokens. `cmd/gavel` reaches it through the
  `authHTTPClient` package variable.
- **2026-10-13: M0 done.**
  - `printAuthIdentity(cmd, identity, migrate)`: login migrates, whoami doesn't.
    `IdentityStore.Ready` (`to_regclass('public.users')`) skips Sync on an unmigrated DB.
  - Tests: the `whoami` end-to-end test against `clerktest`, the migrate-option table test, and
    `Ready` before and after migrations (embedded PG).
  - `api/auth_test.go` and `auth_failure_test.go` now use `clerktest`, which records posted forms
    (`Forms`), enforces grants, and has `DenyLogin`/`SignWithKeyID`.
    `TestRejectDiscoveryAndDeviceURLs` keeps its own server, since it models a malicious issuer.
  - Open: embedded-PG tests print initdb/postgres logs (commons-db `StartEmbedded`), which breaks
    the pristine-output rule. Fix: `commonsdb.EmbeddedConfig.Logger` exists, so M1 adds one shared
    embedded-PG test helper that buffers that output and logs it only when the test fails.
- **2026-10-13: test-run permission (Batman, this session).** Scoped test runs don't need
  asking; full test suites and full builds still do.
- **2026-10-13: M1.1 done; embedded-PG helper added.**
  - `tenant/` has `LocalOrgID`, `Scope`, `Local`, `WithScope`, `FromContext` and `IsLocal`, with
    tests.
  - `internal/database/dbtest` has three helpers:
    - `Postgres(t, name)` buffers the server output (`FastTesting` on) and logs it only on failure;
    - `Quiet(db)` turns off gorm error logging;
    - `RequireConstraintViolation(t, err, name)` asserts the exact constraint that fired.
  - `identity_integration_test.go` uses all three, so its intentional FK failures are asserted, not
    printed. One `INF Starting embedded postgres` line remains: it goes through commons' global
    logger.
  - **Decision (Batman):** M1's tests use `dbtest`. Moving the ~20 existing `StartEmbedded` call
    sites (several Ginkgo) onto it is a separate cleanup step after M1.
- **2026-10-13: M1.2 done.**
  - **HCL:** the 20 existing `org_id` columns are now `NOT NULL DEFAULT` the local org, and their
    org FK blocks are gone. `violations` gains `org_id` plus `violations_file_scan_fkey
    (org_id, file_path)`. Every PK and unique key from the M1.2 table includes `org_id`.
  - **`tenancy.hcl`:** six new tables. Their org FKs stay in the HCL: they have no existing rows,
    so the add-column-plus-FK problem doesn't apply.
  - **`095_tenancy_prepare.sql` (pre):** drops `violations_file_path_fkey`. It uses `to_regclass`
    inside the subquery because a `::regclass` cast is folded at plan time and raises on a fresh DB
    even behind an `IS NOT NULL` guard.
  - **`151_tenancy_local_org.sql` (post):** inserts the local org and adds the 21
    `<t>_org_id_fkey` constraints.
  - **Exclusions:** extracted to the package var `schemaExclusions` (`database.go`).
    `TestTenancySchemaBundle` checks that every table in 151's array has its FK excluded, since a
    missed exclusion makes Atlas drop and re-add it on every run.
  - **Upgrade test:** `tenancy_migration_integration_test.go` applies
    `testdata/schema_main/`, a snapshot of `main@edc252b`'s schema with main's exclusions. It seeds
    rows, migrates, and asserts:
    - existing rows land in the local org, and the local org row exists;
    - every tenant table has NOT NULL `org_id` and exactly one org FK;
    - an unknown org is rejected;
    - per-org duplicate keys work (seen_prs, workspaces, global labels, file_scans);
    - the composite violations FK holds and cascades within one org only;
    - a second run is a no-op.
  - **Atlas plan reviewed** (trace log): no column or table drops. The only drops are the old PKs
    and the `linter_executions_unique`, `todo_workspaces_repo_key_key` and
    `todo_labels_global_name_key` indexes, each recreated with `org_id` in the same plan.
  - **Not ours:** captain's own `Migrate` isn't idempotent. Re-running it drops and recreates six
    `captain_*` partial indexes and drops the `captain_sessions_parent_relation_valid` check
    without re-adding it. This happens on main too. Raise it with captain separately.
  - **Expected failure until M1.3:** `TestHCLMigrationFromExistingGitHubSchema` fails in
    `LinterStats.RecordExecution`, whose `ON CONFLICT (linter_name, work_dir, executed_at)` no
    longer matches a unique key.
- **2026-10-13: M1.3 github/cache done.**
  - `Store` has `org` (the local org from `Open`/`Shared`) and `ForOrg`, which shares the DB handle
    and a pointer `writeMu`.
  - All nine models have `OrgID` (PK part) and `UserID`.
  - Every query filters by `org_id`. Every upsert conflicts on `org_id` plus its key and updates
    only the spec's payload columns, so `UpdateAll` is gone except on shared favicons.
  - `pruneOnOpen` stays global, with a comment saying so.
  - `tenancy_test.go` covers every public method across two orgs, including that an upsert keeps an
    existing row's `user_id` (§3 #2).
  - **Open for M3:** `Store.Status()` counts rows table-wide. Once it's served per org in hosted
    mode, it reveals other orgs' row counts. Scope the counts (or hide them) when M3 wires org
    servers.
- **2026-10-13: M1.3 internal/cache done.**
  - `ViolationCache` and `LinterStats` have `org`, `ForOrg`, and a `rows()` helper that every read
    goes through.
  - `ClearCache`, the global branch of `ClearViolations`, and the scan cleanup are now org-scoped
    deletes, so `AllowGlobalUpdate` is gone.
  - `GetStats` size is commented "storage size across all orgs".
  - Both `OnConflict` targets include `org_id`.
  - **Pre-existing bug fixed:** `GetAllViolations`, `GetViolationsBySource` and
    `GetViolationsBySources` ordered by unquoted `column`, a PostgreSQL reserved word, so they
    always failed on Postgres. Nothing called them under test before. They now use
    `violationOrder` with `"column"` quoted.
  - **Test fixture fix:** `TestIdentitySchemaAndPersistence` and
    `TestHCLMigrationFromExistingGitHubSchema` built their "legacy" schema by `AutoMigrate`-ing the
    current `githubcache` structs. Those now carry `OrgID` primary keys, so the fixtures stopped
    being legacy and wrote zero-UUID orgs. `internal/database/legacy_models_test.go` freezes the
    pre-tenancy struct shapes from main@edc252b, and both tests use them.
  - `internal/database`, `internal/cache` and `github/cache` pass with embedded PG.
- **2026-10-13: M1.3 taskhistory done.**
  - `Store` has `org` and `ForOrg`.
  - `importRecord` inserts `org_id`, and its `DO UPDATE` is guarded by
    `task_run_history.org_id = EXCLUDED.org_id`. When another org imports the same run ID, the
    import is a silent no-op, not an overwrite. That's acceptable because run IDs are globally
    unique.
  - `ListRuns` and `Snapshot` filter by `org_id`. `Prune` stays global.
  - `tenancy_test.go` is a plain Go test beside the Ginkgo suite, because `dbtest` needs a
    `testing.TB`.
- **2026-10-13: M1.3 todos/native done.**
  - **Filtering:** `Repository` has `org`, `ForOrg`, and `inTx` (same org over a transaction).
    Every read is filtered per query, as the spec asks, through two fragments:
    - `inOrgWorkspace(col)` for `workspace_id` columns;
    - `inOrgIssue(col)` for `issue_id` columns.
  - **Central helpers:** `getIssue`, `lockIssue`, `lockExecutionIssue` and `issueWorkspace` take
    `org`. Every write that locks an issue is therefore org-checked, and another org's issue is
    `ErrNotFound`.
  - **Workspace checks:** `requireWorkspace` guards `CreateIssue` and workspace-scoped
    `SetLabelDefinition`/`DeleteLabelDefinition`. Without it, `DeleteLabelDefinition` given another
    org's workspace ID would strip that label from all of that workspace's issues.
  - **Labels:** global labels conflict on `(org_id, name)`, and reads filter `todo_labels.org_id`
    directly.
  - **LaunchCoordinator:** uses `c.repository.org`, not a separate field: same effect, no
    duplicate state. Its four in-transaction `NewRepository(captainTx.Gorm())` calls are now
    `c.repository.inTx(...)`.
  - **Attribution:** `requestUser(tx)` reads `tenant.FromContext(tx.Statement.Context).UserID`. It
    fills:
    - `todo_workspaces.user_id`;
    - `todo_issues.created_by_user_id`;
    - `todo_issue_events.actor_user_id`, including label-strip events;
    - `todo_issue_prompt_runs.requested_by_user_id`, including `ExecutionIntegration`'s insert;
    - `todo_issue_plans.created_by_user_id`;
    - `todo_labels.updated_by_user_id`, also refreshed on conflict.
  - **Session owners:** `Repository.CreateOwnedSession` (in `ownership_sessions.go`) wraps every
    `CreateOrGetSession`, 9 sites including `todos/runtime/lifecycle_runs.go`, and calls
    `recordSessionOwner` (`ON CONFLICT DO NOTHING`). The runtime site isn't transactional, but both
    writes are idempotent.
  - **Tests:** `tenancy_test.go` covers every exported Repository/ExecutionIntegration method
    across two orgs, plus attribution and session owners. A mutation check (removing the filter
    from `ListEvents`, then from `ClaimPromptRunOwner`) failed the test naming each method, so the
    coverage is real. `todos/native`, `todos/runtime` and `todos/portable` pass.
  - **Open:** `todos/portable` builds its own `native.NewRepository(tx)`, so it imports and exports
    as the local org. That's fine for the CLI; route it through the org if hosted mode ever exposes
    it.
- **2026-10-13: M1.3 runtime and identity done; M1 code complete.**
  - `todos/runtime`: `Open` and `New` act for the local org. `OpenForOrg(ctx, options, scope)`
    resolves the shared DB and delegates to the new `NewForOrg(ctx, db, options, scope)`.
    `NewForOrg` exists because `database.Shared` caches one process-wide handle that tests can't
    reset, so a test through `Open` depends on test order. M3's per-org factory will also want to
    own its handle. `OpenGlobal` stays local.
  - `api`:
    - `Ownership.OrgID` is `uuid.UUID`, and the local org when the identity has no active org.
      Clerk mode rejects org-less requests before Sync (M3).
    - `upsertOrg` is shared by `Sync` and the new `SyncMemberships`, which replaces the user's
      `org_memberships` rows and keeps dropped orgs.
    - `Membership{OrgID, OrgSlug, OrgName, Role}` is defined in `clerk_organizations.go`, for M2's
      `UserMemberships`.
  - `api/OWNERSHIP.md` describes the local org, the NOT NULL org keys, the scoped stores and the
    security boundary.
  - Passing with embedded PG: `internal/database`, `internal/cache`, `github/cache`,
    `internal/taskhistory`, `todos/native`, `todos/runtime`, `todos/portable` and `api`.
- **2026-10-13: M1 accepted.**
  - `go build ./...` succeeds.
  - `go test ./...`: 69 packages pass. The only failure was two `procfile` supervisor specs hitting
    their 3s `Eventually` timeouts under full-suite load. `procfile` then passed 3/3 in isolation
    and in the embedded run.
  - `GAVEL_DB_EMBEDDED_TEST=1 go test ./...`: all 70 packages pass.
  - **Environment note:** `/tmp` is an 8 GB tmpfs, too small for `go test ./...`'s parallel build
    output. Full runs use `GOTMPDIR=$HOME/.cache/gavel-gotmp`. The first attempts also failed on a
    full root disk until Batman freed space.
  - **Watch:** `procfile` supervisor specs are load-sensitive. If they fail again under a full run,
    raise them as a flaky-test issue.
- **2026-10-13: Rebased PR1 onto origin/main (93639d4).** Batman asked for the rebase, preferring
  main where it doesn't collide with ours, and decided the collisions:
  - **Commit-stat cache:** main replaced `commit_stat_caches`/`commit_stat_cursors` with
    `git_range_stats` and deleted `github/cache/commit_stats.go`. We took the deletion and dropped
    those tables from our HCL, 151, the exclusions and the tests.
  - **Captain FKs (Batman):** main's rule is that gavel installs no DDL on Captain's tables (it
    switched its own FKs to Captain delete guards). Our cascade FKs from the ownership sidecars
    went with `150_identity_captain_constraints.sql`. Every migration now deletes sidecar rows
    whose Captain resource is gone (`sweepOwnershipOrphans`, a Go step after
    `registerCaptainDeleteGuards`). It is Go rather than a `-- runs: always` script because
    main's bundle test forbids always-run scripts (Batman's call). Databases that ran this branch
    lose the old FKs through Atlas, since they're no longer excluded.
  - **Git state (Batman):** main's `git_repos` gets `org_id`/`user_id`, unique
    `(org_id, root_dir)`, and an org FK in 151. `gitstate.Store` has `ForOrg`. Every read is
    filtered, and writes that bump a repo require it to be the org's (`gitstate.ErrNotFound`). The
    tracker uses the local org until M3 wires hosted mode.
  - **Embedded-PG tests (Batman):** main moved to `commons-db/dbtest` (always on, no env gate). Our
    tests use `dbtest.ForT`, and `internal/database/dbtest` keeps only `Quiet` and
    `RequireConstraintViolation`.
  - **Main's new todos code** is org-scoped: parent/child issues (a parent in another org is
    `ErrNotFound`, so its workspace isn't revealed), landings, `ResolveComment`,
    `RecordAskAnswer`, the multi-workspace `CountIssuesByStatus`, `LookupWorkspaces`, and the plan
    coordinator links. A rejected link rolls back Captain's plan change in the same transaction.
  - Main creates Captain sessions through `sessiontree.Ensure`/`EnsureTx` and admits runs through
    Captain's `promptrun.Recording`, so `CreateOwnedSession` became `recordSessionOwners`. It is
    called in those link hooks and in `AttachPromptRun`, and records each session's root as well.
  - Main's link hooks built repositories with `NewRepository(tx.Gorm())`, which is the local org;
    they now use the coordinator's repository.
  - `internal/database/testdata/schema_main` is re-snapshotted from 93639d4, with main's exclusions.
  - Lockfiles, `dist/` and workspace files are main's (main already has ajv and jsonc-parser).
- **2026-10-13: M2.0 spike, CLI access token format.** Inspected the access token from a
  `gavel login` against Batman's Clerk dev instance (`adequate-wolf-6.clerk.accounts.dev`), with a
  script that never printed the raw token.
  - It is a JWT: header `{"alg":"RS256","kid":"ins_…","typ":"at+jwt"}`. Its `kid` is in the
    instance's `/.well-known/jwks.json`, and the RS256 signature verifies against that key.
  - Payload claims: `iss` (the instance), `sub` (Clerk user ID), `client_id` (the public OAuth
    client), `scope` (`openid profile email offline_access`), `jti` (`oat_…`), `iat`, `nbf`, `exp`.
  - **No `aud`, no email, no org claims.** The org therefore comes from `X-Gavel-Org`, as planned.
  - Lifetime is 24 hours (`exp - iat = 86400`). A refresh token is issued.
  - **Decision (per the rule):** verify locally with the JWKS. `/oauth/userinfo` isn't needed.
    The verifier must check `typ` and `client_id`, because ID tokens and session tokens share the
    signing keys (M2.1 updated).
  - Follow-up decision with Batman: also ask Clerk whether the token is still valid, cached for
    up to 10m with a fixed (not sliding) expiry, to bound revocation lag (M2.1, §15).
  - The userinfo probe wasn't conclusive and didn't need to be: the token had expired an hour
    earlier (401). Separately, Cloudflare in front of Clerk blocked Python-urllib's default user
    agent (`403 error code: 1010`). Go's default client worked for `gavel login`, but if a Clerk
    call ever returns that 403, the user agent is the cause.
- **2026-10-14: PR split changed to two PRs.** Batman's call: M0–M1 and M2–M6 ship together as
  one PR from `api-auth` (its M0–M1 commits stay as they are); M7–M10 get their own PR on a
  separate branch. The delivery list in the 2026-10-13 planning entry is updated.
- **2026-10-14: Atlas plan for the `git_repos` key change reviewed.** Captured with `LOG_LEVEL=trace`
  on `TestTenancyMigrationFromMainSchema`. On main's schema, Atlas runs: `DROP INDEX
  git_repos_root_dir_key`; `ADD COLUMN user_id uuid NULL, ADD COLUMN org_id uuid NOT NULL DEFAULT
  <local org>` plus the user FK (a constant default, so no table rewrite); `CREATE INDEX
  git_repos_owner_idx`; `CREATE UNIQUE INDEX git_repos_org_root_dir_key (org_id, root_dir)`.
  Existing rows are already unique on `root_dir`, so the new key can't fail. An older binary's
  `ON CONFLICT (root_dir)` stops working once the old index is dropped, which is covered by the
  one-way migration decision.
- **2026-10-14: M2.1 Clerk Backend API endpoints confirmed.** Source: Clerk's OpenAPI spec
  (`clerk/openapi-specs`, `bapi/2026-05-12.yml`, server `https://api.clerk.com/v1`) and Clerk's JS
  backend SDK (`API_URL` + `/v1` + path). Clerk's guide page shows the verify URL without `/v1`;
  the SDK and spec both use `/v1`, so we do too. An unauthenticated probe returns 401 on both, so
  it can't tell them apart.
  - **Verify:** `POST /v1/oauth_applications/access_tokens/verify`, `Authorization: Bearer
    <secret key>`, body `{"access_token": "..."}`. It accepts JWT access tokens.
    - 200 has two shapes: `{object: "clerk_idp_oauth_access_token", id, client_id, subject, scopes,
      revoked, revocation_reason, expired, expiration, ...}`, or `{"active": false}` for an
      inactive JWT.
    - 404 for an unknown token, 400 for a bad request.
    - **Surprise:** a revoked token still returns 200, with `revoked: true`. The verifier accepts
      only a 200 token object with `revoked` and `expired` false, and with `subject` and
      `client_id` matching the JWT's `sub` and `client_id`. `{active:false}` and 404 are
      rejections. Any other status is an error and isn't cached.
  - **Memberships:** `GET /v1/users/{id}/organization_memberships?limit=&offset=` returns
    `{data: [OrganizationMembership], total_count}`. Pagination is by `offset` (`limit` up to 500),
    not a cursor. Each membership has `role` and `organization{id, slug, name}`.
  - **GitHub token:** `GET /v1/users/{id}/oauth_access_tokens/oauth_github` returns a bare JSON
    array (no `data` wrapper) of `{token, expires_at, scopes, provider, ...}`. Clerk refreshes an
    expired OAuth 2.0 token transparently.
- **2026-10-14: M2.1 done.**
  - `api/jwks.go`: `JWKSCache` (10m TTL, an unknown kid refetches at most every 30s, the same key
    validation as before) and `parseSigned`. `AuthClient.verifyIdentity` uses the cache. A JWKS
    fetch failure still surfaces as its own error, with a test that fails when the distinction is
    removed.
  - `api/session_token.go` `SessionVerifier`, `api/access_token.go` `AccessTokenVerifier`,
    `api/github_oidc.go` `GitHubOIDCVerifier`, `ClerkOrganizations.UserMemberships`, and
    `api/clerk_users.go` `GitHubToken`. `ClerkOrganizations` has one Backend API `call` helper,
    which `ResolveOrganization` now uses too. Its error texts say "Clerk Backend API".
  - Access token cache: `commons-db/cache` (go-cache), each entry set with
    `store.WithExpiration(min(ttl, exp - check time))`. Getting an entry never extends it. A
    lifetime of zero or less isn't cached, because go-cache reads 0 as "use the default TTL".
    `github.com/eko/gocache/lib/v4` becomes a direct dependency (`go mod tidy`).
  - `api/clerktest` gained `SessionToken`, `AccessToken` (a unique `jti` per token), `SignWithType`,
    `PublishKey`, `JWKSReads`, `RevokeAccessToken`, `Memberships`, `GitHubToken`,
    `BackendRequests`, and `BackendClient`.
  - Mutation checks: the `at+jwt` guard, the `exp` cap on cache entries, and the local `client_id`
    check are each caught by a test.
  - Tests: `go test ./api/...` plus the `cmd/gavel` auth tests pass. Not yet wired into
    `gavel serve`; that's M3.
- **2026-10-14: Batman on M2.1 follow-ups.**
  - Reject a session that is pending on Clerk's side. A session with no org picked is rejected by
    M3.3 (`409 no_active_org`), as already specced.
  - "Give descriptive error messages wherever you can." Rejections now wrap `ErrInvalidToken`
    with the reason, which the tests assert. A 401/403 from the Backend API adds "check the Clerk
    secret key".
  - Commits: commit whenever a change makes sense on its own, not strictly per milestone.
- **2026-10-14: M3.1 deviations from the spec.**
  - The options type is `DashboardOptions` (`cmd/gavel/dashboard_options.go`), because
    `ServeOptions` already belongs to `gavel ssh serve`. It embeds `PRListOptions`; clicky flattens
    embedded structs (`flags.parseStructFieldsRecursive`).
  - clicky has no env-var tag, so `withEnv(os.Getenv)` fills empty `--clerk-*` flags and reads
    `CLERK_SECRET_KEY` and `GAVEL_SECRET_KEY` into untagged fields, which have no flag.
  - Each flag lands with the milestone that uses it. M3 adds `--auth`, `--public-url`,
    `--clerk-issuer`, `--clerk-publishable-key`, `--clerk-client-id` and `--trusted-proxy`. M5 adds
    the GitHub App flags, and M6 adds `--ssh-addr`, `--ssh-host-key` and `--data-dir`. The
    `--sandbox*`, `--kubeconfig` and `--sandbox-secret-env` flags wait for M9/M10 (PR 2). Also,
    clicky can't tell an explicit `--sandbox` from its default, so "reject `--sandbox` with
    `--auth none`" needs that flag to have no default; M9 should keep that in mind.
  - `Validate` makes no I/O and reports all problems at once (`errors.Join`). It also checks that
    `--public-url` is an origin, that the issuer is https, the `pk_`/`sk_` prefixes, that
    `GAVEL_SECRET_KEY` is base64 of at least 32 bytes, and that each `--trusted-proxy` is a CIDR.
    The "DB required in clerk mode" check runs at startup with `database.Require`, because it
    needs the database.
- **2026-10-14: hosted routes are an allowlist (Batman).** A review of every route in
  `Server.Handler()` found that the spec's denylist (`/api/proc/*`, task control, global settings,
  project diff, `/debug/*`) left host-wide state reachable to any signed-in tenant: Captain's
  session handler (every org's transcripts), the git, test-run and results endpoints (the host's
  checkouts and files), `/api/activity*` (every org's GitHub calls), and `/api/status`'s DSN
  source and table counts. Batman chose the allowlist; M3.3's "Features off" section is rewritten
  with the list. The hosted dashboard therefore starts without the sessions, git and test-run
  views, which return when M9/M10 make them org-aware.
  - **How the gate works:** `hostedRoutes` asks the dashboard's own mux which pattern a request
    resolves to (`mux.Handler(r)`) and allows it only if that registered pattern is listed. A
    first version matched with a separate `ServeMux` of allowed patterns. A test caught that
    `GET /api/projects/{name}` then admitted `/api/projects/git-summary`, which the real mux
    routes to a blocked handler. Resolving through the real mux removes that class of mismatch.
  - With hosted mode on, `Handler()` panics if an allowlist entry isn't a registered pattern, so
    a typo can't turn a route off unnoticed. Every hosted test runs this check.
  - Hub subscriptions go through the gate. The hub accepts a subscription with 204 and then
    closes it with a `<sub>/__closed` frame carrying the 404.
- **2026-10-14: the GitHub client's cache is per org (`github.Options.Cache`).** `cachedGet`, the
  completed-run, job-log and workflow-definition caches all used `cache.Shared()`, the local
  org's store. Job logs and completed runs are served from that cache by ID without calling
  GitHub, so in hosted mode one org could have read another's private CI logs by job ID. On a
  network error, `cachedGet` would also have served another org's stale responses.
  `Options.Cache` is the org's store (`cache.Shared().ForOrg(org)` in the M3.2 factory), and nil
  keeps the local-org cache. `cachedGet` takes the store explicitly. A test covers each path
  with two orgs, and fails if `Options.Cache` is ignored. The github package has no other
  process-wide caches.
- **2026-10-14: `Server` org scoping (M3.2).**
  - `Server.SetOrg(org, db)` takes the org's database handle along with the org (the spec had
    `SetOrg(org)`), because pr/ui wires its stores explicitly instead of through
    `database.Shared`. `cacheStore()` is the org's GitHub cache: `Server.githubCache` (nil means
    `cache.Shared()`), always `ForOrg(org)`. Seen PRs, unread maps and `/api/status` use it.
  - `cache.Store.Status()` shows the DSN and per-table counts only to the local org. Another
    org sees whether the cache works. A store with a zero org, such as the disabled one, counts
    as local.
  - `/api/config` POST for a non-local org goes to `org_settings` (`LoadOrgSearchConfig` and
    `SaveOrgSearchConfig` in `pr/ui/org_settings.go`) synchronously, and errors are reported.
  - Left on the shared or local cache on purpose: repo favicons (public homepage URLs, already
    guarded against private addresses), and the process favicon, test-run syncer,
    `collectTestRuns` and `/api/activity/cache`, whose routes are off in hosted mode and read
    the host's own files.
  - The allowed PR handlers all copy `s.ghOpts` and change only `Repo`, so the org's token
    source and cache carry through. Nothing in `pr/ui` builds a fresh `github.Options{}`.
- **2026-10-14: M5.1 `TokenSource` pulled forward into M3 (commit `7b87e94`).** Before M5, a
  hosted org's `Server` would have used the host's `GITHUB_TOKEN`, `GH_TOKEN` or `auth.json` for
  every PR call. `github.Options.TokenSource`, when set, is the only credential. The M3.2 factory
  gives each org a source that answers "connect GitHub" until M5 mints installation tokens.
  `token()` is the only code path in `github/` and `pr/` that reads host credentials.
- **2026-10-14: `Server` background loops stop with its context (commit `2d0df94`).** The auth
  probe, process metrics and task history import loops never returned, so evicting an org's
  `Server` (M3.2) would have leaked them. A test counts the loop goroutines through
  `runtime.Stack`, with no new dependency.

## 14. Progress checklist

- [x] M0 housekeeping (dist/lockfile noise decision; `whoami` without migrations; `api/clerktest` fake issuer, api auth tests moved onto it)
- [x] M1.1 tenant package
- [x] M1.2 schema plus migrations 095 and 151, idempotency test
- [x] Cleanup after M1: move existing `StartEmbedded` test call sites onto `dbtest` (done on main with `commons-db/dbtest`; ours follow it)
- [x] M1.3 github/cache ForOrg
- [x] M1.3 internal/cache ForOrg
- [x] M1.3 taskhistory ForOrg
- [x] M1.3 todos/native ForOrg plus session_owners
- [x] M1.3 runtime OpenForOrg, IdentityStore memberships
- [x] M2.0 spike: CLI access token (JWT, verified locally)
- [x] M2.1 JWKS cache, session verifier, access verifier, GitHub OIDC verifier, Clerk memberships
- [x] M3.1 ServeOptions plus Validate (`DashboardOptions`)
- [ ] M3.2 OrgServers
- [ ] M3.3 middleware, /api/auth/config, /api/me, features off in hosted mode (incl. execution 501s), spec stripping, attachment headers
- [ ] M4 UI
- [ ] M5.0 spike: GitHub App via Clerk
- [x] M5.1 TokenSource (pulled into M3, `7b87e94`)
- [ ] M5.2 App tokens
- [ ] M5.3 install linking plus repo sync
- [ ] M5.4 org PR polling
- [ ] M5.5 OIDC upload, `gavel ci upload`, action.yml
- [ ] M6.1 SSH keys API/UI (plus CLI if M2.0 allows), `login --server`, `org use`
- [ ] M6.2 push server Authorizer, traversal fix, in-process SSH (no host hooks in clerk mode), tests
- [ ] M7.0 commons-db upstream PR plus bump
- [ ] M7.1 blob package
- [ ] M7.2 snapshots, attachments, recordings, (transcripts: ask), job logs
- [ ] M8 retention
- [ ] M9.1 executors (container, kubernetes, NetworkPolicy)
- [ ] M9.2 dispatcher, run token, run API
- [ ] M9.3 source bundles
- [ ] M9.4 runner exec (test/lint/push-check)
- [ ] M9.5 project actions and push checks via the dispatcher; runner image
- [ ] M10.0 spikes (interfaces, captain ingest export, approvals)
- [ ] M10.1 remote provider, transcripts, hosted agent dispatch, run bundle upload plus serve-side push

## 15. Decisions to revisit

- **`todo_workspace_paths.path` is globally unique (2026-10-13).** The spec keeps org off path
  tables. In hosted mode, paths are server-owned per repo, so orgs shouldn't collide. If they ever
  do, the second org gets a conflict that reveals the path exists. Revisit if user-supplied paths
  ever reach hosted mode.

- **Hosted `commit` / `open-pr` project actions (2026-10-13).** Hosted checkouts come from git
  bundles, so they're clean, and the diff endpoint is off in hosted mode. The dashboard's commit
  action has no selected files to commit. For now it returns
  `400 {"error":"commits come from agent runs in hosted mode"}`, and the only hosted commit path is
  agent runs (M10, bundle upload plus serve-side push). Revisit if a hosted flow needs a human to
  commit or open a PR from the UI, e.g. reviewing an agent's changes before committing them.

- **CLI access-token revocation lag (2026-10-13).** Clerk issues 24-hour `at+jwt` access tokens, and
  local JWKS checks can't see a revoked token. Batman's call: after the local checks, ask Clerk
  whether the token is still valid, and cache the answer in memory for
  `min(exp, auth.token.cache.ttl)`, default 10m, without extending the entry on use. A revoked
  token works for at most one TTL. A sliding expiry was rejected, because a revoked token in
  constant use would never be re-checked. Revisit if Clerk's verify endpoint is slow or
  rate-limited at our request volume, or if multiple server replicas need a shared cache.

- **Pending Clerk sessions are rejected (2026-10-14).** `SessionVerifier` refuses a session token
  whose `sts` is `pending` (a sign-in with unfinished tasks, such as choosing an organization),
  matching Clerk's SDK default of treating it as signed out. Batman confirmed it on 2026-10-14,
  together with M3.3's `409 no_active_org` for a session with no org picked. Revisit if hosted mode adds its own "choose an org" step that should run on a pending
  session.
