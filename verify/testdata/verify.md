---
build: make build install
codeBlocks: [bash]
---

# Verify Command E2E Tests

## Help and Error Cases

### command: verify help

```yaml
exitCode: 0
```

```bash
gavel verify --help
```

- contains: "AI-powered code review"

## Claude CLI

### command: verify with claude

```yaml
exitCode: 0
timeout: 120
```

```bash
which claude || exit 0
gavel verify --model claude
```

### command: verify commit range with claude

```yaml
exitCode: 0
timeout: 120
```

```bash
which claude || exit 0
gavel verify --model claude --range HEAD~1..HEAD
```

### command: verify single file with claude

```yaml
exitCode: 0
timeout: 120
```

```bash
which claude || exit 0
gavel verify --model claude verify/verify.go
```

## Gemini CLI

### command: verify with gemini

```yaml
exitCode: 0
timeout: 120
```

```bash
which gemini || exit 0
gavel verify --model gemini-3-pro-previw
```

## Codex CLI

### command: verify with codex

```yaml
exitCode: 0
timeout: 120
```

```bash
which codex || exit 0
gavel verify --model codex-gpt-5.2
```
