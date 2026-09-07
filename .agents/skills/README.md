# gavel-skills

[Agent Skills](https://agentskills.io/) for the [Gavel](https://github.com/flanksource/gavel) CLI testing framework.

Skills follow the open [Agent Skills specification](https://agentskills.io/specification) and work with any compatible agent including Claude Code, VS Code Copilot, Cursor, Gemini CLI, and others.

## Installation

### Using the skills CLI (recommended)

Install skills into any compatible agent with [`npx skills`](https://github.com/vercel-labs/skills):

```bash
# Install to current project
npx skills add flanksource/gavel

# Install globally (all projects)
npx skills add flanksource/gavel -g

# Preview available skills before installing
npx skills add flanksource/gavel -l

# Install for a specific agent
npx skills add flanksource/gavel -a claude
```

Manage installed skills:

```bash
npx skills list           # show installed skills
npx skills check          # check for updates
npx skills update         # update all skills
npx skills remove         # uninstall skills
```

### Auto-discovery (local development)

Skills in `.agents/skills/` are auto-discovered by compatible agents. Clone the repo and open it — no extra setup needed:

```bash
git clone https://github.com/flanksource/gavel.git
cd gavel
```

### Claude Code plugin

```bash
# Via marketplace
/plugin marketplace add flanksource/gavel
/plugin install gavel-skills@flanksource-gavel

# Or from a local clone
claude plugin install /path/to/gavel/.agents

# For development/testing
claude --plugin-dir /path/to/gavel/.agents
```

Manual configuration — add to the appropriate settings file:

| Scope | File | Shared via git |
|-------|------|----------------|
| User | `~/.claude/settings.json` | No |
| Project | `.claude/settings.json` | Yes |
| Local | `.claude/settings.local.json` | No |

```json
{
  "enabledPlugins": {
    "gavel-skills@flanksource-gavel": true
  }
}
```

After manual changes, reload with `/reload-plugins`.

## Skills

| Skill | Description |
|-------|-------------|
| [gavel-fixture-tester](gavel-fixture-tester/SKILL.md) | Author fixture tests and TODO executable definitions of done |
| [gavel-runner](gavel-runner/SKILL.md) | Run gavel test and lint, focus on a subset, re-run only failures, filter noise with baselines, and pull JSON/markdown results from finished or live runs |
| [gavel-git](gavel-git/SKILL.md) | Use gavel instead of gh/git for pull requests, CI status, and commits — inspect checks with `gavel pr status`, open PRs with AI-generated content, and commit with `gavel commit` |
| [gavel-todos](gavel-todos/SKILL.md) | Manage TODO content and lifecycle, then execute the persisted definition of done with `gavel todos check` |
| [gavel-triage](gavel-triage/SKILL.md) | Shape a TODO backlog for external execution with scope, criteria, and executable definitions of done |
| [gavel-ci-migrator](gavel-ci-migrator/SKILL.md) | Migrate a repo's GitHub Actions lint/test workflows to the `flanksource/gavel` composite action — discover existing jobs, ask the user per-workflow whether to replace or add alongside, rewrite YAML, and verify with actionlint |

### gavel-fixture-tester

Write data-driven CLI tests as markdown files with:

- **YAML front-matter** for build commands, default executables, env vars, and template variables
- **Markdown tables** where each row is a test case with custom column variables
- **Command blocks** for multi-line scripts with setup/teardown
- **CEL expressions** for flexible output assertions (`stdout`, `stderr`, `exitCode`, `json`, `ansi`)
- **TODO Verification content** as the persisted executable definition of done

```bash
gavel fixtures tests.md
gavel fixtures fixtures/**/*.md
```

See [gavel-fixture-tester/SKILL.md](gavel-fixture-tester/SKILL.md) for full documentation and live examples.

### gavel-runner

Drive the everyday test + lint loop:

- **Subset** — `gavel test ./pkg/foo`, `--changed`, `--cache`, `--framework`, framework-aware focus (`-- -run` / `-- --focus`), and single-framework raw pass-through
- **Re-run failures** — `gavel test --failed` (defaults to `.gavel/last.json`)
- **Suppress noise** — `--baseline` against a saved snapshot
- **Output** — `--format "json=…,markdown=…,html=…"` and `gavel summary` for PR-comment-shaped markdown
- **Live runs** — `gavel test --ui` exposes an HTTP+SSE API for snapshot, stream, rerun, stop
- **Timeouts** — `--timeout`, `--test-timeout`, `--lint-timeout`

```bash
gavel test --lint
gavel test --failed
gavel test --ui
```

See [gavel-runner/SKILL.md](gavel-runner/SKILL.md) for the full reference and `jq`/`curl` recipes.

### gavel-git

Reach for `gavel` instead of raw `gh`/`git` for PR, CI, and commit work:

- **PR + CI status** — `gavel pr status [--logs] [--follow]` replaces `gh pr view` / `gh run view` / `gh run list`
- **Open a PR** — `gavel commit -p` or `gavel pr create <SHA>` (AI-generated title/body/branch)
- **Commit** — `gavel commit` (session-scoped, conventional message, hooks); `-i` tree picker, `-A` split by directory, `--fixup`
- **CI → fixes** — `gavel pr status --ai-fix`

```bash
gavel pr status --logs
gavel commit -p
gavel commit -i
```

See [gavel-git/SKILL.md](gavel-git/SKILL.md) for the full intent→command cheatsheet.

### gavel-todos

Manage the TODO lifecycle without starting implementation:

- **Prepare** — create and shape TODOs, attach plans, and manage plan review for the external executor
- **Check** — `gavel todos check` executes the persisted Verification content as the executable definition of done
- **Feed** — `gavel todos sync` for source comments, `gavel todos create` for explicit work

```bash
gavel todos check
gavel todos list
```

See [gavel-todos/SKILL.md](gavel-todos/SKILL.md) for the TODO model, external-execution boundary, plan review, and definition-of-done checks.

### gavel-triage

Turn a pile of titles into a queue that is ready for the external executor:

- **Sweep** — `gavel todos list --format json` carries `markdown_body`, `acceptance_criteria`, and `verification` per item, so `jq` finds title-only, unprovable, and repeatedly-failing work directly
- **Verdict** — every TODO leaves triage as ready, shape, investigate, already-done, or retire
- **Prove, don't guess** — `gavel todos check` decides "already done" from the fixture, never from reading the code
- **Shape** — rewrite the body with problem statement, scope, acceptance criteria, and an executable `## Verification` definition of done
- **Retire** — comment the rationale, link the survivor, then close; there is no merge command, but `todos link` records the pairing

```bash
gavel todos list --format json > .tmp/todos.json
gavel todos check --status unverified
gavel todos edit <id> --body @.tmp/body.md
gavel todos edit <id> --status completed --priority low
gavel todos link <duplicate> <survivor>
```

See [gavel-triage/SKILL.md](gavel-triage/SKILL.md) for the readiness rubric, the durable-vs-projected status model (projected statuses are refused, not silently dropped), and the `depends_on` / `related_to` / derived `blocks` relations.

### gavel-ci-migrator

Move a repo's lint/test pipelines onto the [`flanksource/gavel` GitHub Action](../../README.md#github-action):

- **Discover** existing `golangci-lint`, `go test`, `make lint|test`, `ginkgo`, `gotestsum` jobs across `.github/workflows/`
- **Propose** per-workflow disposition via `AskUserQuestion` — replace, add alongside, or skip; pin the action to `@main`, `@v<tag>`, or `source`
- **Apply** the smallest-blast-radius YAML rewrite — keeps unrelated steps, sets `fetch-depth: 0`, scopes `permissions:` correctly, removes redundant artifact uploads
- **Verify** with `actionlint` (if installed) and `git diff` — never auto-commits

```text
"switch CI to gavel"
"replace golangci-lint-action with gavel"
"migrate workflows to flanksource/gavel"
```

See [gavel-ci-migrator/SKILL.md](gavel-ci-migrator/SKILL.md) for detection cues, canonical replacement snippets, and edge cases (matrix collapse, monorepos, coverage uploads).

## Prerequisites

- [Gavel](https://github.com/flanksource/gavel) installed for running fixture tests (`go install github.com/flanksource/gavel/cmd/gavel@latest`)

## Directory structure

```
.agents/
├── .claude-plugin/
│   └── plugin.json              # Claude Code plugin manifest
├── skills/
│   ├── gavel-fixture-tester/
│   │   └── SKILL.md             # Skill instructions (agentskills.io format)
│   ├── gavel-runner/
│   │   └── SKILL.md             # Skill instructions (agentskills.io format)
│   ├── gavel-git/
│   │   └── SKILL.md             # Skill instructions (agentskills.io format)
│   ├── gavel-todos/
│   │   └── SKILL.md             # Skill instructions (agentskills.io format)
│   ├── gavel-triage/
│   │   └── SKILL.md             # Skill instructions (agentskills.io format)
│   ├── gavel-ci-migrator/
│   │   └── SKILL.md             # Skill instructions (agentskills.io format)
│   └── README.md
.claude-plugin/
└── marketplace.json             # Claude Code marketplace registry
```
