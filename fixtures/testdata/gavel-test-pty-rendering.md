---
terminal: pty
cwd: ../..
env:
  CLICKY_FORCE_INTERACTIVE: "1"
  TERM: "xterm-256color"
timeout: 3m
---

# gavel test — PTY/ANSI rendering audit

Runs the installed `gavel` binary under a real PTY and asserts structural
properties of the ANSI byte stream it emits. Pre-requisite: `make build install`
has produced an up-to-date `gavel` on `$PATH`.

The inner run targets a single small Go package (`./testrunner/parsers`) so
the stream is short and bounded. `--test-timeout 1m` caps each subprocess; the
surrounding fixture `timeout: 3m` caps total wall-clock.

`CLICKY_FORCE_INTERACTIVE=1` keeps clicky's task manager in interactive mode
under the outer `gavel fixtures` driver. Without it, clicky's
`isTestEnvironment()` guard would disable progress/color rendering and the
stream would be uninteresting.

Assertions are precomputed on the PTY capture and exposed as booleans on the
`ansi` map so fixtures don't need to thread ANSI escape bytes through
Markdown/YAML/CEL escape rules. Validations below are joined with `&&` by the
fixture parser.

### command: interactive renderer engages and shuts down cleanly

```bash
gavel test --test-timeout 1m ./testrunner/parsers
```

- cel: exitCode == 0
- cel: ansi.has_any
- cel: ansi.has_color
- cel: ansi.has_updates
- cel: ansi.has_cursor_hide
- cel: ansi.has_cursor_show
- cel: ansi.has_reset
- cel: !ansi.stray_controls
- cel: !ansi.has_duplicates
