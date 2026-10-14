# PR-linked TODOs

Link a TODO explicitly to an existing GitHub or hosted Azure DevOps pull request. The dashboard's **Create from PR** action saves the link; the New Todo form accepts a PR URL, and the detail view supports linking, replacing and removing it.

Linked TODOs display a **PR #number** badge in both list layouts and above the detail tabs. Open **PR status and actions** to fetch the linked PR directly, including repositories outside the dashboard's configured PR list. The panel shows its source and target branches, current revision, checks, unresolved review threads and GitHub Gavel verification artifacts. **Refresh PR status** reloads it on demand.

The panel exposes **Rebase and merge**, **Squash and merge** and **Create a merge commit** for GitHub and Azure DevOps. GitHub also offers approval, auto-merge and updating the source branch by merging the target into it. Each mutation requires confirmation and refreshes the displayed status afterwards; actions are disabled while a TODO turn is running. Rebase and merge completes the PR using the host's rebase strategy. It is separate from updating the source branch. Azure completion uses the [native completion API](https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-requests/update?view=azure-devops-rest-7.1) without bypassing branch policies or deleting the source branch.

```bash
gavel todos create --title "Handle empty parser input" --pr https://github.com/acme/service/pull/27
gavel todos edit <todo-id> --pr https://dev.azure.com/acme/project/_git/service/pullrequest/27
gavel todos edit <todo-id> --clear-pr
```

The `pr` field contains `number` and `url`, with optional `head` and `base` metadata. JSON creation and update requests accept that object. Update requests use `clearPR: true` to remove the link; omitting both fields leaves it unchanged. Form requests accept a PR URL or the same JSON object. Providing both a link and `clearPR` is an error.

```json
{"title":"Handle empty parser input","pr":{"number":27,"url":"https://github.com/acme/service/pull/27"}}
```

Links survive Markdown export/import, project transfer, and merging or deduplicating TODOs. Merging TODOs linked to different PRs fails before retirement. The nullable storage column has no backfill: titles, bodies and verification commands never establish a PR link. Triage must replace generic titles such as “Address feedback on PR #27” with the concrete behavior or defect being addressed.

An implementation run starts a separate worktree at the current PR source head, including the source repository for fork PRs. Gavel commits each completed turn through its commit pipeline, pushes a fast-forward update to that existing PR, and resolves only the native review thread IDs listed in the turn's `addressedThreadIds`. Asking a question or returning a failed result does not publish the turn. A changed remote source branch stops the run rather than overwriting newer work.

The TODO fixture remains the sole verifier. A mandatory runtime check joins any existing local verification:

```bash
gavel pr status https://github.com/acme/service/pull/27 --whole-pr --head-sha HEAD --follow --fail-fast --logs --tail-logs 40 --json
```

`--whole-pr` checks the complete PR at the actual pushed revision: checks and pipeline jobs, merge readiness, and all unresolved native review threads. It cannot be narrowed with `--actions` or `--comments`. Failure logs, verification results, the current PR status and native thread IDs feed Captain's next turn. Missing checks wait for registration; failed checks, a changed head, blocked readiness or unresolved feedback prevent completion. Human approvals and missing credentials still require human action.

Planning does not publish changes. Verification runs inspect the current source head without committing, pushing or resolving threads. Neither execution path merges or approves the PR, creates a replacement PR, or force-pushes. Existing unlinked TODOs retain their normal workflow.
