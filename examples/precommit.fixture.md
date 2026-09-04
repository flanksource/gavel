# Pre-commit checks

Fast checks to run before every commit. This template runs against the bundled
[`examples/sample-app`](sample-app); point the paths at your own code (or add
`changed: true` to scope to what you changed) when you copy it into your project.

    gavel fixtures examples/precommit.fixture.md

Runner-step paths are relative to the git root, like `gavel test` / `gavel lint`.

## Lint

```yaml lint
linters: [golangci-lint]
files: [./examples/sample-app]
```

## Test

```yaml test
paths: [./examples/sample-app]
framework: [go]
test-timeout: 2m
```

## Formatting is clean

```bash
test -z "$(gofmt -l "$GIT_ROOT_DIR/examples/sample-app")"
```
