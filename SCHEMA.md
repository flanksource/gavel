# `.gavel.yaml` Configuration Schema

This document is the complete reference for Gavel's project/user configuration file,
`.gavel.yaml` (also accepted as `.gavel.yml`).

A machine-readable JSON Schema lives alongside this file at
[`gavel.schema.json`](./gavel.schema.json). It is generated from the Go config
types (`verify.GavelConfig`) — the source of truth — so the two never drift. An
annotated example is in [`gavel.yaml.example`](./gavel.yaml.example).

## File locations & merge order

Gavel looks for `.gavel.yaml` in three places and merges them, with later layers
overriding earlier ones:

1. `~/.gavel.yaml` — personal defaults
2. `<git-root>/.gavel.yaml` — repository config
3. `<target-dir>/.gavel.yaml` — directory config, when the target differs from the git root

Inspect the merged result for any path with:

```bash
gavel config [path]
gavel config --resolve [path] # include resolved prompt specs and effective models
```

The merge is **not** a blind overwrite. Scalars are last-write-wins, lists are
usually appended, a couple of booleans are "sticky", and `fixtures.files`
replaces rather than appends. The exact behavior per field is in the tables
below.

## Editor integration

Point your editor's YAML language server at the schema by adding a modeline to
the top of any `.gavel.yaml`:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/flanksource/gavel/main/gavel.schema.json
```

This enables inline completion, hover docs, and validation. Every object in the
schema sets `additionalProperties: false`, so an unknown or misspelled key is
flagged rather than silently ignored.

## Regenerating the schema

`gavel.schema.json` is generated; do not hand-edit it. After changing the config
types, regenerate it:

```bash
go generate .
```

A test (`verify.TestConfigSchema_GoldenMatchesCommitted`) fails if the committed
file is stale, and `verify.TestConfigJSONSchema_CoversStruct` fails if any config
field is left undocumented.

---

## `verify`

Settings for AI verification fixture steps.

| Key | Type | Default | Merge | Description |
| --- | --- | --- | --- | --- |
| `verify.model` | string | `claude` | last non-empty wins | AI CLI / model used by AI fixture steps. Common values: `claude`, `gemini`, `codex`, or a fully qualified model name. |
| `verify.prompt` | string | `""` | last non-empty wins | Repo-specific review policy appended to Gavel's built-in AI fixture prompt. |
| `verify.promptTemplate` | prompt override | built-in | last configured override wins | Complete reviewer prompt supplied inline or from a file. |
| `verify.checks.disabled` | string[] | `[]` | appended | Individual check IDs to disable (e.g. `SEC-1`, `PERF-2`). |
| `verify.checks.disabledCategories` | string[] | `[]` | appended | Whole categories to disable: `completeness`, `code-quality`, `testing`, `consistency`, `security`, `performance`. |

### Prompt overrides

Every registered AI prompt accepts a bare string, an `inline` value, or a `file`
reference. A string is complete dotprompt `.prompt` source. `inline` may also be
a structured Captain `api.Spec`; in that form `prompt.user` becomes the template
body and the remaining fields become prompt frontmatter. Relative files resolve
against the `.gavel.yaml` layer that declared them, and a missing or malformed
file is an error.

```yaml
todos:
  planPrompt:
    inline:
      model: claude-sonnet-5
      effort: high
      prompt:
        system: You produce implementation plans.
        user: Plan this work: {{{body}}}

commit:
  messagePrompt:
    file: .gavel/prompts/commit-message.prompt
