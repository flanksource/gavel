---
codeBlocks: [bash]
---

# Test YAML Expects

### command: ExitCode and stdout

```yaml
exitCode: 0
```

```bash
echo "output text"
```

### command: Stderr validation

```yaml
exitCode: 1
```

```bash
echo "error message" >&2 && exit 1
```

### command: Multiple expectations

```yaml
exitCode: 0
timeout: 10
```

```bash
echo "success"
```
