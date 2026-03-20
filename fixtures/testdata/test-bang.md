### command: Debug config loading

```bash
REPO=$(cd ../.. && pwd)
cat > "$REPO/.gitanalyze.yaml" <<'YAML'
ignore_commit_types:
  - '!feat'
YAML
gavel git analyze --json --path "$REPO" --arg HEAD~5..HEAD --verbose 2>&1
rm -f "$REPO/.gitanalyze.yaml"
```

- cel: stderr.contains('Skipping commit')