```

Prompt override paths are `verify.promptTemplate`,
`commit.{messagePrompt,functionalityRemovedPrompt,compatibilityPrompt,summaryPrompt,groupingPrompt,prContentPrompt}`,
`todos.{runPrompt,planPrompt}`, `status.summaryPrompt`, and
`test.outlineSummaryPrompt`. Use `gavel config --resolve` (`-r`) to see each
prompt's built-in/inline/file source, complete body, declared Captain spec, and
effective model/backend. Structured `--json`/`--yaml` output has the shape
`{config, prompts}`.

## `lint`

Settings for `gavel lint`.

| Key | Type | Default | Merge | Description |
| --- | --- | --- | --- | --- |
| `lint.ignore` | object[] | `[]` | appended | Rules that suppress matching violations (see below). |
| `lint.ignore[].rule` | string | — | — | Match the violation's rule ID. Accepts literals, `*` globs (`"acme-*"`), and `!`-prefixed negations. |
| `lint.ignore[].source` | string | — | — | Match the emitting linter (`golangci-lint`, `eslint`, `betterleaks`, …). Same matcher syntax as `rule`. |
| `lint.ignore[].file` | string | — | — | Match the violation's file path using doublestar globs (`"pkg/**/*.go"`). |
| `lint.linters.<name>.enabled` | bool | linter default | later layer wins per linter | Force a linter on/off. Omit to use the linter's built-in default. |

An ignore rule matches only when **every** populated field matches; an empty
rule (no `rule`/`source`/`file`) never matches.

## `commit`

Settings for `gavel commit`.

| Key | Type | Default | Merge | Description |
| --- | --- | --- | --- | --- |
| `commit.model` | string | inherits `verify.model` | last non-empty wins | AI CLI / model for commit-message generation and compatibility analysis. |
| `commit.hooks` | object[] | `[]` | appended | Hooks run before the commit is written (see below). |
| `commit.hooks[].name` | string | — | — | Display name for the hook. |
| `commit.hooks[].run` | string | — | — | Shell command to execute. |
| `commit.hooks[].files` | string[] | — | — | Glob filter; the hook runs only if a staged file matches. Runs unconditionally when omitted. |
| `commit.gitignore` | string[] | `[]` | appended + deduped | Extra ignore globs applied when selecting files to commit. |
| `commit.allow` | string[] | `[]` | appended + deduped | Paths allowed through even when a broader `commit.gitignore` glob matches (e.g. generated artifacts you intentionally commit). |
| `commit.precommit.mode` | mode | `prompt` | last non-empty wins | Gate for `commit.gitignore` prompts and linked-dependency checks. |
| `commit.linkedDeps.mode` | mode | `prompt` | last non-empty wins | **Deprecated** — superseded by `commit.precommit`. Retained for backward-compatible loading; prefer `commit.precommit.mode`. |
| `commit.compatibility.mode` | mode | `skip` | last non-empty wins | Gate for the AI warning that surfaces removed functionality and backward-compatibility issues. |
| `commit.lint.enabled` | bool | `false` | later layer wins | Run every non-secrets linter over the staged file set before committing. Overridden per run by `--lint`. |
| `commit.lint.secrets` | bool | `true` | later layer wins | Run the betterleaks/secrets linter before committing. Overridden per run by `--lint-secrets`. |
| `commit.tidy.enabled` | bool | `true` | later layer wins | Run `go mod tidy` in every Go module and stage the resulting `go.mod`/`go.sum` changes. Overridden per run by `--tidy`. |

### `mode` values

The `precommit`, `linkedDeps`, and `compatibility` gates share a `mode` type:

| Value | Behavior |
| --- | --- |
| `prompt` | Ask before committing (default for `precommit`/`linkedDeps`). |
| `fail` | Hard-fail the commit when the check triggers. |
| `skip` | Bypass the check (default for `compatibility`). |
| `false` | Alias for `skip`. |

## `fixtures`

Fixture-test discovery for `gavel test`.

| Key | Type | Default | Merge | Description |
| --- | --- | --- | --- | --- |
| `fixtures.enabled` | bool | `false` | sticky — once true in any layer, stays true | Auto-discover fixture files when running `gavel test`. |
| `fixtures.files` | string[] | `["**/*.fixture.md"]` | later non-empty list **replaces** earlier list | Globs used to discover fixtures. |

## `checks`

Post-completion check loop for `gavel todos run --check`: after an agent reports
done, gavel runs the configured tests/lint and feeds any failures back to the
same agent session, re-running until they pass or `maxIterations` is reached.
Opt-in — runs only when enabled here, by a TODO's frontmatter `checks` block, or
by the `--check` flag. Frontmatter overrides these project defaults. Omit `test`
to skip tests and `lint` to skip linting; when enabled with neither set, both run
against changed files.

| Key | Type | Default | Merge | Description |
| --- | --- | --- | --- | --- |
| `checks.enabled` | bool | unset (off) | later set value wins | Turn the loop on. `--check` / frontmatter can force it on. |
| `checks.maxIterations` | int | `3` | later non-zero wins | Maximum agent re-runs before giving up. |
| `checks.test` | object | — | later non-nil replaces | `gavel test` options for the check run. Omit to skip tests. |
| `checks.test.paths` | string[] | — | — | Package paths to test. Empty discovers all. |
| `checks.test.changed` | bool | `false` | — | Only test packages affected by the agent's changes. |
| `checks.test.timeout` | string | — | — | Global wall-clock deadline (e.g. `5m`). |
| `checks.lint` | object | — | later non-nil replaces | `gavel lint` options for the check run. Omit to skip linting. |
| `checks.lint.linters` | string[] | — | — | Linters to run. Empty runs every detected linter. |
| `checks.lint.changed` | bool | `false` | — | Only report new violations versus the base ref. |
| `checks.lint.timeout` | string | — | — | Per-linter deadline (e.g. `5m`). |

## `ssh`

SSH post-receive hook / push backend.

| Key | Type | Default | Merge | Description |
| --- | --- | --- | --- | --- |
| `ssh.cmd` | string | `gavel test --lint` | last non-empty wins | Command executed by the SSH post-receive hook. An empty override inherits the parent value. |

## `pre` / `post`

Top-level hook steps. `pre` runs before the main test/lint pipeline (in
declaration order); `post` runs after as non-blocking cleanup/reporting whose
failures are logged but do not replace the main result.

| Key | Type | Default | Merge | Description |
| --- | --- | --- | --- | --- |
| `pre` / `post` | object[] | `[]` | appended in load order (home → repo → cwd) | List of hook steps. |
| `pre[].name` / `post[].name` | string | — | — | Optional display name for the step. |
| `pre[].run` / `post[].run` | string | — | — | Shell command to execute. |

## `secrets`

betterleaks / gitleaks secret-scanning orchestration. Rule authoring lives in
the TOML files themselves; Gavel only discovers and merges them.

| Key | Type | Default | Merge | Description |
| --- | --- | --- | --- | --- |
| `secrets.disabled` | bool | `false` | sticky OR — once true in any layer, stays true | Disable the betterleaks linter even when the binary is on `PATH`. |
| `secrets.configs` | string[] | `[]` | appended + deduped | Additional `.betterleaks.toml` / `.gitleaks.toml` paths to merge in, beyond those discovered in the home dir, git root, and cwd. Relative paths resolve against the `.gavel.yaml` directory. |

---

## Minimal example

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/flanksource/gavel/main/gavel.schema.json
verify:
  model: claude

commit:
  hooks:
    - name: gofmt
      run: gofmt -w ./...
      files:
        - "**/*.go"
  precommit:
    mode: prompt

fixtures:
  enabled: true

secrets:
  disabled: false
```

See [`gavel.yaml.example`](./gavel.yaml.example) for a fully annotated example
covering every section.
