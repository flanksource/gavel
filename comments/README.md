# Embeddable comments

Gavel owns outcome-oriented comments for DOM anchors, Markdown/templates, captured code ranges and native pull requests. The service includes transactional persistence, provider bindings/publication, attachments, read markers, lint metadata and generated Gavel CLI/API operations. See [STATUS.md](STATUS.md) for validation boundaries and [EMBEDDING.md](EMBEDDING.md) for configuration and runnable usage.

`comments.New` requires a `Store`, `Access`, `Targets` and explicit workflows. Hosts retain their connection pool, tenant/actor identity, authorization and target resolution. Constructors do not provision tables, start workers or bind listeners.

## Storage

Construct `postgres.New(db)` from a host-owned GORM Postgres connection. The tables are part of Gavel's database schema (`internal/database/schema/comment*.hcl`): `gavel serve` migrates them with every other Gavel table, and in-module hosts and tests call `database.Migrate(ctx, dsn)`.

Every `Access.Principal` must return a tenant and stable actor. `Access.Authorize` runs on resolved targets and each action. Target resolvers must validate supported selectors, resource ownership and source revisions. Return `comments.ErrForbidden` for access rejection.

`HumanReview()` keeps `resolved` awaiting a human outcome; `NativePR()` treats resolution as final. Custom workflows explicitly declare their initial status, status phases and transitions. A comment's outcome is independent of a Captain conversation or tool approval.

## API and CLI

Call `comments/clicky.Register(resolve)` once before `clicky.GenerateCLI(root)`, then mount the standard RPC execution routes with the host's chosen prefix. The resolver returns the service for the current context. The registered entity is `comments`; collection creation is a typed action that does not require a pre-existing ID. Mount `Stream(resolve)` and `Attachments(resolve)` behind the same authentication middleware. Wrap the generated document with `OpenAPI(handler, prefix)` to include binary and SSE routes.

Create requests use `target`, `bodyMarkdown`, `workflowId` and `requestId`. The target contains `uri`, `scopeUri`, `kind`, optional `revision`, `label` and host-validated `selector`. Detail returns replies and outcome. Mutating actions require `expectedVersion`; stale writes return 409. Author and tenant fields are derived by the host.

Root create request IDs are tenant-scoped. Reusing one with different input conflicts. Replies and transitions increment the root version and append their event within the same transaction. Event cursors are tenant-local, commit-ordered sequences.

`comments/clicky.Handler(resolve)` provides an isolated HTTP embedding of the registered comments catalog. It includes binary attachments, SSE, OpenAPI and `POST /api/v1/comments/batch`. Each batch is an array of ordinary action requests and commits all actions or none; each step must supply the version produced by the preceding step when mutating the same root twice. PostgreSQL implements the optional `TransactionStore` port required by batches.

The `metadata` action replaces opaque host content; `retarget` changes a target within its existing scope after host authorization. `edit-reply` and `delete-reply` address `messageId`, preserve author/creation metadata, and reject deleting a reply that still has children. `delete` hides a root from listings, detail and attachments without discarding audit history; `restore` reverses that local deletion. Retargeting and root deletion/restoration require a human. Retrying a create request whose result was deleted conflicts.

`gavel launch serve` embeds this API for local workspaces. Its scope is `gavel://workspace/<sha256-of-absolute-root>` and the tenant is the local user. Only loopback callers may use this host; `X-Gavel-Comment-Actor` is a base64url JSON actor assertion from a trusted local UI host. Remote/hosted applications must supply their own authentication and resolver instead. The tables exist once `gavel serve` has migrated the database.

## Lint

`comments/lint.Discover`, `Inspect`, `Prepare` and `Run` wrap Gavel's existing Captain-backed prompt lint pipeline. Hosts must authorize the selected directory and file scope before invoking it. `Prepare` validates prompt identity/version and scope without calling a model. `Run` is an explicit model invocation.

`Materialize` explicitly creates selected findings as local comments. Its run/finding keys make a retry idempotent; a conflicting finding body returns an error. Each root links to its Gavel execution. A partial batch returns its accepted roots plus the error so callers can retry safely. This action never publishes comments remotely.

## Delivery operations

`Service.Queue` authorizes and atomically admits explicit publication or agent-delivery intent with an audit event. `Service.Process` claims one operation, rechecks authorization and calls a host executor. Construction never dispatches work. An executor error or invalid receipt leaves delivery `unknown`; expired in-flight leases also become `unknown` and are not replayed automatically. Applications must reconcile ambiguous external actions before issuing a new request.

`Service.Process` takes a `ProcessRequest{Executor, Lease, OperationID?, Kinds?}`: an empty `OperationID` claims any pending operation of the tenant, and `Kinds` restricts the claim (`ProcessProvider` claims only publication kinds).

Agent delivery is opt-in. A host passes `Options.Delivery` (built by `comments/agents.Delivery(agents.Options{Bridge, Groups, Workdir})`) and optionally `Options.Routes`. `Service.Deliver` queues one `deliver-agent` operation and processes exactly that operation inline, so nothing waits for a worker. The target is `{kind:"session", sessionId}` (deliver into a live Captain session) or `{kind:"run", prompt, runOptions?, groupKey?, fresh?}` (start a managed session; with `groupKey`, reuse the live session recorded for that group in the scope, kept in `comment_agent_sessions`; with `fresh`, always launch a new session and record it as the group's). `fresh` on a session target is `ErrInvalid`; on a route target it starts a new session for every routed comment. Retrying a delivery is another `Deliver` with the latest version and a new request ID; the new `agentRun` replaces the previous one. The executor assigns the thread to the agent, acknowledges it as the agent, and records the hand-off under `content.agentRun` (`{sessionId, requestedAt, state, error?, target{label, channel}}`). A failure is recorded as `agentRun.state:"error"` and returned as `comments.ErrDelivery`. A pending `deliver-agent` operation does not freeze its root, unlike a publication.

Routes (`comment_routes`) deliver new roots automatically: after a root authored by a human (`Actor.Kind == "user"`) is created, the route with the longest `targetPrefix` covering its target URI (equal to it, or continued after a `/` boundary, so `.../team` does not cover `.../team-x`) and workflow is chosen, and delivery runs in a goroutine that never delays the create. A disabled best match opts its subtree out. Only humans manage routes. Routed delivery never holds the create's version: it queues against the latest version, and every executor change goes through `Service.WithLatest`, which re-reads the thread and retries (bounded) when a concurrent writer moved it; a conflict at an unchanged version is returned as is. A route delivery that never reached its executor is recorded through `Delivery.RecordFailure` as `agentRun.state:"error"`. Uploads are append-only and rebase onto the latest version the same way, so the creator's follow-up attachment with the create's version succeeds while a delivery claims the thread.

GitHub and Azure adapters validate captured heads before inline publication and preserve native identities, replies and diff context. All publication is explicit. `ProcessProvider` handles one request for the authenticated worker tenant; `Sync` only reads native state. Published native lifecycle changes use provider operations.

## Verification

Use `gavel test ginkgo ./comments --concurrency 1` against commons-db's isolated `dbtest` databases. Set `COMMONS_DB_URL` to the local test server when the environment's embedded-server configuration is unavailable. Never point these tests at a valuable database. Provider tests use fixture HTTP servers and do not publish live comments.
