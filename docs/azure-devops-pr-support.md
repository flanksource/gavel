# Azure DevOps PR support

Status: approved for implementation.

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

Add focused Ginkgo coverage for URL/remote parsing, authentication precedence/refresh, current-build selection, UUID filters, timeline ordering, log tails, merge policies, API errors, and polling. Exercise creation with local HTTP and bare Git servers. Run affected GitHub, watcher, commit, creation, and project API tests, followed by `make lint` and `make build`. Record the gates in the TODO verification fixture.
