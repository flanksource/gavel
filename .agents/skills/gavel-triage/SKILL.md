---
name: gavel-triage
description: >-
  Triage a Gavel TODO backlog by sorting work into keep, shape, investigate, or retire; detecting completed or duplicate work; setting priority and status;
  and preparing survivors for external execution with scope, acceptance criteria, and an executable definition of done. Use for backlog cleanup, choosing
  the next task, shaping raw TODOs, or investigating stuck work.
allowed-tools: [Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion]
---

# Gavel TODO Triage

Triage decides each TODO's **fate** and prepares survivors for the external executor. An untriaged backlog is a pile of titles that cannot be implemented
reliably.

Division of labour with the sibling skills:

| Skill | Owns |
| --- | --- |
| **gavel-triage** (this) | Deciding what a TODO *is*, whether it still matters, and shaping it for external execution |
| [`gavel-todos`](../gavel-todos/SKILL.md) | TODO creation, lifecycle review, checking, and import/export without starting implementation |
| [`gavel-fixture-tester`](../gavel-fixture-tester/SKILL.md) | Authoring the `## Verification` executable definition of done |

A TODO is **ready** when someone who has never seen it could implement it and Gavel could prove it done without asking a question. Everything below serves
that bar.

## When to use

- "Triage the backlog", "clean up the todos", "what's actionable here", "what should I pick up".
- A backlog has accumulated bare titles from `gavel todos sync` (source `TODO`/`FIXME` comments) or
  quick captures.
- Items sit in `failed` / `unverified` and nobody knows whether they are broken, stale, or already
  fixed on main.
- Before a planning session, a sprint boundary, or handing a queue to the external executor.

## 1. Pull the backlog

```bash
gavel todos list                          # pending work in this workspace
gavel todos list --all                    # every registered project
gavel todos list --done                   # include verified + completed
gavel todos list --status failed          # one status
gavel todos list --since 30d              # created or updated in a window
gavel todos list --group-by directory     # file | directory | repo | all | none
```

`--format json` (identical to `--json`) is the machine-readable form and carries the triage signals
directly. Per-item fields:

