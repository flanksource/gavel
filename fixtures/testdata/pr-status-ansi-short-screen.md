---
cwd: ../..
timeout: 5m
record:
  ansi: {width: 100, height: 24}
---

# gavel pr status — end state on a screen shorter than the report

Companion to `pr-status-ansi.md`, which runs the same command at 120x45 where
the report fits. This file exists as a separate fixture rather than a section
there because the PTY is sized once per file: a per-test `record:` block is
merged into the test but does not resize the terminal the recorder already
allocated.

At 100x24 the ~34-line report overflows, the top scrolls away, and the shell
prompt lands on the last row. The tail must survive intact and free of duplicate
lines — a redraw that miscounted physical rows (a line wider than the terminal
soft-wraps onto extra rows) would stack fragments of the previous frame here,
and `cast.duplicates` is what catches that.

### command: the tail of the report survives an overflowing screen

```yaml
exitCode: 1
```

```bash
gavel pr status --repo flanksource/clicky-ui 61
```

- cel: cast.width == 100 && cast.height == 24
- cel: cast.has_duplicates == false
- cel: cast.final.contains("Show more details")
- cel: cast.final.contains("Comments: (1)")
- cel: ansi.alt_screen == false
- cel: ansi.stray_controls == false
