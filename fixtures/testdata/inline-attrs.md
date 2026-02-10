---
codeBlocks: [bash]
---

# Test Inline Attributes

### command: ExitCode attribute

Command should fail with exit code 1.

```bash exitCode=1
exit 1
```

### Non-Zero exit failure

This should fail:

```bash exitCode=1
exit 0
```

### command: ExitCode with stdout

Command should succeed and output text.

```bash exitCode=0
echo "success output"
```

### command: Timeout attribute

Command should complete within timeout.

```bash exitCode=0 timeout=5
sleep 1 && echo "completed in time"
```

```bash exitCode=0 timeout=5
sleep 10 && echo "did not completed in time"
```
