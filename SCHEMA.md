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
  plan:
    prompt:
      system: You produce implementation plans.
      user: Plan this work: {{{body}}}
    model: claude-sonnet-5
    effort: high

commit:
  messagePrompt:
    file: .gavel/prompts/commit-message.prompt
```

Prompt override paths are `verify.promptTemplate`,
`commit.{messagePrompt,functionalityRemovedPrompt,compatibilityPrompt,summaryPrompt,groupingPrompt,prContentPrompt}`,
`todos.{run,plan,triage}`, `status.summaryPrompt`, and
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

Post-completion check loop inside `gavel todos run`: after an agent reports
done, gavel runs the configured tests/lint and feeds any failures back to the
same agent session, re-running until they pass or `todos.run.workflow.verify.maxIterations` is reached.
Opt-in — runs only when enabled here or by a TODO's frontmatter `checks` block;
there is no per-run flag. Frontmatter overrides these project defaults. Omit `test`
to skip tests and `lint` to skip linting; when enabled with neither set, both run
against changed files.

| Key | Type | Default | Merge | Description |
| --- | --- | --- | --- | --- |
| `checks.enabled` | bool | unset (off) | later set value wins | Append the check steps to every todo's definition of done. A TODO's frontmatter `checks` block can force them on for that todo. |
| `checks.test` | object | — | later non-nil replaces | `gavel test` options for the check run. Omit to skip tests. |
| `checks.test.paths` | string[] | — | — | Package paths to test. Empty discovers all. |
| `checks.test.changed` | bool | `false` | — | Only test packages affected by the agent's changes. |
| `checks.test.timeout` | string | — | — | Global wall-clock deadline (e.g. `5m`). |
| `checks.lint` | object | — | later non-nil replaces | `gavel lint` options for the check run. Omit to skip linting. |
| `checks.lint.linters` | string[] | — | — | Linters to run. Empty runs every detected linter. |
| `checks.lint.changed` | bool | `false` | — | Only report new violations versus the base ref. |
| `checks.lint.timeout` | string | — | — | Per-linter deadline (e.g. `5m`). |

### Who grades the definition of done

A TODO's acceptance criteria are marked by an LLM checklist step, and the spec it
runs as is resolved separately from the agent that did the work — a run never
grades itself. The chain, highest first:

```
request (a `gavel todos check` flag or the dashboard's verification payload)
  → .gavel.yaml todos.verify
  → .gavel.yaml ai:
  → captain
```

`todos.verify` is a Captain `api.Spec` (model, budget, permissions — no prompt:
the checklist is generated from the criteria). The built-in default names **no
model**: it only pins the run mode to agentic (`api.ModeAgent`), because the
grader is told to inspect the change with its own tools and a mode that cannot
read the diff would still return confident-looking verdicts. Which model runs
under that mode comes from configuration like every other operation — first
`.gavel.yaml`, then `~/.captain.yaml` — and a repo that configures neither fails
loudly, telling you to run `gavel configure` or `captain configure`, rather than
silently seeding a specific model.

```yaml
ai:
  model: claude-haiku-4-5   # every other AI operation
todos:
  verify:
    model: claude-code-opus  # …but grade the definition of done with this
    budget:
      maxTurns: 20
```

## `todos`

Settings for `gavel todos run`, `gavel todos check`, and the todo lifecycle.
See [MANUAL.md](MANUAL.md#gavel-todos) for the lifecycle model itself (steps,
`when` predicates, `outcomes`) and the CLI flags that drive it.

| Key | Type | Default | Merge | Description |
| --- | --- | --- | --- | --- |
| `todos.run` | prompt spec | see below | field-wise override | AI spec for the `run` (implement) step. Overrides the `ai:` base field-wise. |
| `todos.plan` | prompt spec | see below | field-wise override | AI spec for the `plan` step (read-only investigation that produces a reviewable plan). |
| `todos.triage` | prompt spec | see below | field-wise override | AI spec for the `triage` step: a read-only pass that compacts a TODO's description and reviews its verification fixture. Auxiliary — never picked automatically, only run with `--step triage`. |
| `todos.verify` | Captain `api.Spec` | `{model: {mode: agent}}`, no model name | field-wise override | Spec the `verify` step (and `gavel todos check`) grades the definition of done as. No `prompt:` field — the checklist is generated from the TODO's acceptance criteria, not a template. |
| `todos.checkConcurrency` | int | `4` | last non-zero wins | How many TODOs `gavel todos check` (and the verification phase after a bulk triage) checks at once. |
| `todos.timeout` | string (Go duration) | unset | last non-empty wins | Wall-clock deadline for a lifecycle step run. A context constraint, not an ordinary spec field — it can only ever lower a budget a step or prompt asked for, never raise one. When nothing in the resolved spec sets a budget timeout at all (this key included), the host stamps its own `30m` ceiling as a last resort. |
| `todos.baseUrl` | string | `""` | last non-empty wins | Absolute origin this gavel dashboard is reachable at (e.g. `https://gavel.example.com`). TODO bodies store attachments as server-relative links; pushing a TODO to an external tracker rewrites them against this origin. |
| `todos.lifecycle` | object | the embedded lifecycle | merged by step name over the built-in lifecycle | Overrides the built-in todo lifecycle. See below. |
| `todos.steps.<name>` | Captain `api.Spec` | unset | field-wise override | Project-level spec layer for a custom lifecycle step (one declared under `todos.lifecycle.steps`), sitting exactly where `todos.run` sits for the built-in run step: above the step's own `spec:` declaration, below the todo's `llm:` block. Naming a built-in step here (`run`, `plan`, `triage`, `verify`) is an error — those keep their typed blocks above. |

`todos.run`, `todos.plan`, and `todos.triage` are prompt specs (bare string,
`inline`, or `file`, same as the other prompt-override fields — see [Prompt
overrides](#prompt-overrides) above). `todos.verify` is a plain Captain spec: it
has no prompt template to override, so a `file:` form here would be a silent
no-op.

### `todos.lifecycle`

Overrides the built-in todo lifecycle (`todos/lifecycle/todos.yaml`): which
prompt runs when, and which status its result lands the todo in. Either a
`file:` path (a lifecycle YAML document, relative to the project root) **or**
an inline `name`/`subject`/`steps` — the two forms are mutually exclusive, and
a config that sets both fails to load rather than silently dropping the inline
definition. A step named like a built-in step (`plan`,
`verify`, `run`, `triage`) replaces it wholesale; a new name is appended. A
`verify` step must survive the merge — every lifecycle must declare one.

| Key | Type | Description |
| --- | --- | --- |
| `todos.lifecycle.file` | string | Path to a lifecycle YAML document, relative to the project root. |
| `todos.lifecycle.name` | string | Name of the lifecycle; defaults to the built-in `todos`. |
| `todos.lifecycle.subject` | map[string]string | Extra subject declarations a custom step's predicates may read, field → CEL type (`string`, `bool`, `int`, `double`, `dyn`, `list<string>`, `list<dyn>`, `map<string,dyn>`). |
| `todos.lifecycle.steps` | object[] | Steps to replace or add: `{name, prompt, envelope, when, inputs, spec, outcomes, auxiliary}`. `outcomes` is a list of `{status, when}` pairs evaluated in order — first true wins, none true is an error. |

A custom step carries its own run configuration inline under its `spec:` field
in `todos.lifecycle.steps`; `todos.steps.<name>` (above) is the per-project
layer on top of that declaration, so a lifecycle file can be shared while each
project picks the model, budget or permissions its custom steps run with.

See [`gavel.yaml.example`](./gavel.yaml.example) for a worked `todos.lifecycle`
example with a custom step.

### Retired `todos.*` keys

These keys were removed when the ad-hoc run-mode/driver model was replaced by
the lifecycle. A `.gavel.yaml` still declaring one fails to load with the exact
message shown (`<path>: <key> is no longer supported; use <replacement>`)
rather than silently running on built-in defaults.

| Removed key | Replacement |
| --- | --- |
| `todos.driver` | `ai.model` with the compact `mode:model:effort` form, e.g. `ai.model: "cli:opus:high"` |
| `todos.prompts` | a lifecycle step under `todos.<step>` |
| `todos.groupBy` | nothing — grouping was removed; runs dispatch per todo |
| `todos.approvals` | `permissions.mode: default` on the step (the dashboard brokers each tool call) |
| `checks.maxIterations` | `todos.run.workflow.verify.maxIterations` |

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
