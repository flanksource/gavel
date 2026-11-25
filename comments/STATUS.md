# Implementation checkpoint

Task: `f02cb19b-a985-4a89-9ec3-6f6fcba57f6c`.

Status: implemented in the linked workspace; full delivery remains unverified on repository lint and published dependency compatibility. No release or external publication has been performed.

Implemented: host-supplied identity/authorization/target interfaces; independent workflow outcomes; transactional Postgres root/replies/events; fresh schema provisioning with foreign-table refusal; optimistic concurrency; request deduplication; explicit outbox admission, leases and uncertain-delivery handling; provider bindings and GitHub/Azure executors/synchronization; attachments and read markers; lint request/reference/finding persistence; captured code selectors; paginated Clicky CLI/API and capability metadata; resumable events; native Gavel prompt discovery/preparation/preview/execution and idempotent finding materialization; Gavel host/CLI registration and a runnable embedding example.

Shared dependencies: Captain gains a presentation-only prepared-prompt preview accessor. Clicky gains int64, standard duration and structured JSON action binding, plus RPC type metadata. The comments implementation itself remains in Gavel. These dependency changes need published releases and matching pins before a standalone downstream consumer can use the new surface.

The supporting changes are committed locally in Captain as `bf56d041` and Clicky as `658486b8`.

Verification so far: 15 comments specs pass with the race detector, including generated HTTP reads/attachments, provider reply metadata round-trips and publication replay; two native provider fixture specs pass; 13 Gavel authorization/catalog specs pass; CLI registration passes; 43 Clicky flag specs and the Captain preview spec pass. Tests are serialized with runner concurrency 1 and Go package parallelism 1. Use the isolated localhost test pool through `COMMONS_DB_URL=postgres://postgres:postgres@localhost:7432/postgres?sslmode=disable` in this environment; dbtest creates disposable databases on that server.

`make build` passes. `make lint` completes its architecture test, vet and tidy but its task ignores golangci-lint's nonzero exit: strict lint still reports six pre-existing findings in untouched `externalissues/jira` sources. The new Gavel code has no reported lint findings.

Scoped Clicky lint (`flags`, `entity`, `rpc`) and Captain CLI lint report zero issues. The embedding example compiles and its `--help` smoke check passes; it was not pointed at an application database.

The tracked definition of done passes 7/9 checks: comments/race/database integration, both native provider HTTP contracts, Gavel authorization, entity catalog, CLI registration and build. Strict repository lint and published dependency compatibility fail, so task `f02cb19b` remains UNVERIFIED. The native lint integration was already blocked in `GOWORK=off` mode by Gavel prompt code referencing Captain APIs absent from its pinned release; the new preview accessor and Clicky binding fix also require coordinated dependency releases and matching pins. No local replace or file override was introduced.

Unrelated working-tree changes were preserved. No xero-cli code, existing-data migration or remote provider publication was performed.
