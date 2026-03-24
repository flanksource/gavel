# gavel-skills

Claude Code skill plugin for the [Gavel](https://github.com/flanksource/gavel) CLI testing framework.

## Installation

```bash
claude plugin install /path/to/gavel/.claude/skills
```

Or add to your project's `.claude/settings.json`:

```json
{
  "plugins": [
    ".claude/skills"
  ]
}
```

## Skills

| Skill | Command | Description |
|-------|---------|-------------|
| gavel-fixture-tester | `/gavel-fixture-tester` | Create and run gavel fixture-based tests using markdown files with command blocks, tables, and CEL assertions |

### gavel-fixture-tester

Write data-driven CLI tests as markdown files with:

- **YAML front-matter** for build commands, default executables, env vars, and template variables
- **Markdown tables** where each row is a test case with custom column variables
- **Command blocks** for multi-line scripts with setup/teardown
- **CEL expressions** for flexible output assertions (`stdout`, `stderr`, `exitCode`, `json`, `ansi`)

```bash
# Run fixtures
gavel fixtures tests.md
gavel fixtures fixtures/**/*.md
```

See [`skills/gavel-fixture-tester/SKILL.md`](skills/gavel-fixture-tester/SKILL.md) for full documentation and live examples.
