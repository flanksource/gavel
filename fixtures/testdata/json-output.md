---
codeBlocks: [bash]
---

# JSON Output Tests

## gavel fixtures --json

### command: valid JSON from gavel fixtures

```yaml
exitCode: 0
timeout: 60
```

```bash
gavel fixtures --json ../../fixtures/testdata/codeblocks.md 2>/dev/null | jq .
```

- cel: exitCode == 0

### command: fixtures JSON has expected fields

```yaml
exitCode: 0
timeout: 60
```

```bash
gavel fixtures --json ../../fixtures/testdata/codeblocks.md 2>/dev/null | jq '.name, (.children | length)'
```

- cel: stdout.contains("codeblocks.md")

### command: fixture test results have status

```yaml
exitCode: 0
timeout: 60
```

```bash
gavel fixtures --json ../../fixtures/testdata/codeblocks.md 2>/dev/null | jq '[.. | .results? // empty | .status]'
```

- contains: PASS
