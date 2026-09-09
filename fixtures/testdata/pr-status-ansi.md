---
cwd: ../..
timeout: 5m
record:
  ansi: {width: 120, height: 45}
---

# gavel pr status — ANSI rendering audit

Runs the installed `gavel` binary under a real PTY against a known merged PR and
asserts structural properties of both the ANSI byte stream and the settled
screen. Pre-requisite: `make build install` has produced an up-to-date `gavel`
on `$PATH`, and the GitHub API is reachable.

`flanksource/clicky-ui#61` is a merged PR whose checks are complete, so the
report is stable: one failing rollup check (CodeQL), two runs of the same
`Storybook` workflow, skipped jobs, and a gavel artifact summary. The failing
check is what makes the expected exit code 1.

`record: ansi` implies `terminal: pty` and exposes the `cast` root — the settled
final screen, and the duplicate-line report that is the tell-tale of a redraw
which left stale content behind. The `ansi` root carries structural booleans
computed over the raw stream, so assertions never have to thread escape bytes
through Markdown/YAML/CEL quoting. Negations are written `== false` rather than
`!expr`, because a leading `!` is a YAML tag and quoting it to escape that turns
the expression into a string literal once the validations are joined with `&&`.

### command: the report renders cleanly on a screen taller than it

The report is ~34 lines, so at height 45 it fits without scrolling and the
settled screen holds all of it.

```yaml
exitCode: 1
```

```bash
gavel pr status --repo flanksource/clicky-ui 61
```

- cel: cast.exit_code == 1
- cel: cast.width == 120 && cast.height == 45
- cel: ansi.has_any && ansi.has_color && ansi.has_reset
- cel: ansi.stray_controls == false
- cel: cast.has_duplicates == false
- cel: cast.final.contains("PR #61")
- cel: cast.final.contains("Show more details")

### command: header drops the meaningless mergeable column

GitHub reports UNKNOWN mergeability once a PR is merged; rendered bare beside
the state it read as an error rather than an absent answer.

```yaml
exitCode: 1
```

```bash
gavel pr status --repo flanksource/clicky-ui 61
```

- cel: cast.final.contains("MERGED")
- cel: cast.final.contains("UNKNOWN") == false
- cel: cast.final.contains("Mergeable:") == false

### command: workflows are disambiguated and skipped jobs carry no duration

Two runs of `Storybook` previously rendered under identical headings with
mirrored children, reading as a duplicated section; skipped jobs showed `(0s)`
as though the work had been measured.

```yaml
exitCode: 1
```

```bash
gavel pr status --repo flanksource/clicky-ui 61
```

- cel: cast.final.contains("Storybook (run ")
- cel: cast.final.contains("(0s)") == false
- cel: cast.final.contains("Delete Storybook Preview")

### command: a failing rollup check links somewhere actionable

`CodeQL` has no jobs to expand and is what drives the non-zero exit code, so it
has to carry its details URL.

```yaml
exitCode: 1
```

```bash
gavel pr status --repo flanksource/clicky-ui 61
```

- cel: cast.final.contains("CodeQL")
- cel: cast.final.contains("https://github.com/flanksource/clicky-ui/runs/")

### command: closed-PR bot notices are not counted as actionable comments

The CodeRabbit "review failed, the pull request is closed" notice and the
pr-preview teardown notice carry no action.

```yaml
exitCode: 1
```

```bash
gavel pr status --repo flanksource/clicky-ui 61
```

- cel: cast.final.contains("Comments: (1)")
- cel: cast.final.contains("The pull request is closed") == false
- cel: cast.final.contains("Preview removed") == false

The overflowing-screen case lives in `pr-status-ansi-short-screen.md`: a
per-test `record:` block does not resize the PTY, because the recorder is
created once for the file.

### command: no terminal-state junk reaches the terminal

Nothing enters the alternate screen, so nothing may emit the sequence that
leaves it — those bytes were previously written twice on every invocation. The
shutdown hooks logged four INF lines to stderr, which a PTY merges into the
report.

```yaml
exitCode: 1
```

```bash
gavel pr status --repo flanksource/clicky-ui 61
```

- cel: ansi.alt_screen == false
- cel: ansi.stray_controls == false
- cel: stdout.contains("All shutdown hooks executed") == false
- cel: stdout.contains("All tasks completed gracefully") == false
