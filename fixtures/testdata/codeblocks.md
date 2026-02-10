---
codeBlocks: [bash]
---

# Test CodeBlocks Filtering

### command: Bash executes, others skipped

Bash block should execute, python should be skipped.

```yaml
exitCode: 0
```

```bash
echo "bash executed"
```

```python
print("should not execute")
```

### command: Unlabeled blocks skipped

Blocks without language labels should be skipped.

```yaml
exitCode: 0
```

```bash
echo "labeled bash"
```

```
echo "unlabeled, should skip"
```
