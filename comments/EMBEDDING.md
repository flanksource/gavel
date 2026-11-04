# Embedding the comments service

Import `github.com/flanksource/gavel/comments` for the core, `comments/postgres` for PostgreSQL, `comments/clicky` for CLI/HTTP/tools, `comments/lint` for Gavel lint prompts and `comments/providers` for native PR adapters. Importing and constructing these packages never performs DDL or external dispatch.

## Host ports

`comments.New` takes these explicit options:

- `Store` uses the host's connection pool. `postgres.New(db)` does not close or provision it.
- `Access` supplies authenticated tenant/actor identity and authorizes every target/action. Scope list permission covers all records/counts in that scope; each returned comment is authorized again. Request bodies cannot choose their tenant or author.
- `Targets` validates ownership, supported selector schemas and captured revisions. A host can namespace templates, DOM surfaces and code without adopting Gavel's project catalog. Preserve original captures when current anchor resolution changes.
- `Workflows` declares status phases/transitions. `HumanReview()` keeps proposed resolution separate from human closure; `NativePR()` has final native resolution.
- `Blobs` and `AttachmentLimit` optionally configure attachments with an explicit positive byte limit. The Postgres store implements `Blobs`. `comments/blob.Store` adapts Gavel's configured filesystem/object store.
- `Providers` optionally supplies GitHub/Azure adapters with per-principal credential resolvers. Bind credentials to the request's tenant and repository.

`CodeSelector` validates version 1, relative paths, inclusive line ranges, code revisions, and both diff commits plus the selected side. Other selector kinds are host-defined JSON. Structured root/reply content and capture metadata round-trip without a Gavel-specific presentation schema.

Provision only empty comments tables with `postgres.Provision(ctx, dsn)`. Existing-data conversion, backfill, filesystem import, xero-cli refactoring and existing playground storage cutover are excluded.

## Gavel host

Gavel registers comments on its CLI, generated API/catalog and assistant tool surface. It resolves the pool and tenant per request and authorizes targets against registered projects. Project scopes are `gavel://project/<URL-escaped-project-name>`; local targets can be namespaced beneath that URI. Native PR URLs must belong to the project's repository inventory.

```sh
gavel comments provision
gavel comments list --scope 'gavel://project/demo' --limit 50
gavel comments dispatch
```

Provisioning and dispatch are explicit local CLI operations. The hosted API uses organization/user identity and its repository inventory. Hosted Azure credentials and lint execution against host files are refused until the host supplies the corresponding authorized integration.

## Generated API

Creation accepts `target`, `bodyMarkdown`, `workflowId`, `requestId`, optional `content` and `links`. The target includes `uri`, `scopeUri`, `kind`, optional `revision`, `label` and JSON `selector`. A link has `relation` and `uri`. Actor/tenant fields are host-derived.

Mutations require `expectedVersion`; stale writes return 409. `reply` accepts an optional same-root `parentId`; `assign` accepts an actor or null; `resolve`/`close` accept `summary` and `evidence`. Root create request IDs are tenant-scoped and conflicting reuse is rejected.

Listing accepts `scope`, optional `target`, `phase`, `unread`, `limit` and `offset`. The API reports pagination totals and row-level `readVersion`/`unread`. `read` monotonically advances the current actor's observed version and rejects future versions. `capabilities` reports workflow/actions, providers/bindings and attachment limits.

Mount `Stream(resolve)` at `GET <prefix>/comments/events?scope=...`; use `after` or `Last-Event-ID` to resume commit-ordered events. Mount attachment handlers at `POST <prefix>/comments/{id}/attachments` and `GET <prefix>/comments/{id}/attachments/{attachment}`. Upload raw bytes with `name`, `expectedVersion` and optional `messageId` query parameters plus `Content-Type`. Downloads authorize the root and use attachment disposition. Failed aggregate admission removes the unlinked blob. Use `OpenAPI(handler, prefix)` around the generated document to include these routes.

## Native publication

Drafts and materialized lint findings remain local. `publish`, `publish-reply`, `resolve-native` and `reopen-native` explicitly admit captured snapshots. Include `provider`, `requestId`, `expectedVersion` and, for replies, `messageId`. Inspect the admitted operation through `operation`. `sync` only reads native state and deduplicates bindings.

GitHub supports general PR comments and review threads; replies target the root review thread. Azure preserves native parent relationships, thread status and iteration/change-tracking context. Native bindings retain payloads and remote-to-local reply identities. Native lifecycle is independent of the current provider account's resolve permission.

Call `Service.ProcessProvider` from a host-controlled worker with an explicit tenant context. Delivery errors and expired in-flight leases become `unknown` and are never replayed automatically. Receipts survive confirmation failures. Reconcile unknown deliveries before creating a new request; remote APIs do not provide an exactly-once guarantee.

## Native lint prompts

`Discover`, `Inspect`, `Prepare`, `Preview` and `RunRecorded` use Gavel's Captain-backed lint pipeline. Preparation and preview validate prompt identity/version, variables and file scope without calling a model. `RunRecorded` explicitly executes the model through Captain's runtime/approval seam.

Optionally register `comments/lint.Register(APIOptions)` to generate `comment-lint` discovery, inspection, preview, execution, detail and materialization operations. The host supplies an authorized directory resolver; an HTTP caller cannot redirect execution to a different directory.

Lint records retain selection, prompt version, native run/session references and findings. Captain owns execution state. Repeating an admitted request returns its record with a conflict instead of launching again. `Materialize` explicitly creates selected findings, idempotently by run/finding identity, with an execution link. Publishing remains separate.

## Runnable host

The [example](../examples/comments/main.go) uses a loopback listener, fixed demonstration identity and a host-owned target namespace. Supply a disposable empty Postgres database:

```sh
go run ./examples/comments --db "$COMMENTS_EXAMPLE_DSN" --provision --port 8099
```

Create a template comment:

```sh
curl --fail-with-body -X POST http://127.0.0.1:8099/api/v1/comments -H 'Content-Type: application/json' -d '{"target":{"uri":"example://workspace/templates/invoice","scopeUri":"example://workspace","kind":"template","revision":"v1"},"bodyMarkdown":"Clarify the final outcome","workflowId":"human-review","requestId":"example-template-1"}'
```

Replace the example's identity/authorization and target resolver with those of the embedding tool. Reuse its pool, blob storage, runtime configuration and credentials. A separate comments daemon is unnecessary.
