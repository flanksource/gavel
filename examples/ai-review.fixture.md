---
ai:
  model: claude-code-opus
verify:
  threshold: 80
---

# AI code review

An AI reviewer inspects the change (the working-tree diff by default) against the
acceptance criteria below and returns a structured score. Checklist items whose
text is a built-in check ID (run `gavel verify --list` to see them, e.g.
`tests-added`, `no-hardcoded-secrets`) run that static check; every other item is
scored by the model as a free-text criterion. Run it with:

    gavel fixtures examples/ai-review.fixture.md

Requires an AI backend — configure one first with `captain configure` (or set
`ANTHROPIC_API_KEY`). Change the model, threshold, prompt, and criteria to fit
your project. Set `verify.scope` to review something other than the working-tree
diff (e.g. a git ref or range).

```prompt
You are a senior engineer reviewing this change. Focus on correctness and test
coverage. Confirm the change does what its description claims, that failure paths
are tested, and that no secrets are introduced. Be concrete — cite files and
lines, and prefer actionable feedback over style nits.
```

## Acceptance Criteria

- [ ] tests-added
- [ ] error-paths-tested
- [ ] no-hardcoded-secrets
- [ ] The change is covered by a test that fails without it
- [ ] User-facing behavior changes are reflected in the README / docs

- cel: json.score >= 80



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
