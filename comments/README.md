# Embeddable comments

Gavel owns outcome-oriented comments for DOM anchors, Markdown/templates, captured code ranges and native pull requests. The service includes transactional persistence, provider bindings/publication, attachments, read markers, lint metadata and generated Gavel CLI/API operations. See [STATUS.md](STATUS.md) for validation boundaries and [EMBEDDING.md](EMBEDDING.md) for configuration and runnable usage.

`comments.New` requires a `Store`, `Access`, `Targets` and explicit workflows. Hosts retain their connection pool, tenant/actor identity, authorization and target resolution. Constructors do not provision tables, start workers or bind listeners.

## Storage

Construct `postgres.New(db)` from a host-owned GORM Postgres connection. Call `postgres.Provision(ctx, dsn)` explicitly for a fresh database. Provisioning uses commons-db's SQLglot migration infrastructure and defines only new comments tables. There are no imports, backfills or legacy-data conversion paths.

Every `Access.Principal` must return a tenant and stable actor. `Access.Authorize` runs on resolved targets and each action. Target resolvers must validate supported selectors, resource ownership and source revisions. Return `comments.ErrForbidden` for access rejection.

`HumanReview()` keeps `resolved` awaiting a human outcome; `NativePR()` treats resolution as final. Custom workflows explicitly declare their initial status, status phases and transitions. A comment's outcome is independent of a Captain conversation or tool approval.

## API and CLI

Call `comments/clicky.Register(resolve)` once before `clicky.GenerateCLI(root)`, then mount the standard RPC execution routes with the host's chosen prefix. The resolver returns the service for the current context. The registered entity is `comments`; collection creation is a typed action that does not require a pre-existing ID. Mount `Stream(resolve)` and `Attachments(resolve)` behind the same authentication middleware. Wrap the generated document with `OpenAPI(handler, prefix)` to include binary and SSE routes.

Create requests use `target`, `bodyMarkdown`, `workflowId` and `requestId`. The target contains `uri`, `scopeUri`, `kind`, optional `revision`, `label` and host-validated `selector`. Detail returns replies and outcome. Mutating actions require `expectedVersion`; stale writes return 409. Author and tenant fields are derived by the host.

Root create request IDs are tenant-scoped. Reusing one with different input conflicts. Replies and transitions increment the root version and append their event within the same transaction. Event cursors are tenant-local, commit-ordered sequences.

## Lint

`comments/lint.Discover`, `Inspect`, `Prepare` and `Run` wrap Gavel's existing Captain-backed prompt lint pipeline. Hosts must authorize the selected directory and file scope before invoking it. `Prepare` validates prompt identity/version and scope without calling a model. `Run` is an explicit model invocation.

`Materialize` explicitly creates selected findings as local comments. Its run/finding keys make a retry idempotent; a conflicting finding body returns an error. Each root links to its Gavel execution. A partial batch returns its accepted roots plus the error so callers can retry safely. This action never publishes comments remotely.

## Delivery operations

`Service.Queue` authorizes and atomically admits explicit publication or agent-delivery intent with an audit event. `Service.Process` claims one operation, rechecks authorization and calls a host executor. Construction never dispatches work. An executor error or invalid receipt leaves delivery `unknown`; expired in-flight leases also become `unknown` and are not replayed automatically. Applications must reconcile ambiguous external actions before issuing a new request.

GitHub and Azure adapters validate captured heads before inline publication and preserve native identities, replies and diff context. All publication is explicit. `ProcessProvider` handles one request for the authenticated worker tenant; `Sync` only reads native state. Published native lifecycle changes use provider operations.

## Verification

Use `gavel test ginkgo ./comments --concurrency 1` against commons-db's isolated `dbtest` databases. Set `COMMONS_DB_URL` to the local test server when the environment's embedded-server configuration is unavailable. Never point these tests at a valuable database. Provider tests use fixture HTTP servers and do not publish live comments.
