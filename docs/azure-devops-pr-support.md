# Azure DevOps PR support

Status: implemented and verified.

Add hosted Azure DevOps support to `gavel pr status` and the shared PR creation flows. Status includes pipeline jobs, failing logs, and mergeability. Authentication supports `AZURE_DEVOPS_EXT_PAT` and an existing Azure CLI login. Support modern and legacy hosted HTTPS/SSH repository URLs.

## Implementation

- Move shared result types into `pr/model` and creation orchestration into `pr/create`; update existing callers without compatibility aliases.
- Introduce a shared provider seam for repository resolution, PR lookup, status, open-PR lookup, default branch, and creation.
- Keep GitHub behavior, artifact enrichment, polling, filters, rendering, and creation base precedence. Preserve native Azure timeline UUIDs and explicit build/check associations.
- Scope Azure builds to the PR repository and current revision; select the latest attempt per pipeline. Fetch failed timeline logs only when requested.
- Derive Azure merge readiness from merge status and blocking policy evaluations. Distinguish ready, conflicting, blocked, pending, and unknown; surface reasons.
- Reuse isolated worktrees, ordered cherry-picks, AI-generated content, draft/base options, and conflict recovery across CLI, commit-push, and project branch-PR creation.
- Fail unsupported Azure review, AI-repair, and auto-merge flags before side effects. Dashboard listing, close/merge actions, and self-hosted Server are excluded.

## Verification

Focused Ginkgo coverage includes URL/remote parsing, authentication precedence/refresh, current-build selection, UUID filters, timeline ordering, log tails, merge policies, API errors, and polling. Creation tests use local HTTP and bare Git servers. The verifier runs affected GitHub, watcher, commit, creation, and project API tests, followed by `make lint` and `make build`.

The executable verifier is [fixtures/azure-devops-pr.fixture.md](../fixtures/azure-devops-pr.fixture.md). Run it with `gavel fixtures fixtures/azure-devops-pr.fixture.md`. It exercises the provider with local HTTP services and bare Git repositories; it does not create a live PR.

TODO `1c5fa474-edda-4a3c-b40b-ec51e6e86015` passed `gavel todos check`: all eight definition-of-done checks passed, including `make lint` and `make build`.

The full commit suite has four unrelated model-alias expectation failures. Concurrent external Git operations left unresolved conflicts in `go.work.sum` and `pr/ui/dist/prui.js`; these are preserved outside the Azure changes.
