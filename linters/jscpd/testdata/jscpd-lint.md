---
codeBlocks: [bash]
timeout: 60s
---

# jscpd Linter Tests

Tests run jscpd directly to verify duplicate detection and ignore flag support.

## Direct jscpd Detection

### command: detects duplicates in sample files

```bash
SAMPLE="$GIT_ROOT_DIR/linters/jscpd/testdata/sample"
OUTDIR=$(mktemp -d)
jscpd --reporters json --output "$OUTDIR" --min-lines 3 --min-tokens 20 "$SAMPLE" >/dev/null 2>&1 || true
cat "$OUTDIR/jscpd-report.json"
rm -rf "$OUTDIR"
```

- cel: json.duplicates.size() >= 2
- cel: json.duplicates.exists(d, d.format == "go")

### command: reports both source files in duplicates

```bash
SAMPLE="$GIT_ROOT_DIR/linters/jscpd/testdata/sample"
OUTDIR=$(mktemp -d)
jscpd --reporters json --output "$OUTDIR" --min-lines 3 --min-tokens 20 "$SAMPLE" >/dev/null 2>&1 || true
cat "$OUTDIR/jscpd-report.json"
rm -rf "$OUTDIR"
```

- cel: json.duplicates.exists(d, d.firstFile.name.contains("file_a"))
- cel: json.duplicates.exists(d, d.secondFile.name.contains("file_b"))

## Ignore Pattern Support

### command: ignore glob excludes test files from scan

```bash
SAMPLE="$GIT_ROOT_DIR/linters/jscpd/testdata/sample"
OUTDIR=$(mktemp -d)
jscpd --reporters json --output "$OUTDIR" --min-lines 3 --min-tokens 20 --ignore "**/*_test.go" "$SAMPLE" >/dev/null 2>&1 || true
cat "$OUTDIR/jscpd-report.json"
rm -rf "$OUTDIR"
```

- cel: json.duplicates.size() == 2

### command: ignore all go files produces no duplicates

```bash
SAMPLE="$GIT_ROOT_DIR/linters/jscpd/testdata/sample"
OUTDIR=$(mktemp -d)
jscpd --reporters json --output "$OUTDIR" --min-lines 3 --min-tokens 20 --ignore "**/*.go" "$SAMPLE" >/dev/null 2>&1 || true
if [ -f "$OUTDIR/jscpd-report.json" ]; then
  cat "$OUTDIR/jscpd-report.json"
else
  echo '{"duplicates":[]}'
fi
rm -rf "$OUTDIR"
```

- cel: json.duplicates.size() == 0
