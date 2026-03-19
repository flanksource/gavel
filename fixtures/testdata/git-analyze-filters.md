# Git Analyze Filter Tests

## CLI Flag Filters

| Name | Command | Exit Code | CEL Validation |
|------|---------|-----------|----------------|
| Commit type filter feat has changes | gavel git analyze --json --path ../.. --commit-types feat --arg HEAD~20..HEAD 2>/dev/null | 0 | stdout.contains('"commit_type": "feat"') |
| Commit type chore returns no changes | gavel git analyze --json --path ../.. --commit-types chore --arg HEAD~20..HEAD 2>/dev/null | 0 | !stdout.contains('"type": "modified"') && !stdout.contains('"type": "added"') |
| Author filter matches | gavel git analyze --json --path ../.. --author moshe --arg HEAD~10..HEAD 2>/dev/null | 0 | stdout.contains('"name": "Moshe Immerman"') |
| Nonexistent type returns no changes | gavel git analyze --json --path ../.. --commit-types nonexistent --arg HEAD~5..HEAD 2>/dev/null | 0 | !stdout.contains('"type": "modified"') && !stdout.contains('"type": "added"') |
| JSON output is valid array | gavel git analyze --json --path ../.. --arg HEAD~3..HEAD 2>/dev/null | 0 | stdout.startsWith('[') && stdout.endsWith(']') |
| Scope filter backend | gavel git analyze --json --path ../.. --scope backend --arg HEAD~20..HEAD 2>/dev/null | 0 | !stdout.contains('"scope": "docs"') |
| Tech filter go | gavel git analyze --json --path ../.. --tech go --arg HEAD~20..HEAD 2>/dev/null | 0 | stdout.contains('"go"') && !stdout.contains('"markdown"') |
| Negation commit type !chore | gavel git analyze --json --path ../.. --commit-types '!chore' --arg HEAD~20..HEAD 2>/dev/null | 0 | !stdout.contains('"commit_type": "chore"') |

## Default Config Filters

| Name | Command | Exit Code | CEL Validation |
|------|---------|-----------|----------------|
| Default noise filters go.sum | gavel git analyze --json --path ../.. --arg HEAD~20..HEAD 2>/dev/null | 0 | !stdout.contains('"go.sum"') |
| Default noise filters lock files | gavel git analyze --json --path ../.. --arg HEAD~20..HEAD 2>/dev/null | 0 | !stdout.contains('.lock"') |

## Include Exclude Filter Sets

| Name | Command | Exit Code | CEL Validation |
|------|---------|-----------|----------------|
| Exclude noise shows go.sum | gavel git analyze --json --path ../.. --exclude noise --arg HEAD~50..HEAD 2>/dev/null | 0 | stdout.contains('"go.sum"') |
| Exclude all defaults shows raw output | gavel git analyze --json --path ../.. --exclude bots --exclude noise --exclude merges --arg HEAD~50..HEAD 2>/dev/null | 0 | stdout.contains('"go.sum"') |

## Custom Config Filters

### command: Config ignore commit types chore and ci

```bash
REPO=$(cd ../.. && pwd)
cat > "$REPO/.gitanalyze.yaml" <<'YAML'
ignore_commit_types:
  - chore
  - ci
YAML
gavel git analyze --json --path "$REPO" --arg HEAD~20..HEAD 2>/dev/null
rm -f "$REPO/.gitanalyze.yaml"
```

- cel: !stdout.contains('"commit_type": "chore"') && !stdout.contains('"commit_type": "ci"')

### command: Config file filter removes go.sum and lock files

```bash
REPO=$(cd ../.. && pwd)
cat > "$REPO/.gitanalyze.yaml" <<'YAML'
ignore_files:
  - '*.lock'
  - 'go.sum'
YAML
gavel git analyze --json --path "$REPO" --arg HEAD~20..HEAD 2>/dev/null
rm -f "$REPO/.gitanalyze.yaml"
```

- cel: !stdout.contains('"go.sum"') && !stdout.contains('.lock"')

### command: Config CEL rule skips merge commits

```bash
REPO=$(cd ../.. && pwd)
cat > "$REPO/.gitanalyze.yaml" <<'YAML'
ignore_commit_rules:
  - cel: "commit.is_merge"
YAML
gavel git analyze --json --path "$REPO" --arg HEAD~50..HEAD 2>/dev/null
rm -f "$REPO/.gitanalyze.yaml"
```

- cel: !stdout.contains('"Merge ')

### command: Config CEL commit.message matches subject

```bash
REPO=$(cd ../.. && pwd)
cat > "$REPO/.gitanalyze.yaml" <<'YAML'
ignore_commit_rules:
  - cel: 'commit.message.startsWith("docs:")'
YAML
gavel git analyze --json --path "$REPO" --arg HEAD~20..HEAD 2>/dev/null
rm -f "$REPO/.gitanalyze.yaml"
```

- cel: !stdout.contains('"subject": "docs:')

### command: Config commit type negation skips non-feat

```bash
REPO=$(cd ../.. && pwd)
cat > "$REPO/.gitanalyze.yaml" <<'YAML'
ignore_commit_types:
  - '!feat'
YAML
gavel git analyze --json --path "$REPO" --arg HEAD~20..HEAD 2>/dev/null
rm -f "$REPO/.gitanalyze.yaml"
```

- cel: stdout.contains('"commit_type": "feat"') && !stdout.contains('"commit_type": "fix"') && !stdout.contains('"commit_type": "docs"')

### command: Config file negation preserves test files

```bash
REPO=$(cd ../.. && pwd)
cat > "$REPO/.gitanalyze.yaml" <<'YAML'
ignore_files:
  - '*.go'
  - '!*_test.go'
YAML
gavel git analyze --json --path "$REPO" --arg HEAD~20..HEAD 2>/dev/null
rm -f "$REPO/.gitanalyze.yaml"
```

- cel: stdout.contains('_test.go"') && !stdout.contains('"file": "main.go"')

### command: Config no extra filters passes everything through

```bash
REPO=$(cd ../.. && pwd)
rm -f "$REPO/.gitanalyze.yaml"
gavel git analyze --json --path "$REPO" --exclude bots --exclude noise --exclude merges --arg HEAD~5..HEAD 2>/dev/null
```

- cel: stdout.contains('"changes"') && stdout.startsWith('[')
