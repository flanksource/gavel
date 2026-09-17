---
name: gavel-todos
description: >-
  Create, inspect, review, verify, and maintain PostgreSQL-backed Gavel TODOs without starting implementation runs. Use for creating or syncing TODOs,
  attaching durable plans, authoring fixture-backed executable definitions of done, reviewing plans, running gavel todos check, inspecting history, or
  importing and exporting portable .todos Markdown.
allowed-tools: [Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion]
---

# Gavel TODOs

Manage Gavel's issue lifecycle without starting coding-agent execution. Treat a TODO as a native PostgreSQL issue with durable body, plan, executable
definition of done, status, and history.

Never invoke `gavel todos run`, including indirectly through `gavel todos plan approve --run`. An external orchestrator owns planning, implementation, and
committing. This skill prepares work for that orchestrator, manages plan review, and executes the definition of done with `gavel todos check`.

Use CLI help as the flag-level source of truth:

```bash
gavel todos --help
gavel todos create --help
gavel todos check --help
gavel todos plan --help
gavel todos steps --help
```

Every todo moves through a lifecycle: an ordered set of steps, each a captain
prompt plus a CEL `when` predicate (deciding whether the step applies now) and
ordered `outcomes` (deciding which status a finished run lands the todo in).
`gavel todos steps` lists the lifecycle's steps, or — given a todo — which
steps apply to it now and which one `gavel todos run` would pick next; this
skill never invokes `run`, but `steps` is safe to use for inspection.

## Inspect and select work

```bash
gavel todos list                         # pending work in this workspace
gavel todos list --done                  # include verified and completed work
gavel todos list --all                   # aggregate registered projects
gavel todos get 3f2a1b                   # short ID, full ID, title, or alias
gavel todos steps 3f2a1b                 # where this todo stands in the lifecycle
```

Runtime TODOs live only in PostgreSQL. `.todos` Markdown is a portable
interchange format used by explicit `todos import` and `todos export` commands;
it is not an alternate runtime provider.

Do not hard-wrap TODO, plan, or verification Markdown at 80 columns. Use 160 characters as the line-length limit.

## Create a TODO

Create simple work directly, or attach body, plan, and verification Markdown:

```bash
gavel todos create "Fix flaky parser"
gavel todos create "Fix parser" --body @description.md
gavel todos create "Fix parser" --plan @plan.md
gavel todos create "Fix parser" --plan @plan.md --status approved
gavel todos create "Fix parser" --verification @verification.md
```

Apply these input rules:

- Pass `--body`, `--plan`, and `--verification` as inline text or `@path`.
- Resolve relative file references from `--cwd`; use `\@text` for a literal
  leading `@`.
- Omit an outer `## Verification` heading from `--verification` content.
- Move top-level H1/H2 Verification sections out of `--body` into the dedicated
  verification fixture.
- When both inputs are present, keep explicit `--verification` first and append
  the fixture extracted from `--body`.
- Do not use the retired `--body-file` flag for create, edit, or comment; use
  `--body @path`. `todos reopen --comment-file` remains supported.

A supplied plan is persisted as an immutable Captain revision selected on the
TODO. Without `--status approved`, it awaits review and the TODO appears in
`review`. Use `--status approved` only with `--plan` when the supplied plan has
already been reviewed; the TODO remains `pending` and is ready for the external
executor.

## Author the executable definition of done

The persisted Verification content is the TODO's **executable definition of done**. Before authoring or changing it, read and follow
[`gavel-fixture-tester`](../gavel-fixture-tester/SKILL.md); that skill owns the fixture format, test/lint fences, command blocks, tables, CEL assertions, and
acceptance-checklist guidance. Do not duplicate its fixture specification here.

- When passing a standalone file with `--verification @verification.md`, omit the outer `## Verification` heading.
- When embedding the fixture in `--body` or portable `.todos` Markdown, keep it under `## Verification`.
- Author and inspect the standalone file with `gavel fixtures outline .tmp/verification.md`, then execute it with
  `gavel fixtures .tmp/verification.md` before attaching it to the TODO.
- Keep acceptance criteria observable and encode executable evidence in the fixture rather than relying on prose-only claims.

## Review externally produced plans

- Use `gavel todos plan approve <todo>` to approve a reviewed plan without starting implementation. Never add `--run`.
- Use `gavel todos plan reject <todo>` to clear a reviewed plan and return the
  TODO to pending.
- Use `gavel todos plan revise <todo> --feedback "..."` to resume the plan
  session, incorporate feedback, and return the revision to review.
- Hand pending or approved work back to the external orchestrator; do not start implementation from this skill.
- Use [`gavel-runner`](../gavel-runner/SKILL.md) for focused test/lint iteration
  outside the TODO lifecycle.

## Execute the definition of done

Use `check` when implementation exists or the gate must be rerun:

```bash
gavel todos check
gavel todos check 3f2a1b
gavel todos check --timeout 10m
```

`check` executes configured test/lint steps, the persisted executable definition of done, and the acceptance-criteria checklist through one fixture/CEL
pipeline. Passing work becomes `verified`; failing work becomes `unverified`, with evidence recorded on the issue and dashboard.

The gavel fixture attached to a TODO is the **only** verifier — there is no
separate scoring or review pass. `check` is that fixture's document run
standalone (it is the lifecycle's `verify` step, dispatched by name instead of
picked automatically); an implementation run performs the same fixture in-loop.
Author the definition of done as executable fixture content, not prose the
agent is trusted to self-report against.

## Maintain and exchange TODOs

| Command | Purpose |
| --- | --- |
| `gavel todos sync [paths...]` | Import source `TODO`/`FIXME` comments |
| `gavel todos edit <todo> --body @body.md` | Replace issue content |
| `gavel todos comment <todo> --body @comment.md` | Add a comment |
| `gavel todos reopen <todo>` | Reopen completed work |
| `gavel todos transfer <todo> --to <project>` | Move work to another project |
| `gavel todos import [files...]` | Import portable `.todos` Markdown |
| `gavel todos export [ids...]` | Export native TODOs as portable Markdown |

There is no `gavel todos criteria` subcommand. Keep acceptance criteria under `## Acceptance Criteria` in the body and update them with
`gavel todos edit <todo> --body @body.md`.

Use [`gavel-git`](../gavel-git/SKILL.md) when TODO work continues into pull
request creation, CI inspection, or an explicitly requested commit workflow.

## Safety rules

- Inspect the selected TODO and its executable definition of done before editing or checking it.
- Never start planning or implementation execution; hand ready work to the external orchestrator.
- Never use `gavel todos plan approve --run`.
- Use `todos check`, not the removed score-based `todos verify` workflow.