| Field | Use in triage |
| --- | --- |
| `short_id`, `id`, `title`, `workspace` | Identity |
| `status` | **Projected** state — see [status model](#the-status-model-read-before-writing-status) |
| `provider_state` | **Durable** state: `open`, `draft`, `verified`, `closed` |
| `execution_state` | Captain's view: `idle`, `running`, `failed`, `verification_failed`, … |
| `priority` | `high` \| `medium` \| `low` |
| `created`, `last_run` | Staleness |
| `attempts` | Repeated-failure signal — high attempts with no progress means the TODO is wrong, not the code |
| `run_mode` | `run` or `plan` — a TODO stuck in `plan` never got approved |
| `markdown_body` | **Absent ⇒ title-only, not implementable** |
| `acceptance_criteria` | `[{text}]`. **Absent ⇒ no definition of done** |
| `verification` | Parsed executable definition of done. **Absent ⇒ nothing can prove it done** |

The global `--filter` takes a CEL expression over those fields:

```bash
gavel todos list --filter 'priority == "high"'
gavel todos list --filter 'attempts > 2'
```

### Readiness sweep

```bash
gavel todos list --format json > .tmp/todos.json

jq 'length' .tmp/todos.json                                          # backlog size
jq '[.[] | select(has("markdown_body") | not)] | length'   .tmp/todos.json   # title-only
jq '[.[] | select(has("acceptance_criteria") | not)] | length' .tmp/todos.json
jq '[.[] | select(has("verification") | not)] | length'    .tmp/todos.json

# Title-only items — always the first triage batch:
jq -r '.[] | select(has("markdown_body") | not) | "\(.short_id)  \(.title)"' .tmp/todos.json

# Shaped but unprovable — has criteria, no fixture:
jq -r '.[] | select(has("acceptance_criteria") and (has("verification") | not))
       | "\(.short_id)  \(.title)"' .tmp/todos.json

# Burned attempts — the spec is probably the bug:
jq -r '.[] | select(.attempts > 1) | "\(.short_id)  attempts=\(.attempts)  \(.title)"' .tmp/todos.json

# Stale: no run in 30 days (compare against today's date, not a hardcoded one):
jq -r --arg cut "$(date -v-30d +%Y-%m-%d 2>/dev/null || date -d '30 days ago' +%Y-%m-%d)" \
  '.[] | select(.last_run < $cut) | "\(.short_id)  \(.last_run[0:10])  \(.title)"' .tmp/todos.json
```

## 2. Read one TODO

```bash
gavel todos get 1b495fb9        # short id, full id, title, or alias
```

`get` renders body, event history, plan, and verification as pretty text. **It ignores `--format
json`** — it always prints the pretty detail. For machine-readable single-TODO detail, use the
dashboard API (below), which returns the readiness booleans `hasPlan` and `hasVerification`
alongside `criteria`, `body`, `verificationMarkdown`, `planStatus`, and `events`.

```bash
PORT="$(cat ~/.config/gavel/pr-ui.port 2>/dev/null || echo 9092)"
curl -s "http://localhost:$PORT/api/todos/item?ref=1b495fb9&dir=$PWD" | jq '{
  title, status, priority, planStatus, hasPlan, hasVerification,
  criteriaCount: (.criteria | length)
}'
```

Read the **event history** before judging a TODO — `attempts`, `last_run`, and the event kinds tell
you whether it failed on the spec, on the environment, or was never started.

## 3. Assign a verdict

Every TODO leaves triage with exactly one of these:

| Verdict | When | Action |
| --- | --- | --- |
| **Ready** | Body, scope, criteria, and fixture all present and still accurate | Leave `pending`; correct the priority if wrong |
| **Shape** | Real work, under-specified (title-only, no criteria, no fixture) | Rewrite the body — [§4](#4-shape-a-todo-for-external-execution) |
| **Investigate** | Real problem, solution genuinely unknown | Leave pending and hand it to the external planner |
| **Already done** | Suspect the code has moved on | `gavel todos check <ref>` — let the fixture decide |
| **Retire** | Obsolete, duplicate, or won't-fix | Comment the reason, then close — [§6](#6-retire-and-dedupe) |

**Never guess "already done".** `gavel todos check <ref>` executes the persisted definition of done through the fixture/CEL pipeline. Passing work becomes
`verified`; failing work becomes `unverified` with evidence on the issue. A check needs a fixture, so shape a TODO with no `verification` before checking it.

```bash
gavel todos check 1b495fb9
gavel todos check 1b495fb9               # re-test a TODO that failed its gate
```

## 4. Shape a TODO for external execution

Rewrite the whole body — `gavel todos edit` replaces it, it does not append.

```bash
gavel todos edit 1b495fb9 --title "Default /todos/new status to pending, matching the dialog"
gavel todos edit 1b495fb9 --body @.tmp/body.md
```

`--body` takes inline text or `@path`; relative paths resolve from `--cwd`, and `\@text` escapes a
literal leading `@`. Write the file with the Write tool, not a heredoc.

A shaped body has four parts — problem statement, then these headings:

````markdown
The TODO ask state currently offers only **Send & resume**. If the user cannot answer the pending
question, manually moving the TODO to pending does not resolve the waiting Captain prompt run, so
the runtime projection returns the TODO to ask.

## Acceptance Criteria

- [ ] Every TODO ask form shows **Reject question** left of **Send & resume**.
- [ ] Rejecting posts `rejected: true` without requiring text or a selection.
- [ ] A failed rejection preserves the draft and displays the request error.

## Scope

- Extend the internal answer input with `rejected?: boolean`.
- Reuse the existing backend behavior; do not add an endpoint, status, or dialog.

## Verification

---
timeout: 30m
codeBlocks: [test, lint]
ai: {}
verify:
  scope: diff
  threshold: 80
---

```yaml test
paths: [./pr/ui]
framework: [vitest]
```

```yaml lint
changed: true
fix: false
```
````

Rules that make the difference between an execution-ready TODO and a wish:

- **Acceptance criteria are checkboxes, one observable behaviour each.** Each unchecked item is
  scored against the implementation diff. "Works correctly" is unscoreable; "invalid input exits
  non-zero with the offending value in the message" is.
- **Scope is a fence.** State what must *not* change — it is what stops an agent redesigning a
  subsystem.
- **The `## Verification` fixture is the executable definition of done.** Do not invent a parallel checking flow. Author it with
  [`gavel-fixture-tester`](../gavel-fixture-tester/SKILL.md), inspect it with `gavel fixtures outline .tmp/verification.md`, and execute it standalone with
  `gavel fixtures .tmp/verification.md` while iterating.
- Prefer `verify.scope: diff` so review is bounded to the TODO's own changes.
- A verification fixture supplied to `todos create --verification` omits the outer `## Verification`
  heading; inside a `--body` it keeps it.
- Do not hard-wrap TODO Markdown at 80 columns; use 160 characters as the line-length limit.

For an item whose *solution* is unclear rather than its *specification*, do not invent a plan during triage. Leave it pending and hand it to the external
planner. After a reviewed plan appears, triage may request revision or reject it:

```bash
gavel todos plan revise 1b495fb9 --feedback "..."
gavel todos plan reject 1b495fb9            # clear the plan, back to pending
```

## 5. Change status and priority

### The status model (read before writing status)

`status` is two different things wearing one name. Only the **durable** half is writable; the rest is
projected from Captain execution and any attempt to set it is declined.

| Status | Kind | Writable? |
| --- | --- | --- |
| `draft` | Durable → `draft` | Yes — parked, not ready for external execution |
| `pending` | Durable → `open` | Yes — the actionable queue |
| `verified` | Durable → `verified` | Yes — definition of done passed |
| `completed` | Durable → `closed` | Yes — closed |
| `skipped` | Durable → `verified` | Yes — precondition already satisfied |
| `in_progress`, `review`, `ask`, `failed`, `unverified` | **Projected** from Captain | **No** |

So "mark it failed" is not a triage move — `failed` and `unverified` are reports from the last run.
Both the CLI and the API reject them loudly rather than accepting a write that would change nothing.
To retire failing work, close it (`completed`) with a comment explaining why. To retry it, fix the specification and return it to the external executor.

### Applying the change

`gavel todos edit` sets any combination of `--title`, `--body`, `--status`, and `--priority`, and
requires at least one:

```bash
gavel todos edit 1b495fb9 --priority high
gavel todos edit 1b495fb9 --status completed
gavel todos edit 1b495fb9 --title "Fix parser panic" --priority low
```

A projected status is refused with the assignable set named in the error, so a mistake is visible
rather than silent:

```
$ gavel todos edit 1b495fb9 --status failed
Error: status "failed" is projected from the last run and cannot be assigned;
assignable statuses: draft, pending, verified, completed, skipped
```

`--status` and `--priority` do not carry a rationale. When the change needs one, comment first
(§6) or use the dashboard API, which takes the transition and the comment in one request:

```bash
PORT="$(cat ~/.config/gavel/pr-ui.port 2>/dev/null || echo 9092)"

curl -s -X PATCH "http://localhost:$PORT/api/todos/item?ref=1b495fb9&dir=$PWD" \
  -H 'Content-Type: application/json' \
  -d '{"status":"completed","comment":"Superseded by 8f1fb6e9 — same defect, wider scope."}'
```

The endpoint accepts any combination of `status`, `priority`, `title`, `body`, `comment`, and
requires at least one; `dir` is the workspace path. It rejects projected statuses with `400`, and
needs a running dashboard (`gavel serve`, or `gavel pr list --ui`, default port 9092). Filtering is
unaffected — `--status in_progress` reads fine, it just cannot be written.

`gavel todos reopen <ref>` is the transition worth preferring over `edit --status pending`: it
returns a completed TODO to the queue *and* takes the rationale inline, via `--comment` /
`--comment-file`.

## 6. Retire and dedupe

Every retirement gets a **comment first, then the close** — a closed TODO with no rationale is
indistinguishable from a lost one.

```bash
gavel todos comment 1b495fb9 "Obsolete: the /todos/new route was deleted in 1060a1b0."
gavel todos comment 1b495fb9 --body @.tmp/rationale.md
```

There is **no merge command** — the survivor absorbs the duplicate's content by hand, and a link
records the pairing:

1. Pick the survivor — the one with the better body, criteria, or fixture, not the older id.
2. Fold anything unique from the duplicate into the survivor's body with `todos edit`.
3. Link the two so the pairing survives the close: `gavel todos link <duplicate> <survivor>`.
4. Comment on both, naming the other's short id in each direction.
5. Close the duplicate: `gavel todos edit <duplicate> --status completed`.

### Links

Two relations are writable; the third is derived and read-only.

| Relation | Meaning | Writable? |
| --- | --- | --- |
| `related_to` | Symmetric, non-blocking — duplicates, overlapping scope, read-together work (the default) | Yes |
| `depends_on` | `<ref>` is blocked until `<target-ref>` is verified or completed | Yes |
| `blocks` | The reverse view of someone else's `depends_on` | No — add `depends_on` from the blocked TODO |

```bash
gavel todos link 1b495fb9 8f1fb6e9                          # related_to
gavel todos link 1b495fb9 8f1fb6e9 --relation depends-on    # 1b495fb9 waits on 8f1fb6e9
gavel todos links 1b495fb9                                  # every edge, from this TODO's side
gavel todos links 1b495fb9 --format json
gavel todos unlink 1b495fb9 8f1fb6e9 --relation depends-on
```

Dependency cycles, cross-workspace links, duplicate edges, and self-links are all rejected — a link
that does not appear is an error, never a silent no-op. Before ordering a sprint, read the
dependencies: a `pending` TODO whose `depends_on` target is unverified is not actually ready, whatever
its status says.

Candidate pairs surface from title overlap and shared paths:

```bash
jq -r '.[] | "\(.short_id)  \(.title)"' .tmp/todos.json | sort -k2
```

A TODO in the wrong project moves rather than closes:

```bash
gavel todos transfer 1b495fb9 --to <project>
```

## 7. A backlog pass, end to end

```bash
# 1. Snapshot and size the problem.
gavel todos list --format json > .tmp/todos.json
jq 'length' .tmp/todos.json

# 2. Anything claiming to be done? Let the fixtures decide before humans read them.
gavel todos check 1b495fb9

# 3. Title-only items — shape or retire, one batch.
jq -r '.[] | select(has("markdown_body") | not) | "\(.short_id)  \(.title)"' .tmp/todos.json

# 4. Shaped-but-unprovable — attach fixtures.
jq -r '.[] | select(has("acceptance_criteria") and (has("verification") | not))
       | "\(.short_id)  \(.title)"' .tmp/todos.json

# 5. Re-prioritise the survivors, then report the ready queue for external execution.
gavel todos list --filter 'priority == "high"'
```

Report the pass as counts per verdict plus the ids retired, not as a per-item narrative.

## Gotchas

- **`gavel todos edit` carries no rationale.** `--status`/`--priority` change state silently; pair
  them with `todos comment`, or use the PATCH endpoint's `comment` field to do both atomically.
- **`gavel todos get` ignores `--format json`** and always prints pretty text. Use
  `/api/todos/item?ref=…&dir=…` for machine-readable single-TODO detail.
- **`failed` and `unverified` are not writable.** They are projections of the last run, and both the
  CLI and the API reject them — but they filter fine (`todos list --status failed`).
- **`blocks` cannot be written.** It is the reverse view of `depends_on`; add the dependency from the
  blocked TODO instead.
- **There is no `wontfix` or `cancelled` status** reachable from the CLI or API. Retirement is
  `completed` plus a comment carrying the reason.
- **There is no `gavel todos criteria` subcommand.** Acceptance criteria live in the body under
  `## Acceptance Criteria` and are edited with `todos edit --body`.
- **`todos edit --body` replaces the whole body.** Read the current body first and re-emit it with
  your additions, or you will silently drop the fixture.
- **The `--body-file` flag is retired** for create/edit/comment — use `--body @path`. Only
  `todos reopen --comment-file` still takes a file flag.
- **`todos check` needs a fixture.** A TODO with no `verification` cannot be resolved by checking;
  shape it first.
- **The PATCH endpoint needs a running dashboard.** Read the port from
  `~/.config/gavel/pr-ui.port`, defaulting to 9092.
- Use `.tmp/` for scratch snapshots, never `/tmp`.

## Safety rules

- Triage reads and reclassifies; it does not plan or implement. Hand ready work to the external orchestrator instead of starting execution.
- Never close a TODO you have not read in full, including its event history.
- Retiring work is the user's call when the reason is judgement rather than fact. "The route was
  deleted" is fact — close it. "This is probably not worth doing" is judgement — ask, batching every
  such item into one question.
- Do not rewrite a body to match what the code already does. If the TODO's premise is stale, retire
  it and say so; silently re-scoping it to be trivially satisfiable defeats the definition of done.
- Do not hand-edit the TODO tables in Postgres. Every mutation goes through the CLI or the API so
  events, versions, and Captain projections stay consistent.
