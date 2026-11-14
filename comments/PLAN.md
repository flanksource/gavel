# Embeddable comments service

Status: implemented in the linked workspace; repository lint and published dependency compatibility remain unverified. See STATUS.md for the tracked 7/9 gate result.

Gavel owns `comments`, `comments/postgres`, `comments/clicky`, `comments/lint` and provider adapters. Hosts supply database connections, tenant/actor identity, authorization, targets, workflows and attachments. Comments own final outcomes independently of TODOs and Captain conversations.

Scope is new code and empty-table provisioning. Existing-data conversion, backfills, imports, cutover and xero-cli changes are excluded.

## Delivery

1. Public models, workflows, transactional Postgres storage and fresh provisioning.
2. Generated Clicky CLI/API, authorization, activity and attachments.
3. Native Gavel lint prompt discovery, preview, execution and idempotent findings.
4. Explicit GitHub/Azure publication, replies, resolution and synchronization.
5. Gavel registration and external embedding examples.

## Invariants

- Tenant scope applies to every record and operation; actors come from host identity.
- Versions reject stale writes; events and admitted operations commit with mutations.
- Final outcomes require final workflow phases; reopening retains prior outcomes in history.
- PR drafts never publish implicitly. Ambiguous deliveries require reconciliation.
- Captured targets and revisions survive changes in anchor resolution.
- Lint findings become local comments; publication is explicit.

## Verification

Use isolated commons-db databases, Ginkgo and fixture HTTP transports. Cover concurrent writes, tenant isolation, atomic events, workflow outcomes, code selectors, structured payload round-trips, fresh provisioning and repeated provisioning, operation recovery, lint provenance and provider deduplication. Run owning-repository lint/build and released-module compatibility checks. No live provider publication during verification.
