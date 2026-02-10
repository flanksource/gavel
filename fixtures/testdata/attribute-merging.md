---
codeBlocks: [bash]
---

# Test Attribute Merging

### command: Inline exitCode overrides YAML

Inline exitCode=0 should override YAML exitCode=1.

```yaml
exitCode: 1
timeout: 30
```

```bash exitCode=0
echo "inline exitCode wins"
```

### command: Non-conflicting fields merge

Inline provides exitCode, YAML provides timeout - both should apply.

```yaml
timeout: 20
```

```bash exitCode=0
echo "merged output"
```
