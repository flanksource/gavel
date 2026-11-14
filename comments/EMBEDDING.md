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
- `Delivery` optionally opts the host into agent delivery. Build it with `comments/agents.Delivery(agents.Options{Bridge, Groups, Workdir})`: `Bridge` reaches Captain (`agents.Captain{}` in-process, or a fake such as `agents/agentstest.Bridge` in tests), `Groups` is the Postgres store (it records the session each delivery group reuses), and `Workdir` resolves the directory a target's managed sessions run in and that a scope's session listing is narrowed to. Optional `BaseURL` (an absolute origin such as `http://127.0.0.1:8080`) makes the attachment links in a delivered thread absolute; without it they are `/api/v1/comments/{id}/attachments/{attachment}` paths.
- `Routes` optionally stores comment routes (the Postgres store implements it). Routes require `Delivery`.

`CodeSelector` validates version 1, relative paths, inclusive line ranges, code revisions, and both diff commits plus the selected side. Other selector kinds are host-defined JSON. Structured root/reply content and capture metadata round-trip without a Gavel-specific presentation schema.

The comments tables are part of Gavel's database schema (`internal/database/schema/comment*.hcl`). `gavel serve` migrates them with every other Gavel table, and in-module hosts call `database.Migrate(ctx, dsn)`. The playground consumes the comments API through its host projection without importing old JSON comments.

## Gavel host

Gavel registers comments on its CLI, generated API/catalog and assistant tool surface. It resolves the pool and tenant per request and authorizes targets against registered projects. Project scopes are `gavel://project/<URL-escaped-project-name>`; local targets can be namespaced beneath that URI. Native PR URLs must belong to the project's repository inventory.

`gavel edit` threads live under the project holding the file, as `gavel://project/<name>/files/<path>`. An agent replies as itself by exporting `GAVEL_COMMENT_ACTOR` (base64url JSON `{kind, id, label}`), which only the local tenant honours.

```sh
gavel comments list --scope 'gavel://project/demo' --limit 50
GAVEL_COMMENT_ACTOR=<actor> gavel comments reply <id> --expectedVersion <version> --bodyMarkdown 'Done'
gavel comments dispatch
```

Dispatch is an explicit local CLI operation. The hosted API uses organization/user identity and its repository inventory. Hosted Azure credentials and lint execution against host files are refused until the host supplies the corresponding authorized integration.

## Generated API

Creation accepts `target`, `bodyMarkdown`, `workflowId`, `requestId`, optional `content` and `links`. The target includes `uri`, `scopeUri`, `kind`, optional `revision`, `label` and JSON `selector`. A link has `relation` and `uri`. Actor/tenant fields are host-derived.

Mutations require `expectedVersion`; stale writes return 409. `reply` accepts an optional same-root `parentId`; `assign` accepts an actor or null; `resolve`/`close` accept `summary` and `evidence`. Root create request IDs are tenant-scoped and conflicting reuse is rejected.

Listing accepts `scope`, optional `target`, `phase`, `unread`, `limit` and `offset`. The API reports pagination totals and row-level `readVersion`/`unread`. `read` monotonically advances the current actor's observed version and rejects future versions. `capabilities` reports workflow/actions, providers/bindings and attachment limits.

Mount `Stream(resolve)` at `GET <prefix>/comments/events?scope=...`; use `after` or `Last-Event-ID` to resume commit-ordered events. Mount attachment handlers at `POST <prefix>/comments/{id}/attachments` and `GET <prefix>/comments/{id}/attachments/{attachment}`. Upload raw bytes with `name`, `expectedVersion` and optional `messageId` query parameters plus `Content-Type`. An attachment only appends and is checked against the latest thread, so `expectedVersion` is the version the uploader last saw: changes made since (such as an automatic agent delivery assigning the new thread) do not conflict, while a final thread or a version the thread has not reached answers 409. A client can attach to a comment it just created with the create's version. Downloads authorize the root and use attachment disposition. Failed aggregate admission removes the unlinked blob. Use `OpenAPI(handler, prefix)` around the generated document to include these routes.

## Agent delivery and routes

`Handler(resolve)` mounts these routes; a host mounting the generated routes itself calls `RegisterDeliveryRoutes(router, resolve)`. A host without `Delivery` answers them 501.

- `GET <prefix>/comments/agent-sessions[?scope=<scopeUri>][&all=1]` lists live sessions as `[{id, source, title, model, cwd, status, inCmux}]`. With `scope` and without `all=1`, only sessions whose cwd holds, or is inside, the scope's workdir are listed.
- `POST <prefix>/comments/{id}/deliver` with `{expectedVersion, requestId, target}` delivers the thread and returns the updated comment, whose `content.agentRun` records the hand-off. A stale version answers 409; a delivery that ran and failed answers 502 with the error, which is also recorded as `agentRun.state:"error"`. Malformed `runOptions`, and `fresh` on a session target, answer 400 before anything is queued. To retry a hand-off that errored (or start over after one that went through), POST again with the thread's latest version and a new `requestId`: the earlier operation is finished, so nothing blocks the new one, and its `agentRun` replaces the previous one.
- A run target with `fresh:true` always launches a new managed session, never reusing its `groupKey`'s; the new session is recorded as the group's, so the group's later non-fresh deliveries go to it. On a stored route target, `fresh` starts a new session for every routed comment.
- `POST <prefix>/comments/{id}/deliver/preview` with `{target}` (a run target) renders the run without claiming the thread or launching: `{model, mode, resolution, validationError?, warnings?, tools:[{name, description, source}]}`.
- `GET <prefix>/comments/routes?scope=<scopeUri>` lists routes, `PUT <prefix>/comments/routes` saves one (no `id` creates it), and `DELETE <prefix>/comments/routes/{routeId}` deletes one (204). A route is `{id, scopeUri, targetPrefix, workflowId, target, enabled}`; its `targetPrefix` must lie within its `scopeUri`.

