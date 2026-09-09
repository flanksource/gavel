# gavel — agent notes

Shared ways of working, the gavel todo workflow, and global skills come from the root ~/.agents/AGENTS.md.

## Skills
- [gavel-ci-migrator](.agents/skills/gavel-ci-migrator/SKILL.md) — migrate a repo's GitHub Actions lint/test workflows onto the flanksource/gavel composite action: discover existing golangci-lint / go test / make jobs, ask per-workflow whether to replace or add, rewrite YAML, verify with actionlint without auto-committing
- [gavel-fixture-tester](.agents/skills/gavel-fixture-tester/SKILL.md) — author fixture tests and TODO executable definitions of done
- [gavel-git](.agents/skills/gavel-git/SKILL.md) — use gavel instead of gh/git for pull requests, CI status, and commits: check PR checks with `gavel pr status`, open PRs with AI-generated content, commit with `gavel commit`
- [gavel-runner](.agents/skills/gavel-runner/SKILL.md) — run gavel test and lint, focus on a subset, re-run only failures (defaults to .gavel/last.json), filter noise with baselines, and pull JSON/markdown results from finished or live runs
- [gavel-todos](.agents/skills/gavel-todos/SKILL.md) — manage TODO content and lifecycle without starting implementation, then execute the persisted
  definition of done with `gavel todos check`

## Memory
- [Config Resolution & Prompt Registry](.agents/memory/config-and-prompts.md) — `gavel config --resolve`, `prompts/registry` as the prompt-descriptor source of truth, inline override shapes, and clicky prompt routing (module-vs-package replace gotcha)
- [TODO Lifecycle, Plans & Session Storage](.agents/memory/todo-lifecycle-plans-and-sessions.md) — session-centric turns/prompt-run modeling on Captain's native lifecycle, inline `plan.content`, and the session DSN precedence resolver
- [TODO Providers, CRUD & Rendering](.agents/memory/todo-providers-and-crud.md) — `todos/provider.go` as the abstraction point, JSON-first issue CRUD, `/todos/new` create surfaces, Todos shell/detail seams, and markdown rendering via clicky-ui + streamdown
- [Todo Run/Plan UI & Run Context](.agents/memory/todo-run-ui.md) — `/api/todos/run/context` catalog path, split-button + remembered run options, approval reuse, and Captain `whoami` parity/reconciliation
- [cmux TODO Execution](.agents/memory/cmux-execution.md) — workspace/surface ref handling, prompt-file delivery, send + send-key Enter submission, and hook-session idle polling with backoff
- [PR UI Dashboard & Shared clicky-ui Components](.agents/memory/pr-ui-and-shared-components.md) — StatusIndicator consolidation, process metrics/menubar/Wails tray, PR diff seams, ListMenu/MenubarView shared primitives, and duplicate React/React Query resolution fixes
- [Commit Workflow & PR Content](.agents/memory/commit-workflow.md) — Captain prompt bridge for PR content, 40-rune title limit, interactive AI summaries, the linked-deps commit gate, and `task build:prod`
- [Fixtures, Test History & Linter Integration](.agents/memory/fixtures-testing-and-linters.md) — fixture runner/Ginkgo seams, `gavel fixtures outline` + AI criteria nodes, `.gavel/run-*.json` test history, and signal-based linter auto-activation (react-doctor)
- [commons-db Logging & GitHub Cache](.agents/memory/database-and-logging.md) — SQL classification and `SchemaChangeSession` live in commons-db/gorm, exact log-level overrides, and the Postgres-backed GitHub cache wiring
