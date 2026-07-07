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
| pr fix help has examples | pr fix --help | stdout.contains("Examples") |
| pr list help has examples | pr list --help | stdout.contains("Examples") |
| todos parent help has examples | todos --help | stdout.contains("Examples") |
| todos run help has examples | todos run --help | stdout.contains("Examples") |
| commit help has examples | commit --help | stdout.contains("Examples") |

## todos verify is the AI-review surface

| Name | Args | CEL |
|------|------|-----|
| verify help names acceptance criteria | todos verify --help | stdout.contains("acceptance criteria") |
| verify help notes it replaces gavel verify | todos verify --help | stdout.contains("gavel verify") |

## Help no longer uses the old fork name

| Name | Args | CEL |
|------|------|-----|
| test help is not forked | test --help | !stdout.contains("arch-unit") |
| test help mentions gavel test | test --help | stdout.contains("gavel test") |