A new human-authored root is delivered through the route with the longest `targetPrefix` covering its target URI: the URI equals the prefix, or continues it after a `/` (a prefix ending in `/` covers everything beneath it). `.../pages/team` covers `.../pages/team/members` but not `.../pages/team-x`. A disabled best match opts its subtree out. Routed delivery runs in the background after the create responds and works on the thread's latest version: it queues against it, and each of its own changes (assign, acknowledge, `agentRun`) re-reads the thread and retries when a concurrent write, such as the creator's screenshot upload, moved it. A delivery that cannot be queued after bounded retries is recorded as `agentRun.state:"error"`.

A delivered thread (the live-session message, and the managed session's `message` variable) carries where it is (`content.page.url`, else `content.playground.element.pageUrl`, else the file or target), its anchor, and, when present, `content.page.title`, the element's component and source (`content.playground.element`), its locators and HTML (the selector's `locators`/`html`, else the element's; HTML trimmed to 4 KB), links to the thread's attachments, then the root and each reply. Every field is optional; a known field of the wrong JSON shape fails the delivery.

`runOptions` is `{presets?, spec?, variables?, instructions?}`. A run's host variables (`dir`, `scope`, `target`, `actor`, `gavelExecutable`, `message`) override its own, and every managed session gets the `list_threads`, `reply_to_thread` and `resolve_thread` tools for threads on its target.

```sh
gavel comments deliver <id> --session <session-id>
gavel comments deliver <id> --prompt editor-comments --group page:home --run-options '{"instructions":"Keep it short"}'
gavel comments deliver <id> --prompt editor-comments --group page:home --fresh
gavel comments routes set --scope 'gavel://project/demo' --prefix 'gavel://project/demo/pages/' --session <session-id>
gavel comments routes list --scope 'gavel://project/demo'
gavel comments routes delete <route-id>
```

The dashboard (`gavel serve`), `gavel launch serve` and `gavel edit` all opt in with the in-process Captain bridge. `comments/testhost` serves the API with a fake bridge whose one live session is `fake-session`.

## Native publication

Drafts and materialized lint findings remain local. `publish`, `publish-reply`, `resolve-native` and `reopen-native` explicitly admit captured snapshots. Include `provider`, `requestId`, `expectedVersion` and, for replies, `messageId`. Inspect the admitted operation through `operation`. `sync` only reads native state and deduplicates bindings.

GitHub supports general PR comments and review threads; replies target the root review thread. Azure preserves native parent relationships, thread status and iteration/change-tracking context. Native bindings retain payloads and remote-to-local reply identities. Native lifecycle is independent of the current provider account's resolve permission.

Call `Service.ProcessProvider` from a host-controlled worker with an explicit tenant context. Delivery errors and expired in-flight leases become `unknown` and are never replayed automatically. Receipts survive confirmation failures. Reconcile unknown deliveries before creating a new request; remote APIs do not provide an exactly-once guarantee.

## Native lint prompts

`Discover`, `Inspect`, `Prepare`, `Preview` and `RunRecorded` use Gavel's Captain-backed lint pipeline. Preparation and preview validate prompt identity/version, variables and file scope without calling a model. `RunRecorded` explicitly executes the model through Captain's runtime/approval seam.

Optionally register `comments/lint.Register(APIOptions)` to generate `comment-lint` discovery, inspection, preview, execution, detail and materialization operations. The host supplies an authorized directory resolver; an HTTP caller cannot redirect execution to a different directory.

Lint records retain selection, prompt version, native run/session references and findings. Captain owns execution state. Repeating an admitted request returns its record with a conflict instead of launching again. `Materialize` explicitly creates selected findings, idempotently by run/finding identity, with an execution link. Publishing remains separate.

## Runnable host

The [example](../examples/comments/main.go) uses a loopback listener, fixed demonstration identity and a host-owned target namespace. Supply a disposable Postgres database; `--migrate` applies Gavel's schema to it:

```sh
go run ./examples/comments --db "$COMMENTS_EXAMPLE_DSN" --migrate --port 8099
```

Create a template comment:

```sh
curl --fail-with-body -X POST http://127.0.0.1:8099/api/v1/comments -H 'Content-Type: application/json' -d '{"target":{"uri":"example://workspace/templates/invoice","scopeUri":"example://workspace","kind":"template","revision":"v1"},"bodyMarkdown":"Clarify the final outcome","workflowId":"human-review","requestId":"example-template-1"}'
```

Replace the example's identity/authorization and target resolver with those of the embedding tool. Reuse its pool, blob storage, runtime configuration and credentials. A separate comments daemon is unnecessary.
