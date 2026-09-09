---
build: go build ./examples/sample-app
---

# Pre-release gate

Comprehensive checks before cutting a release, run against the bundled
[`examples/sample-app`](sample-app). The `build:` front-matter compiles it first
(from the git root); a failed build skips the rest. Point the paths at your own
code when you copy this into your project.

    gavel fixtures examples/pre-release.fixture.md

## Full test suite

```yaml test
paths: [./examples/sample-app]
framework: [go]
pre-build: true
test-timeout: 5m
show-passed: true
```

## Lint

```yaml lint
files: [./examples/sample-app]
timeout: 5m
```
