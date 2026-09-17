---
name: gavel-ci-migrator
description: Migrate a repo's GitHub Actions lint and test workflows onto the flanksource/gavel composite action — discover existing golangci-lint / go test / make lint|test jobs, ask the user per-workflow whether to replace, add alongside, or skip, rewrite YAML with the right permissions and checkout depth, and verify with actionlint without auto-committing
allowed-tools: [Read, Write, Edit, Bash, Glob, Grep, AskUserQuestion]
---

# Gavel CI Migrator

Switch a repository's GitHub Actions lint/test pipelines to call the [`flanksource/gavel`](https://github.com/flanksource/gavel) composite action instead of running `golangci-lint` / `go test` / `make lint` directly. The action runs `gavel test --lint`, uploads a JSON+HTML+log artifact bundle, and posts a sticky PR comment with a markdown summary.

For *running* gavel locally, defer to [`gavel-runner`](../gavel-runner/SKILL.md). For *writing* fixture tests, defer to [`gavel-fixture-tester`](../gavel-fixture-tester/SKILL.md). This skill is exclusively about migrating CI pipelines.

The skill is **read-only until the user explicitly picks a disposition** for each workflow. It never auto-commits.

## When to Use

- "switch CI to gavel" / "use the gavel action for lint and test"
- "migrate this repo's workflows to flanksource/gavel"
- "replace golangci-lint-action with gavel"
- After adding a `.gavel.yaml` to a repo that still runs lint/test the old way

## When NOT to Use

- Repo has no `.github/workflows/` directory — nothing to migrate.
- Repo is not Go-based — gavel today targets Go test/lint discovery; do not propose migration for pure JS/TS/Python repos.
- The workflow already uses `flanksource/gavel@…` — instead, offer to **pin or bump the version**, not re-migrate.
- The user wants to add gavel as a *brand-new* workflow and there is nothing existing to migrate — direct them to copy the snippet from the gavel `README.md` "GitHub Action" section; this skill is for moving off existing lint/test jobs.

## Phase 1 — Discover

Inventory CI before proposing anything.

1. List workflow files:
   - Glob `.github/workflows/*.yml` and `.github/workflows/*.yaml`.
   - Read each file fully — do not skim. Workflows are short.

2. For each file, classify every step in every job. Detection cues:

   **Lint cues** (any of these → step is a lint step):
   - `golangci-lint run`, `golangci/golangci-lint-action`
   - `staticcheck`, `revive`, `errcheck`
   - `gofmt -l`, `gofumpt -l`, `goimports -l`
   - `go vet`
   - `make lint`, `task lint`, `mage lint`

   **Test cues** (any of these → step is a test step):
   - `go test`, `gotestsum`, `ginkgo run`, `ginkgo -r`
   - `make test`, `task test`, `mage test`
   - Any step whose `name:` matches `/test|coverage|spec/i` and runs Go

   **Already-gavel cues** (skip — see "When NOT to Use"):
   - `uses: flanksource/gavel@…`
   - `gavel test`, `gavel lint`, `gavel test --lint` in a `run:` block

3. For each workflow, record:
   - Filename and triggering events (`on: push`, `on: pull_request`, …)
   - Job name(s) and `runs-on:`
   - `actions/setup-go` version source (`go-version-file:` or pinned `go-version:`)
   - `actions/checkout` `fetch-depth:` (gavel needs `0` for `--changed` and PR diff)
   - Current `permissions:` block
   - Matrix dimensions if any (OS × Go version)
   - Artifact-upload steps tied to lint/test output (will become redundant)

4. Summarize findings to the user before asking anything. Example (do not paraphrase — show real values):

   ```
   Found 3 workflows:
   - .github/workflows/lint.yml  → 1 lint job (golangci-lint-action v9), no PR comment
   - .github/workflows/test.yml  → 1 test job (go test ./...), uploads coverage.out
   - .github/workflows/release.yml → release-only, no lint/test (will skip)
   ```

## Phase 2 — Propose

Use `AskUserQuestion` with structured options. Two questions, asked together (single `AskUserQuestion` call, two entries in `questions`):

### Question 1 — per-workflow disposition

Multi-select. One option per detected lint/test workflow. Labels:
- `lint.yml: Replace with gavel args:"lint"`
- `lint.yml: Add gavel alongside (keep current job)`
- `lint.yml: Skip`
- `test.yml: Replace with gavel args:"test"`
- `test.yml: Add gavel alongside (keep current job)`
- `test.yml: Skip`

If a single workflow contains **both** a lint job and a test job, collapse to a `Replace with gavel args:"test --lint"` option instead of offering two replacements.

### Question 2 — action version pin

