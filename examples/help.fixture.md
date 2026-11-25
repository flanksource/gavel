---
exec: "{{.executablePath}}"
---

# CLI help contract

Asserts that the upgraded `--help` pages stay comprehensive and free of the old
fork name. `{{.executablePath}}` is the gavel binary running this fixture, so the
tests exercise its own help output.

    gavel fixtures examples/help.fixture.md

## Global and command flags

| Name | Args | Exit code | CEL |
|------|------|-----------|-----|
| Root publishes global flags and discovers commands | --help | 0 | stdout.contains("Available Commands:") && stdout.contains("--cwd") && stdout.contains("--no-progress") && !stdout.contains("--work-dir") |
| Lint shows command flags and points to root help | lint --help | 0 | stdout.contains("--linters") && !stdout.contains("--no-progress") && stdout.contains("Use gavel --help to show arguments available to all commands.") |
| Cobra help uses framework-specific usage | help test go | 0 | stdout.contains("gavel test go [flags] [path...] [-- go test args...]") && stdout.contains("Use gavel --help to show arguments available to all commands.") && !stdout.contains("--no-progress") |
| Custom fixture help uses the same footer | fixtures --help | 0 | stderr.contains("Use gavel --help to show arguments available to all commands.") && !stderr.contains("Maximum concurrent tasks") |
| Subcommands still accept global flags | version --no-color | 0 | stdout.contains("gavel ") |

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

## Options-struct help reaches cobra

Clicky only wires an Options struct's long help into `cmd.Long` when it satisfies
`entity.Help` exactly (`Help() api.Textable`). A near-miss signature compiles and
is silently ignored, and an inherited parent help func shadows a subcommand's own
page — both leave `--help` with nothing but the flag list.

| Name | Args | CEL |
|------|------|-----|
| pr status help has examples | pr status --help | stdout.contains("Examples") |
| proc start help describes the daemon | proc start --help | stdout.contains("detached background daemon") |
| proc status help describes the control socket | proc status --help | stdout.contains("control socket") |
| system install help covers auth | system install --help | stdout.contains("GitHub authentication") |
| system status help names the shared status source | system status --help | stdout.contains("single source of truth") |
| ssh install help names systemd | ssh install --help | stdout.contains("systemd") |
| ui serve help documents auto-stop | ui serve --help | stdout.contains("auto-stop") |
| repomap view help has examples | repomap view --help | stdout.contains("EXAMPLES") |
| test outline help lists every source | test outline --help | stdout.contains("Markdown fixture") |
| test history help explains the columns | test history --help | stdout.contains("pass rate") |
| fixtures outline help is not the parent page | fixtures outline --help | stdout.contains("only parses fixture files") |

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
