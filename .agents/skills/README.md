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
│   └── README.md
.claude-plugin/
└── marketplace.json             # Claude Code marketplace registry
```