Single-select:
- `@main` — track latest (good for the gavel repo's own self-test and for early adopters)
- `@v<latest-tag>` — pinned to the most recent release tag (**recommended for downstream repos**; fetch the latest tag with `gh release view -R flanksource/gavel --json tagName -q .tagName`)
- `source` — use a pre-installed gavel binary on `$PATH`; only relevant when the calling workflow builds gavel from source first (this is what the gavel repo's own `.github/workflows/gavel-action.yml` does)

### Show the diff *before* asking

For each workflow the user might replace, render the proposed before/after directly in the message — not a summary. Use the canonical invocation from gavel's `action.yml` and `README.md` as the template. Example output to the user:

```yaml
# .github/workflows/lint.yml — proposed replacement

# BEFORE
- name: golangci-lint
  uses: golangci/golangci-lint-action@v9
  with:
    version: v2.11.4
    args: --timeout=10m --tests=false

# AFTER
- uses: flanksource/gavel@v<TAG>
  with:
    args: lint
    fail-on-error: "true"
```

## Phase 3 — Apply

Once the user has picked, rewrite the workflow files. Use `Edit` (preferred) or `Write`. **Never** use heredoc / `cat <<EOF`.

### Rewrite rules

- **Smallest blast radius**: replace only the steps that map to gavel. Do not rewrite a whole job if only the lint step is being migrated. Preserve unrelated steps (release uploads, docker pushes, slack notifications, codecov uploads that the team wants to keep).
- **Checkout**: ensure `actions/checkout@v4` precedes the gavel step with `fetch-depth: 0`. Add or update if needed.
- **Permissions**: add/merge:
  ```yaml
  permissions:
    contents: read
    pull-requests: write  # only when comment: "true"
  ```
  Never widen beyond what's needed. If `comment: "false"` is chosen, drop `pull-requests: write`.
- **`actions/setup-go`**: remove **only** if no remaining step in the same job uses the Go toolchain directly. The gavel action handles its own binary; it does not need an explicit Go toolchain step. If the job still has a non-gavel Go step, keep `setup-go`.
- **Redundant artifact uploads**: if the workflow had `actions/upload-artifact` for `coverage.out` / lint output that gavel now produces, remove those steps. Keep uploads for unrelated outputs.
- **Make targets**: delete `make lint` / `make test` *invocations* in CI. **Do not touch the Makefile itself** — local developers still use those targets.
- **Matrix builds**: gavel currently targets a single OS per invocation. If the existing matrix is OS × Go-version, collapse to a single Ubuntu job for the gavel step. Mention this collapse to the user in the diff so they can object. Do not silently delete a matrix.

### Canonical replacements

**Lint-only job** (golangci-lint-action → gavel):
```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0
- uses: flanksource/gavel@v<TAG>
  with:
    args: lint
    fail-on-error: "true"
```

**Test-only job** (`go test ./...` → gavel):
```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0
- uses: flanksource/gavel@v<TAG>
  with:
    args: test
    fail-on-error: "true"
```

**Combined lint + test** (preferred when both exist):
```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0
- uses: flanksource/gavel@v<TAG>
  with:
    args: test --lint
    fail-on-error: "true"
```

**Self-hosted / build-from-source** (mirrors gavel's own `.github/workflows/gavel-action.yml`):
```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0
- uses: actions/setup-go@v5
  with:
    go-version-file: go.mod
    cache: true
- name: Build gavel from source
  run: |
    go install ./cmd/gavel
- uses: flanksource/gavel@v<TAG>
  with:
    args: test --lint
    version: source
```

### Authoring constraints

- 2-space YAML indent, no trailing whitespace, no tabs.
- No `continue-on-error: true` added to silence gavel failures. If a user asks for that, push back — the right fix is to fix the failure, not hide it.
- No `.gavel.yaml` edits to suppress violations as part of this migration.

## Phase 4 — Verify

After the rewrites:

1. **YAML lint**: if `actionlint` is on `$PATH`, run it on changed files. If it's not, say so explicitly — do not silently skip.
   ```bash
   command -v actionlint && actionlint .github/workflows/*.yml
   ```

2. **Show the final diff**:
   ```bash
   git diff -- .github/workflows/
   ```

3. **Suggest validation**:
   - Push to a branch and let the live workflow run prove correctness.
   - Or run `act -j <jobname>` locally if `act` is installed (note: `act` doesn't perfectly replicate GitHub-hosted runners; treat as a smoke test).

4. **Do not commit**. Leave the changes in the working tree. The human reviews and commits.

## Edge Cases

- **Matrix collapse**: if you collapse a `strategy.matrix: { os: [ubuntu, macos, windows] }` → single Ubuntu, surface the loss to the user before applying. Some teams genuinely need multi-OS test coverage; the right answer there is "skip migration of this workflow" until gavel supports matrices natively.
- **Monorepos with multiple Go modules**: one gavel step per module, using `working-directory:`. Distinct `comment-header:` per module so each module gets its own sticky PR comment instead of one overwriting the other.
- **Coverage uploads (codecov, coveralls)**: gavel currently doesn't replace these. If the existing workflow runs `go test -coverprofile=coverage.out` followed by `codecov/codecov-action`, keep the test step *and* add gavel alongside, or wait until coverage is part of gavel's artifact set before replacing.
- **`actions/setup-go` with `cache: true`**: dropping `setup-go` also drops Go-module caching. If the existing job had `cache: true` and you're removing `setup-go`, note that the first gavel run will be slower until the runner warms its own cache. Not a blocker, but worth flagging.
- **Already-pinned golangci version**: if the team has pinned `golangci-lint-action@v9 with version: v2.11.4`, capture that pin in the migration diff so the user knows they're switching to gavel's bundled lint config (gavel reads `.golangci.yml` if present — confirm it exists before replacing).

## Reference

- gavel `action.yml` — composite action definition (inputs, outputs)
- gavel `README.md` "GitHub Action" section — canonical minimal + full example
- gavel `.github/workflows/gavel-action.yml` — the `version: source` self-test pattern
- gavel `.github/workflows/lint.yml`, `.github/workflows/test.yml` — pre-migration baseline shapes
