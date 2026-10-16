# Identity and ownership schema

## Users and organizations

`users` and `orgs` have internal UUID primary keys. Clerk identities are unique by
`(issuer, clerk_user_id)` and `(issuer, clerk_org_id)` respectively. Development
and production issuers therefore cannot collide. An org's slug is optional,
unique within its issuer when populated, and can change without changing FKs.

When PostgreSQL is configured, `gavel login` applies migrations, and both
`gavel login` and `gavel whoami` upsert the verified user and selected org.
`whoami` never migrates: on a database whose identity tables don't exist yet it
displays the signed claims without recording them. Neither command stores
credentials in PostgreSQL or assigns existing resources to the user. Without a
configured DB, authentication still works and only signed identity claims are
displayed. `IdentityStore.Sync` returns internal `Ownership` UUIDs; an identity
with no active org acts for the local org. It does not infer a membership,
role, or project permission.

`IdentityStore.SyncMemberships` replaces a user's `org_memberships` rows (org
plus Clerk role) with the given set and creates or refreshes each org under the
user's issuer. A dropped membership removes the row and keeps the org.

## The local organization

Every tenant-owned row belongs to exactly one org. Without authentication
(`gavel serve --auth none` and the CLI), that org is the built-in **local org**
`00000000-0000-0000-0000-000000000001` (`issuer`, `clerk_org_id` and `slug` all
`local`). Tenant tables default `org_id` to it, so rows written by code that
doesn't know about orgs land there.

## Organization slug

`whoami` adds `organization_slug` immediately after `organization_id`, including
an empty string when unknown. A signed `org_slug` claim is used when present.
With a configured database, the existing `orgs.slug` is a fallback and a missing
claim does not erase it.

Clerk's OAuth documentation guarantees `org_id`, not `org_slug`. To resolve a
missing slug automatically on a **trusted backend**, configure the store:

```go
store := api.IdentityStore{
    DB: db,
    OrganizationResolver: api.ClerkOrganizations{
        SecretKey: serverSideClerkSecret,
    },
}
identity, ownership, err := store.Sync(ctx, verifiedIdentity)
```

Use a resolver for the same Clerk issuer/instance as the verified identity.
It fetches the canonical slug/name from Clerk's Backend API and rejects a
mismatched organization ID. The backend must authenticate the request before
calling Sync. This resolver is deliberately **not wired to a CLI secret-key
flag or environment variable**. A future hosted API must invoke it and expose
the enriched identity to its authenticated CLI client. In the current local CLI,
an absent claim and absent stored slug remain empty; it cannot obtain private
backend metadata using only the OAuth client ID.

## Existing table coverage

| Tables | Ownership/reference model |
| --- | --- |
| `todo_workspaces` | `org_id NOT NULL` (repo key unique per org); nullable `user_id` FK |
| `todo_workspace_paths`, `todo_issue_aliases`, `todo_issue_relationships` | Inherit via workspace FK |
| `todo_issues` | Inherit org via workspace; `created_by_user_id` FK |
| `todo_labels` | `org_id NOT NULL`, matching the workspace's org for workspace labels; global names unique per org; `updated_by_user_id` FK |
| `todo_issue_events` | Inherit through issue; `actor_user_id` FK, retaining system/agent actor text |
| `todo_issue_prompt_runs` | Inherit through issue; `requested_by_user_id` FK |
| `todo_issue_plans` | Inherit through issue; `created_by_user_id` FK |
| HTTP, workflow-run, job-log, workflow-definition caches | `org_id NOT NULL`, leading every primary key; nullable `user_id` FK (may contain private data) |
| Seen PRs, commit caches/cursors, test-run caches/cursors | `org_id NOT NULL`, leading every primary key; nullable `user_id` FK |
| `file_scans` | `org_id NOT NULL`; primary key `(org_id, file_path)`; nullable `user_id` FK |
| `violations` | `org_id NOT NULL`; FK `(org_id, file_path)` to `file_scans` |
| `linter_executions`, `debounce_metadata` | `org_id NOT NULL`, part of the unique/primary key; nullable `user_id` FK |
| `task_run_history` | `org_id NOT NULL` (run IDs stay globally unique); nullable `user_id` FK |
| `git_repos` | `org_id NOT NULL`; root dir unique per org; nullable `user_id` FK |
| `git_worktrees`, `git_branches`, `git_range_stats` | Inherit via `repo_id` |
| `todo_run_landings` | Inherit through issue |
| `org_memberships`, `org_settings`, `repos`, `github_installations`, `sandbox_runs` | `org_id NOT NULL` FK |
| `user_ssh_keys` | Owned by the user; a key authenticates its user in every org they belong to |
| `favicon_caches` | Public metadata; intentionally shared |
| Captain sessions | Gavel-owned `session_owners` sidecar with session/user/org FKs |
| Captain turns, runs/iterations, plans/revisions, model calls, messages, events, artifacts, sources, processes, turn requests, MCP credentials | Inherit through the session/run/plan links |
| Captain API tokens, runtime profiles/presets | Gavel-owned `api_token_owners`, `runtime_profile_owners`, `runtime_preset_owners` sidecars |
| Captain git-agent tasks/attempts | `git_agent_task_owners` sidecar; attempts inherit task ownership |
| Migration bookkeeping | Infrastructure metadata; intentionally shared |

Captain-owned schemas are not modified, and gavel installs no DDL on Captain's
tables: sidecars hold Captain IDs without foreign keys. A sidecar row whose
Captain resource is deleted owns nothing, and every migration removes it
(`sweepOwnershipOrphans`). Gavel records a `session_owners` row for every Captain
session it creates, and for its root, inside the transaction that creates it.
Org deletion is restricted while owned resources remain; user deletion nulls
attribution without deleting history.

The `org_id` FKs on pre-existing tables are installed by
`151_tenancy_local_org.sql` after the local org row exists, and are excluded
from Atlas inspection so the HCL doesn't drop them.

## Scoped stores and the security boundary

Every store reads and writes one org. Its constructor and process-wide
singleton use the local org; `ForOrg(org)` returns a copy over another org:

- `github/cache.Store`, `internal/cache.ViolationCache` and `LinterStats`,
  `internal/taskhistory.Store`, `git/gitstate.Store` and `todos/native.Repository` filter every query
  by org, set `org_id` on every insert, and never update `org_id`, `user_id`
  or `created_*` on conflict.
- `todos/runtime.OpenForOrg` and `NewForOrg` build a provider over one org.
- Another org's workspace, issue or Captain link reads as `ErrNotFound`.
- Writes take their user attribution from `tenant.FromContext(ctx).UserID`.

Global maintenance deliberately spans every org: the HTTP-cache prune, the
task-history prune, and the violation cache's storage-size statistic.

A user/org FK proves referential integrity, **not authorization**. Scoping
stores by org is not the same as checking that the caller belongs to that org;
membership enforcement is the job of the authenticated server. There is no
PostgreSQL row-level security.
