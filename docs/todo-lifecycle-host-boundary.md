# Captain-owned approvals for TODO runs

Status: implemented in the local Captain and Gavel checkouts, updated 2026-09-28. Focused tests, Captain race tests, both builds, Captain lint, and a read-only dashboard smoke check pass. Broader Gavel tests and lint have failures outside this approval seam; an interactive approval was not created against the existing local TODO database.

## Problem and owner boundary

An attended TODO run previously asked Gavel's dashboard to construct an approval broker before `promptrun.Run` admitted the Captain prompt run. The broker read that run back from Captain's store to recover its approval timeout. When the row had not been admitted yet, dispatch failed with `step run: run <id> was not admitted: captain prompt run not found: <id>`.

The broker has no independent process to start. It is a per-run callback backed by Captain's durable approval rows. Captain already owns run admission, run state, resolved permissions, provider construction, and the run deadline, so Captain should bind this callback after admission. Gavel should opt an attended run into approvals and keep the dashboard's approval list and answer endpoints.

## Implemented API and lifecycle

1. Add `promptrun.Input.Approvals *ApprovalOptions` with `RequestedBy`. `nil` means no Captain approval broker. `Preflight` validates the requester, event sink, provider ownership, and conflicting callback; it permits a missing recording because Gavel previews before admission. `Run` additionally requires `Input.Record` before admitting anything. Invalid combinations fail without creating a prompt run.
2. In `promptrun.Run`, admit the run first. In execution, after the effective run timeout is installed and before building the provider, construct `approval.Broker` with the admitted run's session ID and run ID, recording database, requester, and event notification through `Input.OnEvent`. Parse `Resolved.Spec.Permissions.ApprovalTimeout`; use Captain's existing provider timeout when unspecified. Set the broker's absolute deadline from the effective run context. Install `broker.CanUseTool` in `Input.Config`.
3. Keep the callback bound to this run's broker. Do not add a global callback or use Gavel's `ExecutorContext` to recover identities. The run-level binding propagates cancellation into each callback context and preserves the broker's concurrent request, hold/release, answer, and expiry semantics.
4. Keep Codex's authenticated MCP endpoint alive across turns, but resolve each tool call's context from the active Codex turn through `callertools.Options.ContextForCall`. Permission callbacks and tool handlers receive that turn's values, deadline, and cancellation. A call without an active turn fails explicitly; ending one turn does not cancel the endpoint needed by the next turn. The provider-owned background context controls only the endpoint lifetime.
5. Make Gavel's attended dashboard entry points set the opt-in on the run input. Remove `todos.ApprovalBroker`, the `RunOptions.Broker` and downstream factory plumbing, `stepInput.brokerFactory`, `stepInput.canUseTool`, and the dashboard's `todoApprovalBroker`, `approvalWindow`, `runDeadline`, and `notifyApproval`. Keep the dashboard's durable listing, resolve action, and notification rendering. Route Captain's `EventPermission` fields, including approval ID, tool input, and terminal reason, into the existing transcript and notification seam.
6. In Gavel's recording callback, call `promptrun.DefaultOutcome(result, runErr, stopped)` first. Preserve its cancelled, failed, waiting, and verifier verdict classifications. Apply Gavel's envelope and definition-of-done mapping only after a completed run. Take usage and cost from Captain's final `Result`, avoiding event-sum drift. Keep Gavel's own TODO linking, comments, commit workflow, and envelope rendering.
7. Update the TODO preflight documentation to state that preview validates the approval opt-in but neither admits a run nor constructs its broker. Dispatch binds the broker after admission.

## Verification and order of work

1. Preserve current work in both dirty checkouts. Read scoped diffs before each edit. Add focused failing Ginkgo coverage first. The missing-run path is covered by a Captain test that admits a run, binds approvals, and resolves a fresh-context caller tool against its durable row.
2. Implement Captain's opt-in and test validation before admission, broker identity and timeout, Claude and cmux provider paths, concurrent approvals, deadline, cancellation, and event delivery. Fix Codex caller tools at the shared MCP runtime seam so each call uses the active turn context while the endpoint persists across turns; test callback values, deadline, cancellation, and context resolution when no turn is active.
3. Switch Gavel's attended paths to the opt-in and remove the broker factory. Test attended versus unattended execution, approval notifications and answers, and the reported missing-run failure.
4. Update Gavel's recorded outcome classification. Test cancellation, run errors, questions, failed verification, failed envelope and definition of done, successful completion, and final usage/cost.
5. Run focused Go tests, then `make lint` and `make build` in each changed repository. Check local dashboard rendering with agent-browser. Use an isolated scripted provider for a full interactive approval smoke, since the available local dashboard uses an existing TODO database. Report unrelated gate failures separately from the changed seams.

## Verification recorded

- Captain `go test ./pkg/promptrun ./pkg/ai/approval -count=1`, `make lint`, and `make build` passed. The new binding test admits a run, resolves an approval raised with `context.Background()`, then cancels a second approval through the run context.
- The Codex context follow-up passed `go test ./pkg/ai/callertools ./pkg/ai/provider`, `go test -race ./pkg/ai/callertools ./pkg/ai/provider`, `make lint`, and `make build` in Captain. Focused Ginkgo tests verify per-turn values and deadlines, approval cancellation, and that Codex supplies no call context when no turn is active.
- Gavel focused lifecycle, entity, and dashboard approval tests passed, and `make build` passed. The new lifecycle test checks approval ID, input, and expiry reason in the transcript; the outcome tests cover cancellation, waiting, and final usage/cost.
- A newly built Gavel dashboard served `/todos/f1e27b11?project=gavel&tab=session` in agent-browser. Its session tab rendered successfully. No run or approval was started against the existing local database.
- Broader Gavel tests still fail in runtime-default composition, API-mode sandbox preflight, and unrelated entity catalog routes. `make lint` reports unused declarations in untouched `pr/ui/chat.go` and `pr/ui/entity_routes.go`, then `go mod tidy` cannot find `captain/pkg/plans` and `captain/pkg/sessiontree` in the published Captain version.

No schema or HTTP API change was made. The new Captain opt-in is a typed Go API used by recorded, attended runs.
