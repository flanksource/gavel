---
exec: "{{.executablePath}}"
---

# CLI help contract

Asserts that the upgraded `--help` pages stay comprehensive and free of the old
fork name. `{{.executablePath}}` is the gavel binary running this fixture, so the
tests exercise its own help output.

    gavel fixtures examples/help.fixture.md

## Commands expose structured Examples

| Name | Args | CEL |
|------|------|-----|
| pr parent help mentions status | pr --help | stdout.contains("status") |
| pr create help has examples | pr create --help | stdout.contains("Examples") |
| pr list help has examples | pr list --help | stdout.contains("Examples") |
| todos parent help has examples | todos --help | stdout.contains("Examples") |
| todos run help has examples | todos run --help | stdout.contains("Examples") |
| todos import help names PostgreSQL | todos import --help | stdout.contains("PostgreSQL") |
| todos export help names PostgreSQL | todos export --help | stdout.contains("PostgreSQL") |
| commit help has examples | commit --help | stdout.contains("Examples") |

## todos check is the definition-of-done surface

| Name | Args | Exit Code | CEL |
|------|------|-----------|-----|
| check help names definition of done | todos check --help | 0 | stdout.contains("definition of done") |
| removed verify command is rejected | todos verify | 1 | stderr.contains('unknown command "verify"') |

## Help no longer uses the old fork name

| Name | Args | CEL |
|------|------|-----|
| test help is not forked | test --help | !stdout.contains("arch-unit") |
| test help mentions gavel test | test --help | stdout.contains("gavel test") |
