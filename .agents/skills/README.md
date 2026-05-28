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
| [gavel-fixture-tester](gavel-fixture-tester/SKILL.md) | Create and run fixture-based tests using markdown files with command blocks, tables, and CEL assertions |
| [gavel-runner](gavel-runner/SKILL.md) | Run gavel test and lint, focus on a subset, re-run only failures, filter noise with baselines, and pull JSON/markdown results from finished or live runs |
| [gavel-ci-migrator](gavel-ci-migrator/SKILL.md) | Migrate a repo's GitHub Actions lint/test workflows to the `flanksource/gavel` composite action — discover existing jobs, ask the user per-workflow whether to replace or add alongside, rewrite YAML, and verify with actionlint |

### gavel-fixture-tester

Write data-driven CLI tests as markdown files with:

- **YAML front-matter** for build commands, default executables, env vars, and template variables
- **Markdown tables** where each row is a test case with custom column variables
- **Command blocks** for multi-line scripts with setup/teardown
- **CEL expressions** for flexible output assertions (`stdout`, `stderr`, `exitCode`, `json`, `ansi`)

```bash
gavel fixtures tests.md
gavel fixtures fixtures/**/*.md
```

See [gavel-fixture-tester/SKILL.md](gavel-fixture-tester/SKILL.md) for full documentation and live examples.

### gavel-runner

Drive the everyday test + lint loop:

- **Subset** — `gavel test ./pkg/foo`, `--changed`, `--cache`, `--framework`, runner pass-through (`-- --focus`)
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
│   ├── gavel-ci-migrator/
│   │   └── SKILL.md             # Skill instructions (agentskills.io format)
│   └── README.md
.claude-plugin/
└── marketplace.json             # Claude Code marketplace registry
```
